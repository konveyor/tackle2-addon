package ssh

import (
	"context"
	"fmt"
	"os"
	pathlib "path"
	"strings"
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-addon/command"
	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/nas"
)

var (
	addon  = hub.Addon
	Dir    = ""
	SSHDir = ""
)

func init() {
	Dir, _ = os.Getwd()
	SSHDir = pathlib.Join(Dir, ".ssh")
}

// Agent agent.
type Agent struct {
}

// Start the ssh-agent.
func (r *Agent) Start() (err error) {
	pid := os.Getpid()
	addon.Activity("[SSH] Home (directory): %s", Dir)
	socket := fmt.Sprintf("/tmp/agent.%d", pid)
	cmd := command.New("/usr/bin/ssh-agent")
	cmd.Env = append(os.Environ(), "HOME="+Dir)
	cmd.Options.Add("-a", socket)
	err = cmd.Run()
	if err != nil {
		return
	}
	_ = os.Setenv("SSH_AUTH_SOCK", socket)
	err = nas.MkDir(SSHDir, 0700)
	if err != nil {
		return
	}

	addon.Activity("[SSH] Agent started.")

	return
}

// Add ssh key.
func (r *Agent) Add(id *api.Identity, host string) (err error) {
	if id.Key == "" {
		return
	}
	addon.Activity("[SSH] Adding key: %s", id.Name)
	suffix := fmt.Sprintf("id_%d", id.ID)
	path := pathlib.Join(
		SSHDir,
		suffix)
	f, err := os.OpenFile(
		path,
		os.O_RDWR|os.O_CREATE,
		0600)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
		return
	}
	_, err = f.Write([]byte(r.format(id.Key)))
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
	}
	_ = f.Close()
	ask, err := r.writeAsk(id)
	if err != nil {
		return
	}
	ctx, fn := context.WithTimeout(
		context.TODO(),
		time.Second)
	defer fn()
	cmd := command.New("/usr/bin/ssh-add")
	cmd.Env = append(
		os.Environ(),
		"DISPLAY=:0",
		"SSH_ASKPASS="+ask,
		"HOME="+Dir)
	cmd.Options.Add(path)
	err = cmd.RunWith(ctx)
	if err != nil {
		return
	}
	addon.Activity("[FILE] Created %s.", path)
	return
}

// Ensure key formatting.
func (r *Agent) format(in string) (out string) {
	if in != "" {
		out = strings.TrimSpace(in) + "\n"
	}
	return
}

// writeAsk writes script that returns the key password.
func (r *Agent) writeAsk(id *api.Identity) (path string, err error) {
	path = "/tmp/ask.sh"
	f, err := os.OpenFile(
		path,
		os.O_RDWR|os.O_CREATE,
		0700)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
		return
	}
	defer func() {
		_ = f.Close()
	}()
	script := "#!/bin/sh\n"
	script += "echo " + id.Password
	_, err = f.Write([]byte(script))
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
	}
	return
}
