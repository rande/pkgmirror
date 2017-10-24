// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
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
			s.Id = fmt.Sprintf("pkgmirror.git.%s", code)
			s.Icon = conf.Icon
			s.Type = "git"
			s.Name = code
			s.SourceUrl = conf.Server
			s.TargetUrl = fmt.Sprintf("%s/git/%s", config.PublicServer, conf.Server)
			s.Enabled = conf.Enabled
			s.Usage = fmt.Sprintf(`
You can also download a zip file with the following url:

    %s/path/repository/REFENCE.zip

The reference can be anything: a branch, a tag or a commit. Please note, tag and commit are
stored on dedicated cache location.

You can clone repository with the following command:

    git clone %s/path/repository.git

`, s.TargetUrl, s.TargetUrl)

			d = append(d, s)
		}

		for code, conf := range config.Npm {
			s := &ServiceMirror{}
			s.Id = fmt.Sprintf("pkgmirror.npm.%s", code)
			s.Icon = conf.Icon
			s.Type = "npm"
			s.Name = code
			s.SourceUrl = conf.Server
			s.TargetUrl = fmt.Sprintf("%s/npm/%s", config.PublicServer, code)
			s.Enabled = conf.Enabled
			s.Usage = fmt.Sprintf(`
You need to set the registry to:

    npm set registry %s

That's it! Now any packages will be retrieve from the mirror. Only downloaded archive files will
be stored on a dedicated cache location.

Please note, the configuration is global to all projects running in the current environment.

`, s.TargetUrl)
			d = append(d, s)
		}

		for code, conf := range config.Composer {
			s := &ServiceMirror{}
			s.Id = fmt.Sprintf("pkgmirror.composer.%s", code)
			s.Icon = conf.Icon
			s.Type = "composer"
			s.Name = code
			s.SourceUrl = conf.Server
			s.TargetUrl = fmt.Sprintf("%s/composer/%s", config.PublicServer, code)
			s.Enabled = conf.Enabled
			s.Usage = fmt.Sprintf(`
You need to declare the mirror in your composer.json file:

    "repositories":[
        { "packagist": false },
        { "type": "composer", "url": "%s"}
    ],

That's it!

Please note, the composer mirror alter github path to point to the local git mirror. Make sure
the github mirror is properly configured.
`, s.TargetUrl)

			d = append(d, s)
		}

		for code, conf := range config.Bower {
			s := &ServiceMirror{}
			s.Id = fmt.Sprintf("pkgmirror.bower.%s", code)
			s.Icon = conf.Icon
			s.Type = "bower"
			s.Name = code
			s.SourceUrl = conf.Server
			s.TargetUrl = fmt.Sprintf("%s/bower/%s", config.PublicServer, code)
			s.Enabled = conf.Enabled
			s.Usage = fmt.Sprintf(`
You need to declare the mirror in your .bowerrc file:

    {
        "registry": {
            "search": ["%s"],
            "register": "%s"
        }
    }

That's it!

Please note, the bower mirror alter github path to point to the local git mirror. Make sure
the github mirror is properly configured.
`, s.TargetUrl, s.TargetUrl)

			d = append(d, s)
		}

		for code, conf := range config.Static {
			s := &ServiceMirror{}
			s.Id = fmt.Sprintf("pkgmirror.static.%s", code)
			s.Icon = conf.Icon
			s.Type = "static"
			s.Name = code
			s.SourceUrl = conf.Server
			s.TargetUrl = fmt.Sprintf("%s/static/%s", config.PublicServer, code)
			s.Enabled = conf.Enabled
			s.Usage = fmt.Sprintf(`
You just need to reference the server as %s/myfile.zip, the static handle will retrieve the file
from %s/myfile.zip and store a copy on the mirror server.
`, s.TargetUrl, s.SourceUrl)

			d = append(d, s)
		}

		pkgmirror.Serialize(w, d)
	}
}

func Api_GET_Sse(app *goapp.App) func(w http.ResponseWriter, r *http.Request) {
	brk := app.Get("pkgmirror.sse.broker").(*pkgmirror.SseBroker)

	return brk.Handler
}

func Api_GET_Ping(app *goapp.App) func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	}
}
