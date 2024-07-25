/*
Package slogassert provides a slog Handler that allows testing that
expected logging messages were made in your test code.

# Normal Usage

Normal usage looks like this:

	func TestSomething(t *testing.T) {
	     // This automatically registers a Cleanup function to assert
	     // that all log messages are accounted for.
	     handler := slogassert.New(t, slog.LevelWarn)
	     logger := slog.New(handler)

	     // inject the logger into your test code and run it

	     // Now start asserting things:
	     handler.AssertSomeOf("some log message")

	     // often useful to finish up with an assertion that
	     // all log messages have been accounted for:
	     handler.AssertEmpty()
	}

A variety of assertions at varying levels of detail are available on
the Handler.
*/
package slogassert

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"
)

// Handler implements the slog.Handler interface, with additional
// methods for testing.
//
// All methods on this Handler are thread-safe.
type Handler struct {
	// only the top-level logger should be doing the recording,
	// all children loggers need to defer farther down
	parent *Handler

	wrapped slog.Handler

	leveler      slog.Leveler
	currentGroup []string
	// group -> attrs in that group, "" = default group
	attrs *groupedAttrs

	m           sync.Mutex
	logMessages []LogMessage

	t Tester
}

// The Tester interface defines the incoming testing interface.
//
// The standard library *testing.T and *testing.B values conform to this
// already.
//
// If your testing library doesn't have an equivalent of Helper, it is fine
// to implement it as a no-op.
type Tester interface {
	Helper()
	Fatalf(string, ...any)
}

// New creates a new testing logger, logging with the given level.
//
// If wrapped is not nil, Handle calls will be passed down to that
// handler as well.
//
// It is recommended to generally call defer handler.AssertEmpty() on
// the result of this call.
func New(t Tester, leveler slog.Leveler, wrapped slog.Handler) *Handler {
	if t == nil {
		panic("t must not be nil for a slogtest.Handler")
	}
	handler := &Handler{
		leveler: leveler,
		attrs:   &groupedAttrs{groups: map[string]*groupedAttrs{}},
		t:       t,
		wrapped: wrapped,
	}
	return handler
}

// WithAttrs implements slog.Handler, creating a sub-handler with the given
// hard-coded attributes.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handler := h.child()
	handler.attrs.set(h.currentGroup, attrs...)

	if h.wrapped != nil {
		handler.wrapped = h.wrapped.WithAttrs(attrs)
	}

	return handler
}

// WithGroup implements slog.Handler, creating a new handler that will group
// everything into the given group.
func (h *Handler) WithGroup(name string) slog.Handler {
	handler := h.child()
	handler.currentGroup = append(append([]string{},
		h.currentGroup...), name)

	if h.wrapped != nil {
		handler.wrapped = h.wrapped.WithGroup(name)
	}

	return handler
}

// Enabled implements slog.Handler, reporting back to slog whether or
// not the handler is enabled for this level of log message.
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.leveler.Level()
}

// Handle implements slog.Handler, recording a log message into the
// root handler.
func (h *Handler) Handle(ctx context.Context, record slog.Record) error {
	lm := LogMessage{
		Message:    record.Message,
		Level:      record.Level,
		Stacktrace: string(debug.Stack()),
		Attrs:      map[string]slog.Value{},
		Time:       record.Time,
	}

	var f func(group []string, attr slog.Attr) bool
	f = func(group []string, attr slog.Attr) bool {
		val := attr.Value.Resolve()
		switch val.Kind() {
		case slog.KindGroup:
			groupName := attr.Key
			attrs := val.Group()
			newGroups := append(append([]string{}, group...), groupName)
			for _, attr := range attrs {
				f(newGroups, attr)
			}

		default:
			lm.Attrs[encgroups(group, attr.Key)] = val
		}
		return true
	}

	record.Attrs(func(attr slog.Attr) bool {
		return f(h.currentGroup, attr)
	})

	root := h.root()
	root.m.Lock()
	h.attrs.runOn(f)

	root.logMessages = append(root.logMessages, lm)
	root.m.Unlock()

	if h.wrapped != nil {
		return h.wrapped.Handle(ctx, record)
	}

	return nil
}

