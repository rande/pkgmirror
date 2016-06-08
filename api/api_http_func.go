// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package api

import (
	"fmt"
	"net/http"

	"github.com/rande/goapp"
	"github.com/rande/pkgmirror"
	"golang.org/x/net/context"
)

func Api_GET_MirrorServices(app *goapp.App) func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	config := app.Get("config").(*pkgmirror.Config)

	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		d := []*ServiceMirror{}

		for code, conf := range config.Git {
			s := &ServiceMirror{}
			s.Icon = conf.Icon
			s.Type = "git"
			s.Name = code
			s.SourceUrl = conf.Server
			s.TargetUrl = fmt.Sprintf("%s/git/%s", config.PublicServer, conf.Server)
			s.Enabled = conf.Enabled

			d = append(d, s)
		}

		for code, conf := range config.Npm {
			s := &ServiceMirror{}
			s.Icon = conf.Icon
			s.Type = "npm"
			s.Name = code
			s.SourceUrl = conf.Server
			s.TargetUrl = fmt.Sprintf("%s/npm/%s", config.PublicServer, code)
			s.Enabled = conf.Enabled

			d = append(d, s)
		}

		for code, conf := range config.Composer {
			s := &ServiceMirror{}
			s.Icon = conf.Icon
			s.Type = "composer"
			s.Name = code
			s.SourceUrl = conf.Server
			s.TargetUrl = fmt.Sprintf("%s/composer/%s", config.PublicServer, code)
			s.Enabled = conf.Enabled

			d = append(d, s)
		}

		pkgmirror.Serialize(w, d)
	}
}
