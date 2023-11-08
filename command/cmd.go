/*
Package command provides support for addons to
executing (CLI) commands.
*/
package command

import (
	"context"
	"fmt"
	hub "github.com/konveyor/tackle2-hub/addon"
	"os/exec"
	"regexp"
	"strings"
)

var (
	addon = hub.Addon
)

// AuthMask basic auth mask.
var AuthMask = MaskPattern{
	Regex:       regexp.MustCompile(`://[^:]+:[^@]+@`),
	Replacement: "://###:###@",
}

//
// Command execution.
type Command struct {
	Options  Options
	Mask     Mask
	ErrorMap ErrorMap
	Path     string
	Dir      string
	Output   string
	Silent   bool
}

//
// Run executes the command.
// The command and output are both reported in
// task Report.Activity.
func (r *Command) Run() (err error) {
	err = r.RunWith(context.TODO())
	return
}

//
// RunWith executes the command with context.
// The command and output are both reported in
// task Report.Activity.
func (r *Command) RunWith(ctx context.Context) (err error) {
	r.activity(
		"Running: %s %s",
		r.Path,
		r.Options.String(r.Mask))
	cmd := exec.CommandContext(ctx, r.Path, r.Options...)
	cmd.Dir = r.Dir
	output, err := cmd.CombinedOutput()
	r.Output = r.Mask.Apply(string(output))
	if err != nil {
		err = r.ErrorMap.Error(err, r.Output)
		r.activity(
			"%s failed: %s.\n%s",
			r.Path,
			err.Error(),
			string(r.Output))
	} else {
		r.activity("succeeded.")
	}
	return
}

//
// activity reports activity.
func (r *Command) activity(s string, x ...any) {
	if !r.Silent {
		s += "[CMD] "
		addon.Activity(s, x...)
	}
}

//
// Options are CLI options.
type Options []string

//
//
func (a *Options) String(mask Mask) (s string) {
	var masked []string
	for _, option := range *a {
		masked = append(masked, mask.Apply(option))
	}
	s = strings.Join(masked, " ")
	return
}

//
// Add adds option.
func (a *Options) Add(option string, s ...string) {
	*a = append(*a, option)
	*a = append(*a, s...)
}

//
// Addf adds option.
func (a *Options) Addf(option string, x ...any) {
	*a = append(*a, fmt.Sprintf(option, x...))
}

type MaskPattern struct {
	Regex       *regexp.Regexp
	Replacement string
}

//
// Apply returns masked sting.
func (m *MaskPattern) Apply(in string) (out string) {
	out = m.Regex.ReplaceAllString(in, m.Replacement)
	return
}

//
// Mask collection of mask patterns.
type Mask []MaskPattern

//
// Apply returns masked sting.
func (f *Mask) Apply(in string) (out string) {
	out = in
	for _, p := range *f {
		out = p.Apply(out)
	}
	return
}

//
// ErrorPattern defines errors found in output.
type ErrorPattern struct {
	Regex *regexp.Regexp
	Error func(s string) error
}

func (r *ErrorPattern) Find(output string) (err error) {
	matched := r.Regex.Find([]byte(output))
	if matched != nil {
		err = r.Error(string(matched))
	}
	return
}

//
// ErrorMap finds/builds errors using pattern matching.
type ErrorMap []ErrorPattern

//
// MapError returns masked options.
func (r *ErrorMap) Error(in error, output string) (out error) {
	out = in
	for _, p := range *r {
		err := p.Find(output)
		if err != nil {
			out = err
			break
		}
	}
	return
}
