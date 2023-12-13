package command

import (
	"time"
)

const (
	// Backoff rate increment.
	Backoff = time.Millisecond * 100
	// MaxBackoff max backoff.
	MaxBackoff = 10 * Backoff
	// MinBackoff minimum backoff.
	MinBackoff = Backoff
)

//
// OutputFilter filter output.
type OutputFilter func(in []byte) (out []byte)

//
// Writer records command output.
type Writer struct {
	Filter   OutputFilter
	reporter *Reporter
	buffer   []byte
	backoff  time.Duration
	end      chan any
	ended    chan any
}

//
// Write command output.
func (w *Writer) Write(p []byte) (n int, err error) {
	if w.Filter == nil {
		w.Filter = func(in []byte) (out []byte) {
			out = in
			return
		}
	}
	n = len(p)
	p = w.Filter(p)
	w.buffer = append(w.buffer, p...)
	switch w.reporter.Verbosity {
	case LiveOutput:
		if w.ended == nil {
			w.end = make(chan any)
			w.ended = make(chan any)
			go w.report()
		}
	}
	return
}

//
// End of writing.
func (w *Writer) End() {
	if w.end == nil {
		return
	}
	close(w.end)
	<-w.ended
	close(w.ended)
	w.end = nil
}

//
// report in task Report.Activity.
// Rate limited.
func (w *Writer) report() {
	w.backoff = MinBackoff
	ended := false
	for {
		select {
		case <-w.end:
			ended = true
		case <-time.After(w.backoff):
		}
		n := w.reporter.Output(w.buffer, true)
		w.adjustBackoff(n)
		if ended {
			break
		}
	}
	w.reporter.Output(w.buffer, false)
	w.ended <- true
}

//
// adjustBackoff adjust the backoff as needed.
// incremented when output reported.
// decremented when no outstanding output reported.
func (w *Writer) adjustBackoff(reported int) {
	if reported > 0 {
		if w.backoff < MaxBackoff {
			w.backoff += Backoff
		}
	} else {
		if w.backoff > MinBackoff {
			w.backoff -= Backoff
		}
	}
}
