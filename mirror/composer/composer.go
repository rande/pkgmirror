// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package composer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/rande/goapp"
	"github.com/rande/pkgmirror"
	"github.com/rande/pkgmirror/mirror/git"
)

type ComposerConfig struct {
	SourceServer string
	Code         []byte
	Path         string
}

func NewComposerService() *ComposerService {
	return &ComposerService{
		Config: &ComposerConfig{
			SourceServer: "https://packagist.org",
			Code:         []byte("packagist"),
			Path:         "./data/composer",
		},
	}
}

type ComposerService struct {
	DB        *bolt.DB
	Config    *ComposerConfig
	Logger    *log.Entry
	GitConfig *git.GitConfig
	lock      bool
}

func (ps *ComposerService) Init(app *goapp.App) error {
	var err error

	ps.Logger.Info("Init")

	ps.DB, err = bolt.Open(fmt.Sprintf("%s/%s.db", ps.Config.Path, ps.Config.Code), 0600, &bolt.Options{
		Timeout:  1 * time.Second,
		ReadOnly: false,
	})

	if err != nil {
		return err
	}

	return ps.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(ps.Config.Code)

		return err
	})
}

func (ps *ComposerService) Serve(state *goapp.GoroutineState) error {
	ps.Logger.Info("Starting Composer Service")

	for {
		ps.Logger.Info("Starting a new sync...")

		ps.SyncPackages()
		ps.UpdateEntryPoints()
		ps.CleanPackages()

		ps.Logger.Info("Wait before starting a new sync...")
		time.Sleep(60 * 15 * time.Second)
	}
}

func (ps *ComposerService) End() error {
	return nil
}

func (ps *ComposerService) SyncPackages() error {
	logger := ps.Logger.WithFields(log.Fields{
		"action": "SyncPackages",
	})

	logger.Info("Starting SyncPackages")

	dm := pkgmirror.NewWorkerManager(10, func(id int, data <-chan interface{}, result chan interface{}) {
		for raw := range data {
			pkg := raw.(PackageInformation)

			p := &PackageResult{}

			logger.WithFields(log.Fields{
				"package":     pkg.Package,
				"source_hash": pkg.HashSource,
				"worker":      id,
			}).Debug("Load loading package information")

			url := fmt.Sprintf("%s/p/%s", ps.Config.SourceServer, pkg.GetSourceKey())

			if err := pkgmirror.LoadRemoteStruct(url, p); err != nil {
				logger.WithFields(log.Fields{
					"package": pkg.Package,
					"url":     url,
					"error":   err.Error(),
				}).Error("Error loading package information")

				continue
			}

			pkg.PackageResult = *p

			result <- pkg
		}
	})

	dm.ResultCallback(func(data interface{}) {
		pkg := data.(PackageInformation)

		ps.savePackage(&pkg)
	})

	dm.Start()

	pr := &PackagesResult{}

	logger.Info("Loading packages.json")

	if err := pkgmirror.LoadRemoteStruct(fmt.Sprintf("%s/packages.json", ps.Config.SourceServer), pr); err != nil {
		logger.WithFields(log.Fields{
			"path":  "packages.json",
			"error": err.Error(),
		}).Error("Error loading packages.json")

		return err // an error occurs avoid empty file
	}

	for provider, sha := range pr.ProviderIncludes {
		path := strings.Replace(provider, "%hash%", sha.Sha256, -1)

		logger := logger.WithFields(log.Fields{
			"provider": provider,
			"hash":     sha.Sha256,
		})

		logger.Info("Loading provider information")

		pr := &ProvidersResult{}

		if err := pkgmirror.LoadRemoteStruct(fmt.Sprintf("%s/%s", ps.Config.SourceServer, path), pr); err != nil {
			logger.WithField("error", err.Error()).Error("Error loading provider information")
		} else {
			logger.Debug("End loading provider information")
		}

		for name, sha := range pr.Providers {
			p := PackageInformation{
				Server:  string(ps.Config.Code),
				Package: name,
				Exist:   false,
			}

			logger := logger.WithFields(log.Fields{
				"package": name,
			})

			logger.Debug("Analysing package")

			ps.DB.View(func(tx *bolt.Tx) error {
				b := tx.Bucket(ps.Config.Code)
				data := b.Get([]byte(p.Package))

				p.Exist = false

				if err := json.Unmarshal(data, &p); err == nil {
					p.Exist = p.HashSource == sha.Sha256
				}

				p.HashSource = sha.Sha256

				return nil
			})

			if !p.Exist {
				logger.Info("Add/Update new package")

				dm.Add(p)
			} else {
				logger.Debug("Skipping package")
			}
		}
	}

	logger.Info("Wait for download to complete")

	dm.Wait()

	return nil
}

