// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pkgmirror

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sync"
)


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
	// compress data for saving bytes ...
	buf := bytes.NewBuffer([]byte(""))
	if gz, err := gzip.NewWriterLevel(buf, gzip.BestCompression); err != nil {
		return nil, err
	} else {
		if _, err := gz.Write(data); err != nil {
			return nil, err
		}

		gz.Close()
	}

	return buf.Bytes(), nil
}

func NewWorkerManager(process int, processCallback FuncProcess) *workerManager {
	return &workerManager{
		count:           process,
		processCallback: processCallback,
		add:             make(chan interface{}),
		result:          make(chan interface{}),
		wg:              sync.WaitGroup{},
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
				dm.wg.Add(1)
				dm.resultCallback(raw)
				dm.wg.Done()
			}
		}()
	}
}

func (dm *workerManager) Add(raw interface{}) {
	dm.add <- raw
}

func (dm *workerManager) Wait() {
	close(dm.add)

	dm.wg.Wait()

	close(dm.result)
}

func (dm *workerManager) ResultCallback(fn FuncResult) {
	dm.resultCallback = fn
}