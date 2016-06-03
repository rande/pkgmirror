// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package npm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/rande/goapp"
	"github.com/rande/gonode/core/vault"
	"github.com/rande/pkgmirror"
	"github.com/rande/pkgmirror/mirror/git"
)

var (
	NPM_ARCHIVE = regexp.MustCompile(`http(s|):\/\/([\w\.]+)\/(.*)`)
)

type NpmConfig struct {
	SourceServer    string
	PublicServer    string
	FallbackServers []string
	Path            string
	Code            []byte
}

func NewNpmService() *NpmService {
	return &NpmService{
		Config: &NpmConfig{
			SourceServer: "https://registry.npmjs.org",
			Code:         []byte("npm"),
			Path:         "./data/npm",
		},
		dbLock: &sync.Mutex{},
	}
}

type NpmService struct {
	DB        *bolt.DB
	Config    *NpmConfig
	Logger    *log.Entry
	GitConfig *git.GitConfig
	Vault     *vault.Vault
	lock      bool
	dbLock    *sync.Mutex
}

func (ns *NpmService) Init(app *goapp.App) error {
	var err error

	ns.Logger.Info("Init")

	ns.DB, err = bolt.Open(fmt.Sprintf("%s/%s.db", ns.Config.Path, ns.Config.Code), 0600, &bolt.Options{
		Timeout:  1 * time.Second,
		ReadOnly: false,
	})

	if err != nil {
		return err
	}

	return ns.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(ns.Config.Code)

		return err
	})
}

func (ns *NpmService) Serve(state *goapp.GoroutineState) error {
	ns.Logger.Info("Starting Npm Service")

	for {
		ns.Logger.Info("Starting a new sync...")

		ns.SyncPackages()

		ns.Logger.Info("Wait before starting a new sync...")
		time.Sleep(60 * 15 * time.Second)
	}
}

func (ns *NpmService) End() error {
	return nil
}

func (ns *NpmService) SyncPackages() error {
	logger := ns.Logger.WithFields(log.Fields{
		"action": "SyncPackages",
	})

	logger.Info("Starting SyncPackages")

	p := make(map[string]*json.RawMessage)

	logger.WithFields(log.Fields{
		"url": fmt.Sprintf("%s/-/all", ns.Config.SourceServer),
	}).Info("Load all packages")

	f, _ := os.Create(fmt.Sprintf("%s/all.json", ns.Config.Path))

	resp, err := http.Get(fmt.Sprintf("%s/-/all", ns.Config.SourceServer))

	if err != nil {
		logger.WithError(err).Error("Unable to download npm packages")

		return err
	}

	io.Copy(f, resp.Body)

	resp.Body.Close()

	if err := pkgmirror.LoadStruct(fmt.Sprintf("%s/all.json", ns.Config.Path), &p); err != nil {
		logger.WithError(err).Error("Unable to load all npm packages")
	}

	logger.WithFields(log.Fields{
		"url": fmt.Sprintf("%s/-/all", ns.Config.SourceServer),
	}).Info("End loading packages's metadata")

	dm := pkgmirror.NewWorkerManager(40, func(id int, data <-chan interface{}, result chan interface{}) {
		for raw := range data {
			sp := raw.(ShortPackageDefinition)

			p := &FullPackageDefinition{}
			url := fmt.Sprintf("%s/%s", ns.Config.SourceServer, sp.Name)

			logger.WithFields(log.Fields{
				"package": sp.Name,
				"worker":  id,
				"url":     url,
			}).Debug("Loading package information")

			if err := pkgmirror.LoadRemoteStruct(url, p); err != nil {
				logger.WithFields(log.Fields{
					"package": sp.Name,
					"error":   err.Error(),
					"worker":  id,
				}).Error("Error loading package information")

				continue
			}

			sp.FullPackageDefinition = *p
			result <- sp
		}
	})

	dm.ResultCallback(func(data interface{}) {
		pkg := data.(ShortPackageDefinition)

		ns.savePackage(&pkg)
	})

	dm.Start()

	for name, raw := range p {
		if name == "_updated" {
			continue
		}

		sp := &ShortPackageDefinition{}
		tp := &ShortPackageDefinition{}

		if err := json.Unmarshal(*raw, sp); err != nil {
			logger.WithFields(log.Fields{
				"error":   err,
				"package": name,
			}).Error("Unable to unmarshal remote data")

			continue
		}

		store := true
		ns.DB.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(ns.Config.Code)

			if err := json.Unmarshal(b.Get([]byte(fmt.Sprintf("%s.meta", name))), tp); err == nil {
				if tp.Time.Modified == sp.Time.Modified {
					store = false
				}
			}

			return nil
		})

		if store {
			logger.WithFields(log.Fields{
				"package": name,
			}).Info("Add/Update package to process")

			dm.Add(*sp)
		} else {
			logger.WithFields(log.Fields{
				"package": name,
			}).Debug("Skip package")
		}
	}

	logger.Info("Wait worker to complete")

	dm.Wait()

	return nil
}