func (ps *ComposerService) Get(key string) ([]byte, error) {
	var data []byte

	ps.Logger.WithFields(log.Fields{
		"action": "Get",
		"key":    key,
	}).Info("Get raw data")

	err := ps.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(ps.Config.Code)

		raw := b.Get([]byte(key))

		if len(raw) == 0 {
			return pkgmirror.EmptyKeyError
		}

		data = make([]byte, len(raw))

		copy(data, raw)

		return nil
	})

	return data, err
}

func (ps *ComposerService) GetPackage(key string) (*PackageInformation, error) {
	pi := &PackageInformation{}

	ps.Logger.WithFields(log.Fields{
		"action": "Get",
		"key":    key,
	}).Info("Get package data")

	err := ps.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(ps.Config.Code)

		raw := b.Get([]byte(key))

		if len(raw) == 0 {
			return pkgmirror.EmptyKeyError
		}

		return json.Unmarshal(raw, pi)
	})

	return pi, err
}

// This method generates the different entry points required by a repository.
//
func (ps *ComposerService) UpdateEntryPoints() error {
	if ps.lock {
		return pkgmirror.SyncInProgressError
	}

	ps.lock = true

	defer func() {
		ps.lock = false
	}()

	logger := ps.Logger.WithFields(log.Fields{
		"action": "UpdateEntryPoints",
	})

	logger.Info("Start")

	pkgResult := &PackagesResult{}
	if err := pkgmirror.LoadRemoteStruct(fmt.Sprintf("%s/packages.json", ps.Config.SourceServer), pkgResult); err != nil {
		logger.WithFields(log.Fields{
			"path":  "packages.json",
			"error": err.Error(),
		}).Error("Error loading packages.json")

		return err // an error occurs avoid empty file
	}

	logger.Info("packages.json loaded")

	providers := map[string]*ProvidersResult{}

	for provider, sha := range pkgResult.ProviderIncludes {
		pr := &ProvidersResult{}

		logger.WithField("provider", provider).Info("Load provider")

		if err := pkgmirror.LoadRemoteStruct(fmt.Sprintf("%s/%s", ps.Config.SourceServer, strings.Replace(provider, "%hash%", sha.Sha256, -1)), pr); err != nil {
			ps.Logger.WithField("error", err.Error()).Error("Error loading provider information")
		}

		providers[provider] = pr

		// iterate packages from each provider
		for name := range pr.Providers {
			ps.DB.View(func(tx *bolt.Tx) error {
				b := tx.Bucket(ps.Config.Code)
				data := b.Get([]byte(name))

				pi := &PackageInformation{}
				if err := json.Unmarshal(data, pi); err != nil {
					return err
				}

				// https://github.com/golang/go/issues/3117
				p := providers[provider].Providers[name]
				p.Sha256 = pi.HashTarget
				providers[provider].Providers[name] = p

				return nil
			})
		}

		// save provider file
		data, err := json.Marshal(providers[provider])

		if err != nil {
			ps.Logger.WithFields(log.Fields{
				"provider": provider,
				"error":    err,
			}).Error("Unable to marshal provider information")
		}

		hash := sha256.Sum256(data)

		// https://github.com/golang/go/issues/3117
		p := pkgResult.ProviderIncludes[provider]
		p.Sha256 = hex.EncodeToString(hash[:])
		pkgResult.ProviderIncludes[provider] = p

		path := fmt.Sprintf("%s", strings.Replace(provider, "%hash%", p.Sha256, -1))

		ps.DB.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(ps.Config.Code)
			b.Put([]byte(path), data)

			ps.Logger.WithFields(log.Fields{
				"provider": provider,
				"path":     path,
			}).Debug("Save provider")

			return nil
		})
	}

	//pr.ProviderIncludes = providerIncludes
	pkgResult.ProvidersURL = fmt.Sprintf("/composer/%s%s", ps.Config.Code, pkgResult.ProvidersURL)
	pkgResult.Notify = fmt.Sprintf("/composer/%s%s", ps.Config.Code, pkgResult.Notify)
	pkgResult.NotifyBatch = fmt.Sprintf("/composer/%s%s", ps.Config.Code, pkgResult.NotifyBatch)
	pkgResult.Search = fmt.Sprintf("/composer/%s%s", ps.Config.Code, pkgResult.Search)

	ps.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(ps.Config.Code)
		data, _ := json.Marshal(pkgResult)
		b.Put([]byte("packages.json"), data)

		ps.Logger.Info("Save packages.json")

		return nil
	})

	ps.Logger.Info("End UpdateEntryPoints")

	return nil
}

