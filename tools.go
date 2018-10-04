// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pkgmirror

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

func OpenDatabaseWithBucket(basePath string, bucket []byte) (db *bolt.DB, err error) {
	if err = os.MkdirAll(basePath, 0755); err != nil {
		return
	}

	path := fmt.Sprintf("%s/%s.db", basePath, bucket)

	db, err = bolt.Open(path, 0600, &bolt.Options{
		Timeout:  1 * time.Second,
		ReadOnly: false,
	})

	if err != nil {
		return
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)

		return err
	})

	return
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
	cpt := 0
	for {
		if err := loadRemoteStruct(url, v); err != nil {
			cpt++

			if cpt > 5 {
				return err
			}
		} else {
			return nil
		}
	}
}

func loadRemoteStruct(url string, v interface{}) error {
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

func Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	if writer, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed); err != nil {
		return nil, err
	} else if _, err := writer.Write(data); err != nil {
		return nil, err
	} else {
		writer.Close()

		return buf.Bytes(), nil
	}
}

func Decompress(data []byte) ([]byte, error) {
	if reader, err := gzip.NewReader(bytes.NewBuffer(data)); err != nil {
		return nil, err
	} else if data, err := ioutil.ReadAll(reader); err != nil {
		return nil, err
	} else if err := reader.Close(); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

func Unmarshal(data []byte, v interface{}) error {
	if data, err := Decompress(data); err != nil {
		return err
	} else if err := json.Unmarshal(data, v); err != nil {
		return err
	} else {
		return nil
	}
}

func Marshal(v interface{}) ([]byte, error) {
	if data, err := json.Marshal(v); err != nil {
		return nil, err
	} else if data, err := Compress(data); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

func NewWorkerManager(process int, processCallback FuncProcess) *workerManager {
	return &workerManager{
		count:           process,
		processCallback: processCallback,
		add:             make(chan interface{}),
		result:          make(chan interface{}),
		wg:              sync.WaitGroup{},
		resultDone:      make(chan bool),
	}
}

type FuncProcess func(id int, data <-chan interface{}, result chan interface{})
type FuncResult func(raw interface{})

type workerManager struct {
	add             chan interface{}
	result          chan interface{}
	count           int
	lock            bool
	processCallback FuncProcess
	resultCallback  FuncResult
	wg              sync.WaitGroup
	resultDone      chan bool
}

func (dm *workerManager) Start() {
	// force count to the number of worker
	dm.wg.Add(dm.count)

	for i := 0; i < dm.count; i++ {
		go func(id int) {
			dm.processCallback(id, dm.add, dm.result)
			dm.wg.Done()
		}(i)
	}

	if dm.resultCallback != nil {
		// if we get result increment wg by one
		go func() {
			for raw := range dm.result {
				dm.resultCallback(raw)
			}

			dm.resultDone <- true
		}()
	}
}

func (dm *workerManager) Add(raw interface{}) {
	dm.add <- raw
}

func (dm *workerManager) Wait() {
	// close task related actions
	close(dm.add) // close for range loop
	dm.wg.Wait()  // wait for other to

	// close result related actions
	close(dm.result)

	if dm.resultCallback != nil {
		<-dm.resultDone
	}
}

func (dm *workerManager) ResultCallback(fn FuncResult) {
	dm.resultCallback = fn
}

func Serialize(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	err := encoder.Encode(data)

	return err
}

func GetStateChannel(id string, primary chan State) chan State {
	ch := make(chan State)

	go func() {
		for {
			select {
			case s := <-ch:
				s.Id = id
				primary <- s
			}

		}
	}()

	return ch
}
