package repository

import (
	"errors"
	"fmt"
	"hash/fnv"
	urllib "net/url"
	"os"
	pathlib "path"
	"strconv"
	"strings"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-addon/command"
	"github.com/konveyor/tackle2-addon/ssh"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/nas"
)

// Git repository.
type Git struct {
	Authenticated
	Remote Remote
	Path   string
}

// Validate settings.
func (r *Git) Validate() (err error) {
	u := GitURL{}
	err = u.With(r.Remote.URL)
	if err != nil {
		return
	}
	switch u.Scheme {
	case "http":
		if !r.Insecure {
			err = errors.New("http URL used with git.insecure.enabled = FALSE")
			return
		}
	}
	return
}

// Fetch clones the repository.
func (r *Git) Fetch() (err error) {
	url := r.URL()
	addon.Activity("[GIT] Home (directory): %s", r.home())
	addon.Activity("[GIT] Cloning: %s", url.String())
	_ = nas.RmDir(r.Path)
	if r.Identity.ID != 0 {
		addon.Activity(
			"[GIT] Using credentials (id=%d) %s.",
			r.Identity.ID,
			r.Identity.Name)
	}
	err = nas.MkDir(r.home(), 0755)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = r.writeConfig()
	if err != nil {
		return
	}
	err = r.writeCreds()
	if err != nil {
		return
	}
	agent := ssh.Agent{}
	err = agent.Add(&r.Identity, url.Host)
	if err != nil {
		return
	}
	cmd := r.git()
	cmd.Options.Add("clone")
	cmd.Options.Add("--depth", "1")
	if r.Remote.Branch != "" {
		cmd.Options.Add("--single-branch")
		cmd.Options.Add("--branch", r.Remote.Branch)
	}
	cmd.Options.Add(url.String(), r.Path)
	err = cmd.Run()
	if err != nil {
		return
	}
	err = r.checkout()
	return
}

// Branch creates a branch with the given name if not exist and switch to it.
func (r *Git) Branch(ref string) (err error) {
	cmd := r.git()
	cmd.Dir = r.Path
	cmd.Options.Add("checkout", ref)
	err = cmd.Run()
	if err != nil {
		cmd = command.New("/usr/bin/git")
		cmd.Dir = r.Path
		cmd.Options.Add("checkout", "-b", ref)
	}
	r.Remote.Branch = ref
	err = cmd.Run()
	return
}

// addFiles adds files to staging area.
func (r *Git) addFiles(files []string) (err error) {
	cmd := r.git()
	cmd.Dir = r.Path
	cmd.Options.Add("add", files...)
	err = cmd.Run()
	return
}

// Commit files and push to remote.
func (r *Git) Commit(files []string, msg string) (err error) {
	err = r.addFiles(files)
	if err != nil {
		return err
	}
	cmd := r.git()
	cmd.Dir = r.Path
	cmd.Options.Add("commit")
	cmd.Options.Add("--allow-empty")
	cmd.Options.Add("-m", msg)
	err = cmd.Run()
	if err != nil {
		return err
	}
	err = r.push()
	return
}

// Head returns HEAD commit.
func (r *Git) Head() (commit string, err error) {
	cmd := r.git()
	cmd.Dir = r.Path
	cmd.Options.Add("rev-parse")
	cmd.Options.Add("HEAD")
	err = cmd.Run()
	if err != nil {
		return
	}
	commit = string(cmd.Output())
	commit = strings.TrimSpace(commit)
	return
}

// git returns git command.
func (r *Git) git() (cmd *command.Command) {
	cmd = command.New("/usr/bin/git")
	cmd.Env = append(
		os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_TRACE_SETUP=1",
		"GIT_TRACE=1",
		"HOME="+r.home())
	return
}

// push changes to remote.
func (r *Git) push() (err error) {
	cmd := r.git()
	cmd.Dir = r.Path
	cmd.Options.Add("push", "origin", "HEAD")
	err = cmd.Run()
	return
}

// URL returns the parsed URL.
func (r *Git) URL() (u GitURL) {
	u = GitURL{}
	_ = u.With(r.Remote.URL)
	return
}

