// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package static

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/rande/goapp"
	"github.com/rande/gonode/core/vault"
	"github.com/rande/pkgmirror"
)

func NewStaticService() *StaticService {
	return &StaticService{
		Config: &StaticConfig{
			Path:         "./data/static",
			SourceServer: "http://localhost",
		},
		Vault: &vault.Vault{
			Algo: "no_op",
			Driver: &vault.DriverFs{
				Root: "./cache/git",
			},
		},
	}
}

type StaticConfig struct {
	SourceServer string
	Path         string
	Code         []byte
}

type StaticService struct {
	DB        *bolt.DB
	Config    *StaticConfig
	Logger    *log.Entry
	Vault     *vault.Vault
	StateChan chan pkgmirror.State
}

func (gs *StaticService) Init(app *goapp.App) error {
	os.MkdirAll(string(filepath.Separator)+gs.Config.Path, 0755)

	var err error

	gs.DB, err = bolt.Open(fmt.Sprintf("%s/%s.db", gs.Config.Path, gs.Config.Code), 0600, &bolt.Options{
		Timeout:  1 * time.Second,
		ReadOnly: false,
	})

	if err != nil {
		return err
	}

	return gs.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(gs.Config.Code)

		return err
	})
}

func (gs *StaticService) Serve(state *goapp.GoroutineState) error {
	// nothing to do, do sync feature available

	return nil
}

func (gs *StaticService) WriteArchive(w io.Writer, path string) (*StaticFile, error) {
	logger := gs.Logger.WithFields(log.Fields{
		"path":   path,
		"action": "WriteArchive",
	})

	vaultKey := fmt.Sprintf("%s", path)
	bucketKey := vaultKey

	url := fmt.Sprintf("%s/%s", gs.Config.SourceServer, path)

	file := &StaticFile{}

	err := gs.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(gs.Config.Code)

		data := b.Get([]byte(bucketKey))

		if len(data) == 0 {
			return pkgmirror.EmptyDataError
		}

		if err := json.Unmarshal(data, file); err != nil {
			return err
		}

		return nil
	})

	if err == pkgmirror.EmptyDataError {
		file.Url = url
	} else {
		return nil, err
	}

	if !gs.Vault.Has(vaultKey) {
		logger.Info("Create vault entry")

		var wg sync.WaitGroup
		var err error
		var data []byte

		pr, pw := io.Pipe()
		wg.Add(1)

		go func() {
			meta := vault.NewVaultMetadata()
			meta["path"] = path

			if _, err := gs.Vault.Put(vaultKey, meta, pr); err != nil {
				logger.WithError(err).Info("Error while writing into vault")

				gs.Vault.Remove(vaultKey)
			}

			wg.Done()
		}()

		if err = gs.downloadStatic(pw, file); err != nil {
			logger.WithError(err).Info("Error while writing archive")

			pw.Close()
			pr.Close()

			gs.Vault.Remove(vaultKey)

			return nil, err
		} else {
			pw.Close()
		}

		wg.Wait()

		pr.Close()

		err = gs.DB.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(gs.Config.Code)

			data, err = json.Marshal(file)

			if err != nil {
				return err
			}

			if err := b.Put([]byte(bucketKey), data); err != nil {
				return err
			}

			return nil
		})
	}

	logger.Info("Read vault entry")
	if _, err := gs.Vault.Get(vaultKey, w); err != nil {
		return nil, err
	}

	return file, nil
}

func (gs *StaticService) downloadStatic(w io.Writer, file *StaticFile) error {
	logger := gs.Logger.WithFields(log.Fields{
		"url":    file.Url,
		"action": "writeArchive",
	})

	logger.Info("Start downloading the remote static file")

	resp, err := http.Get(file.Url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return pkgmirror.ResourceNotFoundError
	}

	if resp.StatusCode != 200 {
		return pkgmirror.HttpError
	}

	written, err := io.Copy(w, resp.Body)

	if err != nil {
		logger.WithError(err).Error("Error while writing input stream to the target stream")

		return err
	}

	file.Size = written
	file.Header = resp.Header
	file.DownloadAt = time.Now()

	logger.Info("Complete downloading the remote static file")

	return nil
}
