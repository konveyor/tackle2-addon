package command

import (
	"strings"
)

//
// Verbosity.
const (
	// Disabled reports: NOTHING.
	Disabled = -2
	// Error reports: error.
	Error = -1
	// Default reports: error, started, succeeded.
	Default = 0
	// LiveOutput reports: error, started, succeeded, output (live).
	LiveOutput = 1
)

//
// ReportFilter filter reported output.
type ReportFilter func(in string) (out string)

//
// Reporter activity reporter.
type Reporter struct {
	Filter    ReportFilter
	Verbosity int
	index     int
}

//
// Run reports command started in task Report.Activity.
func (r *Reporter) Run(path string, options Options) {
	switch r.Verbosity {
	case Disabled:
	case Error:
	case Default,
		LiveOutput:
		addon.Activity(
			"[CMD] Running: %s %s",
			path,
			strings.Join(options, " "))
	}
}

//
// Succeeded reports command succeeded in task Report.Activity.
func (r *Reporter) Succeeded(path string) {
	switch r.Verbosity {
	case Disabled:
	case Error:
	case Default,
		LiveOutput:
		addon.Activity("[CMD] %s succeeded.", path)
	}
}

//
// Error reports command failed in task Report.Activity.
func (r *Reporter) Error(path string, err error, output []byte) {
	if len(output) == 0 {
		return
	}
	switch r.Verbosity {
	case Disabled:
	case Error,
		Default:
		addon.Activity(
			"[CMD] %s failed: %s.\n%s",
			path,
			err.Error(),
			output)
	case LiveOutput:
		addon.Activity(
			"[CMD] %s failed: %s.",
			path,
			err.Error())
	}
}

//
// Output reports command output in task Report.Activity.
// Returns the number of bytes reported.
func (r *Reporter) Output(buffer []byte, delimited bool) (reported int) {
	if r.Filter == nil {
		r.Filter = func(in string) (out string) {
			out = in
			return
		}
	}
	switch r.Verbosity {
	case Disabled:
	case Error:
	case Default:
	case LiveOutput:
		if r.index >= len(buffer) {
			return
		}
		batch := string(buffer[r.index:])
		if delimited {
			end := strings.LastIndex(batch, "\n")
			if end != -1 {
				batch = batch[:end]
				output := r.Filter(batch)
				addon.Activity("> %s", output)
				reported = len(output)
				r.index += len(batch)
				r.index++
			}
		} else {
			output := r.Filter(batch)
			addon.Activity("> %s", output)
			reported = len(batch)
			r.index = len(buffer)
		}
	}
	return
}
