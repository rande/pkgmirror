// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"net/http"
	"regexp"

	"goji.io"
	"goji.io/pattern"
	"golang.org/x/net/context"
)

var (
	GIT_PATTTERN_URL = regexp.MustCompile(`\/git\/(.*)\/([\w\d]{40}|(.*))\.zip`)
)

func NewGitPat() goji.Pattern {
	return &GitPat{}
}

type GitPat struct {
}

func (pp *GitPat) Match(ctx context.Context, r *http.Request) context.Context {
	if results := GIT_PATTTERN_URL.FindStringSubmatch(r.URL.Path); len(results) == 0 {
		return nil
	} else {
		return &gitPatMatch{ctx, results[1], results[2], "zip"}
	}
}

type gitPatMatch struct {
	context.Context
	Path   string
	Ref    string
	Format string
}

func (m gitPatMatch) Value(key interface{}) interface{} {

	switch key {
	case pattern.AllVariables:
		return map[pattern.Variable]string{
			"path":   m.Path,
			"ref":    m.Ref,
			"format": m.Format,
		}
	case pattern.Variable("path"):
		return m.Path
	case pattern.Variable("ref"):
		return m.Ref
	case pattern.Variable("format"):
		return m.Format
	}

	return m.Context.Value(key)
}
