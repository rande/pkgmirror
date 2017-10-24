// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pkgmirror

type ComposerConfig struct {
	Server  string
	Enabled bool
	Icon    string
}

type BowerConfig struct {
	Server  string
	Enabled bool
	Icon    string
}

type NpmConfig struct {
	Server    string
	Enabled   bool
	Icon      string
	Fallbacks []*struct {
		Server string
	}
}

type GitConfig struct {
	Server  string
	Enabled bool
	Icon    string
	Clone   string
}

type StaticConfig struct {
	Server  string
	Enabled bool
	Icon    string
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
	Bower          map[string]*BowerConfig
	Static         map[string]*StaticConfig
}