// writeConfig writes config file.
func (r *Git) writeConfig() (err error) {
	path := pathlib.Join(r.home(), ".gitconfig")
	f, err := os.Create(path)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
		return
	}
	proxy, err := r.proxy()
	if err != nil {
		return
	}
	s := "[user]\n"
	s += "name = Konveyor Dev\n"
	s += "email = konveyor-dev@googlegroups.com\n"
	s += "[credential]\n"
	s += "helper = store --file="
	s += pathlib.Join(r.home(), ".git-credentials")
	s += "\n"
	s += "[http]\n"
	s += fmt.Sprintf("sslVerify = %t\n", !r.Insecure)
	if proxy != "" {
		s += fmt.Sprintf("proxy = %s\n", proxy)
	}
	_, err = f.Write([]byte(s))
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
	}
	_ = f.Close()
	addon.Activity("[FILE] Created %s.", path)
	return
}

// writeCreds writes credentials (store) file.
func (r *Git) writeCreds() (err error) {
	if r.Identity.User == "" || r.Identity.Password == "" {
		return
	}
	path := pathlib.Join(r.home(), ".git-credentials")
	f, err := os.Create(path)
	if err != nil {
		err = liberr.Wrap(
			err,
			"path",
			path)
		return
	}
	url := r.URL()
	for _, scheme := range []string{
		"https",
		"http",
	} {
		entry := scheme
		entry += "://"
		if r.Identity.User != "" {
			entry += r.Identity.User
			entry += ":"
		}
		if r.Identity.Password != "" {
			entry += r.Identity.Password
			entry += "@"
		}
		entry += url.Host
		_, err = f.Write([]byte(entry + "\n"))
		if err != nil {
			err = liberr.Wrap(
				err,
				"path",
				path)
			break
		}
	}
	_ = f.Close()
	addon.Activity("[FILE] Created %s.", path)
	return
}

// proxy builds the proxy.
func (r *Git) proxy() (proxy string, err error) {
	kind := ""
	url := r.URL()
	switch url.Scheme {
	case "http":
		kind = "http"
	case "https",
		"git@github.com":
		kind = "https"
	default:
		return
	}
	p, err := addon.Proxy.Find(kind)
	if err != nil || p == nil || !p.Enabled {
		return
	}
	for _, h := range p.Excluded {
		if h == url.Host {
			return
		}
	}
	addon.Activity(
		"[GIT] Using proxy (%d) %s.",
		p.ID,
		p.Kind)
	auth := ""
	if p.Identity != nil {
		var id *api.Identity
		id, err = addon.Identity.Get(p.Identity.ID)
		if err != nil {
			return
		}
		auth = fmt.Sprintf(
			"%s:%s@",
			id.User,
			id.Password)
	}
	proxy = fmt.Sprintf(
		"http://%s%s",
		auth,
		p.Host)
	if p.Port > 0 {
		proxy = fmt.Sprintf(
			"%s:%d",
			proxy,
			p.Port)
	}
	return
}

// checkout ref.
func (r *Git) checkout() (err error) {
	branch := r.Remote.Branch
	if branch == "" {
		return
	}
	dir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(dir)
	}()
	_ = os.Chdir(r.Path)
	cmd := r.git()
	cmd.Options.Add("checkout", branch)
	err = cmd.Run()
	return
}

// home returns the Git home directory path.
func (r *Git) home() (home string) {
	h := fnv.New32a()
	_, _ = h.Write([]byte(r.Remote.URL))
	n := h.Sum32()
	digest := strconv.FormatUint(uint64(n), 16)
	home = pathlib.Join(
		Dir,
		".git",
		digest)
	return
}

// GitURL git clone URL.
type GitURL struct {
	Raw    string
	Scheme string
	Host   string
	Path   string
}

// With populates the URL.
func (r *GitURL) With(u string) (err error) {
	r.Raw = u
	parsed, pErr := urllib.Parse(u)
	if pErr == nil {
		r.Scheme = parsed.Scheme
		r.Host = parsed.Host
		r.Path = parsed.Path
		return
	}
	notValid := liberr.New("URL not valid.")
	part := strings.Split(u, ":")
	if len(part) != 2 {
		err = notValid
		return
	}
	r.Host = part[0]
	r.Path = part[1]
	part = strings.Split(r.Host, "@")
	if len(part) != 2 {
		err = notValid
		return
	}
	r.Host = part[1]
	return
}

// String representation.
func (r *GitURL) String() string {
	return r.Raw
}
