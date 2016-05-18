package pkgmirror

import (
	"time"
	"encoding/json"
)

type PackagesResult struct {
	Packages         []interface{} `json:"packages"`
	Notify           string `json:"notify"`
	NotifyBatch      string `json:"notify-batch"`
	ProvidersURL     string `json:"providers-url"`
	Search           string `json:"search"`
	ProviderIncludes map[string]struct {
		Sha256 string `json:"sha256"`
	} `json:"provider-includes"`
}

type ProvidersResult struct {
	Providers map[string]struct {
		Sha256 string `json:"sha256"`
	} `json:"providers"`
}

type Package struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	Keywords          []string `json:"keywords"`
	Homepage          string `json:"homepage"`
	Version           string `json:"version"`
	VersionNormalized string `json:"version_normalized"`
	License           []string `json:"license"`
	Authors           []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"authors"`
	Source            struct {
						  Type      string `json:"type"`
						  URL       string `json:"url"`
						  Reference string `json:"reference"`
					  } `json:"source"`
	Dist              struct {
						  Type      string `json:"type"`
						  URL       string `json:"url"`
						  Reference string `json:"reference"`
						  Shasum    string `json:"shasum"`
					  } `json:"dist"`
	Type              string `json:"type"`
	Time              time.Time `json:"time"`
	Autoload          json.RawMessage `json:"autoload"`
	Require           map[string]string `json:"require"`
	RequireDevmap     map[string]string `json:"require-dev"`
	UID               int `json:"uid"`
}

type PackageResult struct {
	Packages map[string]map[string]Package `json:"packages"`
}