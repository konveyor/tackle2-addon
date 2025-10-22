/*
Package command provides support for addons to
executing (CLI) commands.
*/
package command

import (
	"context"

	"path"

	"github.com/konveyor/tackle2-addon/logging"
	hub "github.com/konveyor/tackle2-hub/addon"
	hubcmd "github.com/konveyor/tackle2-hub/command"
)

var (
	addon = hub.Addon
)

func init() {
	hubcmd.Log = logging.New()
}

type Options = hubcmd.Options

// New returns a command.
func New(path string) (cmd *Command) {
	cmd = &Command{}
	cmd.Path = path
	return
}

// Command execution.
type Command struct {
	hubcmd.Command
	reporter Reporter
}

// Run executes the command.
// The command and output are both reported in
// task Report.Activity.
func (r *Command) Run() (err error) {
	err = r.RunWith(context.TODO())
	return
}

// RunWith executes the command with context.
// The command and output are both reported in
// task Report.Activity.
func (r *Command) RunWith(ctx context.Context) (err error) {
	writer := &Writer{}
	writer.reporter = &r.reporter
	r.Writer = &Writer{}
	output := path.Base(r.Path) + ".output"
	r.reporter.file, err = addon.File.Touch(output)
	if err != nil {
		return
	}
	r.reporter.Run(r.Path, r.Options)
	addon.Attach(r.reporter.file)
	defer func() {
		writer.End()
		if err != nil {
			r.reporter.Error(r.Path, err, writer.buffer)
		} else {
			r.reporter.Succeeded(r.Path, writer.buffer)
		}
	}()
	err = r.Command.RunWith(ctx)
	return
}

// Output returns the command output.
func (r *Command) Output() (b []byte) {
	return r.Writer.(*Writer).buffer
}
