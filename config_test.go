// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pkgmirror

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

func Test_Config(t *testing.T) {
	c := &Config{}

	confStr := `
DataDir = "/var/lib/pkgmirror"
PublicServer = "https://mirror.example.com"
InternalServer = "localhost:8000"
`

	_, err := toml.Decode(confStr, c)

	assert.NoError(t, err)
	assert.Equal(t, "/var/lib/pkgmirror", c.DataDir)
}
