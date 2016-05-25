package pkgmirror

import (
	"encoding/json"
	"fmt"
	"time"
)

type PackagesResult struct {
	Packages         json.RawMessage `json:"packages"`
	Notify           string          `json:"notify"`
	NotifyBatch      string          `json:"notify-batch"`
	ProvidersURL     string          `json:"providers-url"`
	Search           string          `json:"search"`
	ProviderIncludes map[string]struct {
		Sha256 string `json:"sha256"`
	} `json:"provider-includes"`
}

type ProvidersResult struct {
	Providers map[string]struct {
		Sha256 string `json:"sha256"`
	} `json:"providers"`
	Code string `json:"-"`
}

// package description
type Package struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	Keywords          []string `json:"keywords"`
	Homepage          string   `json:"homepage"`
	Version           string   `json:"version"`
	VersionNormalized string   `json:"version_normalized"`
	License           []string `json:"license"`
	Authors           []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"authors"`
	Source struct {
		Type      string `json:"type"`
		URL       string `json:"url"`
		Reference string `json:"reference"`
	} `json:"source"`
	Dist struct {
		Type      string `json:"type"`
		URL       string `json:"url"`
		Reference string `json:"reference"`
		Shasum    string `json:"shasum"`
	} `json:"dist"`
	Type          string            `json:"type"`
	Time          time.Time         `json:"time"`
	Autoload      *json.RawMessage  `json:"autoload"`
	Require       map[string]string `json:"require"`
	RequireDevmap map[string]string `json:"require-dev"`
	UID           int               `json:"uid"`
}

// used to load the packages.json file
type PackageResult struct {
	Packages map[string]map[string]*Package `json:"packages"`
}

type PackageInformation struct {
	PackageResult PackageResult `json:"-"`
	Package       string        `json:"package"`
	Exist         bool          `json:"-"`
	HashSource    string        `json:"hash_source"`
	HashTarget    string        `json:"hash_target"`
}

func (pi *PackageInformation) GetSourceKey() string {
	return fmt.Sprintf("p/%s$%s.json", pi.Package, pi.HashSource)
}

func (pi *PackageInformation) GetTargetKey() string {
	return fmt.Sprintf("p/%s$%s.json", pi.Package, pi.HashTarget)
}
