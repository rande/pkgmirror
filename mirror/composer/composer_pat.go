// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package composer

import (
	"fmt"
	"net/http"
	"regexp"

	"goji.io"
	"goji.io/pattern"
	"golang.org/x/net/context"
)

func NewPackagePat(code string) goji.Pattern {
	return &PackagePat{
		Pattern: regexp.MustCompile(fmt.Sprintf(`\/composer\/%s\/p\/([^\/]*)\/([^\/]*)\$([^\/]*)\.json`, code)),
	}
}

type PackagePat struct {
	Pattern *regexp.Regexp
}

func (pp *PackagePat) Match(ctx context.Context, r *http.Request) context.Context {
	if results := pp.Pattern.FindStringSubmatch(r.URL.Path); len(results) == 0 {
		return nil
	} else {
		return &packagePatMatch{ctx, results[1], results[2], results[3], "json"}
	}
}

func NewPackageInfoPat(code string) goji.Pattern {
	return &PackageInfoPat{
		Pattern: regexp.MustCompile(fmt.Sprintf(`\/composer\/%s\/p\/([^\/]*)\/([^\/]*)(.json|)`, code)),
	}
}

type PackageInfoPat struct {
	Pattern *regexp.Regexp
}

func (pp *PackageInfoPat) Match(ctx context.Context, r *http.Request) context.Context {
	if results := pp.Pattern.FindStringSubmatch(r.URL.Path); len(results) == 0 {
		return nil
	} else {
		format := "html"

		if len(results[3]) > 0 {
			format = results[3][1:]
		}

		return &packagePatMatch{ctx, results[1], results[2], "", format}
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
