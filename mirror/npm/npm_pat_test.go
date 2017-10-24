// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package npm

import (
	"testing"

	"net/http"

	"github.com/stretchr/testify/assert"
	"goji.io/pattern"
	"golang.org/x/net/context"
)

type TestVersion struct {
	Url     string
	Package string
	Version string
}

func mustReq(method, path string) (context.Context, *http.Request) {
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		panic(err)
	}
	ctx := pattern.SetPath(context.Background(), req.URL.EscapedPath())

	return ctx, req
}

func Test_Npm_Pat(t *testing.T) {

	cases := []struct{ Url, Package, Version string }{
		{"/npm/npm/aspace/-/aspace-0.0.1.tgz", "aspace", "0.0.1"},
		{"/npm/npm/@type%2fnode/-/node-6.0.90.tgz", "@type%2fnode", "6.0.90"},
		{"/npm/npm/dateformat/-/dateformat-1.0.2-1.2.3.tgz", "dateformat", "1.0.2-1.2.3"},
	}

	matcher := NewArchivePat("npm")

	for _, p := range cases {
		c, r := mustReq("GET", p.Url)

		result := matcher.Match(c, r)

		assert.NotNil(t, result)
		assert.Equal(t, p.Package, result.Value(pattern.Variable("package")))
		assert.Equal(t, p.Version, result.Value(pattern.Variable("version")))
		assert.Equal(t, "tgz", result.Value(pattern.Variable("format")))
	}
}

func Test_Npm_Pat_AllVariables(t *testing.T) {
	p := NewArchivePat("npm")

	c, r := mustReq("GET", "/npm/npm/aspace/-/aspace-0.0.1.tgz")

	result := p.Match(c, r)

	assert.NotNil(t, result)

	vars := result.Value(pattern.AllVariables).(map[pattern.Variable]string)

	assert.Equal(t, "aspace", vars["package"])
	assert.Equal(t, "0.0.1", vars["version"])
	assert.Equal(t, "tgz", vars["format"])
}

func Test_Npm_Pat_OtherVariable(t *testing.T) {
	p := NewArchivePat("npm")

	c, r := mustReq("GET", "/npm/npm/aspace/-/aspace-0.0.1.tgz")

	result := p.Match(c, r)

	assert.NotNil(t, result)

	assert.Nil(t, result.Value(pattern.Variable("foo")))
}
