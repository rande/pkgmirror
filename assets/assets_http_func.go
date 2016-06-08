// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package assets

import (
	"net/http"
	"strings"

	"fmt"

	"github.com/rande/goapp"
	"golang.org/x/net/context"
)

var contentTypes = map[string]string{
	"js":    "application/javascript",
	"css":   "text/css",
	"svg":   "image/svg+xml",
	"eot":   "application/vnd.ms-fontobject",
	"woff":  "application/x-font-woff",
	"woff2": "application/font-woff2",
	"ttf":   "application/x-font-ttf",
	"png":   "image/png",
	"jpg":   "image/jpg",
	"gif":   "image/gif",
}

func Assets_GET_File(app *goapp.App) func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		path := ""
		if r.URL.Path == "/" {
			path = "index.html"
		} else {
			path = r.URL.Path[1:]
		}

		if asset, err := Asset(path); err != nil {
			w.WriteHeader(404)
			w.Write([]byte(fmt.Sprintf("<html><head><title>Page not found</title></head><body><h1>Page not found</h1><div>Page: %s</div></body></html>", path)))
		} else {
			ext := path[(strings.LastIndex(path, ".") + 1):]

			if _, ok := contentTypes[ext]; ok {
				w.Header().Set("Content-Type", contentTypes[ext])
			}

			w.Write(asset)
		}
	}
}
