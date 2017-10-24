// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pkgmirror

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/NYTimes/gziphandler"
	log "github.com/Sirupsen/logrus"
	"github.com/bakins/logrus-middleware"
	"github.com/rande/goapp"
	"goji.io"
)

func GetApp(conf *Config, l *goapp.Lifecycle) (*goapp.App, error) {

	app := goapp.NewApp()

	// init logger
	logger := log.New()
	if level, err := log.ParseLevel(conf.LogLevel); err != nil {
		return app, errors.New(fmt.Sprintf("Unable to parse the log level: %s", conf.LogLevel))
	} else {
		logger.Level = level
	}

	if len(conf.DataDir) == 0 {
		return app, errors.New("Please configure DataDir")
	}

	app.Set("logger", func(app *goapp.App) interface{} {
		return logger
	})

	app.Set("config", func(app *goapp.App) interface{} {
		return conf
	})

	app.Set("mux", func(app *goapp.App) interface{} {
		m := goji.NewMux()

		m.Use(func(h http.Handler) http.Handler {
			gzip := gziphandler.GzipHandler(h)

			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				skip := false
				if len(r.URL.Path) > 4 && r.URL.Path[1:4] == "npm" {
					skip = true
				}

				if len(r.URL.Path) > 8 && r.URL.Path[1:9] == "composer" {
					skip = true
				}

				if r.URL.Path == "/api/sse" {
					skip = true
				}

				if skip {
					h.ServeHTTP(w, r)
				} else {
					gzip.ServeHTTP(w, r)
				}
			})
		})

		m.Use(func(h http.Handler) http.Handler {
			lm := &logrusmiddleware.Middleware{
				Logger: logger,
				Name:   "pkgmirror",
			}

			return lm.Handler(h, "http")
		})

		return m
	})

	return app, nil
}
