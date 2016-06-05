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
		app.Set("mirror.git", func(app *goapp.App) interface{} {
			logger := app.Get("logger").(*log.Logger)

			s := NewGitService()
			s.Config.Server = config.PublicServer
			s.Config.DataDir = fmt.Sprintf("%s/git", config.DataDir)
			s.Vault = &vault.Vault{
				Algo: "no_op",
				Driver: &vault.DriverFs{
					Root: fmt.Sprintf("%s/git", config.CacheDir),
				},
			}
			s.Logger = logger.WithFields(log.Fields{
				"handler": "git",
			})
			s.Init(app)

			return s
		})

		return nil
	})

	l.Run(func(app *goapp.App, state *goapp.GoroutineState) error {
		//c.Ui.Info(fmt.Sprintf("Start Git Sync (server: %s/git)", config.PublicServer))

		s := app.Get("mirror.git").(pkgmirror.MirrorService)
		s.Serve(state)

		return nil
	})

	l.Run(func(app *goapp.App, state *goapp.GoroutineState) error {
		//c.Ui.Info(fmt.Sprintf("Start HTTP Server (bind: %s)", config.InternalServer))

		logger := app.Get("logger").(*log.Logger)
		mux := app.Get("mux").(*goji.Mux)
		gitService := app.Get("mirror.git").(*GitService)

		mux.HandleFuncC(NewGitPat(), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			path := pat.Param(ctx, "path")
			ref := pat.Param(ctx, "ref")

			logger.WithFields(log.Fields{
				"path": path,
				"ref":  ref,
			}).Info("Zip archive")

			w.Header().Set("Content-Type", "application/zip")
			if err := gitService.WriteArchive(w, fmt.Sprintf("%s.git", path), ref); err != nil {
				pkgmirror.SendWithHttpCode(w, 500, err.Error())
			}
		})

		mux.HandleFuncC(pat.Get("/git/*"), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			logger.WithFields(log.Fields{
				"path": r.URL.Path[5:],
			}).Info("Git fetch")

			if err := gitService.WriteFile(w, r.URL.Path[5:]); err != nil {
				w.WriteHeader(http.StatusNotFound)
			}
		})

		return nil
	})
}
