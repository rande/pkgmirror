// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package api

type ServiceMirror struct {
	Type      string
	Name      string
	SourceUrl string
	TargetUrl string
	Icon      string
	Enabled bool
}
