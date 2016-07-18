// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mirror

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/rande/pkgmirror/mirror/git"
	"github.com/rande/pkgmirror/test"
	"github.com/stretchr/testify/assert"
)

func Test_Git_Clone_Existing_Repo(t *testing.T) {
	optin := &test.TestOptin{Git: true}

	test.RunHttpTest(t, optin, func(args *test.Arguments) {
		assert.NoError(t, os.RemoveAll("/tmp/pkgmirror/foo"))

		gitService := args.App.Get("pkgmirror.git.local").(*git.GitService)

		url := fmt.Sprintf("%s/git/local/foo.git", args.TestServer.URL)

		cmd := exec.Command(gitService.Config.Binary, "clone", url, "/tmp/pkgmirror/foo")

		assert.NoError(t, cmd.Start())
		assert.NoError(t, cmd.Wait())
	})
}

func Test_Git_Clone_Non_Existing_Repo(t *testing.T) {
	optin := &test.TestOptin{Git: true}

	test.RunHttpTest(t, optin, func(args *test.Arguments) {
		assert.NoError(t, os.RemoveAll("/tmp/pkgmirror/foo"))

		gitService := args.App.Get("pkgmirror.git.local").(*git.GitService)

		assert.False(t, gitService.Has("foobar.git"))
		url := fmt.Sprintf("%s/git/local/foobar.git", args.TestServer.URL)

		cmd := exec.Command(gitService.Config.Binary, "clone", url, "/tmp/pkgmirror/foo")

		assert.NoError(t, cmd.Start())
		assert.NoError(t, cmd.Wait())

		assert.True(t, gitService.Has("foobar.git"))
	})
}

func Test_Git_Has(t *testing.T) {
	optin := &test.TestOptin{Git: true}

	test.RunHttpTest(t, optin, func(args *test.Arguments) {
		assert.NoError(t, os.RemoveAll("/tmp/pkgmirror/foo"))

		gitService := args.App.Get("pkgmirror.git.local").(*git.GitService)

		assert.True(t, gitService.Has("foo.git"))
	})
}

func Test_Git_Download_Master_Archive(t *testing.T) {
	optin := &test.TestOptin{Git: true}

	test.RunHttpTest(t, optin, func(args *test.Arguments) {
		res, _ := test.RunRequest("GET", fmt.Sprintf("%s/git/local/foo/master.zip", args.TestServer.URL))

		assert.Equal(t, 200, res.StatusCode)
		assert.Equal(t, "application/zip", res.Header.Get("Content-Type"))
	})
}

func Test_Git_Download_Tag_Archive(t *testing.T) {
	optin := &test.TestOptin{Git: true}

	test.RunHttpTest(t, optin, func(args *test.Arguments) {
		res, _ := test.RunRequest("GET", fmt.Sprintf("%s/git/local/foo/0.0.1.zip", args.TestServer.URL))

		assert.Equal(t, 200, res.StatusCode)
		assert.Equal(t, "application/zip", res.Header.Get("Content-Type"))
	})
}

func Test_Git_Download_Sha1_Archive(t *testing.T) {
	optin := &test.TestOptin{Git: true}

	test.RunHttpTest(t, optin, func(args *test.Arguments) {
		res, _ := test.RunRequest("GET", fmt.Sprintf("%s/git/local/foo/9b9cc9573693611badb397b5d01a1e6645704da7.zip", args.TestServer.URL))

		assert.Equal(t, 200, res.StatusCode)
		assert.Equal(t, "application/zip", res.Header.Get("Content-Type"))
	})
}

func Test_Git_Download_Non_Existant_Archive(t *testing.T) {
	optin := &test.TestOptin{Git: true}

	test.RunHttpTest(t, optin, func(args *test.Arguments) {
		res, _ := test.RunRequest("GET", fmt.Sprintf("%s/git/local/bar/master.zip", args.TestServer.URL))

		assert.Equal(t, 500, res.StatusCode)
		assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	})
}
