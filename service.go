// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pkgmirror

import "github.com/rande/goapp"

type MirrorService interface {
	Init(app *goapp.App) error
	Serve(state *goapp.GoroutineState) error
	End() error
}
