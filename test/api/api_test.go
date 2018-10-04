// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package api

import (
	"fmt"
	"testing"

	//"encoding/json"

	//"github.com/rande/pkgmirror/api"
	"github.com/rande/pkgmirror/test"
	"github.com/stretchr/testify/assert"
)

func Test_Api_Ping(t *testing.T) {
	optin := &test.TestOptin{}

	test.RunHttpTest(t, optin, func(args *test.Arguments) {
		res, err := test.RunRequest("GET", fmt.Sprintf("%s/api/ping", args.TestServer.URL))

		assert.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)
		assert.Equal(t, []byte("pong"), res.GetBody())
	})
}

//func Test_Api_List(t *testing.T) {
//	optin := &test.TestOptin{true, true, true, true}
//
//	test.RunHttpTest(t, optin, func(args *test.Arguments) {
//		res, err := test.RunRequest("GET", fmt.Sprintf("%s/api/mirrors", args.TestServer.URL))
//
//		assert.NoError(t, err)
//		assert.Equal(t, 200, res.StatusCode)
//
//		mirrors := []*api.ServiceMirror{}
//
//		data := res.GetBody()
//
//		err = json.Unmarshal(data, &mirrors)
//		assert.NoError(t, err)
//
//		assert.Equal(t, 4, len(mirrors))
//	})
//}
