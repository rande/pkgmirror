// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package static

import (
	"net/http"
	"time"
)

type StaticFile struct {
	Header     http.Header
	Url        string
	DownloadAt time.Time
	Size       int64
}
