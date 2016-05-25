package pkgmirror

import "github.com/rande/goapp"

type MirrorService interface {
	Init(app *goapp.App) error
	Serve(state *goapp.GoroutineState) error
	End() error
}
