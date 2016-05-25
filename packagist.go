package pkgmirror

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/rande/goapp"
)

type PackagistConfig struct {
	Server string
	Code   []byte
}

func NewPackagistService() *PackagistService {
	return &PackagistService{
		Config: &PackagistConfig{
			Server: "https://packagist.org",
			Code:   []byte("packagist"),
		},
	}
}

type PackagistService struct {
	DB              *bolt.DB
	Config          *PackagistConfig
	DownloadManager *DownloadManager
	Logger          *log.Entry
	GitConfig       *GitConfig
}

func (ps *PackagistService) Init(app *goapp.App) error {
	var err error

	ps.Logger.Info("Init")

	ps.DB, err = bolt.Open(fmt.Sprintf("./data/packagist_%s.db", ps.Config.Code), 0600, &bolt.Options{
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

func (ps *PackagistService) Serve(state *goapp.GoroutineState) error {
	ps.Logger.Info("Starting Packagist Service")

	for {
		ps.SyncMirror(state)
		ps.Logger.Info("Wait before starting a new sync...")
		time.Sleep(60 * time.Second)
	}
}

func (ps *PackagistService) End() error {

	return nil
}

func (ps *PackagistService) SyncMirror(state *goapp.GoroutineState) {
	ps.Logger.Info("Starting PackagistService Sync Mirror")

	dm := &DownloadManager{
		Add:   make(chan PackageInformation),
		Count: 15,
	}
	pr := &PackagesResult{}

	url := fmt.Sprintf("%s/packages.json", ps.Config.Server)

	ps.Logger.Debug("Loading packages.json")

	if err := LoadRemoteStruct(url, pr); err != nil {
		ps.Logger.WithFields(log.Fields{
			"path":  "packages.json",
			"error": err.Error(),
		}).Error("Error loading packages.json")
	} else {
		ps.Logger.WithFields(log.Fields{
			"path": "packages.json",
		}).Info("End loading packages.json")
	}

	var wg sync.WaitGroup
	var lock sync.Mutex

	PackageListener := make(chan PackageInformation)

	go dm.Wait(state, func(id int, done chan<- struct{}, pkgs <-chan PackageInformation) {
		for pkg := range pkgs {
			p := &PackageResult{}

			ps.Logger.WithFields(log.Fields{
				"package": pkg.Package,
				"id":      id,
			}).Info("Retrieve package information")

			if err := LoadRemoteStruct(fmt.Sprintf("%s/%s", ps.Config.Server, pkg.GetSourceKey()), p); err != nil {
				ps.Logger.WithFields(log.Fields{
					"package": pkg.Package,
					"error":   err.Error(),
				}).Error("Error loading package information")
			}

			pkg.PackageResult = *p

			PackageListener <- pkg
		}
	})

	go func(db *bolt.DB) {
		for {
			select {
			case pkg := <-PackageListener:
				lock.Lock()
				db.Update(func(tx *bolt.Tx) error {
					b := tx.Bucket(ps.Config.Code)

					logger := ps.Logger.WithFields(log.Fields{
						"package": pkg.Package,
						"path":    pkg.GetTargetKey(),
					})

					for _, version := range pkg.PackageResult.Packages[pkg.Package] {
						version.Dist.URL = GitRewriteArchive(ps.GitConfig, version.Dist.URL)
						version.Source.URL = GitRewriteRepository(ps.GitConfig, version.Source.URL)
					}

					data, _ := json.Marshal(pkg.PackageResult)
					sha := sha256.Sum256(data)
					pkg.HashTarget = hex.EncodeToString(sha[:])

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
				lock.Unlock()

				wg.Done()
			}
		}
	}(ps.DB)

	providers := map[string]*ProvidersResult{}

	for provider, sha := range pr.ProviderIncludes {
		path := strings.Replace(provider, "%hash%", sha.Sha256, -1)

		logger := ps.Logger.WithFields(log.Fields{
			"provider": provider,
			"path":     path,
		})

		logger.Info("Loading provider information")

		pr := &ProvidersResult{}

		if err := LoadRemoteStruct(fmt.Sprintf("%s/%s", ps.Config.Server, path), pr); err != nil {
			log.WithField("error", err.Error()).Error("Error loading provider information")
		} else {
			log.Debug("End loading provider information")
		}

		providers[provider] = pr

		for name, sha := range pr.Providers {
			p := PackageInformation{
				Package: name,
				Exist:   false,
			}

			lock.Lock()
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
			lock.Unlock()

			logger := ps.Logger.WithFields(log.Fields{
				"provider": provider,
				"package":  name,
			})

			if !p.Exist {
				logger.Info("Add new package")

				wg.Add(1)
				dm.Add <- p
			} else {
				logger.Debug("Skipping package")
			}
		}
	}

	ps.Logger.Info("Wait for download to complete")

	wg.Wait()

	ps.Logger.Info("Update provider files and packages.json")

	for provider := range pr.ProviderIncludes {
		for name := range providers[provider].Providers {
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

		// save provider
		data, _ := json.Marshal(providers[provider])
		hash := sha256.Sum256(data)

		// https://github.com/golang/go/issues/3117
		p := pr.ProviderIncludes[provider]
		p.Sha256 = hex.EncodeToString(hash[:])
		pr.ProviderIncludes[provider] = p

		path := strings.Replace(provider, "%hash%", p.Sha256, -1)

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

	ps.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(ps.Config.Code)
		data, _ := json.Marshal(pr)
		b.Put([]byte("packages.json"), data)

		ps.Logger.Info("Save packages.json")

		return nil
	})

	ps.Logger.Info("End update cycle")
}

func (ps *PackagistService) Get(key string) ([]byte, error) {
	var data []byte

	err := ps.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(ps.Config.Code)

		raw := b.Get([]byte(key))

		data = make([]byte, len(raw))

		copy(data, raw)

		return nil
	})

	return data, err
}
