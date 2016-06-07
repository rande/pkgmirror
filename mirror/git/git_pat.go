// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"fmt"
	"net/http"
	"regexp"

	"goji.io"
	"goji.io/pattern"
	"golang.org/x/net/context"
)

func NewGitPat(hostname string) goji.Pattern {
	return &GitPat{
		Hostname: hostname,
		Pattern:  regexp.MustCompile(fmt.Sprintf(`\/git\/%s\/(.*)\/([\w\d]{40}|(.*))\.zip`, hostname)),
	}
}

type GitPat struct {
	Hostname string
	Pattern  *regexp.Regexp
}

func (pp *GitPat) Match(ctx context.Context, r *http.Request) context.Context {
	if results := pp.Pattern.FindStringSubmatch(r.URL.Path); len(results) == 0 {
		return nil
	} else {
		return &gitPatMatch{ctx, pp.Hostname, results[1], results[2], "zip"}
	}
}

type gitPatMatch struct {
	context.Context
	Hostname string
	Path     string
	Ref      string
	Format   string
}

func (m gitPatMatch) Value(key interface{}) interface{} {

	switch key {
	case pattern.AllVariables:
		return map[pattern.Variable]string{
			"hostname": m.Hostname,
			"path":     m.Path,
			"ref":      m.Ref,
			"format":   m.Format,
		}
	case pattern.Variable("hostname"):
		return m.Hostname
	case pattern.Variable("path"):
		return m.Path
	case pattern.Variable("ref"):
		return m.Ref
	case pattern.Variable("format"):
		return m.Format
	}

	return m.Context.Value(key)
}