func (ns *NpmService) savePackage(pkg *ShortPackageDefinition) error {
	return ns.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(ns.Config.Code)

		logger := ns.Logger.WithFields(log.Fields{
			"package": pkg.Name,
		})

		logger.Info("Save package information")

		if data, err := json.Marshal(pkg); err != nil {
			return err
		} else {
			if err := b.Put([]byte(fmt.Sprintf("%s.meta", pkg.Name)), data); err != nil {
				logger.WithError(err).Error("Unable to save package meta")
			} else {
				logger.Debug("Save package meta")
			}
		}

		for _, version := range pkg.FullPackageDefinition.Versions {
			if results := NPM_ARCHIVE.FindStringSubmatch(version.Dist.Tarball); len(results) > 0 {
				version.Dist.Tarball = fmt.Sprintf("%s/npm/%s", ns.Config.PublicServer, results[3])
			} else {
				logger.WithFields(log.Fields{
					"error":   "regexp does not match",
					"tarball": version.Dist.Tarball,
				}).Error("Unable to find host")
			}
		}

		data, _ := json.Marshal(&pkg.FullPackageDefinition)

		data, err := pkgmirror.Compress(data)

		if err != nil {
			logger.WithError(err).Error("Unable to compress data")

			return err
		}

		// store the path
		if err := b.Put([]byte(pkg.Name), data); err != nil {
			logger.WithError(err).Error("Error updating/creating definition")

			return err
		} else {
			logger.Debug("Save package")
		}

		return nil
	})
}

func (ns *NpmService) Get(key string) ([]byte, error) {
	var data []byte

	ns.Logger.WithFields(log.Fields{
		"action": "Get",
		"key":    key,
	}).Info("Get raw data")

	err := ns.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(ns.Config.Code)

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

func (ns *NpmService) WriteArchive(w io.Writer, pkg, version string) error {

	logger := ns.Logger.WithFields(log.Fields{
		"package": pkg,
		"version": version,
		"action":  "WriteArchive",
	})

	vaultKey := fmt.Sprintf("%s/%s", pkg, version)

	if !ns.Vault.Has(vaultKey) {
		url := fmt.Sprintf("%s/%s/-/%s-%s.tgz", ns.Config.SourceServer, pkg, pkg, version)

		logger.WithField("url", url).Info("Create vault entry")

		resp, err := http.Get(url)

		if err != nil {
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return pkgmirror.ResourceNotFoundError
		}

		meta := vault.NewVaultMetadata()
		meta["path"] = pkg
		meta["version"] = version

		if _, err := ns.Vault.Put(vaultKey, meta, resp.Body); err != nil {
			logger.WithError(err).Info("Error while writing into vault")

			ns.Vault.Remove(vaultKey)

			return err
		}
	}

	logger.Info("Read vault entry")
	if _, err := ns.Vault.Get(vaultKey, w); err != nil {
		return err
	}

	return nil
}
