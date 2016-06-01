// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package npm

import (
	"net/http"
	"regexp"

	"goji.io"
	"goji.io/pattern"
	"golang.org/x/net/context"
)

var (
	PAT_ARCHIVE_URL = regexp.MustCompile(`\/npm\/([\w\d\.-]+)\/-\/([\w\d\.-]+)-((0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:(?:\d*[A-Za-z-][0-9A-Za-z-]*|(?:0|[1-9]\d*))\.)*(?:\d*[A-Za-z-][0-9A-Za-z-]*|(?:0|[1-9]\d*))))?(?:\+((?:(?:[0-9A-Za-z-]+)\.)*[0-9A-Za-z-]+))?)\.(tgz)`)
)

func NewArchivePat() goji.Pattern {
	return &PackagePat{}
}

type PackagePat struct {
}

func (pp *PackagePat) Match(ctx context.Context, r *http.Request) context.Context {
	if results := PAT_ARCHIVE_URL.FindStringSubmatch(r.URL.Path); len(results) == 0 {
		return nil
	} else {
		return &packagePatMatch{ctx, results[1], results[3], "tgz"}
	}
}

type packagePatMatch struct {
	context.Context
	Package string
	Version string
	Format  string
}

func (m packagePatMatch) Value(key interface{}) interface{} {

	switch key {
	case pattern.AllVariables:
		return map[pattern.Variable]string{
			"package": m.Package,
			"version": m.Version,
			"format":  m.Format,
		}
	case pattern.Variable("version"):
		return m.Version
	case pattern.Variable("package"):
		return m.Package
	case pattern.Variable("format"):
		return m.Format
	}

	return m.Context.Value(key)
}
