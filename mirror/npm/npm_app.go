// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package npm

import (
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

		for name, conf := range config.Npm {

			if !conf.Enabled {
				continue
			}

			app.Set(fmt.Sprintf("pkgmirror.npm.%s", name), func(name string, conf *pkgmirror.NpmConfig) func(app *goapp.App) interface{} {
				return func(app *goapp.App) interface{} {
					s := NewNpmService()
					s.Config.Path = fmt.Sprintf("%s/npm", config.DataDir)
					s.Config.PublicServer = config.PublicServer
					s.Config.SourceServer = conf.Server
					s.Config.Code = []byte(name)
					s.Logger = logger.WithFields(log.Fields{
						"handler": "npm",
						"server":  s.Config.SourceServer,
						"code":    name,
					})
					s.Vault = &vault.Vault{
						Algo: "no_op",
						Driver: &vault.DriverFs{
							Root: fmt.Sprintf("%s/npm/%s_packages", config.DataDir, name),
						},
					}
					s.StateChan = pkgmirror.GetStateChannel(fmt.Sprintf("pkgmirror.npm.%s", name), app.Get("pkgmirror.channel.state").(chan pkgmirror.State))

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
		//c.Ui.Info(fmt.Sprintf("Start HTTP Server (bind: %s)", config.InternalServer))

		for name, conf := range config.Npm {
			if !conf.Enabled {
				continue
			}

			ConfigureHttp(name, conf, app)
		}

		return nil
	})

	for name, conf := range config.Npm {
		if !conf.Enabled {
			continue
		}

		l.Run(func(name string, conf *pkgmirror.NpmConfig) func(app *goapp.App, state *goapp.GoroutineState) error {
			return func(app *goapp.App, state *goapp.GoroutineState) error {
				//c.Ui.Info(fmt.Sprintf("Start Npm Sync (server: %s/npm)", config.PublicServer))
				s := app.Get(fmt.Sprintf("pkgmirror.npm.%s", name)).(*NpmService)
				s.Serve(state)

				return nil
			}
		}(name, conf))
	}
}

func ConfigureHttp(name string, conf *pkgmirror.NpmConfig, app *goapp.App) {
	mux := app.Get("mux").(*goji.Mux)
	npmService := app.Get(fmt.Sprintf("pkgmirror.npm.%s", name)).(*NpmService)

	mux.HandleFuncC(NewArchivePat(name), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "Content-Type: application/octet-stream")
		if err := npmService.WriteArchive(w, pat.Param(ctx, "package"), pat.Param(ctx, "version")); err != nil {
			pkgmirror.SendWithHttpCode(w, 500, err.Error())
		}
	})

	mux.HandleFuncC(pat.Get(fmt.Sprintf("/npm/%s/*", name)), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		pkg := r.URL.Path[6+len(name):]

		if data, err := npmService.Get(pkg); err != nil {
			pkgmirror.SendWithHttpCode(w, 404, err.Error())
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Encoding", "gzip")
			w.Write(data)
		}
	})
}
