package pkgmirror

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/rande/goapp"
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

type DownloadManager struct {
	Add   chan PackageInformation
	Count int
}

func (dm *DownloadManager) Wait(state *goapp.GoroutineState, fn func(id int, done chan<- struct{}, urls <-chan PackageInformation)) {
	done := make(chan struct{})
	defer close(done)

	var wg sync.WaitGroup

	wg.Add(dm.Count)

	pkgs := make(chan PackageInformation)

	for i := 0; i < dm.Count; i++ {
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
					"package": pkg.Package,
				}).Debug("Append new package")

				pkgs <- pkg

			case <-state.In:
				log.Info("Exiting, reason: state.In signal received")
				return
			case <-done:
				log.Info("Exiting, reason: done signal received")
				return
			}
		}
	}()

	wg.Wait()

	close(pkgs)
}

func SendWithHttpCode(res http.ResponseWriter, code int, message string) {
	res.Header().Set("Content-Type", "application/json")

	res.WriteHeader(code)

	status := "KO"
	if code >= 200 && code < 300 {
		status = "OK"
	}

	data, _ := json.Marshal(map[string]string{
		"status":  status,
		"message": message,
	})

	res.Write(data)
}
