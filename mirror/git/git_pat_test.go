// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"net/http"
	"testing"

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

func Test_Composer_Pat_Archive(t *testing.T) {
	p := NewGitPat("github.com")

	c, r := mustReq("GET", "/git/github.com/kevinlebrun/colors.php/cb9b6666a2dfd9b6074b4a5caec7902fe3033578.zip")

	result := p.Match(c, r)

	assert.NotNil(t, result)
	assert.Equal(t, "github.com", result.Value(pattern.Variable("hostname")))
	assert.Equal(t, "kevinlebrun/colors.php", result.Value(pattern.Variable("path")))
	assert.Equal(t, "cb9b6666a2dfd9b6074b4a5caec7902fe3033578", result.Value(pattern.Variable("ref")))
	assert.Equal(t, "zip", result.Value(pattern.Variable("format")))
}

func Test_Git_Pat_AllVariables(t *testing.T) {
	p := NewGitPat("github.com")

	c, r := mustReq("GET", "/git/github.com/kevinlebrun/colors.php/cb9b6666a2dfd9b6074b4a5caec7902fe3033578.zip")

	result := p.Match(c, r)

	assert.NotNil(t, result)

	vars := result.Value(pattern.AllVariables).(map[pattern.Variable]string)

	assert.Equal(t, "github.com", vars["hostname"])
	assert.Equal(t, "kevinlebrun/colors.php", vars["path"])
	assert.Equal(t, "cb9b6666a2dfd9b6074b4a5caec7902fe3033578", vars["ref"])
	assert.Equal(t, "zip", vars["format"])
}

func Test_Git_Pat_OtherVariable(t *testing.T) {
	p := NewGitPat("github.com")

	c, r := mustReq("GET", "/git/github.com/kevinlebrun/colors.php/cb9b6666a2dfd9b6074b4a5caec7902fe3033578.zip")

	result := p.Match(c, r)

	assert.NotNil(t, result)

	assert.Nil(t, result.Value(pattern.Variable("foo")))
}
