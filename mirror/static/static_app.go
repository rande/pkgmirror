// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package static

import (
	"bytes"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/rande/goapp"
	"github.com/rande/gonode/core/vault"
	"github.com/rande/pkgmirror"
	"goji.io"
	"goji.io/pat"
	"golang.org/x/net/context"
)

func ConfigureApp(config *pkgmirror.Config, l *goapp.Lifecycle) {

	l.Register(func(app *goapp.App) error {
		logger := app.Get("logger").(*log.Logger)

		v := &vault.Vault{
			Algo: "no_op",
			Driver: &vault.DriverFs{
				Root: fmt.Sprintf("%s/static", config.CacheDir),
			},
		}

		for name, conf := range config.Static {
			if !conf.Enabled {
				continue
			}

			app.Set(fmt.Sprintf("pkgmirror.static.%s", name), func(name string, conf *pkgmirror.StaticConfig) func(app *goapp.App) interface{} {
				return func(app *goapp.App) interface{} {
					s := NewStaticService()
					s.Vault = v
					s.Config.SourceServer = conf.Server
					s.Config.Path = fmt.Sprintf("%s/static", config.DataDir)
					s.Config.Code = []byte(name)
					s.Logger = logger.WithFields(log.Fields{
						"handler": "static",
						"code":    name,
					})
					s.StateChan = pkgmirror.GetStateChannel(fmt.Sprintf("pkgmirror.static.%s", name), app.Get("pkgmirror.channel.state").(chan pkgmirror.State))

					if err := s.Init(app); err != nil {
						panic(err)
					}

					return s
				}
			}(name, conf))
		}

		return nil
	})

	l.Prepare(func(app *goapp.App) error {
		for name, conf := range config.Static {
			if !conf.Enabled {
				continue
			}

			ConfigureHttp(name, conf, app)
		}

		return nil
	})

	for name, conf := range config.Static {
		if !conf.Enabled {
			continue
		}

		l.Run(func(name string) func(app *goapp.App, state *goapp.GoroutineState) error {
			return func(app *goapp.App, state *goapp.GoroutineState) error {
				s := app.Get(fmt.Sprintf("pkgmirror.static.%s", name)).(pkgmirror.MirrorService)
				s.Serve(state)

				return nil
			}
		}(name))
	}
}

func ConfigureHttp(name string, conf *pkgmirror.StaticConfig, app *goapp.App) {
	staticService := app.Get(fmt.Sprintf("pkgmirror.static.%s", name)).(*StaticService)

	mux := app.Get("mux").(*goji.Mux)

	mux.HandleFuncC(pat.Get(fmt.Sprintf("/static/%s/*", name)), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[9+len(name):]

		// this will consumes too much memory with large files.
		buf := bytes.NewBuffer([]byte(""))

		if file, err := staticService.WriteArchive(buf, path); err != nil {
			code := 500
			if err == pkgmirror.ResourceNotFoundError {
				code = 404
			}

			pkgmirror.SendWithHttpCode(w, code, err.Error())
		} else {

			// copy header
			for name := range file.Header {
				if name == "Content-Length" {
					continue
				}

				w.Header().Set(name, file.Header.Get(name))
			}

			w.WriteHeader(200)

			buf.WriteTo(w)
		}
	})
}
