// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mirror

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/rande/goapp"
	"github.com/rande/pkgmirror/test"
	"github.com/stretchr/testify/assert"
)

func Test_Git_Download_File(t *testing.T) {
	test.RunHttpTest(t, func(t *testing.T, ts *httptest.Server, app *goapp.App) {
		res, _ := test.RunRequest("GET", fmt.Sprintf("%s/git/local/foo.git/HEAD", ts.URL))

		assert.Equal(t, 200, res.StatusCode)
		data := res.GetBody()

		assert.Equal(t, []byte("ref: refs/heads/master\n"), data, "Unable to d/l HEAD file")
	})
}

func Test_Git_Download_Master_Archive(t *testing.T) {
	test.RunHttpTest(t, func(t *testing.T, ts *httptest.Server, app *goapp.App) {
		res, _ := test.RunRequest("GET", fmt.Sprintf("%s/git/local/foo/master.zip", ts.URL))

		assert.Equal(t, 200, res.StatusCode)
		assert.Equal(t, "application/zip", res.Header.Get("Content-Type"))
	})
}

func Test_Git_Download_Tag_Archive(t *testing.T) {
	test.RunHttpTest(t, func(t *testing.T, ts *httptest.Server, app *goapp.App) {
		res, _ := test.RunRequest("GET", fmt.Sprintf("%s/git/local/foo/0.0.1.zip", ts.URL))

		assert.Equal(t, 200, res.StatusCode)
		assert.Equal(t, "application/zip", res.Header.Get("Content-Type"))
	})
}

func Test_Git_Download_Sha1_Archive(t *testing.T) {
	test.RunHttpTest(t, func(t *testing.T, ts *httptest.Server, app *goapp.App) {
		res, _ := test.RunRequest("GET", fmt.Sprintf("%s/git/local/foo/9b9cc9573693611badb397b5d01a1e6645704da7.zip", ts.URL))

		assert.Equal(t, 200, res.StatusCode)
		assert.Equal(t, "application/zip", res.Header.Get("Content-Type"))
	})
}

func Test_Git_Download_Non_Existant_Archive(t *testing.T) {
	test.RunHttpTest(t, func(t *testing.T, ts *httptest.Server, app *goapp.App) {
		res, _ := test.RunRequest("GET", fmt.Sprintf("%s/git/local/bar/master.zip", ts.URL))

		assert.Equal(t, 500, res.StatusCode)
		assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	})
}
