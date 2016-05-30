// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package composer

import (
	"net/http"
	"regexp"

	"goji.io"
	"goji.io/pattern"
	"golang.org/x/net/context"
)

var (
	PACKAGE_URL = regexp.MustCompile(`\/packagist\/p\/([^\/]*)\/([^\/]*)\$([^\/]*)\.json`)
)

func NewPackagePat() goji.Pattern {
	return &PackagePat{}
}

type PackagePat struct {
}

func (pp *PackagePat) Match(ctx context.Context, r *http.Request) context.Context {
	if results := PACKAGE_URL.FindStringSubmatch(r.URL.Path); len(results) == 0 {
		return nil
	} else {
		return &packagePatMatch{ctx, results[1], results[2], results[3], "json"}
	}
}

type packagePatMatch struct {
	context.Context
	Vendor  string
	Package string
	Ref     string
	Format  string
}

func (m packagePatMatch) Value(key interface{}) interface{} {

	switch key {
	case pattern.AllVariables:
		return map[pattern.Variable]string{
			"vendor":  m.Vendor,
			"package": m.Package,
			"ref":     m.Ref,
			"format":  m.Format,
		}
	case pattern.Variable("vendor"):
		return m.Vendor
	case pattern.Variable("package"):
		return m.Package
	case pattern.Variable("ref"):
		return m.Ref
	case pattern.Variable("format"):
		return m.Format
	}

	return m.Context.Value(key)
}
