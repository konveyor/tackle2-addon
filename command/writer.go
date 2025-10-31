package command

import (
	"fmt"
	"io"
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

// Writer reports command output.
// Provides both io.Reader and io.Writer.
// Command output is buffered (rate-limited) and reported.
type Writer struct {
	reporter *Reporter
	buffer   []byte
	backoff  time.Duration
	end      chan any
	ended    chan any
	read     int
}

// Write command output.
func (w *Writer) Write(p []byte) (n int, err error) {
	n = len(p)
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

// Read the buffer.
func (w *Writer) Read(p []byte) (n int, err error) {
	if w.read >= len(w.buffer) {
		err = io.EOF
		return
	}
	n = copy(p, w.buffer[w.read:])
	w.read += n
	return
}

// Seek to read position.
// Provides io.Seeker.
func (w *Writer) Seek(offset int64, whence int) (n int64, err error) {
	switch whence {
	case io.SeekStart:
		n = offset
	case io.SeekCurrent:
		n = int64(w.read) + offset
	case io.SeekEnd:
		n = int64(len(w.buffer)) + offset
	default:
		err = fmt.Errorf("whence not valid: %d", whence)
		return
	}
	if n < 0 || n > int64(len(w.buffer)) {
		err = fmt.Errorf("out of bounds")
		return
	}
	w.read = int(n)
	return
}

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

// Reporter returns the reporter.
func (w *Writer) Reporter() *Reporter {
	return w.reporter
}

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
		n := w.reporter.Output(w.buffer)
		w.adjustBackoff(n)
		if ended && n == 0 {
			break
		}
	}
	w.ended <- true
}

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
