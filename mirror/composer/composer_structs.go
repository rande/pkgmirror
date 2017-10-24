// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package composer

import (
	"encoding/json"
	"fmt"
	"time"
)

type ProviderInclude map[string]struct {
	Sha256 string `json:"sha256"`
}

type PackagesResult struct {
	Packages         json.RawMessage `json:"packages"`
	Notify           string          `json:"notify"`
	NotifyBatch      string          `json:"notify-batch"`
	ProvidersURL     string          `json:"providers-url"`
	Search           string          `json:"search"`
	ProviderIncludes ProviderInclude `json:"provider-includes"`
}

type ProvidersResult struct {
	Providers map[string]struct {
		Sha256 string `json:"sha256"`
	} `json:"providers"`
	Code string `json:"-"`
}

// package description
type Package struct {
	Name              string           `json:"name,omitempty"`
	Abandoned         *json.RawMessage `json:"abandoned,omitempty"`
	Description       string           `json:"description,omitempty"`
	Keywords          []string         `json:"keywords,omitempty"`
	Homepage          string           `json:"homepage,omitempty"`
	Version           string           `json:"version,omitempty"`
	VersionNormalized string           `json:"version_normalized,omitempty"`
	License           []string         `json:"license,omitempty"`
	Bin               []string         `json:"bin,omitempty"`
	Authors           []struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Homepage string `json:"homepage"`
		Role     string `json:"role"`
	} `json:"authors,omitempty"`
	Source struct {
		Type      string `json:"type"`
		URL       string `json:"url"`
		Reference string `json:"reference"`
	} `json:"source,omitempty"`
	Dist struct {
		Type      string `json:"type"`
		URL       string `json:"url"`
		Reference string `json:"reference"`
		Shasum    string `json:"shasum"`
	} `json:"dist,omitempty"`
	Extra      *json.RawMessage `json:"extra,omitempty"`
	TargetDir  string           `json:"target-dir,omitempty"`
	Type       string           `json:"type,omitempty"`
	Time       time.Time        `json:"time,omitempty"`
	Autoload   *json.RawMessage `json:"autoload,omitempty"`
	Replace    *json.RawMessage `json:"replace,omitempty"`
	Conflict   *json.RawMessage `json:"conflict,omitempty"`
	Provide    *json.RawMessage `json:"provide,omitempty"`
	Require    *json.RawMessage `json:"require,omitempty"`
	RequireDev *json.RawMessage `json:"require-dev,omitempty"`
	Suggest    *json.RawMessage `json:"suggest,omitempty"`
	UID        int              `json:"uid,omitempty"`
}

// used to load the packages.json file
type PackageResult struct {
	Packages map[string]map[string]*Package `json:"packages"`
}

type PackageInformation struct {
	Server        string        `json:"server"`
	PackageResult PackageResult `json:"-"`
	Package       string        `json:"package"`
	Exist         bool          `json:"-"`
	HashSource    string        `json:"hash_source"`
	HashTarget    string        `json:"hash_target"`
}

func (pi *PackageInformation) GetSourceKey() string {
	return fmt.Sprintf("%s$%s.json", pi.Package, pi.HashSource)
}

func (pi *PackageInformation) GetTargetKey() string {
	return fmt.Sprintf("%s$%s", pi.Package, pi.HashTarget)
}
