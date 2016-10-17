// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mirror

import (
	"fmt"
	"testing"

	"github.com/rande/pkgmirror/test"
	"github.com/stretchr/testify/assert"
)

func Test_Static_Get_Valid_File(t *testing.T) {
	optin := &test.TestOptin{Static: true}

	test.RunHttpTest(t, optin, func(args *test.Arguments) {
		// check the original file exist on the remote server
		res, err := test.RunRequest("GET", fmt.Sprintf("%s/static/file.txt", args.MockedServer.URL))

		assert.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)
		assert.Equal(t, "This is a sample test file.", string(res.GetBody()))

		// get the proxied file

		res, err = test.RunRequest("GET", fmt.Sprintf("%s/static/static/file.txt", args.TestServer.URL))

		assert.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)
		assert.Equal(t, "This is a sample test file.", string(res.GetBody()))
	})
}

func Test_Static_Get_Invalid_File(t *testing.T) {
	optin := &test.TestOptin{Static: true}

	test.RunHttpTest(t, optin, func(args *test.Arguments) {
		// check the original file exist on the remote server
		res, err := test.RunRequest("GET", fmt.Sprintf("%s/static/static/non-existant-file.txt", args.TestServer.URL))

		assert.NoError(t, err)
		assert.Equal(t, 404, res.StatusCode)
	})
}
