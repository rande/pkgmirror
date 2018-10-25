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
	DB            *bolt.DB
	Config        *NpmConfig
	Logger        *log.Entry
	GitConfig     *git.GitConfig
	Vault         *vault.Vault
	lock          bool
	dbLock        *sync.Mutex
	StateChan     chan pkgmirror.State
	BoltCompacter *pkgmirror.BoltCompacter
}

func (ns *NpmService) Init(app *goapp.App) (err error) {
	ns.Logger.Info("Init")

	ns.Logger.WithFields(log.Fields{
		"basePath": ns.Config.Path,
		"name":     ns.Config.Code,
	}).Info("Init bolt db")

	if err := ns.openDatabase(); err != nil {
		return err
	}

	return ns.optimize()
}

func (ns *NpmService) openDatabase() (err error) {
	if ns.DB, err = pkgmirror.OpenDatabaseWithBucket(ns.Config.Path, ns.Config.Code); err != nil {
		ns.Logger.WithFields(log.Fields{
			log.ErrorKey: err,
			"path":       ns.Config.Path,
			"bucket":     string(ns.Config.Code),
			"action":     "Init",
		}).Error("Unable to open the internal database")

		return err
	}

	return nil
}

func (ns *NpmService) optimize() error {
	ns.lock = true

	path := ns.DB.Path()

	ns.DB.Close()

	if err := ns.BoltCompacter.Compact(path); err != nil {
		return err
	}

	err := ns.openDatabase()

	ns.lock = false

	return err
}

func (ns *NpmService) Serve(state *goapp.GoroutineState) error {
	ns.Logger.Info("Starting Npm Service")

	syncEnd := make(chan bool)

	iteration := 0

	sync := func() {
		ns.Logger.Info("Starting a new sync...")

		ns.SyncPackages()

		iteration++

		// optimize every 10 iteration
		if iteration > 9 {
			ns.Logger.Info("Starting database optimization")
			ns.optimize()
			iteration = 0
		}

		syncEnd <- true
	}

	// go func() {
	// 	// Grab the initial stats.
	// 	prev := ns.DB.Stats()

	// 	for {
	// 		// Wait for 10s.
	// 		time.Sleep(10 * time.Second)

	// 		// Grab the current stats and diff them.
	// 		stats := ns.DB.Stats()
	// 		diff := stats.Sub(&prev)

	// 		ns.Logger.WithFields(log.Fields{
	// 			"FreeAlloc":     diff.FreeAlloc,
	// 			"FreePageN":     diff.FreePageN,
	// 			"PendingPageN":  diff.PendingPageN,
	// 			"FreelistInuse": diff.FreelistInuse,
	// 			"TxN":           diff.TxN,
	// 			"PageCount":     diff.TxStats.PageCount,
	// 			"PageAlloc":     diff.TxStats.PageAlloc,
	// 			"CursorCount":   diff.TxStats.CursorCount,
	// 			"NodeCount":     diff.TxStats.NodeCount,
	// 			"NodeDeref":     diff.TxStats.NodeDeref,
	// 			"Rebalance":     diff.TxStats.Rebalance,
	// 			"RebalanceTime": diff.TxStats.RebalanceTime,
	// 			"Split":         diff.TxStats.Split,
	// 			"Spill":         diff.TxStats.Spill,
	// 			"SpillTime":     diff.TxStats.SpillTime,
	// 			"Write":         diff.TxStats.Write,
	// 			"WriteTime":     diff.TxStats.WriteTime,
	// 		}).Info("Dump stats")

	// 		// Save stats for the next loop.
	// 		prev = stats
	// 	}
	// }()

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
	if ns.lock {
		return pkgmirror.DatabaseLockedError
	}

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

			fields := log.Fields{
				"package":         currentPkg.Name,
				"currentRev":      currentPkg.Rev,
				"remoteRev":       remotePkg.Rev,
				"worker":          id,
				"currentReleases": currentPkg.ReleasesAvailable,
				"remoteReleases":  len(remotePkg.Versions),
			}

			if currentPkg.Rev != remotePkg.Rev || currentPkg.ReleasesAvailable != len(remotePkg.Versions) {
				logger.WithFields(fields).Debug("Updating package information")

				result <- *remotePkg
			} else {
				logger.WithFields(fields).Debug("Revisions are equal, nothing to update")
			}
		}
	})

	dm.ResultCallback(func(data interface{}) {
		pkg := data.(FullPackageDefinition)

		err := ns.savePackage(&pkg)

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

	logger.Info("Wait for download to complete")

	ns.StateChan <- pkgmirror.State{
		Message: "Wait for download to complete",
		Status:  pkgmirror.STATUS_RUNNING,
	}

	dm.Wait()

	return nil
}

func (ns *NpmService) savePackage(pkg *FullPackageDefinition) error {
	if ns.lock {
		return pkgmirror.DatabaseLockedError
	}

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
		ID:                pkg.ID,
		Rev:               pkg.Rev,
		Name:              pkg.Name,
		ReleasesAvailable: len(pkg.Versions),
	}

	if meta, err = json.Marshal(shortPkg); err != nil {
		return err
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

		return err
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

	return err
}

func (ns *NpmService) loadPackage(name string) (*FullPackageDefinition, error) {
	// handle scoped package
	name = strings.Replace(name, "/", "%2f", -1)

	logger := ns.Logger.WithFields(log.Fields{
		"action": "loadPackage",
		"name":   name,
	})

	logger.Debug("Load remote data")

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

	if ns.lock {
		return data, pkgmirror.DatabaseLockedError
	}

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
		logger.Info("Package does not exist")

		if err := ns.UpdatePackage(key); err != nil {
			return data, err
		}

		return ns.Get(key)
	}

	return data, err
}

func (ns *NpmService) UpdatePackage(key string) error {
	if ns.lock {
		return pkgmirror.DatabaseLockedError
	}

	pkg, err := ns.loadPackage(key)

	if err != nil {
		return err
	}

	return ns.savePackage(pkg)
}

func (ns *NpmService) WriteArchive(w io.Writer, pkg, version string) error {
	if ns.lock {
		return pkgmirror.DatabaseLockedError
	}

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
