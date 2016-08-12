// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/rande/goapp"
	"github.com/rande/gonode/core/vault"
	"github.com/rande/pkgmirror"
)

var (
	BITBUCKET_ARCHIVE = regexp.MustCompile(`http(s|):\/\/([\w-\.]+)\/([\w\.\d-]+)\/([\w-\.\d]+)\/get\/([\w]+)\.zip`)
	GITHUB_ARCHIVE    = regexp.MustCompile(`http(s|):\/\/api\.([\w-\.]+)\/repos\/([\w\.\d-]+)\/([\w\.\d-]+)\/zipball\/([\w]+)`)
	GITLAB_ARCHIVE    = regexp.MustCompile(`http(s|):\/\/([\w-\.]+)\/([\w-\.\d]+)\/([\w-\.\d]+)\/repository\/archive.zip\?ref=([\w]+)`)

	GIT_REPOSITORY = regexp.MustCompile(`^(((git|http(s|)):\/\/|git@))([\w-\.]+@|)([\w-\.]+)(\/|:)([\w-\.\/]+?)(\.git|)$`)
	SVN_REPOSITORY = regexp.MustCompile(`(svn:\/\/(.*)|(.*)\.svn\.(.*))`)

	CACHEABLE_REF = regexp.MustCompile(`([\w\d]{40}|[\w\d]+\.[\w\d]+\.[\w\d]+(-[\w\d]+|))`)
)

func NewGitService() *GitService {
	return &GitService{
		Config: &GitConfig{
			DataDir:      "./data/git",
			Binary:       "git",
			SourceServer: "git@github.com:%s",
			PublicServer: "http://localhost:8000",
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
	PublicServer string
	SourceServer string
	Server       string
	DataDir      string
	Binary       string
	Clone        string
}

type GitService struct {
	DB        *bolt.DB
	Config    *GitConfig
	Logger    *log.Entry
	Vault     *vault.Vault
	StateChan chan pkgmirror.State
}

func (gs *GitService) Init(app *goapp.App) error {
	os.MkdirAll(string(filepath.Separator)+gs.Config.DataDir, 0755)

	return nil
}

func (gs *GitService) Serve(state *goapp.GoroutineState) error {
	syncEnd := make(chan bool)

	sync := func() {
		gs.Logger.Info("Starting a new sync...")

		gs.syncRepositories(fmt.Sprintf("%s/%s", gs.Config.DataDir, gs.Config.Server))

		syncEnd <- true
	}

	// start the first sync
	go sync()

	for {
		select {
		case <-state.In:
			return nil

		case <-syncEnd:
			gs.StateChan <- pkgmirror.State{
				Message: "Wait for a new run",
				Status:  pkgmirror.STATUS_HOLD,
			}

			gs.Logger.Info("Wait before starting a new sync...")

			// we recursively call sync unless a state.In comes in to exist the current
			// go routine (ie, the Serve function). This might not close the sync processus
			// completely. We need to have a proper channel (queue mode) for git fetch.
			// This will probably make this current code obsolete.
			go func() {
				time.Sleep(60 * time.Second)
				sync()
			}()
		}
	}
}

func (gs *GitService) syncRepositories(service string) {
	gs.Logger.WithFields(log.Fields{
		"action":  "SyncRepositories",
		"datadir": service,
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
			"path":   path[len(service):],
			"action": "SyncRepositories",
		})

		gs.StateChan <- pkgmirror.State{
			Message: fmt.Sprintf("Fetch %s", path[len(service):]),
			Status:  pkgmirror.STATUS_RUNNING,
		}

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

		gs.Logger.WithFields(log.Fields{
			"path":   path,
			"action": "SyncRepositories",
		}).Debug("Complete the fetch command")
	}
}

func (gs *GitService) WriteArchive(w io.Writer, path, ref string) error {
	if CACHEABLE_REF.Match([]byte(ref)) {
		return gs.cacheArchive(w, path, ref)
	} else {
		return gs.writeArchive(w, path, ref)
	}
}

func (gs *GitService) cacheArchive(w io.Writer, path, ref string) error {
	logger := gs.Logger.WithFields(log.Fields{
		"path":   path,
		"ref":    ref,
		"action": "cacheArchive",
	})

	vaultKey := fmt.Sprintf("%s:%s/%s", gs.Config.Server, path, ref)

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

				gs.Vault.Remove(vaultKey)
			}

			wg.Done()
		}()

		if err := gs.writeArchive(pw, path, ref); err != nil {
			logger.WithError(err).Info("Error while writing archive")

			pw.Close()
			pr.Close()

			gs.Vault.Remove(vaultKey)

			return err
		} else {
			pw.Close()
		}

		wg.Wait()

		pr.Close()
	}

	logger.Info("Read vault entry")
	if _, err := gs.Vault.Get(vaultKey, w); err != nil {
		return err
	}

	return nil
}

