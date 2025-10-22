package ssh

import (
	"github.com/konveyor/tackle2-addon/logging"

	"github.com/konveyor/tackle2-hub/ssh"
)

func init() {
	ssh.Log = logging.New()
}

// Agent agent.
type Agent = ssh.Agent
