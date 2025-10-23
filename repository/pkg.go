package repository

import (
	"os"
	"path"

	"github.com/konveyor/tackle2-addon/command"
	"github.com/konveyor/tackle2-addon/sink"
	hub "github.com/konveyor/tackle2-hub/addon"
	hubscm "github.com/konveyor/tackle2-hub/scm"
)

var (
	addon = hub.Addon
	Dir   = ""
)

func init() {
	Dir, _ = os.Getwd()
	hubscm.Log = hubscm.Log.WithSink(sink.New(true))
}

type Remote = hubscm.Remote
type SCM = hubscm.SCM
type Subversion = hubscm.Subversion
type Git = hubscm.Git

func init() {
	hubscm.NewCommand = command.New
}

// New SCM repository factory.
func New(destDir string, remote *Remote, option ...any) (r SCM, err error) {
	var insecure bool
	switch remote.Kind {
	case "subversion":
		insecure, err = addon.Setting.Bool("svn.insecure.enabled")
		if err != nil {
			return
		}
		svn := &Subversion{}
		svn.Home = path.Join(Dir, svn.Id())
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
		git.Home = path.Join(Dir, git.Id())
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
