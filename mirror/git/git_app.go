// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package git

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

		vault := &vault.Vault{
			Algo: "no_op",
			Driver: &vault.DriverFs{
				Root: fmt.Sprintf("%s/git", config.CacheDir),
			},
		}

		for name, conf := range config.Git {
			if !conf.Enabled {
				continue
			}

			app.Set(fmt.Sprintf("pkgmirror.git.%s", name), func(name string, conf *pkgmirror.GitConfig) func(app *goapp.App) interface{} {

				return func(app *goapp.App) interface{} {
					s := NewGitService()
					s.Config.Server = conf.Server
					s.Config.PublicServer = config.PublicServer
					s.Config.DataDir = fmt.Sprintf("%s/git", config.DataDir)
					s.Vault = vault
					s.Logger = logger.WithFields(log.Fields{
						"handler": "git",
						"code":    name,
					})
					s.StateChan = pkgmirror.GetStateChannel(fmt.Sprintf("pkgmirror.git.%s", name), app.Get("pkgmirror.channel.state").(chan pkgmirror.State))
					s.Init(app)

					return s
				}
			}(name, conf))
		}

		return nil
	})

	l.Prepare(func(app *goapp.App) error {
		for name, conf := range config.Git {
			if !conf.Enabled {
				continue
			}

			ConfigureHttp(name, conf, app)
		}

		return nil
	})

	for name, conf := range config.Git {
		if !conf.Enabled {
			continue
		}

		l.Run(func(name string) func(app *goapp.App, state *goapp.GoroutineState) error {

			return func(app *goapp.App, state *goapp.GoroutineState) error {
				s := app.Get(fmt.Sprintf("pkgmirror.git.%s", name)).(pkgmirror.MirrorService)
				s.Serve(state)

				return nil
			}
		}(name))
	}
}

func ConfigureHttp(name string, conf *pkgmirror.GitConfig, app *goapp.App) {
	mux := app.Get("mux").(*goji.Mux)

	gitService := app.Get(fmt.Sprintf("pkgmirror.git.%s", name)).(*GitService)

	mux.HandleFuncC(NewGitPat(conf.Server), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		if err := gitService.WriteArchive(w, fmt.Sprintf("%s.git", pat.Param(ctx, "path")), pat.Param(ctx, "ref")); err != nil {
			pkgmirror.SendWithHttpCode(w, 500, err.Error())
		}
	})

	mux.HandleFuncC(pat.Get(fmt.Sprintf("/git/%s/*", conf.Server)), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		if err := gitService.WriteFile(w, r.URL.Path[6+len(conf.Server):]); err != nil {
			w.WriteHeader(http.StatusNotFound)
		}
	})
}
