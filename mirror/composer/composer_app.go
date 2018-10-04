// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package composer

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

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

		for name, conf := range config.Composer {
			if !conf.Enabled {
				continue
			}

			app.Set(fmt.Sprintf("pkgmirror.composer.%s", name), func(name string, conf *pkgmirror.ComposerConfig) func(app *goapp.App) interface{} {
				return func(app *goapp.App) interface{} {

					s := NewComposerService()

					if u, err := url.Parse(conf.Server); err != nil {
						panic(err)
					} else {
						s.Config.BasePublicServer = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
					}

					s.Config.Path = fmt.Sprintf("%s/composer", config.DataDir)
					s.Config.PublicServer = config.PublicServer
					s.Config.SourceServer = conf.Server

					s.Config.Code = []byte(name)
					s.Logger = logger.WithFields(log.Fields{
						"handler": "composer",
						"server":  s.Config.SourceServer,
						"code":    name,
					})
					s.StateChan = pkgmirror.GetStateChannel(fmt.Sprintf("pkgmirror.composer.%s", name), app.Get("pkgmirror.channel.state").(chan pkgmirror.State))

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
		// BC Compatible
		mux := app.Get("mux").(*goji.Mux)
		mux.HandleFuncC(pat.Get("/packagist/*"), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/composer"+r.URL.EscapedPath(), http.StatusMovedPermanently)
		})

		for name, conf := range config.Composer {
			if !conf.Enabled {
				continue
			}

			ConfigureHttp(name, conf, app)
		}

		return nil
	})

	for name, conf := range config.Composer {
		if !conf.Enabled {
			continue
		}

		l.Run(func(name string) func(app *goapp.App, state *goapp.GoroutineState) error {
			return func(app *goapp.App, state *goapp.GoroutineState) error {
				s := app.Get(fmt.Sprintf("pkgmirror.composer.%s", name)).(pkgmirror.MirrorService)
				s.Serve(state)

				return nil
			}
		}(name))
	}
}

// http://localhost:8000/composer/drupal8/drupal/provider-2011-2%24e22123ab0815d43cedb1309f7ad7b803127ac9679f7aaa9b281cf768f6806ae2.json

func ConfigureHttp(name string, conf *pkgmirror.ComposerConfig, app *goapp.App) {
	mux := app.Get("mux").(*goji.Mux)
	logger := app.Get("logger").(*log.Logger).WithFields(log.Fields{
		"handler": "composer",
		"code":    name,
	})

	composerService := app.Get(fmt.Sprintf("pkgmirror.composer.%s", name)).(*ComposerService)

	mux.HandleFuncC(pat.Get(fmt.Sprintf("/composer/%s(/|)", name)), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, fmt.Sprintf("/composer/%s/packages.json", name), http.StatusMovedPermanently)
	})

	mux.HandleFuncC(NewPackagePat(name), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFuncC(NewPackageInfoPat(name), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		if pi, err := composerService.GetPackage(fmt.Sprintf("%s/%s", pat.Param(ctx, "vendor"), pat.Param(ctx, "package"))); err != nil {
			pkgmirror.SendWithHttpCode(w, 404, err.Error())
		} else {
			switch pat.Param(ctx, "format") {
			case "html":
				http.Redirect(w, r, fmt.Sprintf("/composer/%s/p/%s.json", name, pi.GetTargetKey()), http.StatusFound)
			}
		}
	})

	baseUrlLen := len(fmt.Sprintf("/composer/%s/", name))

	// catch all for this element (drupal element)
	mux.HandleFuncC(pat.Get(fmt.Sprintf("/composer/%s/*", name)), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		logger.Debug(r.URL.Path)

		format := "html"
		url := r.URL.Path[baseUrlLen:]
		key := url
		hash := ""
		ref := ""

		if i := strings.Index(url, "."); i > 0 {
			format = url[i+1:]
			key = url[:i]
		}

		if i := strings.Index(key, "$"); i > 0 {
			hash = key[i+1:]
			ref = key[:i]
		}

		logger.WithFields(log.Fields{
			"url":    url,
			"format": format,
			"key":    key,
			"ref":    ref,
			"hash":   hash,
		}).Debug(url)

		if data, err := composerService.Get(url); err != nil {
			pkgmirror.SendWithHttpCode(w, 404, err.Error())
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Encoding", "gzip")
			w.Write(data)
		}
	})
}
