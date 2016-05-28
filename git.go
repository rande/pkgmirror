// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pkgmirror

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/rande/goapp"
	"github.com/rande/gonode/core/vault"
)

var (
	BITBUCKET_ARCHIVE = regexp.MustCompile(`http(s|):\/\/([\w-\.]+)\/([\w-]+)\/([\w-]+)\/get\/([\w]+)\.zip`)
	GITHUB_ARCHIVE    = regexp.MustCompile(`http(s|):\/\/api\.([\w-\.]+)\/repos\/([\w-]+)\/([\w-]+)\/zipball\/([\w]+)`)
	GITLAB_ARCHIVE    = regexp.MustCompile(`http(s|):\/\/([\w-\.]+)\/([\w-]+)\/([\w-]+)\/repository\/archive.zip\?ref=([\w]+)`)

	GIT_REPOSITORY = regexp.MustCompile(`^(((git|http(s|)):\/\/|git@))([\w-\.]+@|)([\w-\.]+)(\/|:)([\w-\.\/]+?)(\.git|)$`)
	SVN_REPOSITORY = regexp.MustCompile(`(svn:\/\/(.*)|(.*)\.svn\.(.*))`)

	CACHEABLE_REF = regexp.MustCompile(`([\w\d]{40}|[\w\d]+\.[\w\d]+\.[\w\d]+(-[\w\d]+|))`)
)

func NewGitService() *GitService {
	return &GitService{
		Config: &GitConfig{
			Code:    []byte("git"),
			DataDir: "./data/git",
			Binary:  "git",
			Server:  "http://localhost:8000",
		},
		Vault: &vault.Vault{
			Algo: "no_op",
			Driver: &vault.DriverFs{
				Root: "./cache/git",
			},
		},
	}
}

type GitConfig struct {
	Server  string
	Code    []byte
	DataDir string
	Binary  string
}

type GitService struct {
	DB     *bolt.DB
	Config *GitConfig
	Logger *log.Entry
	Vault  *vault.Vault
}

func (gs *GitService) Init(app *goapp.App) error {
	os.MkdirAll(string(filepath.Separator)+gs.Config.DataDir, 0755)

	return nil
}

func (gs *GitService) Serve(state *goapp.GoroutineState) error {
	for {
		gs.Logger.Info("Starting a new sync...")

		gs.SyncServices()

		gs.Logger.Info("Wait before starting a new sync...")
		time.Sleep(60 * time.Second)
	}
}

func (gs *GitService) End() error {
	return nil
}

func (gs *GitService) SyncServices() {
	// require structure
	// hostname/vendor/project.git
	glob := fmt.Sprintf("%s/*", gs.Config.DataDir)
	services, _ := filepath.Glob(glob)

	gs.Logger.WithFields(log.Fields{
		"glob":     glob,
		"action":   "SyncServices",
		"services": services,
	}).Info("Sync repositories")

	var wg sync.WaitGroup

	for _, path := range services {
		wg.Add(1)

		go gs.SyncRepositories(path, wg)
	}

	wg.Wait()
}

func (gs *GitService) SyncRepositories(service string, wg sync.WaitGroup) {
	gs.Logger.WithFields(log.Fields{
		"action":  "SyncRepositories",
		"service": service,
	}).Info("Sync service's repositories")

	searchPaths := []string{
		fmt.Sprintf("%s/*.git", service),
		fmt.Sprintf("%s/*/*.git", service),
		fmt.Sprintf("%s/*/*/*.git", service),
	}

	paths := []string{}
	for _, searchPath := range searchPaths {
		if p, err := filepath.Glob(searchPath); err != nil {
			continue
		} else {
			paths = append(paths, p...)
		}
	}

	for _, path := range paths {
		logger := gs.Logger.WithFields(log.Fields{
			"path":   path,
			"action": "SyncRepositories",
		})

		logger.Info("Sync repository")

		cmd := exec.Command(gs.Config.Binary, "fetch")
		cmd.Dir = path

		if err := cmd.Start(); err != nil {
			logger.WithError(err).Error("Error while starting the fetch command")

			continue
		}

		if err := cmd.Wait(); err != nil {
			logger.WithError(err).Error("Error while waiting the fetch command")

			continue
		}

		cmd = exec.Command(gs.Config.Binary, "update-server-info")
		cmd.Dir = path

		if err := cmd.Start(); err != nil {
			logger.WithError(err).Error("Error while starting the update-server-info command")

			continue
		}

		if err := cmd.Wait(); err != nil {
			logger.WithError(err).Error("Error while waiting the update-server-info command")

			continue
		}

		gs.Logger.WithFields(log.Fields{
			"path":   path,
			"action": "SyncRepositories",
		}).Debug("Complete the fetch and update-server-info commands")
	}

	wg.Done()
}

