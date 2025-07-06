package repository

import (
	"os"

	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/pkg/errors"
)

var (
	addon = hub.Addon
	Dir   = ""
)

func init() {
	Dir, _ = os.Getwd()
}

type Remote = api.Repository

// New SCM repository factory.
// Options:
// - *api.Ref
// - api.Ref
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
	for _, opt := range option {
		err = r.Use(opt)
		if err != nil {
			return
		}
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
	Use(option any) (err error)
}

// Authenticated repository.
type Authenticated struct {
	Identity api.Identity
	Insecure bool
}

// Use option.
// Options:
// - *api.Ref
// - api.Ref
// - *api.Identity
// - api.Identity
func (a *Authenticated) Use(option any) (err error) {
	var id *api.Identity
	switch opt := option.(type) {
	case *api.Ref:
		if opt == nil {
			return
		}
		id, err = addon.Identity.Get(opt.ID)
		if err != nil {
			return
		}
		a.Identity = *id
	case api.Ref:
		id, err = addon.Identity.Get(opt.ID)
		if err != nil {
			return
		}
		a.Identity = *id
	case *api.Identity:
		if opt != nil {
			a.Identity = *opt
		}
	case api.Identity:
		a.Identity = opt
	default:
		err = errors.Errorf("Invalid option: %T", opt)
	}
	return
}
