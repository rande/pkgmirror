// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mirror

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/rande/pkgmirror/mirror/composer"
	"github.com/rande/pkgmirror/test"
	"github.com/stretchr/testify/assert"
)

func Test_Composer_Get_PackagesJson(t *testing.T) {
	optin := &test.TestOptin{Composer: true}

	test.RunHttpTest(t, optin, func(args *test.Arguments) {
		// wait for the synchro to complete
		time.Sleep(1 * time.Second)

		res, err := test.RunRequest("GET", fmt.Sprintf("%s/composer/comp/packages.json", args.TestServer.URL))

		assert.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)
	})
}

func Test_Composer_Redirect(t *testing.T) {
	optin := &test.TestOptin{Composer: true}

	test.RunHttpTest(t, optin, func(args *test.Arguments) {
		// wait for the synchro to complete
		time.Sleep(1 * time.Second)

		res, err := test.RunRequest("GET", fmt.Sprintf("%s/composer/comp/p/symfony/framework-standard-edition", args.TestServer.URL))

		assert.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)

		v := &composer.PackageResult{}
		err = json.Unmarshal(res.GetBody(), v)

		assert.NoError(t, err)

		assert.Equal(t, "symfony/framework-standard-edition", v.Packages["symfony/framework-standard-edition"]["2.1.x-dev"].Name)
	})
}
