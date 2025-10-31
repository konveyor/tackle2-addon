/*
Package command provides support for addons to
executing (CLI) commands.
*/
package command

import (
	"context"

	"path"

	"github.com/konveyor/tackle2-addon/sink"
	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/command"
)

var (
	addon = hub.Addon
)

func init() {
	command.Log = command.Log.WithSink(sink.New(false))
}

type Options = command.Options

// New returns a command.
func New(p string) (cmd *command.Command) {
	cmd = command.New(p)
	reporter := &Reporter{}
	writer := &Writer{}
	writer.reporter = reporter
	cmd.Begin = func() (err error) {
		cmd.Writer = writer
		output := path.Base(cmd.Path) + ".output"
		reporter.file, err = addon.File.Touch(output)
		if err != nil {
			return
		}
		reporter.Run(cmd.Path, cmd.Options)
		addon.Attach(reporter.file)
		return
	}
	cmd.End = func() {
		writer.End()
		if cmd.Error != nil {
			reporter.Error(cmd.Path, cmd.Error, writer.buffer)
		} else {
			reporter.Succeeded(cmd.Path, writer.buffer)
		}
	}
	return
}

// Command execution.
type Command struct {
	command.Command
	Reporter Reporter
}

// Run executes the command.
// The command and output are both reported in
// task Report.Activity.
func (r *Command) Run() (err error) {
	err = r.RunWith(context.TODO())
	return
}
