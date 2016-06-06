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
		app.Set("mirror.npm", func(app *goapp.App) interface{} {
			logger := app.Get("logger").(*log.Logger)

			s := NewNpmService()
			s.Config.Path = fmt.Sprintf("%s/npm", config.DataDir)
			s.Config.PublicServer = config.PublicServer
			s.Logger = logger.WithFields(log.Fields{
				"handler": "npm",
				"server":  s.Config.SourceServer,
			})
			s.Vault = &vault.Vault{
				Algo: "no_op",
				Driver: &vault.DriverFs{
					Root: fmt.Sprintf("%s/npm/packages", config.DataDir),
				},
			}
			s.Init(app)

			return s
		})

		return nil
	})

	l.Prepare(func(app *goapp.App) error {
		//c.Ui.Info(fmt.Sprintf("Start HTTP Server (bind: %s)", config.InternalServer))

		logger := app.Get("logger").(*log.Logger)
		mux := app.Get("mux").(*goji.Mux)
		npmService := app.Get("mirror.npm").(*NpmService)

		mux.HandleFuncC(NewArchivePat(), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			pkg := pat.Param(ctx, "package")
			version := pat.Param(ctx, "version")

			logger.WithFields(log.Fields{
				"package": pkg,
				"version": version,
			}).Info("Zip archive")

			w.Header().Set("Content-Type", "Content-Type: application/octet-stream")
			if err := npmService.WriteArchive(w, pkg, version); err != nil {
				pkgmirror.SendWithHttpCode(w, 500, err.Error())
			}
		})

		mux.HandleFuncC(pat.Get("/npm/*"), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			pkg := r.URL.Path[5:]

			if data, err := npmService.Get(pkg); err != nil {
				pkgmirror.SendWithHttpCode(w, 404, err.Error())
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Content-Encoding", "gzip")
				w.Write(data)
			}
		})

		return nil
	})

	l.Run(func(app *goapp.App, state *goapp.GoroutineState) error {
		//c.Ui.Info(fmt.Sprintf("Start Npm Sync (server: %s/npm)", config.PublicServer))

		s := app.Get("mirror.npm").(pkgmirror.MirrorService)
		s.Serve(state)

		return nil
	})
}
