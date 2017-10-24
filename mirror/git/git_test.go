// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Expectation struct {
	Expected string
	Value    string
}

func Test_Archive_Rewrite_Github(t *testing.T) {
	publicServer := "https://mirrors.localhost"

	values := []*Expectation{
		{"https://mirrors.localhost/git/github.com/sonata-project/exporter/b9098b5007c525a238ddf44d578b8efae7bccc72.zip", "https://api.github.com/repos/sonata-project/exporter/zipball/b9098b5007c525a238ddf44d578b8efae7bccc72"},
		{"https://mirrors.localhost/git/github.com/kevinlebrun/colors.php/6d7140aeedef46c97c2324f09b752c599ef17dac.zip", "https://api.github.com/repos/kevinlebrun/colors.php/zipball/6d7140aeedef46c97c2324f09b752c599ef17dac"},
	}

	for _, v := range values {
		assert.Equal(t, v.Expected, GitRewriteArchive(publicServer, v.Value))
	}

}

func Test_Archive_Rewrite_Bitbucket(t *testing.T) {
	publicServer := "https://mirrors.localhost"

	path := GitRewriteArchive(publicServer, "https://bitbucket.org/sonata-project/exporter/get/b9098b5007c525a238ddf44d578b8efae7bccc72.zip")
	assert.Equal(t, "https://mirrors.localhost/git/bitbucket.org/sonata-project/exporter/b9098b5007c525a238ddf44d578b8efae7bccc72.zip", path)
}

func Test_Archive_Rewrite_Gitlab(t *testing.T) {
	publicServer := "https://mirrors.localhost"

	path := GitRewriteArchive(publicServer, "https://gitlab.example.com/sonata-project/exporter/repository/archive.zip?ref=b9098b5007c525a238ddf44d578b8efae7bccc72")
	assert.Equal(t, "https://mirrors.localhost/git/gitlab.example.com/sonata-project/exporter/b9098b5007c525a238ddf44d578b8efae7bccc72.zip", path)
}

func Test_Repository_Rewrite_Git(t *testing.T) {
	publicServer := "https://mirrors.localhost"

	values := []*Expectation{
		{"https://mirrors.localhost/git/github.com/DavidForest/ImgBundle.git", "git@github.com:DavidForest/ImgBundle.git"},
		{"https://mirrors.localhost/git/github.com/sonata-project/exporter.git", "https://github.com/sonata-project/exporter.git"},
		{"https://mirrors.localhost/git/bitbucket.org/foo/bar.git", "https://bitbucket.org/foo/bar"},
		{"https://mirrors.localhost/git/github.com/xstudios/flames.git", "git://github.com/xstudios/flames.git"},
		{"https://mirrors.localhost/git/git.kootstradevelopment.nl/r.kootstra/stackinstance-bundles-mailer-bundle.git", "http://git.kootstradevelopment.nl/r.kootstra/stackinstance-bundles-mailer-bundle.git"},
		{"https://mirrors.localhost/git/github.com/zyncro/bower-videogular-themes-default.git", "https://github.com/zyncro/bower-videogular-themes-default.git"},
	}

	for _, v := range values {
		assert.Equal(t, v.Expected, GitRewriteRepository(publicServer, v.Value))
	}
}

func Test_Repository_Rewrite_SVN(t *testing.T) {
	publicServer := "https://mirrors.localhost"

	values := []*Expectation{
		{"https://m10s.svn.beanstalkapp.com/m10s-common", "https://m10s.svn.beanstalkapp.com/m10s-common"},
		{"svn://localhost/path/to/project", "svn://localhost/path/to/project"},
	}

	for _, v := range values {
		assert.Equal(t, v.Expected, GitRewriteRepository(publicServer, v.Value))
	}
}
