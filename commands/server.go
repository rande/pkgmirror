// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package commands

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/mitchellh/cli"
	"github.com/rande/goapp"
	"github.com/rande/pkgmirror"
	"github.com/rande/pkgmirror/api"
	"github.com/rande/pkgmirror/assets"
	"github.com/rande/pkgmirror/mirror/bower"
	"github.com/rande/pkgmirror/mirror/composer"
	"github.com/rande/pkgmirror/mirror/git"
	"github.com/rande/pkgmirror/mirror/npm"
	"github.com/rande/pkgmirror/mirror/static"
	"goji.io"
	"goji.io/pat"
)

type ServerCommand struct {
	Ui       cli.Ui
	Verbose  bool
	Commands map[string]cli.CommandFactory
	ConfFile string
	LogLevel string
}

func (c *ServerCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("run", flag.ContinueOnError)
	cmdFlags.Usage = func() {
		c.Ui.Output(c.Help())
	}

	cmdFlags.BoolVar(&c.Verbose, "verbose", false, "")
	cmdFlags.StringVar(&c.LogLevel, "log-level", "warning", "The log level")
	cmdFlags.StringVar(&c.ConfFile, "file", "/etc/pkgmirror.toml", "The configuration file")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	config := &pkgmirror.Config{
		CacheDir: fmt.Sprintf("%s/pkgmirror", os.TempDir()),
		LogLevel: "info",
	}

	if _, err := toml.DecodeFile(c.ConfFile, config); err != nil {
		c.Ui.Error(fmt.Sprintf("Unable to parse configuration file: %s", c.ConfFile))

		return 1
	}

	c.Ui.Info("Configure app")

	l := goapp.NewLifecycle()

	app, err := pkgmirror.GetApp(config, l)

	if err != nil {
		c.Ui.Error(err.Error())

		return 1
	}

	composer.ConfigureApp(config, l)
	git.ConfigureApp(config, l)
	npm.ConfigureApp(config, l)
	bower.ConfigureApp(config, l)
	static.ConfigureApp(config, l)
	api.ConfigureApp(config, l)
	assets.ConfigureApp(config, l)

	l.Run(func(app *goapp.App, state *goapp.GoroutineState) error {
		c.Ui.Info(fmt.Sprintf("Start HTTP Server (bind: %s)", config.InternalServer))

		mux := app.Get("mux").(*goji.Mux)

		//mux.HandleFunc(pat.Get("/debug/pprof/cmdline"), http.HandlerFunc(pprof.Cmdline))
		//mux.HandleFunc(pat.Get("/debug/pprof/profile"), http.HandlerFunc(pprof.Profile))
		//mux.HandleFunc(pat.Get("/debug/pprof/symbol"), http.HandlerFunc(pprof.Symbol))

		mux.HandleFunc(pat.Get("/debug/pprof/*"), http.HandlerFunc(pprof.Index))

		http.ListenAndServe(config.InternalServer, mux)

		return nil
	})

	c.Ui.Info("Start app lifecycle")

	return l.Go(app)
}

func (c *ServerCommand) Synopsis() string {
	return "Run the mirroring server."
}

func (c *ServerCommand) Help() string {
	return strings.TrimSpace(`
Usage: pkgmirror run [options]

  Run the mirror server

Options:
  -file               The configuration file (default: /etc/pkgmirror.toml)
  -verbose            Add verbose information to the output
  -log-level          Log level (defaul: warning)
                      possible values: debug, info, warning, error, fatal and panic
`)
}
