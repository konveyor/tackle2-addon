package repository

import (
	"os"

	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
)

var (
	addon   = hub.Addon
	HomeDir = ""
	Dir     = ""
)

func init() {
	HomeDir, _ = os.UserHomeDir()
	Dir, _ = os.Getwd()
}

type Remote = api.Repository

// New SCM repository factory.
func New(destDir string, remote *Remote, identity *api.Ref) (r SCM, err error) {
	var insecure bool
	switch remote.Kind {
	case "subversion":
		insecure, err = addon.Setting.Bool("svn.insecure.enabled")
		if err != nil {
			return
		}
		svn := &Subversion{}
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
		git.Path = destDir
		git.Remote = *remote
		git.Insecure = insecure
		r = git
	}
	err = r.Validate()
	if err != nil {
		return
	}
	err = r.Use(identity)
	if err != nil {
		return
	}
	return
}

// SCM interface.
type SCM interface {
	Validate() (err error)
	Fetch() (err error)
	Branch(ref string) (err error)
	Commit(files []string, msg string) (err error)
	Head() (commit string, err error)
	Use(identity *api.Ref) (err error)
}

// Authenticated repository.
type Authenticated struct {
	Identity api.Identity
	Insecure bool
}

// Use identity (ref) resolves the reference and sets the identity.
func (a *Authenticated) Use(identity *api.Ref) (err error) {
	if identity == nil {
		return
	}
	id, err := addon.Identity.Get(identity.ID)
	if err != nil {
		return
	}
	a.Identity = *id
	return
}
