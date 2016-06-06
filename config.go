// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pkgmirror

type ComposerConfig struct {
	Server string
}

type NpmConfig struct {
	Server    string
	Enabled   bool
	Fallbacks []*struct {
		Server string
	}
}

type GitConfig struct {
	Server string
	Clone  string
}

type Config struct {
	DataDir        string
	LogDir         string
	CacheDir       string
	PublicServer   string
	InternalServer string
	LogLevel       string
	Composer       map[string]*ComposerConfig
	Npm            map[string]*NpmConfig
	Git            map[string]*GitConfig
}
