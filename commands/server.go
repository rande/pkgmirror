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
	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	"github.com/rande/goapp"
	"github.com/rande/gonode/core/vault"
	"github.com/rande/pkgmirror"
	"github.com/rande/pkgmirror/mirror/composer"
	"github.com/rande/pkgmirror/mirror/git"
	"goji.io"
	"goji.io/pat"
	"golang.org/x/net/context"
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
	}

	if _, err := toml.DecodeFile(c.ConfFile, config); err != nil {
		c.Ui.Error(fmt.Sprintf("Unable to parse configuration file: %s", c.ConfFile))

		return 1
	}

	app := goapp.NewApp()

	logger := log.New()

	if c.Verbose {
		logger.Level = log.DebugLevel
	}

	if !c.Verbose {
		if level, err := log.ParseLevel(c.LogLevel); err != nil {
			c.Ui.Error(fmt.Sprintf("Unable to parse the log level: %s", c.LogLevel))

			return 1
		} else {
			logger.Level = level
		}
	}

	if len(config.DataDir) == 0 {
		c.Ui.Error("Please configure DataDir")

		return 1
	}

	c.Ui.Info("Configure app")

	l := goapp.NewLifecycle()
	l.Register(func(app *goapp.App) error {
		app.Set("mux", func(app *goapp.App) interface{} {
			m := goji.NewMux()

			m.UseC(func(h goji.Handler) goji.Handler {
				return goji.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
					logger.WithFields(log.Fields{
						"request_method":      r.Method,
						"request_url":         r.URL.String(),
						"request_remote_addr": r.RemoteAddr,
						"request_host":        r.Host,
					}).Info("Receive HTTP request")

					//t1 := time.Now()
					h.ServeHTTPC(ctx, w, r)
					//t2 := time.Now()
				})
			})

			return m
		})

		app.Set("logger", func(app *goapp.App) interface{} {
			return logger
		})

		app.Set("mirror.composer", func(app *goapp.App) interface{} {
			s := composer.NewComposerService()
			s.Config.Path = fmt.Sprintf("%s/composer", config.DataDir)
			s.GitConfig = app.Get("mirror.git").(*git.GitService).Config
			s.Logger = logger.WithFields(log.Fields{
				"handler": "composer",
				"server":  s.Config.SourceServer,
			})
			s.Init(app)

			return s
		})

		app.Set("mirror.git", func(app *goapp.App) interface{} {
			s := git.NewGitService()
			s.Config.Server = config.PublicServer
			s.Config.DataDir = fmt.Sprintf("%s/git", config.DataDir)
			s.Vault = &vault.Vault{
				Algo: "no_op",
				Driver: &vault.DriverFs{
					Root: fmt.Sprintf("%s/git", config.CacheDir),
				},
			}
			s.Logger = logger.WithFields(log.Fields{
				"handler": "git",
			})
			s.Init(app)

			return s
		})

		return nil
	})

	l.Run(func(app *goapp.App, state *goapp.GoroutineState) error {
		c.Ui.Info(fmt.Sprintf("Start Composer Sync (ref: %s/packagist)", config.PublicServer))

		s := app.Get("mirror.composer").(pkgmirror.MirrorService)
		s.Serve(state)

		return nil
	})

	l.Run(func(app *goapp.App, state *goapp.GoroutineState) error {
		c.Ui.Info(fmt.Sprintf("Start Git Sync (server: %s/git)", config.PublicServer))

		s := app.Get("mirror.git").(pkgmirror.MirrorService)
		s.Serve(state)

		return nil
	})

	l.Run(func(app *goapp.App, state *goapp.GoroutineState) error {
		c.Ui.Info(fmt.Sprintf("Start HTTP Server (bind: %s)", config.InternalServer))

		mux := app.Get("mux").(*goji.Mux)
		composerService := app.Get("mirror.composer").(*composer.ComposerService)
		gitService := app.Get("mirror.git").(*git.GitService)

		mux.HandleFuncC(pat.Get("/packagist"), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/packagist/packages.json", http.StatusMovedPermanently)
		})

		mux.HandleFuncC(pat.Get("/packagist/packages.json"), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if data, err := composerService.Get("packages.json"); err != nil {
				pkgmirror.SendWithHttpCode(w, 500, err.Error())
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(data)
			}
		})

		mux.HandleFuncC(pat.Get("/packagist/p/:ref.json"), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if data, err := composerService.Get(fmt.Sprintf("p/%s.json", pat.Param(ctx, "ref"))); err != nil {
				pkgmirror.SendWithHttpCode(w, 404, err.Error())
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(data)
			}
		})

		mux.HandleFuncC(composer.NewPackagePat(), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			pkg := fmt.Sprintf("%s/%s$%s", pat.Param(ctx, "vendor"), pat.Param(ctx, "package"), pat.Param(ctx, "ref"))

			if refresh := r.FormValue("refresh"); len(refresh) > 0 {
				w.Header().Set("Content-Type", "application/json")

				if err := composerService.UpdatePackage(pkg); err != nil {
					pkgmirror.SendWithHttpCode(w, 500, err.Error())
				} else {
					pkgmirror.SendWithHttpCode(w, 200, "Package updated")
				}

				return
			}

			if data, err := composerService.Get(pkg); err != nil {
				pkgmirror.SendWithHttpCode(w, 404, err.Error())
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Content-Encoding", "gzip")
				w.Write(data)
			}
		})

		mux.HandleFuncC(composer.NewPackageInfoPat(), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if pi, err := composerService.GetPackage(fmt.Sprintf("%s/%s", pat.Param(ctx, "vendor"), pat.Param(ctx, "package"))); err != nil {
				pkgmirror.SendWithHttpCode(w, 404, err.Error())
			} else {
				switch pat.Param(ctx, "format") {
				case "html":
					http.Redirect(w, r, fmt.Sprintf("/packagist/p/%s.json", pi.GetTargetKey()), http.StatusFound)
				}
			}
		})

		mux.HandleFuncC(git.NewGitPat(), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			path := pat.Param(ctx, "path")
			ref := pat.Param(ctx, "ref")

			logger.WithFields(log.Fields{
				"path": path,
				"ref":  ref,
			}).Info("Zip archive")

			w.Header().Set("Content-Type", "application/zip")
			if err := gitService.WriteArchive(w, fmt.Sprintf("%s.git", path), ref); err != nil {
				pkgmirror.SendWithHttpCode(w, 500, err.Error())
			}
		})

		mux.HandleFuncC(pat.Get("/git/*"), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			logger.WithFields(log.Fields{
				"path": r.URL.Path[5:],
			}).Info("Git fetch")

			if err := gitService.WriteFile(w, r.URL.Path[5:]); err != nil {
				w.WriteHeader(http.StatusNotFound)
			}
		})

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
