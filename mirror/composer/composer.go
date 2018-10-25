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
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/rande/goapp"
	"github.com/rande/pkgmirror"
	"github.com/rande/pkgmirror/mirror/git"
)

type ComposerConfig struct {
	SourceServer     string
	PublicServer     string
	BasePublicServer string
	Code             []byte
	Path             string
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
	DB            *bolt.DB
	Config        *ComposerConfig
	Logger        *log.Entry
	lock          bool
	StateChan     chan pkgmirror.State
	ProvidersURL  string
	BoltCompacter *pkgmirror.BoltCompacter
}

func (ps *ComposerService) getPackageUrl(pi *PackageInformation) string {
	return fmt.Sprintf("%s%s", ps.Config.BasePublicServer, ps.getPackageKey(pi))
}

func (ps *ComposerService) getPackageKey(pi *PackageInformation) string {
	// /8/%package%$%hash%.json
	var key = ps.ProvidersURL

	key = strings.Replace(key, "%package%", pi.Package, -1)
	key = strings.Replace(key, "%hash%", pi.HashSource, -1)

	return key
}

func (ps *ComposerService) Init(app *goapp.App) (err error) {
	ps.Logger.Info("Init")

	if err := ps.openDatabase(); err != nil {
		return err
	}

	return ps.optimize()
}

