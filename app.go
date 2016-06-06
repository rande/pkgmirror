// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pkgmirror

import (
	"errors"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/rande/goapp"
	"goji.io"
	"golang.org/x/net/context"
)

func GetApp(conf *Config) (*goapp.App, error) {

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

		m.UseC(func(h goji.Handler) goji.Handler {
			return goji.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
				logger.WithFields(log.Fields{
					"request_method":      r.Method,
					"request_url":         r.URL.String(),
					"request_remote_addr": r.RemoteAddr,
					"request_host":        r.Host,
				}).Info("Receive HTTP request")

				//t1 := time.Now()
				h.ServeHTTPC(ctx, w, r)
				//t2 := time.Now()
			})
		})

		return m
	})

	return app, nil
}
