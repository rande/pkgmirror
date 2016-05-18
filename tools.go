package pkgmirror

import (
	"github.com/boltdb/bolt"
	"fmt"
	"time"
	"encoding/json"
	"bytes"
	"os"
	"net/http"
	"io"
	log "github.com/Sirupsen/logrus"
	"sync"
)

func LoadDB(name string) (*bolt.DB, error) {
	db, err := bolt.Open(fmt.Sprintf("%s.db", name), 0600, &bolt.Options{
		Timeout: 1 * time.Second,
	})

	return db, err
}

func LoadStruct(file string, v interface{}) error {
	r, err := os.Open(file)

	if err != nil {
		return err
	}

	buf := bytes.NewBuffer([]byte(""))
	buf.ReadFrom(r)

	err = json.Unmarshal(buf.Bytes(), v)

	if err != nil {
		return err
	}

	return nil
}

func LoadRemoteStruct(url string, v interface{}) error {
	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	buf := bytes.NewBuffer([]byte(""))

	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(buf.Bytes(), v)

	if err != nil {
		return err
	}

	return nil
}

func NewDownloadManager() *DownloadManager {
	return &DownloadManager{
		Add: make(chan PackageInformation),
	}
}

type DownloadManager struct {
	Add     chan PackageInformation
}

func (dm *DownloadManager) Wait(c int, fn func(id int, done chan<- struct{}, urls <-chan PackageInformation)) {
	done := make(chan struct{})
	defer close(done)

	var wg sync.WaitGroup

	wg.Add(c)

	pkgs := make(chan PackageInformation)

	for i := 0; i < c; i++ {
		go func(id int) {
			fn(id, done, pkgs)
			wg.Done()
		}(i)
	}

	go func() {
		// receive url
		for {
			select {
			case pkg := <-dm.Add:
				log.WithFields(log.Fields{
					"url": pkg.Url,
				}).Info("Append new urls")

				pkgs <- pkg

			case <-done:
				log.WithFields(log.Fields{
				}).Info("Exiting proxy receiving urls")
				return
			}
		}
	}()

	log.WithFields(log.Fields{
		"worker": c,
	}).Info("Waiting for action")

	wg.Wait()
}
