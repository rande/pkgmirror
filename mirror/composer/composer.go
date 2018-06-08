// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package composer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	PublicServer string
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
	lock      bool
	StateChan chan pkgmirror.State
}

func (ps *ComposerService) Init(app *goapp.App) (err error) {
	ps.Logger.Info("Init")

	if ps.DB, err = pkgmirror.OpenDatabaseWithBucket(ps.Config.Path, ps.Config.Code); err != nil {
		ps.Logger.WithFields(log.Fields{
			"error":  err,
			"path":   ps.Config.Path,
			"bucket": string(ps.Config.Code),
			"action": "Init",
		}).Error("Unable to open the internal database")
	}

	return
}

func (ps *ComposerService) Serve(state *goapp.GoroutineState) error {
	ps.Logger.Info("Starting Composer Service")

	syncEnd := make(chan bool)

	sync := func() {
		ps.Logger.Info("Starting a new sync...")

		ps.SyncPackages()
		ps.UpdateEntryPoints()
		ps.CleanPackages()

		syncEnd <- true
	}

	// start the first sync
	go sync()

	for {
		select {
		case <-state.In:
			ps.DB.Close()
			return nil

		case <-syncEnd:
			ps.StateChan <- pkgmirror.State{
				Message: "Wait for a new run",
				Status:  pkgmirror.STATUS_HOLD,
			}

			ps.Logger.Info("Wait before starting a new sync...")

			// we recursively call sync unless a state.In comes in to exist the current
			// go routine (ie, the Serve function). This might not close the sync processus
			// completely. We need to have a proper channel (queue mode) for git fetch.
			// This will probably make this current code obsolete.
			go func() {
				time.Sleep(60 * 15 * time.Second)
				sync()
			}()
		}
	}
}

func (ps *ComposerService) SyncPackages() error {
	logger := ps.Logger.WithFields(log.Fields{
		"action": "SyncPackages",
	})

	logger.Info("Starting SyncPackages")

	ps.StateChan <- pkgmirror.State{
		Message: "Syncing packages",
		Status:  pkgmirror.STATUS_RUNNING,
	}

	dm := pkgmirror.NewWorkerManager(10, func(id int, data <-chan interface{}, result chan interface{}) {
		for raw := range data {
			pkg := raw.(PackageInformation)

			p := &PackageResult{}

			url := fmt.Sprintf("%s/p/%s", ps.Config.SourceServer, pkg.GetSourceKey())

			logger.WithFields(log.Fields{
				"package":     pkg.Package,
				"source_hash": pkg.HashSource,
				"worker":      id,
				"url":         url,
			}).Debug("Load loading package information")

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

	ps.StateChan <- pkgmirror.State{
		Message: "Loading packages.json",
		Status:  pkgmirror.STATUS_RUNNING,
	}

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

	ps.StateChan <- pkgmirror.State{
		Message: "Wait for download to complete",
		Status:  pkgmirror.STATUS_RUNNING,
	}

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

	ps.StateChan <- pkgmirror.State{
		Message: "Update entry points",
		Status:  pkgmirror.STATUS_RUNNING,
	}

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

	ps.StateChan <- pkgmirror.State{
		Message: "End update entry points",
		Status:  pkgmirror.STATUS_RUNNING,
	}

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

		if len(data) == 0 {
			return pkgmirror.EmptyKeyError
		}

		if err := json.Unmarshal(data, pkg); err == nil {
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

		return err
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

		for name := range pkg.PackageResult.Packages {
			for _, version := range pkg.PackageResult.Packages[name] {
				version.Dist.URL = git.GitRewriteArchive(ps.Config.PublicServer, version.Dist.URL)
				version.Source.URL = git.GitRewriteRepository(ps.Config.PublicServer, version.Source.URL)
			}
		}

		ps.StateChan <- pkgmirror.State{
			Message: fmt.Sprintf("Save package information: %s", pkg.Package),
			Status:  pkgmirror.STATUS_RUNNING,
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

func (ps *ComposerService) SearchPackage(r *http.Request) ([]byte, error) {
	ps.Logger.Info(fmt.Sprintf("Searching for %s", r.URL.RawQuery))

	resp, err := http.Get(fmt.Sprintf("%s/search.json?%s", ps.Config.SourceServer, r.URL.RawQuery))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

func (ps *ComposerService) CleanPackages() error {

	logger := ps.Logger.WithFields(log.Fields{
		"action": "CleanPackages",
	})

	logger.Info("Start cleaning ...")

	ps.StateChan <- pkgmirror.State{
		Message: "Start cleaning packages",
		Status:  pkgmirror.STATUS_RUNNING,
	}

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

	ps.StateChan <- pkgmirror.State{
		Message: "End cleaning packages",
		Status:  pkgmirror.STATUS_RUNNING,
	}

	return nil
}