func (gs *GitService) WriteArchive(w io.Writer, path, ref string) error {
	if CACHEABLE_REF.Match([]byte(ref)) {
		gs.cacheArchive(w, path, ref)
	} else {
		gs.writeArchive(w, path, ref)
	}

	return nil
}

func (gs *GitService) cacheArchive(w io.Writer, path, ref string) error {
	logger := gs.Logger.WithFields(log.Fields{
		"path":   path,
		"ref":    ref,
		"action": "cacheArchive",
	})

	vaultKey := fmt.Sprintf("%s/%s", path, ref)

	if !gs.Vault.Has(vaultKey) {

		logger.Info("Create vault entry")

		var wg sync.WaitGroup

		pr, pw := io.Pipe()
		wg.Add(1)

		go func() {
			meta := vault.NewVaultMetadata()
			meta["path"] = path
			meta["ref"] = ref

			if _, err := gs.Vault.Put(vaultKey, meta, pr); err != nil {
				logger.WithError(err).Info("Error while writing into vault")
			}
			pr.Close()

			wg.Done()
		}()

		defer pw.Close()

		if err := gs.writeArchive(pw, path, ref); err != nil {
			logger.WithError(err).Info("Error while writing archive")

			return err
		}

		wg.Wait()
	}

	logger.Info("Read vault entry")
	if _, err := gs.Vault.Get(vaultKey, w); err != nil {
		return err
	}

	return nil
}

func (gs *GitService) writeArchive(w io.Writer, path, ref string) error {
	logger := gs.Logger.WithFields(log.Fields{
		"path":   path,
		"action": "writeArchive",
	})

	cmd := exec.Command(gs.Config.Binary, "archive", "--format=zip", ref)
	cmd.Dir = gs.Config.DataDir + string(filepath.Separator) + path

	stdout, _ := cmd.StdoutPipe()

	if err := cmd.Start(); err != nil {
		logger.WithError(err).Error("Error while starting the archive command")

		return err
	}

	if _, err := io.Copy(w, stdout); err != nil {
		logger.WithError(err).Error("Error while reading stdout from the archive command")
	}

	if err := cmd.Wait(); err != nil {
		logger.WithError(err).Error("Error while waiting the archive command")

		return err
	}

	logger.Info("Complete the archive command")

	return nil
}

func (gs *GitService) WriteFile(w io.Writer, path string) error {
	logger := gs.Logger.WithFields(log.Fields{
		"path":   path,
		"action": "WriteFile",
	})

	if f, err := os.Open(gs.Config.DataDir + string(filepath.Separator) + path); err != nil {
		logger.WithError(err).Error("Error while reading file from the fetch command")

		return err
	} else {
		defer f.Close()

		logger.Debug("Sending data to writer")

		io.Copy(w, f)
	}

	return nil
}

func GitRewriteArchive(config *GitConfig, path string) string {
	if results := GITHUB_ARCHIVE.FindStringSubmatch(path); len(results) == 6 {
		return fmt.Sprintf("%s/git/%s/%s/%s/%s.zip", config.Server, results[2], results[3], results[4], results[5])
	}

	if results := BITBUCKET_ARCHIVE.FindStringSubmatch(path); len(results) == 6 {
		return fmt.Sprintf("%s/git/%s/%s/%s/%s.zip", config.Server, results[2], results[3], results[4], results[5])
	}

	if results := GITLAB_ARCHIVE.FindStringSubmatch(path); len(results) == 6 {
		return fmt.Sprintf("%s/git/%s/%s/%s/%s.zip", config.Server, results[2], results[3], results[4], results[5])
	}

	return config.Server
}

func GitRewriteRepository(config *GitConfig, path string) string {

	if results := SVN_REPOSITORY.FindStringSubmatch(path); len(results) > 1 {
		return path // svn not supported
	}

	if results := GIT_REPOSITORY.FindStringSubmatch(path); len(results) > 1 {
		return fmt.Sprintf("%s/git/%s/%s.git", config.Server, results[6], results[8])
	}

	return config.Server
}
