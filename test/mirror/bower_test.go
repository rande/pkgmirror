// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mirror

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/rande/pkgmirror/mirror/bower"
	"github.com/rande/pkgmirror/test"
	"github.com/stretchr/testify/assert"
)

func Test_Bower_Get_Package(t *testing.T) {
	optin := &test.TestOptin{Bower: true}

	test.RunHttpTest(t, optin, func(args *test.Arguments) {
		// wait for the synchro to complete
		time.Sleep(1 * time.Second)

		res, err := test.RunRequest("GET", fmt.Sprintf("%s/bower/bower/packages/non-existant-package", args.TestServer.URL))
		assert.NoError(t, err)
		assert.Equal(t, 404, res.StatusCode)

		res, err = test.RunRequest("GET", fmt.Sprintf("%s/bower/bower/packages/10digit-legal", args.TestServer.URL))
		assert.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)

		v := &bower.Package{}
		err = json.Unmarshal(res.GetBody(), v)

		assert.Equal(t, "http://localhost:8000/git/github.com/10digit/legal.git", v.Url)
		assert.Equal(t, "10digit-legal", v.Name)
	})
}

func Test_Bower_Get_Packages(t *testing.T) {
	optin := &test.TestOptin{Bower: true}

	test.RunHttpTest(t, optin, func(args *test.Arguments) {
		// wait for the synchro to complete
		time.Sleep(1 * time.Second)

		res, err := test.RunRequest("GET", fmt.Sprintf("%s/bower/bower/packages", args.TestServer.URL))
		assert.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)

		data := res.GetBody()

		v := make(bower.Packages, 0)
		err = json.Unmarshal(data, &v)

		assert.Equal(t, 3, len(v))
	})
}
