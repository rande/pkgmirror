// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pkgmirror

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
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
