// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	godebug "runtime/debug"
	"strings"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/rande/goapp"
	"github.com/rande/pkgmirror"
	"github.com/rande/pkgmirror/api"
	"github.com/rande/pkgmirror/mirror/git"
	"github.com/stretchr/testify/assert"
	"goji.io"
)

type Response struct {
	*http.Response
	RawBody  []byte
	bodyRead bool
}

func (r Response) GetBody() []byte {
	var err error

	if !r.bodyRead {
		r.RawBody, err = ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			log.Fatal(err)
		}

		r.bodyRead = true
	}

	return r.RawBody
}

func RunRequest(method string, path string, options ...interface{}) (*Response, error) {
	var body interface{}
	var headers map[string]string

	if len(options) > 0 {
		body = options[0]
	}

	if len(options) > 1 {
		headers = options[1].(map[string]string)
	}

	client := &http.Client{}
	var req *http.Request
	var err error

	switch v := body.(type) {
	case nil:
		req, err = http.NewRequest(method, path, nil)
	case *strings.Reader:
		req, err = http.NewRequest(method, path, v)
	case io.Reader:
		req, err = http.NewRequest(method, path, v)

	case url.Values:
		req, err = http.NewRequest(method, path, strings.NewReader(v.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	default:
		panic(fmt.Sprintf("please add a new test case for %T", body))
	}

	if headers != nil {
		for name, value := range headers {
			req.Header.Set(name, value)
		}
	}

	if err != nil {
		panic(err)
	}

	resp, err := client.Do(req)

	return &Response{Response: resp}, err
}

func RunHttpTest(t *testing.T, f func(t *testing.T, ts *httptest.Server, app *goapp.App)) {
	l := goapp.NewLifecycle()

	config := &pkgmirror.Config{
		DataDir:        "/tmp/pkgmirror/data",
		CacheDir:       "/tmp/pkmirror/cache",
		PublicServer:   "http://localhost:8000",
		InternalServer: "127.0.0.1:8000",
		LogLevel:       "debug",
		Git: map[string]*pkgmirror.GitConfig{
			"local": {
				Server:  "local",
				Enabled: true,
				Icon:    "https://assets-cdn.github.com/images/modules/logos_page/GitHub-Mark.png",
				Clone:   "",
			},
		},
	}

	app, err := pkgmirror.GetApp(config, l)

	api.ConfigureApp(config, l)
	git.ConfigureApp(config, l)

	assert.NoError(t, err)
	assert.NotNil(t, app)

	l.Run(func(app *goapp.App, state *goapp.GoroutineState) error {

		mux := app.Get("mux").(*goji.Mux)

		ts := httptest.NewServer(mux)

		defer func() {
			state.Out <- goapp.Control_Stop

			ts.Close()

			if r := recover(); r != nil {
				assert.Equal(t, false, true, fmt.Sprintf("RunHttpTest: Panic recovered, message=%s\n\n%s", r, string(godebug.Stack()[:])))
			}
		}()

		f(t, ts, app)

		ts.CloseClientConnections()
		ts.Close()

		return nil
	})

	l.Go(app)
}
