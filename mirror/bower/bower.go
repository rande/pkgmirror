// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package bower

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/rande/goapp"
	"github.com/rande/pkgmirror"
	"github.com/rande/pkgmirror/mirror/git"
)

type BowerConfig struct {
	SourceServer string
	PublicServer string
	Code         []byte
	Path         string
}

func NewBowerService() *BowerService {
	return &BowerService{
		Config: &BowerConfig{
			SourceServer: "https://registry.bower.io",
			Code:         []byte("bower"),
			Path:         "./data/bower",
		},
	}
}

type BowerService struct {
	DB        *bolt.DB
	Config    *BowerConfig
	Logger    *log.Entry
	lock      bool
	StateChan chan pkgmirror.State
}

func (bs *BowerService) Init(app *goapp.App) (err error) {
	bs.Logger.Info("Init")

	if bs.DB, err = pkgmirror.OpenDatabaseWithBucket(bs.Config.Path, bs.Config.Code); err != nil {
		bs.Logger.WithFields(log.Fields{
			log.ErrorKey: err,
			"path":       bs.Config.Path,
			"bucket":     string(bs.Config.Code),
			"action":     "Init",
		}).Error("Unable to open the internal database")
	}

	return
}

func (bs *BowerService) Serve(state *goapp.GoroutineState) error {
	bs.Logger.Info("Starting Bower Service")

	syncEnd := make(chan bool)

	sync := func() {
		bs.Logger.Info("Starting a new sync...")

		bs.SyncPackages()

		syncEnd <- true
	}

	// start the first sync
	go sync()

	for {
		select {
		case <-state.In:
			bs.DB.Close()
			return nil

		case <-syncEnd:
			bs.StateChan <- pkgmirror.State{
				Message: "Wait for a new run",
				Status:  pkgmirror.STATUS_HOLD,
			}

			bs.Logger.Info("Wait before starting a new sync...")

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

func (bs *BowerService) SyncPackages() error {
	logger := bs.Logger.WithFields(log.Fields{
		"action": "SyncPackages",
	})

	logger.Info("Starting SyncPackages")

	bs.StateChan <- pkgmirror.State{
		Message: "Syncing packages",
		Status:  pkgmirror.STATUS_RUNNING,
	}

	pkgs := make(Packages, 0)

	logger.Info("Loading bower packages")

	bs.StateChan <- pkgmirror.State{
		Message: "Loading packages list",
		Status:  pkgmirror.STATUS_RUNNING,
	}

	if err := pkgmirror.LoadRemoteStruct(fmt.Sprintf("%s/packages", bs.Config.SourceServer), &pkgs); err != nil {
		logger.WithFields(log.Fields{
			"path":       "packages",
			log.ErrorKey: err.Error(),
		}).Error("Error loading bower packages list")

		return err // an error occurs avoid empty file
	}

	logger.Info("End loading packages information!")

	for _, pkg := range pkgs {
		bs.DB.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(bs.Config.Code)

			logger := bs.Logger.WithFields(log.Fields{
				"package": pkg.Name,
			})

			saved := &Package{}
			data := b.Get([]byte(pkg.Name))

			if len(data) > 0 {
				if err := json.Unmarshal(data, saved); err != nil {
					logger.WithError(err).Info("Error while unmarshaling current package")
				} else {
					if saved.SourceUrl == pkg.Url {
						logger.Debug("Skip package!")

						return nil // same package no change, avoid io
					}
				}
			}

			bs.StateChan <- pkgmirror.State{
				Message: fmt.Sprintf("Save package information: %s", pkg.Name),
				Status:  pkgmirror.STATUS_RUNNING,
			}

			pkg.SourceUrl = pkg.Url
			pkg.Url = git.GitRewriteRepository(bs.Config.PublicServer, pkg.Url)

			data, _ = json.Marshal(pkg)

			// store the path
			if err := b.Put([]byte(pkg.Name), data); err != nil {
				logger.WithError(err).Error("Error updating/creating definition")

				return err
			}

			logger.Info("Package saved!")

			return nil
		})
	}

	bs.StateChan <- pkgmirror.State{
		Message: "End package synchronisation",
		Status:  pkgmirror.STATUS_HOLD,
	}

	return nil
}

func (bs *BowerService) Get(name string) ([]byte, error) {
	var data []byte

	bs.Logger.WithFields(log.Fields{
		"action": "Get",
		"key":    name,
	}).Info("Get package data")

	err := bs.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bs.Config.Code)

		raw := b.Get([]byte(name))

		if len(raw) == 0 {
			return pkgmirror.EmptyKeyError
		}

		data = make([]byte, len(raw))

		copy(data, raw)

		return nil
	})

	return data, err
}

func (bs *BowerService) WriteList(w io.Writer) error {
	err := bs.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bs.Config.Code)

		c := b.Cursor()

		k, v := c.First()

		w.Write([]byte{'['})

		if k == nil {
			w.Write([]byte{']'})

			return nil
		}

		w.Write(v)

		for k, v := c.Next(); k != nil; k, v = c.Next() {
			w.Write([]byte{','})
			w.Write(v)
		}
		w.Write([]byte{']'})

		return nil
	})

	return err
}
