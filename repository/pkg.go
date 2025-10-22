package repository

import (
	"os"

	"github.com/konveyor/tackle2-addon/logging"
	hub "github.com/konveyor/tackle2-hub/addon"
	hubcmd "github.com/konveyor/tackle2-hub/command"
	hubscm "github.com/konveyor/tackle2-hub/scm"
)

var (
	addon = hub.Addon
	Dir   = ""
)

func init() {
	Dir, _ = os.Getwd()
	hubcmd.Log = logging.New()
	hubscm.Log = logging.New()
}

type Remote = hubscm.Remote
type SCM = hubscm.SCM
type Subversion = hubscm.Subversion
type Git = hubscm.Git

// New SCM repository factory.
// Options:
// - Insecure
// - *api.Identity
// - api.Identity
func New(destDir string, remote *Remote, option ...any) (r SCM, err error) {
	var insecure bool
	switch remote.Kind {
	case "subversion":
		insecure, err = addon.Setting.Bool("svn.insecure.enabled")
		if err != nil {
			return
		}
		svn := &Subversion{}
		svn.HomeRoot = Dir
		svn.Path = destDir
		svn.Remote = *remote
		svn.Insecure = insecure
		r = svn
	default:
		insecure, err = addon.Setting.Bool("git.insecure.enabled")
		if err != nil {
			return
		}
		git := &Git{}
		git.HomeRoot = Dir
		git.Path = destDir
		git.Remote = *remote
		git.Insecure = insecure
		r = git
	}
	err = r.Validate()
	if err != nil {
		return
	}
	for _, opt := range option {
		err = r.Use(opt)
		if err != nil {
			return
		}
	}
	return
}
