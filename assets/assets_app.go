// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package assets

import (
	"github.com/rande/goapp"
	"github.com/rande/pkgmirror"
	"goji.io"
	"goji.io/pat"
)

// required by go-bindata
var rootDir = "./gui/build"

func ConfigureApp(config *pkgmirror.Config, l *goapp.Lifecycle) {
	l.Prepare(func(app *goapp.App) error {
		mux := app.Get("mux").(*goji.Mux)
		mux.HandleFuncC(pat.Get("/*"), Assets_GET_File(app))

		return nil
	})
}
