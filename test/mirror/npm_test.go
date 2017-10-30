// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mirror

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/rande/pkgmirror/mirror/npm"
	"github.com/rande/pkgmirror/test"
	"github.com/stretchr/testify/assert"
)

func Test_Npm_Get_Standard_Package(t *testing.T) {

	optin := &test.TestOptin{Npm: true}

	test.RunHttpTest(t, optin, func(args *test.Arguments) {
		res, err := test.RunRequest("GET", fmt.Sprintf("%s/npm/angular-oauth", args.MockedServer.URL)) // Mocked server

		assert.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)

		res, err = test.RunRequest("GET", fmt.Sprintf("%s/npm/npm/non-existant-package", args.TestServer.URL))
		assert.NoError(t, err)
		assert.Equal(t, 404, res.StatusCode)

		res, err = test.RunRequest("GET", fmt.Sprintf("%s/npm/npm/angular-nvd3-nb", args.TestServer.URL))
		assert.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)

		v := &npm.FullPackageDefinition{}
		err = json.Unmarshal(res.GetBody(), v)

		assert.Equal(t, "angular-nvd3-nb", v.Name)
		assert.Equal(t, "http://localhost:8000/npm/npm/angular-nvd3-nb/-/angular-nvd3-nb-1.0.5-nb.tgz", v.Versions["1.0.5-nb"].Dist.Tarball)

		// download tar file
		url := strings.Replace(v.Versions["1.0.5-nb"].Dist.Tarball, "http://localhost:8000", args.TestServer.URL, -1)

		res, err = test.RunRequest("GET", url)

		assert.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)
		assert.Equal(t, 19497, len(res.GetBody()))
	})
}
func Test_Npm_Download_Scoped_Package_Archive(t *testing.T) {

	optin := &test.TestOptin{Npm: true}

	test.RunHttpTest(t, optin, func(args *test.Arguments) {
		url := strings.Replace("http://localhost:8000/npm/npm/@types/react/-/react-0.0.0.tgz", "http://localhost:8000", args.TestServer.URL, -1)

		res, err := test.RunRequest("GET", url)

		assert.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)
		assert.Equal(t, 25276, len(res.GetBody()))
	})
}
