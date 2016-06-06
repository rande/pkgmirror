// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package composer

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

func Test_Composer_Pat_Definition(t *testing.T) {
	p := NewPackagePat("packagist")

	c, r := mustReq("GET", "/composer/packagist/p/kevinlebrun/colors.php%24f8ef02dddbd0bb7f78a2775e7188415e128d7b147f2a5630784c75cfc46a1a7e.json")

	result := p.Match(c, r)

	assert.NotNil(t, result)
	assert.Equal(t, "kevinlebrun", result.Value(pattern.Variable("vendor")))
	assert.Equal(t, "colors.php", result.Value(pattern.Variable("package")))
	assert.Equal(t, "f8ef02dddbd0bb7f78a2775e7188415e128d7b147f2a5630784c75cfc46a1a7e", result.Value(pattern.Variable("ref")))
	assert.Equal(t, "json", result.Value(pattern.Variable("format")))
}

func Test_Composer_Pat_PackageInformation(t *testing.T) {
	p := NewPackageInfoPat("packagist")

	c, r := mustReq("GET", "/composer/packagist/p/kevinlebrun/colors.php")

	result := p.Match(c, r)

	assert.NotNil(t, result)
	assert.Equal(t, "kevinlebrun", result.Value(pattern.Variable("vendor")))
	assert.Equal(t, "colors.php", result.Value(pattern.Variable("package")))
	assert.Equal(t, "html", result.Value(pattern.Variable("format")))
}
