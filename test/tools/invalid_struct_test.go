// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package tools

import (
	"fmt"
	"testing"

	"github.com/rande/pkgmirror"
	"github.com/rande/pkgmirror/test"
	"github.com/stretchr/testify/assert"
)

func Test_Invalid_Struct(t *testing.T) {
	optin := &test.TestOptin{}

	test.RunHttpTest(t, optin, func(args *test.Arguments) {

		fake := &struct {
			Foo string
		}{}

		err := pkgmirror.LoadRemoteStruct(fmt.Sprintf("%s/invalid.json", args.MockedServer.URL), fake)

		assert.Error(t, err)
	})
}