func (ps *ComposerService) Serve(state *goapp.GoroutineState) error {
	ps.Logger.Info("Starting Composer Service")

	syncEnd := make(chan bool)

	iteration := 0

	sync := func() {
		ps.Logger.Info("Starting a new sync...")

		ps.SyncPackages()
		ps.UpdateEntryPoints()
		ps.CleanPackages()

		iteration++

		// optimize every 10 iteration
		if iteration > 9 {
			ps.Logger.Info("Starting database optimization")
			ps.optimize()
			iteration = 0
		}

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

func (ps *ComposerService) openDatabase() (err error) {
	if ps.DB, err = pkgmirror.OpenDatabaseWithBucket(ps.Config.Path, ps.Config.Code); err != nil {
		ps.Logger.WithFields(log.Fields{
			log.ErrorKey: err,
			"path":       ps.Config.Path,
			"bucket":     string(ps.Config.Code),
			"action":     "Init",
		}).Error("Unable to open the internal database")

		return err
	}

	return nil
}

func (ps *ComposerService) optimize() error {
	ps.lock = true

	path := ps.DB.Path()

	ps.DB.Close()

	if err := ps.BoltCompacter.Compact(path); err != nil {
		return err
	}

	err := ps.openDatabase()

	ps.lock = false

	return err
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

			logger.WithFields(log.Fields{
				"package":     pkg.Package,
				"source_hash": pkg.HashSource,
				"worker":      id,
				"url":         pkg.Url,
			}).Debug("Load loading package information")

			if err := pkgmirror.LoadRemoteStruct(pkg.Url, p); err != nil {
				logger.WithFields(log.Fields{
					"package": pkg.Package,
					"url":     pkg.Url,
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

	packagesResult := &PackagesResult{}

	logger.Info("Loading packages.json")

	ps.StateChan <- pkgmirror.State{
		Message: "Loading packages.json",
		Status:  pkgmirror.STATUS_RUNNING,
	}

	if err := pkgmirror.LoadRemoteStruct(fmt.Sprintf("%s/packages.json", ps.Config.SourceServer), packagesResult); err != nil {
		logger.WithFields(log.Fields{
			"path":  "packages.json",
			"error": err.Error(),
		}).Error("Error loading packages.json")

		return err // an error occurs avoid empty file
	}

	ps.ProvidersURL = packagesResult.ProvidersURL

	for provider, sha := range packagesResult.ProviderIncludes {
		path := strings.Replace(provider, "%hash%", sha.Sha256, -1)

		logger := logger.WithFields(log.Fields{
			"provider":      provider,
			"provider_hash": sha.Sha256,
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

				if err := pkgmirror.Unmarshal(data, &p); err != nil && len(data) > 0 {
					logger.WithFields(log.Fields{
						"error": err,
						"data":  data,
					}).Error("Unable to unmarshal package information")
				} else {
					p.Exist = p.HashSource == sha.Sha256
				}

				p.HashSource = sha.Sha256

				return nil
			})

			p.Url = ps.getPackageUrl(&p)

			logger = logger.WithFields(log.Fields{
				"package_hash": p.HashSource,
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

	logger := ps.Logger.WithFields(log.Fields{
		"action": "GetPackage",
		"key":    key,
	})

	logger.Info("Get package data")

	err := ps.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(ps.Config.Code)

		data := b.Get([]byte(key))

		if len(data) == 0 {
			return pkgmirror.EmptyKeyError
		} else if err := pkgmirror.Unmarshal(data, pi); err != nil {
			logger.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Error while unmarshalling package")

			return err
		} else {
			return nil
		}
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

	logger.Debug("packages.json loaded")

	providers := map[string]*ProvidersResult{}

	for provider, sha := range pkgResult.ProviderIncludes {
		pr := &ProvidersResult{}

		url := fmt.Sprintf("%s/%s", ps.Config.SourceServer, strings.Replace(provider, "%hash%", sha.Sha256, -1))

		logger.WithFields(log.Fields{
			"provider": provider,
			"url":      url,
		}).Debug("Load provider")

		if err := pkgmirror.LoadRemoteStruct(url, pr); err != nil {
			ps.Logger.WithFields(log.Fields{
				"provider": provider,
				"url":      url,
				"error":    err,
			}).Error("Error loading provider information")
		}

		providers[provider] = pr

		// iterate packages from each provider
		for name := range pr.Providers {
			ps.DB.View(func(tx *bolt.Tx) error {
				b := tx.Bucket(ps.Config.Code)
				data := b.Get([]byte(name))

				pi := &PackageInformation{}
				if err := pkgmirror.Unmarshal(data, pi); err != nil {
					logger.WithFields(log.Fields{
						"error": err.Error(),
						"data":  string(data),
						"name":  name,
					}).Error("Error while unmarshalling provider package")

					return err
				}

				// https://github.com/golang/go/issues/3117
				p := providers[provider].Providers[name]
				p.Sha256 = pi.HashTarget
				providers[provider].Providers[name] = p

				return nil
			})
		}

		// save provider file, cannot compress *yet* as we need the sha1 from the uncompressed json file.
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

			if data, err := pkgmirror.Marshal(providers[provider]); err != nil {
				logger.WithError(err).Error("Unable to marshal provider data")

				return err
			} else if err := b.Put([]byte(path), data); err != nil {
				logger.WithError(err).Error("Unable to store provider data")

				return err
			}

			ps.Logger.WithFields(log.Fields{
				"provider": provider,
				"path":     path,
			}).Debug("Save provider")

			return nil
		})
	}

	//pr.ProviderIncludes = providerIncludes
	pkgResult.ProvidersURL = fmt.Sprintf("/composer/%s/p/%%package%%$%%hash%%.json", ps.Config.Code)
	pkgResult.Notify = fmt.Sprintf("/composer/%s/downloads/%%package%%", ps.Config.Code)
	pkgResult.NotifyBatch = fmt.Sprintf("/composer/%s/downloads", ps.Config.Code)
	pkgResult.Search = fmt.Sprintf("/composer/%s/search.json?q=%%query%%&type=%%type%%", ps.Config.Code)

	ps.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(ps.Config.Code)
		data, _ := pkgmirror.Marshal(pkgResult)
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

	logger := ps.Logger.WithFields(log.Fields{
		"package": pkg.Package,
		"action":  "UpdatePackage",
	})

	logger.Info("Explicit reload package information")

	if err := ps.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(ps.Config.Code)
		data := b.Get([]byte(pkg.Package))

		if len(data) == 0 {
			return pkgmirror.EmptyKeyError
		} else if err := pkgmirror.Unmarshal(data, pkg); err == nil {
			return err
		}

		return nil
	}); err != nil {
		return err // unknown package
	}

	pkg.Url = ps.getPackageUrl(pkg)

	pkg.PackageResult = PackageResult{}

	if err := pkgmirror.LoadRemoteStruct(pkg.Url, &pkg.PackageResult); err != nil {
		logger.WithFields(log.Fields{
			"error": err.Error(),
			"url":   pkg.Url,
		}).Error("Error loading package information")

		return err
	}

	if err := ps.savePackage(pkg); err != nil {
		logger.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Unable to save package")

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

		if data, err := pkgmirror.Compress(data); err != nil {
			logger.WithError(err).Error("Unable to compress package data")

			return err
		} else if err := b.Put([]byte(pkg.GetTargetKey()), data); err != nil {
			logger.WithError(err).Error("Error updating/creating definition")

			return err
		} else if data, err := pkgmirror.Marshal(pkg); err != nil {
			logger.WithError(err).Error("Unable to marshal package definition data")

			return err
		} else if err := b.Put([]byte(pkg.Package), data); err != nil {
			logger.WithError(err).Error("Error updating/creating hash definition")

			return err
		}

		return nil
	})
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
		} else if err := pkgmirror.Unmarshal(data, pkgResult); err != nil {
			logger.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Error while decompressing packages.json")

			return err
		}

		// Sample iteration over key
		//  - drupal/a11n_code_example
		//  - drupal/a11n_code_example$e3147979055c65820731c0ebae9a9f989f7d8a52bb3dbb036e2bdc393127528b
		//  - drupal/a11n_code_form
		//  - drupal/a11n_code_form$86bd5df758fbfa094e356700d0d645bc41707e57b1ee5ac471e54386687a53e7
		//  - drupal/a11n_code_rest
		//  - drupal/a11n_code_rest$48329b3394c64d16784baebd626797454475a560b753f29fa5c64b9fa38795fc
		//  - drupal/a12_best_bets
		//  - drupal/a12_best_bets$6037a7856189bf0e9b4d22a44b8587f299ca483fa31585165c52bab221050524
		//  - drupal/a12_connect
		//  - drupal/a12_connect$a90ba85da88bd35f0e389cc7fbca4126ede378976862d22fbb14c4505436def9
		//  - drupal/a_hole
		var pi *PackageInformation
		var pr *ProvidersResult

		b.ForEach(func(k, v []byte) error {
			name := string(k)
			if i := strings.Index(name, "$"); i > 0 { // sha1 package, ie: drupal/a11n_code_example$e3147979055c65820731c0ebae9a9f989f7d8a52bb3dbb036e2bdc393127528b
				if pr != nil {
					// iterate over provider list from packages.json
					// and remove provide with no reference
					for provider, sha := range pkgResult.ProviderIncludes {
						// find the provider but the sha1 does not match (old one) delete
						if name[0:i+1] == provider[0:i+1] && name[i+1:len(name)-5] != sha.Sha256 {
							logger.WithFields(log.Fields{
								"provider":     pr.Code,
								"hash_target":  sha.Sha256,
								"hash_current": name[i+1 : len(name)-5],
							}).Info("Delete provider definition")

							b.Delete(k)
						}
					}
				} else if pi != nil {
					if name[0:i] == pi.Package && pi.HashTarget != name[i+1:] {
						logger.WithFields(log.Fields{
							"package":      pi.Package,
							"hash_target":  pi.HashTarget,
							"hash_current": name[i+1:],
						}).Info("Delete package definition")

						b.Delete(k)
					}
				} else {
					logger.WithField("key", name).Error("Orphan reference")
				}
			} else { // load the current active package or provider, ie: drupal/a11n_code_example
				pr = &ProvidersResult{}
				pi = &PackageInformation{}

				pkgmirror.Unmarshal(v, pi)
				if len(pi.Package) > 0 {
					// logger.WithField("package", name).Debug("Unmarshal PackageInformation")
					pr = nil

					return nil
				}

				pkgmirror.Unmarshal(v, pr)
				if len(pr.Code) > 0 {
					logger.WithField("provider", name).Debug("Unmarshal ProvidersResult")
					pi = nil

					return nil
				}

				logger.WithField("key", name).Debug("Unable to unmarshal data")

				pi = nil
				pr = nil
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
