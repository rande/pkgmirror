// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package composer

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/rande/goapp"
	"github.com/rande/pkgmirror"
	"github.com/rande/pkgmirror/mirror/git"
	"goji.io"
	"goji.io/pat"
	"golang.org/x/net/context"
)

func ConfigureApp(config *pkgmirror.Config, l *goapp.Lifecycle) {

	l.Register(func(app *goapp.App) error {
		app.Set("mirror.composer", func(app *goapp.App) interface{} {
			logger := app.Get("logger").(*log.Logger)

			s := NewComposerService()
			s.Config.Path = fmt.Sprintf("%s/composer", config.DataDir)
			s.GitConfig = app.Get("mirror.git").(*git.GitService).Config
			s.Logger = logger.WithFields(log.Fields{
				"handler": "composer",
				"server":  s.Config.SourceServer,
			})
			s.Init(app)

			return s
		})

		return nil
	})

	l.Run(func(app *goapp.App, state *goapp.GoroutineState) error {
		//c.Ui.Info(fmt.Sprintf("Start Composer Sync (ref: %s/packagist)", config.PublicServer))

		s := app.Get("mirror.composer").(pkgmirror.MirrorService)
		s.Serve(state)

		return nil
	})

	l.Run(func(app *goapp.App, state *goapp.GoroutineState) error {
		//c.Ui.Info(fmt.Sprintf("Start HTTP Server (bind: %s)", config.InternalServer))

		mux := app.Get("mux").(*goji.Mux)
		composerService := app.Get("mirror.composer").(*ComposerService)

		mux.HandleFuncC(pat.Get("/packagist"), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/packagist/packages.json", http.StatusMovedPermanently)
		})

		mux.HandleFuncC(pat.Get("/packagist/packages.json"), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if data, err := composerService.Get("packages.json"); err != nil {
				pkgmirror.SendWithHttpCode(w, 500, err.Error())
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(data)
			}
		})

		mux.HandleFuncC(pat.Get("/packagist/p/:ref.json"), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if data, err := composerService.Get(fmt.Sprintf("p/%s.json", pat.Param(ctx, "ref"))); err != nil {
				pkgmirror.SendWithHttpCode(w, 404, err.Error())
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(data)
			}
		})

		mux.HandleFuncC(NewPackagePat(), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			pkg := fmt.Sprintf("%s/%s$%s", pat.Param(ctx, "vendor"), pat.Param(ctx, "package"), pat.Param(ctx, "ref"))

			if refresh := r.FormValue("refresh"); len(refresh) > 0 {
				w.Header().Set("Content-Type", "application/json")

				if err := composerService.UpdatePackage(pkg); err != nil {
					pkgmirror.SendWithHttpCode(w, 500, err.Error())
				} else {
					pkgmirror.SendWithHttpCode(w, 200, "Package updated")
				}

				return
			}

			if data, err := composerService.Get(pkg); err != nil {
				pkgmirror.SendWithHttpCode(w, 404, err.Error())
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Content-Encoding", "gzip")
				w.Write(data)
			}
		})

		mux.HandleFuncC(NewPackageInfoPat(), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if pi, err := composerService.GetPackage(fmt.Sprintf("%s/%s", pat.Param(ctx, "vendor"), pat.Param(ctx, "package"))); err != nil {
				pkgmirror.SendWithHttpCode(w, 404, err.Error())
			} else {
				switch pat.Param(ctx, "format") {
				case "html":
					http.Redirect(w, r, fmt.Sprintf("/packagist/p/%s.json", pi.GetTargetKey()), http.StatusFound)
				}
			}
		})

		return nil
	})
}
