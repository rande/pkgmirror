package pkgmirror

import (
	"bytes"
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

type DownloadManager struct {
	Add   chan PackageInformation
	Count int
	Done  chan struct{}
}

func (dm *DownloadManager) Wait(fn func(id int, urls <-chan PackageInformation)) {
	var wg sync.WaitGroup

	wg.Add(dm.Count)

	pkgs := make(chan PackageInformation)

	for i := 0; i < dm.Count; i++ {
		go func(id int) {
			fn(id, pkgs)
			wg.Done()
		}(i)
	}

	go func() {
		// receive url
		for {
			select {
			case pkg := <-dm.Add:
				pkgs <- pkg

			case <-dm.Done:
				close(pkgs)
				return
			}
		}
	}()

	wg.Wait()
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