func (gs *GitService) dataFolder() string {
	return gs.Config.DataDir + string(filepath.Separator) + gs.Config.Server
}

func (gs *GitService) writeArchive(w io.Writer, path, ref string) error {
	logger := gs.Logger.WithFields(log.Fields{
		"path":   path,
		"action": "writeArchive",
	})

	cmd := exec.Command(gs.Config.Binary, "archive", "--format=zip", ref)
	cmd.Dir = gs.dataFolder() + string(filepath.Separator) + path

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

func (gs *GitService) Has(path string) bool {
	gitPath := gs.dataFolder() + string(filepath.Separator) + path

	has := true
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		has = false
	}

	gs.Logger.WithFields(log.Fields{
		"path":   gitPath,
		"action": "Has",
		"has":    has,
	}).Debug("Has repository?")

	return has
}

func (gs *GitService) Clone(path string) error {
	gitPath := gs.dataFolder() + string(filepath.Separator) + path
	remote := strings.Replace(gs.Config.Clone, "{path}", path, -1)

	if gs.Config.Clone == remote {
		// same key, no replacement
		return pkgmirror.SameKeyError
	}

	logger := gs.Logger.WithFields(log.Fields{
		"path":   path,
		"action": "Clone",
		"remote": remote,
	})

	logger.Info("Starting cloning remote repository")

	cmd := exec.Command(gs.Config.Binary, "clone", "--mirror", remote, gitPath)

	logger.WithField("cmd", cmd.Args).Debug("Run command")

	if err := cmd.Start(); err != nil {
		logger.WithError(err).Error("Error while starting to clone the remote repository")

		return err
	}

	if err := cmd.Wait(); err != nil {
		logger.WithError(err).Error("Error while cloning the remote repository")

		return err
	}

	return nil
}

func GitRewriteArchive(publicServer, path string) string {
	if results := GITHUB_ARCHIVE.FindStringSubmatch(path); len(results) == 6 {
		return fmt.Sprintf("%s/git/%s/%s/%s/%s.zip", publicServer, results[2], results[3], results[4], results[5])
	}

	if results := BITBUCKET_ARCHIVE.FindStringSubmatch(path); len(results) == 6 {
		return fmt.Sprintf("%s/git/%s/%s/%s/%s.zip", publicServer, results[2], results[3], results[4], results[5])
	}

	if results := GITLAB_ARCHIVE.FindStringSubmatch(path); len(results) == 6 {
		return fmt.Sprintf("%s/git/%s/%s/%s/%s.zip", publicServer, results[2], results[3], results[4], results[5])
	}

	return publicServer
}

func GitRewriteRepository(publicServer, path string) string {
	if results := SVN_REPOSITORY.FindStringSubmatch(path); len(results) > 1 {
		return path // svn not supported
	}

	if results := GIT_REPOSITORY.FindStringSubmatch(path); len(results) > 1 {
		return fmt.Sprintf("%s/git/%s/%s.git", publicServer, results[6], results[8])
	}

	return publicServer
}
