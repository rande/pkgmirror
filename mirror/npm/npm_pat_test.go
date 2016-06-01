// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
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

func mustReq(method, path string) (context.Context, *http.Request) {
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		panic(err)
	}
	ctx := pattern.SetPath(context.Background(), req.URL.EscapedPath())

	return ctx, req
}

func Test_Npm_Pat_Archive(t *testing.T) {
	p := &PackagePat{}

	c, r := mustReq("GET", "/npm/aspace/-/aspace-0.0.1.tgz")

	result := p.Match(c, r)

	assert.NotNil(t, result)
	assert.Equal(t, "aspace", result.Value(pattern.Variable("package")))
	assert.Equal(t, "0.0.1", result.Value(pattern.Variable("version")))
	assert.Equal(t, "tgz", result.Value(pattern.Variable("format")))
}
