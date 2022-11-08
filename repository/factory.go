package repository

import (
	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
	"os"
)

var (
	addon   = hub.Addon
	HomeDir = ""
)

func init() {
	HomeDir, _ = os.UserHomeDir()
}

type SoftError = hub.SoftError

//
// New SCM repository factory.
func New(destDir string, application *api.Application) (r Repository, err error) {
	kind := application.Repository.Kind
	switch kind {
	case "subversion":
		r = &Subversion{}
	default:
		r = &Git{}
	}
	r.With(destDir, application)
	err = r.Validate()
	return
}

//
// Repository interface.
type Repository interface {
	With(path string, application *api.Application)
	Fetch() (err error)
	Validate() (err error)
	CreateBranch(name string) (err error)
	UseBranch(name string) (err error)
	AddFiles(files []string) (err error)
	Commit(msg string) (err error)
	Push() (err error)
}

//
// SCM - source code manager.
type SCM struct {
	Application *api.Application
	Path        string
}

//
// With settings.
func (r *SCM) With(path string, application *api.Application) {
	r.Application = application
	r.Path = path
}
