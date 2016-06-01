// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pkgmirror

import (
	"errors"
)

var (
	SyncInProgressError   = errors.New("A synchronization is already running")
	EmptyKeyError         = errors.New("No value available")
	ResourceNotFoundError = errors.New("Resource not found")
)