func (h *Handler) root() *Handler {
	for h.parent != nil {
		h = h.parent
	}
	return h
}

func (h *Handler) child() *Handler {
	return &Handler{
		parent:       h,
		currentGroup: append([]string{}, h.currentGroup...),
		attrs:        h.attrs.clone(),
		leveler:      h.leveler,
		wrapped:      h.wrapped,
	}
}

// LogMessage is a struct for storing the log messages picked up by
// slogassert's handler.
type LogMessage struct {
	Message    string
	Level      slog.Level
	Stacktrace string
	// key is the slash-encoded group path to this value
	Attrs map[string]slog.Value
	// this package deliberately ignores this, but passing
	// testing/slogtest requires us to store this
	Time time.Time
}

// Print is a default method that can dump a LogMessage out to a
// writer; this is used by slogassert to print unasserted log messages.
func (lm *LogMessage) Print(w io.Writer) {
	msg := strings.Builder{}
	msg.WriteString("--------\nmessage:    ")
	msg.WriteString(lm.Message)
	msg.WriteString("\nlevel:      ")
	msg.WriteString(lm.Level.String())
	msg.WriteString("\nattributes:\n")
	keys := []string{}
	for attrKey := range lm.Attrs {
		keys = append(keys, attrKey)
	}
	sort.Strings(keys)
	for _, key := range keys {
		val := lm.Attrs[key]
		msg.WriteString("  ")
		msg.WriteString(key)
		msg.WriteString(" -> (")
		msg.WriteString(val.Kind().String())
		msg.WriteString(") ")
		msg.WriteString(fmt.Sprintf("%v", val.Any()))
		msg.WriteString("\n")
	}
	msg.WriteString("\nstack trace:\n")
	msg.WriteString(lm.Stacktrace)
	msg.WriteString("\n")
	_, _ = w.Write([]byte(msg.String()))
}

func (lm *LogMessage) clone() LogMessage {
	return LogMessage{
		Message:    lm.Message,
		Level:      lm.Level,
		Stacktrace: lm.Stacktrace,
		Time:       lm.Time,
		Attrs:      maps.Clone(lm.Attrs),
	}
}

type groupedAttrs struct {
	// the attrs at this group level
	attrs []slog.Attr

	// child groups
	groups map[string]*groupedAttrs
}

func (ga *groupedAttrs) set(groupkeys []string, attr ...slog.Attr) {
	target := ga
	for _, group := range groupkeys {
		newTarget := target.groups[group]
		if newTarget == nil {
			target.groups[group] = &groupedAttrs{
				groups: map[string]*groupedAttrs{},
			}
			newTarget = target.groups[group]
		}
		target = newTarget
	}

	target.attrs = append(target.attrs, attr...)
}

func (ga *groupedAttrs) clone() *groupedAttrs {
	new := &groupedAttrs{
		attrs:  append([]slog.Attr{}, ga.attrs...),
		groups: map[string]*groupedAttrs{},
	}
	for group, child := range ga.groups {
		new.groups[group] = child.clone()
	}
	return new
}

// should be run only under handler lock
func (ga *groupedAttrs) runOn(f func([]string, slog.Attr) bool) {
	ga.runOnRecursive(nil, f)
}

func (ga *groupedAttrs) runOnRecursive(
	currGroup []string,
	f func([]string, slog.Attr) bool,
) {
	for _, attr := range ga.attrs {
		res := f(currGroup, attr)
		if !res {
			return
		}
	}

	for group, subGa := range ga.groups {
		newGroup := append(currGroup, group)
		subGa.runOnRecursive(newGroup, f)
	}
}

func dotEncode(s string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(s, "\\", "\\\\"),
		".",
		"\\.",
	)
}

// encgroups unambiguously encodes slices into a single
// string. backslash encoding is a bit klunky, but it's so traditional
// it seems the best choice anyhow.
func encgroups(group []string, key string) string {
	converted := make([]string, len(group)+1)
	for i := 0; i < len(group); i++ {
		converted[i] = dotEncode(group[i])
	}
	converted[len(converted)-1] = dotEncode(key)
	return strings.Join(converted, ".")
}
