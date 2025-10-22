package logging

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	hub "github.com/konveyor/tackle2-hub/addon"
)

var (
	addon = hub.Addon
)

func New() logr.Logger {
	return logr.Logger{}.WithSink(&Sink{})
}

// Sink used to bridge a Logger to the addon.Activity.
type Sink struct{}

func (s *Sink) Init(_ logr.RuntimeInfo) {
}

func (s *Sink) Enabled(_ int) (enabled bool) {
	enabled = true
	return
}

func (s *Sink) Info(_ int, msg string, kv ...any) {
	msg = s.join(msg, kv)
	addon.Activity(msg)
	return
}

func (s *Sink) Error(err error, msg string, kv ...any) {
	msg = s.join(msg, kv)
	msg += "\n"
	msg += err.Error()
	addon.Activity(msg)
	return
}

func (s *Sink) WithValues(_ ...any) (sink logr.LogSink) {
	sink = s
	return
}

func (s *Sink) WithName(_ string) (sink logr.LogSink) {
	sink = s
	return
}

func (s *Sink) join(m string, kv ...any) (joined string) {
	items := []string{}
	for i := range kv {
		if i%2 != 0 {
			key := fmt.Sprintf("%v", kv[i-1])
			v := fmt.Sprintf("%+v", kv[i])
			p := fmt.Sprintf("%s=%s", key, v)
			items = append(items, p)
		}
	}
	joined = m
	if len(items) > 0 {
		joined += ":"
		joined += strings.Join(items, ",")
	}
	return
}
