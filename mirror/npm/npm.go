// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package npm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
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
	StateChan chan pkgmirror.State
}

func (ns *NpmService) Init(app *goapp.App) (err error) {
	ns.Logger.Info("Init")

	ns.Logger.WithFields(log.Fields{
		"basePath": ns.Config.Path,
		"name":     ns.Config.Code,
	}).Info("Init bolt db")

	if ns.DB, err = pkgmirror.OpenDatabaseWithBucket(ns.Config.Path, ns.Config.Code); err != nil {
		ns.Logger.WithFields(log.Fields{
			log.ErrorKey: err,
			"path":       ns.Config.Path,
			"bucket":     string(ns.Config.Code),
			"action":     "Init",
		}).Error("Unable to open the internal database")
	}

	return
}

func (ns *NpmService) Serve(state *goapp.GoroutineState) error {
	ns.Logger.Info("Starting Npm Service")

	syncEnd := make(chan bool)

	sync := func() {
		ns.Logger.Info("Starting a new sync...")

		ns.SyncPackages()

		syncEnd <- true
	}

	// start the first sync
	go sync()

	for {
		select {
		case <-state.In:
			ns.DB.Close()
			return nil

		case <-syncEnd:
			ns.StateChan <- pkgmirror.State{
				Message: "Wait for a new run",
				Status:  pkgmirror.STATUS_HOLD,
			}

			ns.Logger.Info("Wait before starting a new sync...")

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

func (ns *NpmService) SyncPackages() error {
	logger := ns.Logger.WithFields(log.Fields{
		"action": "SyncPackages",
	})

	logger.Info("Starting SyncPackages")

	ns.StateChan <- pkgmirror.State{
		Message: "Fetching packages metadatas",
		Status:  pkgmirror.STATUS_RUNNING,
	}

	logger.Info("Wait worker to complete")

	dm := pkgmirror.NewWorkerManager(10, func(id int, data <-chan interface{}, result chan interface{}) {
		for raw := range data {
			currentPkg := raw.(ShortPackageDefinition)
			remotePkg, err := ns.loadPackage(currentPkg.Name)

			if err != nil {
				logger.WithFields(log.Fields{
					"package":    currentPkg.Name,
					log.ErrorKey: err.Error(),
				}).Error("Error loading package information")

				continue
			}

			if currentPkg.Rev != remotePkg.Rev {
				logger.WithFields(log.Fields{
					"package":    currentPkg.Name,
					"currentRev": currentPkg.Rev,
					"remoteRev":  remotePkg.Rev,
					"worker":     id,
				}).Debug("Updating package information")

				result <- *remotePkg
			} else {
				logger.WithFields(log.Fields{
					"package":    currentPkg.Name,
					"currentRev": currentPkg.Rev,
					"remoteRev":  remotePkg.Rev,
					"worker":     id,
				}).Debug("Revisions are equal, nothing to update")
			}
		}
	})

	dm.ResultCallback(func(data interface{}) {
		pkg := data.(FullPackageDefinition)

		_, err := ns.savePackage(&pkg)

		if err != nil {
			logger.WithFields(log.Fields{
				"package": pkg.Name,
			}).Debug("Error while saving the package")
		}
	})

	dm.Start()

	ns.DB.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket(ns.Config.Code)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			pkg := &ShortPackageDefinition{}

			if len(k) < 5 || string(k[len(k)-5:]) != ".meta" {
				logger.WithFields(log.Fields{
					"package": string(k),
				}).Debug("Skipping non meta entry")

				continue
			}

			logger.WithFields(log.Fields{
				"package": string(k),
			}).Debug("Parsing entry")

			err := json.Unmarshal(v, pkg)

			if err != nil {
				logger.WithFields(log.Fields{
					log.ErrorKey: err,
					"package":    string(k),
				}).Error("Unable to Unmarshal the npm package")

				continue
			}

			dm.Add(*pkg)
		}

		return nil
	})

	//dm.Wait()

	return nil
}

