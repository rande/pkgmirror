// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package npm

import (
	"testing"

	"fmt"

	"github.com/rande/pkgmirror"
	"github.com/stretchr/testify/assert"
)

func Test_Load_Package(t *testing.T) {
	p := &FullPackageDefinition{}

	files := []struct{ Name, File string }{
		{"knwl.js", "knwl.js.json"},
		{"math_example_bulbignz", "math_example_bulbignz.json"},
		{"gulp-app-manager", "gulp-app-manager.json"},
		{"jsontocsv", "jsontocsv.json"},
	}

	for _, f := range files {
		assert.NoError(t, pkgmirror.LoadStruct(fmt.Sprintf("../../fixtures/npm/%s", f.File), p), fmt.Sprintf("Package %s", f.File))

		assert.Equal(t, f.Name, p.Name, fmt.Sprintf("Package %s", f.File))
	}
}
