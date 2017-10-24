// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package bower

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/rande/goapp"
	"github.com/rande/pkgmirror"
	"goji.io"
	"goji.io/pat"
	"golang.org/x/net/context"
)

func ConfigureApp(config *pkgmirror.Config, l *goapp.Lifecycle) {

	l.Register(func(app *goapp.App) error {
		logger := app.Get("logger").(*log.Logger)

		for name, conf := range config.Bower {
			if !conf.Enabled {
				continue
			}

			app.Set(fmt.Sprintf("pkgmirror.bower.%s", name), func(name string, conf *pkgmirror.BowerConfig) func(app *goapp.App) interface{} {
				return func(app *goapp.App) interface{} {
					s := NewBowerService()
					s.Config.Code = []byte(name)
					s.Config.Path = fmt.Sprintf("%s/bower", config.DataDir)
					s.Config.PublicServer = config.PublicServer
					s.Config.SourceServer = conf.Server
					s.Logger = logger.WithFields(log.Fields{
						"handler": "bower",
						"server":  s.Config.SourceServer,
						"code":    name,
					})
					s.StateChan = pkgmirror.GetStateChannel(fmt.Sprintf("pkgmirror.bower.%s", name), app.Get("pkgmirror.channel.state").(chan pkgmirror.State))

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
		for name, conf := range config.Bower {
			if !conf.Enabled {
				continue
			}

			ConfigureHttp(name, conf, app)
		}

		return nil
	})

	for name, conf := range config.Bower {
		if !conf.Enabled {
			continue
		}

		l.Run(func(name string) func(app *goapp.App, state *goapp.GoroutineState) error {
			return func(app *goapp.App, state *goapp.GoroutineState) error {
				s := app.Get(fmt.Sprintf("pkgmirror.bower.%s", name)).(pkgmirror.MirrorService)
				s.Serve(state)

				return nil
			}
		}(name))
	}
}

func ConfigureHttp(name string, conf *pkgmirror.BowerConfig, app *goapp.App) {
	mux := app.Get("mux").(*goji.Mux)
	bowerService := app.Get(fmt.Sprintf("pkgmirror.bower.%s", name)).(*BowerService)

	mux.HandleFuncC(pat.Get(fmt.Sprintf("/bower/%s/packages", name)), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := bowerService.WriteList(w); err != nil {
			pkgmirror.SendWithHttpCode(w, 500, err.Error())
		}
	})

	mux.HandleFuncC(pat.Get(fmt.Sprintf("/bower/%s/packages/:name", name)), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		if data, err := bowerService.Get(fmt.Sprintf("%s", pat.Param(ctx, "name"))); err != nil {
			pkgmirror.SendWithHttpCode(w, 404, err.Error())
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		}
	})
}
