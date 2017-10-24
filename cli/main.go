// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"
	"github.com/rande/pkgmirror/commands"
)

var (
	Version = "0.0.1-Dev"
	RefLog  = "master"
)

func main() {
	ui := &cli.BasicUi{Writer: os.Stdout}

	c := cli.NewCLI("pkgmirror", fmt.Sprintf("%s - %s", Version, RefLog))
	c.Args = os.Args[1:]

	c.Commands = map[string]cli.CommandFactory{
		"run": func() (cli.Command, error) {
			return &commands.ServerCommand{
				Ui: ui,
			}, nil
		},
	}

	exitStatus, _ := c.Run()

	os.Exit(exitStatus)
}
