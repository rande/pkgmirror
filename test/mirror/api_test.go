// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mirror

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"encoding/json"

	"github.com/rande/goapp"
	"github.com/rande/pkgmirror/api"
	"github.com/rande/pkgmirror/test"
	"github.com/stretchr/testify/assert"
)

func Test_Api_Ping(t *testing.T) {
	test.RunHttpTest(t, func(t *testing.T, ts *httptest.Server, app *goapp.App) {
		res, err := test.RunRequest("GET", fmt.Sprintf("%s/api/ping", ts.URL))

		assert.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)
		assert.Equal(t, []byte("pong"), res.GetBody())
	})
}

func Test_Api_List(t *testing.T) {
	test.RunHttpTest(t, func(t *testing.T, ts *httptest.Server, app *goapp.App) {
		res, err := test.RunRequest("GET", fmt.Sprintf("%s/api/mirrors", ts.URL))

		assert.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)

		mirrors := []*api.ServiceMirror{}

		data := res.GetBody()

		err = json.Unmarshal(data, &mirrors)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(mirrors))
	})
}
