package main

import (
	"fmt"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/rande/goapp"
	"github.com/rande/pkgmirror"
	"goji.io"
	"goji.io/pat"
	"golang.org/x/net/context"

	"net/http/pprof"
)

func main() {

	app := goapp.NewApp()

	logger := log.New()
	//logger.Level = log.DebugLevel

	logger.Info("Init GoApp")

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

		app.Set("mirror.packagist", func(app *goapp.App) interface{} {
			s := pkgmirror.NewPackagistService()
			s.GitConfig = app.Get("mirror.git").(*pkgmirror.GitService).Config
			s.Logger = logger.WithFields(log.Fields{
				"handler": "packagist",
				"server":  s.Config.Server,
			})
			s.Init(app)

			return s
		})

		app.Set("mirror.git", func(app *goapp.App) interface{} {
			s := pkgmirror.NewGitService()
			s.Logger = logger.WithFields(log.Fields{
				"handler": "git",
			})
			s.Init(app)

			return s
		})

		return nil
	})

	l.Run(func(app *goapp.App, state *goapp.GoroutineState) error {
		logger.Info("Start Packagist Sync")

		s := app.Get("mirror.packagist").(pkgmirror.MirrorService)
		s.Serve(state)

		return nil
	})

	l.Run(func(app *goapp.App, state *goapp.GoroutineState) error {
		logger.Info("Start Git Sync")

		s := app.Get("mirror.git").(pkgmirror.MirrorService)
		s.Serve(state)

		return nil
	})

	l.Run(func(app *goapp.App, state *goapp.GoroutineState) error {
		logger.Info("Start Goji.io")

		mux := app.Get("mux").(*goji.Mux)
		packagist := app.Get("mirror.packagist").(*pkgmirror.PackagistService)
		git := app.Get("mirror.git").(*pkgmirror.GitService)

		mux.HandleFuncC(pat.Get("/packagist/packages.json"), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if data, err := packagist.Get("packages.json"); err != nil {
				pkgmirror.SendWithHttpCode(w, 500, err.Error())
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(data)
			}
		})

		mux.HandleFuncC(pat.Get("/packagist/p/:ref.json"), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if data, err := packagist.Get(fmt.Sprintf("p/%s.json", pat.Param(ctx, "ref"))); err != nil {
				pkgmirror.SendWithHttpCode(w, 500, err.Error())
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(data)
			}
		})

		mux.HandleFuncC(pat.Get("/packagist/p/:vendor/:package.json"), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			pkg := fmt.Sprintf("%s/%s", pat.Param(ctx, "vendor"), pat.Param(ctx, "package"))

			if refresh := r.FormValue("refresh"); len(refresh) > 0 {
				w.Header().Set("Content-Type", "application/json")

				if err := packagist.UpdatePackage(pkg); err != nil {
					pkgmirror.SendWithHttpCode(w, 500, err.Error())
				} else {
					pkgmirror.SendWithHttpCode(w, 200, "Package updated")
				}

				return
			}

			if data, err := packagist.Get(fmt.Sprintf("%s", pkg)); err != nil {
				pkgmirror.SendWithHttpCode(w, 404, err.Error())
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Content-Encoding", "gzip")
				w.Write(data)
			}
		})

		mux.HandleFuncC(pat.Get("/git/:hostname/:vendor/:package/:ref.zip"), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			hostname := pat.Param(ctx, "hostname")
			pkg := pat.Param(ctx, "package")
			vendor := pat.Param(ctx, "vendor")
			ref := pat.Param(ctx, "ref")

			logger.WithFields(log.Fields{
				"hostname": hostname,
				"package":  pkg,
				"vendor":   vendor,
				"ref":      ref,
			}).Info("Zip archive")

			path := fmt.Sprintf("%s/%s/%s.git", hostname, vendor, pkg)

			w.Header().Set("Content-Type", "application/zip")
			git.WriteArchive(w, path, ref)
		})

		mux.HandleFuncC(pat.Get("/git/*"), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			logger.WithFields(log.Fields{
				"path": r.URL.Path[5:],
			}).Info("Git fetch")

			if err := git.WriteFile(w, r.URL.Path[5:]); err != nil {
				w.WriteHeader(http.StatusNotFound)
			}
		})

		//mux.HandleFunc(pat.Get("/debug/pprof/cmdline"), http.HandlerFunc(pprof.Cmdline))
		//mux.HandleFunc(pat.Get("/debug/pprof/profile"), http.HandlerFunc(pprof.Profile))
		//mux.HandleFunc(pat.Get("/debug/pprof/symbol"), http.HandlerFunc(pprof.Symbol))
		mux.HandleFunc(pat.Get("/debug/pprof/*"), http.HandlerFunc(pprof.Index))

		http.ListenAndServe("localhost:8000", mux)

		return nil
	})

	logger.Info("Start GoApp lifecycle")

	os.Exit(l.Go(app))
}
