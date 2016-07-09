// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package test

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	godebug "runtime/debug"
	"strings"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/rande/goapp"
	"github.com/rande/pkgmirror"
	"github.com/rande/pkgmirror/api"
	"github.com/rande/pkgmirror/mirror/git"
	"github.com/rande/pkgmirror/mirror/npm"
	"github.com/stretchr/testify/assert"
	"goji.io"
)

var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

type Response struct {
	*http.Response
	RawBody  []byte
	bodyRead bool
}

type Arguments struct {
	TestServer   *httptest.Server
	MockedServer *httptest.Server
	App          *goapp.App
	T            *testing.T
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

func RunHttpTest(t *testing.T, f func(args *Arguments)) {

	baseFolder := fmt.Sprintf("%s/pkgmirror/%s", os.TempDir(), RandStringBytesMaskImprSrc(10))

	if err := os.MkdirAll(fmt.Sprintf("%s/data/npm", baseFolder), 0755); err != nil {
		assert.NoError(t, err)
		return
	}
	if err := os.MkdirAll(fmt.Sprintf("%s/data/composer", baseFolder), 0755); err != nil {
		assert.NoError(t, err)
		return
	}
	if err := os.MkdirAll(fmt.Sprintf("%s/data/git", baseFolder), 0755); err != nil {
		assert.NoError(t, err)
		return
	}
	if err := os.MkdirAll(fmt.Sprintf("%s/cache/git", baseFolder), 0755); err != nil {
		assert.NoError(t, err)
		return
	}

	cmd := exec.Command("git", strings.Split(fmt.Sprintf("clone --mirror ../../fixtures/git/foo.bare %s/data/git/local/foo.git", baseFolder), " ")...)

	if err := cmd.Start(); err != nil {
		assert.NoError(t, err)

		return
	}
	if err := cmd.Wait(); err != nil {
		assert.NoError(t, err)

		return
	}

	cmd = exec.Command("git", "update-server-info")
	cmd.Dir = fmt.Sprintf("%s/data/git/local/foo.git", baseFolder)

	if err := cmd.Start(); err != nil {
		assert.NoError(t, err)

		return
	}

	if err := cmd.Wait(); err != nil {
		assert.NoError(t, err)

		return
	}

	l := goapp.NewLifecycle()

	fs := http.FileServer(http.Dir("../../fixtures/mock"))

	ms := httptest.NewServer(fs)

	config := &pkgmirror.Config{
		DataDir:        fmt.Sprintf("%s/data", baseFolder),
		CacheDir:       fmt.Sprintf("%s/cache", baseFolder),
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
		Npm: map[string]*pkgmirror.NpmConfig{
			"npm": {
				Server:  ms.URL + "/npm",
				Enabled: true,
				Icon:    "https://cldup.com/Rg6WLgqccB.svg",
			},
		},
	}

	app, err := pkgmirror.GetApp(config, l)

	api.ConfigureApp(config, l)
	git.ConfigureApp(config, l)
	npm.ConfigureApp(config, l)

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

		f(&Arguments{
			TestServer:   ts,
			MockedServer: ms,
			T:            t,
			App:          app,
		})

		//ms.CloseClientConnections()
		//ms.Close()

		return nil
	})

	l.Go(app)
}