func (ns *NpmService) savePackage(pkg *FullPackageDefinition) ([]byte, error) {
	var data []byte
	var datac []byte
	var meta []byte
	var err error

	logger := ns.Logger.WithFields(log.Fields{
		"package": pkg.Name,
	})

	logger.Info("Save package information")

	ns.StateChan <- pkgmirror.State{
		Message: fmt.Sprintf("Save package information: %s", pkg.Name),
		Status:  pkgmirror.STATUS_RUNNING,
	}

	// create the short version, to avoid storing to many useless information
	shortPkg := &ShortPackageDefinition{
		ID:   pkg.ID,
		Rev:  pkg.Rev,
		Name: pkg.Name,
	}

	if meta, err = json.Marshal(shortPkg); err != nil {
		return data, err
	}

	for _, version := range pkg.Versions {
		if results := NPM_ARCHIVE.FindStringSubmatch(version.Dist.Tarball); len(results) > 0 {
			version.Dist.Tarball = fmt.Sprintf("%s/npm/%s/%s", ns.Config.PublicServer, string(ns.Config.Code), results[3])
		} else {
			logger.WithFields(log.Fields{
				log.ErrorKey: "regexp does not match",
				"tarball":    version.Dist.Tarball,
			}).Error("Unable to find host")
		}
	}

	data, err = json.Marshal(&pkg)
	if err != nil {
		logger.WithError(err).Error("Unable to marshal data")

		return nil, err
	}

	err = ns.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(ns.Config.Code)

		logger.Debug("Saving package meta")

		if err = b.Put([]byte(fmt.Sprintf("%s.meta", pkg.Name)), meta); err != nil {
			logger.WithError(err).Error("Unable to save package meta")

			return err
		}

		datac, err = pkgmirror.Compress(data)

		if err != nil {
			logger.WithError(err).Error("Unable to compress data")

			return err
		}

		if err := b.Put([]byte(pkg.Name), datac); err != nil {
			logger.WithError(err).Error("Error updating/creating definition")

			return err
		}

		logger.Debug("Save package")

		return nil
	})

	return datac, err
}

func (ns *NpmService) loadPackage(name string) (*FullPackageDefinition, error) {

	// handle scoped package
	name = strings.Replace(name, "/", "%2f", -1)

	logger := ns.Logger.WithFields(log.Fields{
		"action": "loadPackage",
		"name":   name,
	})

	logger.Info("Load remote data")

	pkg := &FullPackageDefinition{}

	if err := pkgmirror.LoadRemoteStruct(fmt.Sprintf("%s/%s", ns.Config.SourceServer, name), &pkg); err != nil {
		logger.WithFields(log.Fields{
			"path":       fmt.Sprintf("%s/%s", ns.Config.SourceServer, name),
			log.ErrorKey: err.Error(),
		}).Error("Error loading package definition")

		return nil, err
	}

	if pkg.ID == "" {
		return nil, pkgmirror.InvalidPackageError
	}

	return pkg, nil
}

func (ns *NpmService) Get(key string) ([]byte, error) {
	var data []byte

	logger := ns.Logger.WithFields(log.Fields{
		"action": "Get",
		"key":    key,
	})

	logger.Info("Get raw data")

	err := ns.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(ns.Config.Code)

		raw := b.Get([]byte(key))

		if len(raw) == 0 {
			logger.Info("Package does not exist in local DB")

			return pkgmirror.EmptyKeyError
		}

		data = make([]byte, len(raw))

		copy(data, raw)

		return nil
	})

	// the key is not here, get it from the source
	if err == pkgmirror.EmptyKeyError {
		return ns.updatePackage(key, "")
	}

	if err != nil {
		return data, err
	}

	return data, err
}

func (ns *NpmService) updatePackage(key, rev string) ([]byte, error) {
	pkg, err := ns.loadPackage(key)

	if err != nil {
		return []byte{}, err
	}

	if pkg.Rev == rev { // nothing to update
		return []byte{}, nil
	}

	return ns.savePackage(pkg)
}

func (ns *NpmService) WriteArchive(w io.Writer, pkg, version string) error {
	logger := ns.Logger.WithFields(log.Fields{
		"package": pkg,
		"version": version,
		"action":  "WriteArchive",
	})

	vaultKey := fmt.Sprintf("%s/%s", pkg, version)

	if !ns.Vault.Has(vaultKey) {
		var url string

		if pkg[0] == '@' { // scoped package
			subNames := strings.Split(pkg, "%2f")
			url = fmt.Sprintf("%s/%s/%s/-/%s-%s.tgz", ns.Config.SourceServer, subNames[0], subNames[1], subNames[1], version)
		} else {
			url = fmt.Sprintf("%s/%s/-/%s-%s.tgz", ns.Config.SourceServer, pkg, pkg, version)
		}

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
