// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package composer

import (
	"testing"

	"github.com/rande/pkgmirror"
	"github.com/stretchr/testify/assert"
)

func LoadTestStruct(t *testing.T, file string, v interface{}) {
	err := pkgmirror.LoadStruct(file, v)

	assert.NoError(t, err)
}

func Test_Load_Packages(t *testing.T) {
	p := &PackagesResult{}

	LoadTestStruct(t, "../../fixtures/packagist/packages.json", p)

	assert.Equal(t, "/downloads/", p.NotifyBatch)

	assert.Equal(t, 9, len(p.ProviderIncludes))
}

func Test_Load_Providers(t *testing.T) {
	p := &ProvidersResult{}

	LoadTestStruct(t, "../../fixtures/packagist/p/provider-2013$370af0b17d1ec5b0325bdb0126c9007b69647fafe5df8b5ecf79241e09745841.json", p)

	assert.Equal(t, 7585, len(p.Providers))
}

func Test_Load_Package(t *testing.T) {
	p := &PackageResult{}

	LoadTestStruct(t, "../../fixtures/packagist/p/0n3s3c/baselibrary$3a3dbbc33805b6748f859e8f2c517355f42e2f6d4b71daad077794842dca280c.json", p)

	assert.Equal(t, 1, len(p.Packages))
	assert.Equal(t, 2, len(p.Packages["0n3s3c/baselibrary"]))
	assert.Equal(t, "Library for working with objects in PHP", p.Packages["0n3s3c/baselibrary"]["0.5.0"].Description)
}

func Test_Load_Package_Project(t *testing.T) {
	p := &PackageResult{}

	LoadTestStruct(t, "../../fixtures/packagist/p/symfony/framework-standard-edition$cb64bc5278d2b6bbf7c02ae4b995f3698df1a210dceb509328b4370e13f3ba33.json", p)

	assert.Equal(t, 1, len(p.Packages))
	assert.Equal(t, 158, len(p.Packages["symfony/framework-standard-edition"]))
	assert.Equal(t, "The \"Symfony Standard Edition\" distribution", p.Packages["symfony/framework-standard-edition"]["2.8.x-dev"].Description)
}
