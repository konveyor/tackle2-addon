package ssh

import (
	"os"

	"github.com/konveyor/tackle2-addon/command"
	"github.com/konveyor/tackle2-addon/logging"

	"github.com/konveyor/tackle2-hub/ssh"
)

func init() {
	ssh.Log = logging.New()
	ssh.Home, _ = os.Getwd()
	ssh.NewCommand = command.New
}

// Agent agent.
type Agent = ssh.Agent
