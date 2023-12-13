package main

import (
	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-addon/command"
)

var (
	addon = hub.Addon
)

func main() {
	addon.Run(func() (err error) {
		cmd := command.New("ps")
		cmd.Options.Add("-ef")
		cmd.Reporter.Verbosity = command.LiveOutput
		err = cmd.Run()
		return
	})
}