func (ps *ComposerService) UpdatePackage(name string) error {
	if ps.lock {
		return pkgmirror.SyncInProgressError
	}

	if i := strings.Index(name, "$"); i > 0 {
		name = name[:i]
	}

	pkg := &PackageInformation{
		Package: name,
		Server:  ps.Config.SourceServer,
	}

	ps.Logger.WithFields(log.Fields{
		"package": pkg.Package,
		"action":  "UpdatePackage",
	}).Info("Explicit reload package information")

	err := ps.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(ps.Config.Code)
		data := b.Get([]byte(pkg.Package))

		if err := json.Unmarshal(data, &pkg); err == nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err // unknown package
	}

	pkg.PackageResult = PackageResult{}

	if err := pkgmirror.LoadRemoteStruct(fmt.Sprintf("%s/p/%s", ps.Config.SourceServer, pkg.GetSourceKey()), &pkg.PackageResult); err != nil {
		ps.Logger.WithFields(log.Fields{
			"package": pkg.Package,
			"error":   err.Error(),
			"action":  "UpdatePackage",
		}).Error("Error loading package information")
	}

	if err := ps.savePackage(pkg); err != nil {
		return err
	}

	return ps.UpdateEntryPoints()
}

func (ps *ComposerService) savePackage(pkg *PackageInformation) error {
	return ps.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(ps.Config.Code)

		logger := ps.Logger.WithFields(log.Fields{
			"package": pkg.Package,
			"path":    pkg.GetTargetKey(),
		})

		for _, version := range pkg.PackageResult.Packages[pkg.Package] {
			version.Dist.URL = git.GitRewriteArchive(ps.GitConfig, version.Dist.URL)
			version.Source.URL = git.GitRewriteRepository(ps.GitConfig, version.Source.URL)
		}

		// compute hash
		data, _ := json.Marshal(pkg.PackageResult)
		sha := sha256.Sum256(data)
		pkg.HashTarget = hex.EncodeToString(sha[:])

		data, err := pkgmirror.Compress(data)

		if err != nil {
			logger.WithError(err).Error("Unable to compress data")

			return err
		}

		// store the path
		if err := b.Put([]byte(pkg.GetTargetKey()), data); err != nil {
			logger.WithError(err).Error("Error updating/creating definition")

			return err
		} else {
			data, _ := json.Marshal(pkg)

			if err := b.Put([]byte(pkg.Package), data); err != nil {
				logger.WithError(err).Error("Error updating/creating hash definition")

				return err
			}
		}

		return nil
	})
}

func (ps *ComposerService) CleanPackages() error {

	logger := ps.Logger.WithFields(log.Fields{
		"action": "CleanPackages",
	})

	logger.Info("Start cleaning ...")

	ps.DB.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket(ps.Config.Code)

		pkgResult := &PackagesResult{}
		if data, err := ps.Get("packages.json"); err != nil {
			logger.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Error loading packages.json")

			return err // an error occurs avoid empty file
		} else {
			json.Unmarshal(data, pkgResult)
		}

		var pi *PackageInformation

		b.ForEach(func(k, v []byte) error {
			name := string(k)
			if i := strings.Index(name, "$"); i > 0 {

				if name[0:10] == "p/provider" {
					// skipping
					for provider, sha := range pkgResult.ProviderIncludes {
						if name[0:i+1] == provider[0:i+1] && name[i+1:len(name)-5] != sha.Sha256 {
							logger.WithFields(log.Fields{
								"package":      provider[0:i],
								"hash_target":  sha.Sha256,
								"hash_current": name[i+1 : len(name)-5],
							}).Info("Delete provider definition")

							b.Delete(k)
						}
					}

				} else if name[0:i] == pi.Package {

					if pi.HashTarget != name[i+1:] {
						logger.WithFields(log.Fields{
							"package":      name,
							"hash_target":  pi.HashTarget,
							"hash_current": name[i+1:],
						}).Info("Delete package definition")

						b.Delete(k)
					}
				} else {
					logger.WithField("package", name).Error("Orphan reference")
				}
			} else {
				pi = &PackageInformation{}

				json.Unmarshal(v, pi)
			}

			return nil
		})

		return nil
	})

	return nil
}
