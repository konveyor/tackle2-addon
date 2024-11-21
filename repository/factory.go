package repository

import (
	"os"

	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
)

var (
	addon   = hub.Addon
	HomeDir = ""
)

func init() {
	HomeDir, _ = os.UserHomeDir()
}

type Remote = api.Repository

// New SCM repository factory.
func New(destDir string, remote *Remote, identities []api.Ref) (r SCM, err error) {
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
		svn.Identities = identities
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
		git.Identities = identities
		git.Insecure = insecure
		r = git
	}
	err = r.Validate()
	return
}

// SCM interface.
type SCM interface {
	Validate() (err error)
	Fetch() (err error)
	Branch(ref string) (err error)
	Commit(files []string, msg string) (err error)
	Head() (commit string, err error)
}

// Authenticated repository.
type Authenticated struct {
	Identities []api.Ref
	Insecure   bool
}

// FindIdentity by kind.
func (r *Authenticated) findIdentity(kind string) (matched *api.Identity, found bool, err error) {
	for _, ref := range r.Identities {
		identity, nErr := addon.Identity.Get(ref.ID)
		if nErr != nil {
			err = nErr
			return
		}
		if identity.Kind == kind {
			found = true
			matched = identity
			break
		}
	}
	return
}
