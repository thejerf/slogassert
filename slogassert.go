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

	     // automatically at the end of your function, an
	     // assertion will run that all log messages are accounted
	     // for.
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
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// Handler implements the slog.Handler interface, with additional
// methods for testing.
type Handler struct {
	// only the top-level logger should be doing the recording,
	// all children loggers need to defer farther down
	parent *Handler

	errToInject error

	leveler      slog.Leveler
	currentGroup []string
	// group -> attrs in that group, "" = default group
	attrs *groupedAttrs

	m           sync.Mutex
	logMessages []logMessage

	t *testing.T
}

// New creates a new testing logger, logging with the given level.
//
// This will automatically call a t.Cleanup to assert that the logger
// is empty. If you want to do this manually or not at all, call
// NewWithoutCleanup.
func New(t *testing.T, leveler slog.Leveler) *Handler {
	if t == nil {
		panic("t must not be nil for a slogtest.Handler")
	}
	handler := &Handler{
		leveler: leveler,
		attrs:   &groupedAttrs{groups: map[string]*groupedAttrs{}},
		t:       t,
	}
	t.Cleanup(handler.AssertEmpty)
	return handler
}

// NewWithoutCleanup creates a new testing logger, logging with the given level.
//
// This does not automatically register a cleanup function to assert
// that the logger is empty.
func NewWithoutCleanup(t *testing.T, leveler slog.Leveler) *Handler {
	if t == nil {
		panic("t must not be nil for a slogtest.Handler")
	}
	handler := &Handler{
		leveler: leveler,
		attrs:   &groupedAttrs{groups: map[string]*groupedAttrs{}},
		t:       t,
	}
	return handler
}

// WithAttrs implements slog.Handler, creating a sub-handler with the given
// hard-coded attributes.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handler := h.child()
	handler.attrs.set(h.currentGroup, attrs...)
	return handler
}

// WithGroup implements slog.Handler, creating a new handler that will group
// everything into the given group.
func (h *Handler) WithGroup(name string) slog.Handler {
	handler := h.child()
	handler.currentGroup = append(append([]string{}, h.currentGroup...), name)
	return handler
}

// Enabled implements slog.Handler, reporting back to slog whether or
// not the handler is enabled for this level of log message.
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.leveler.Level()
}

// Handle implements slog.Handler, recording a log message into the
// root handler.
func (h *Handler) Handle(_ context.Context, record slog.Record) error {
	lm := logMessage{
		message:    record.Message,
		level:      record.Level,
		stacktrace: string(debug.Stack()),
		attrs:      map[string]slog.Value{},
		time:       record.Time,
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
			lm.attrs[encgroups(group, attr.Key)] = val
		}
		return true
	}

	record.Attrs(func(attr slog.Attr) bool {
		return f(h.currentGroup, attr)
	})

	root := h.root()
	root.m.Lock()
	root.logMessages = append(root.logMessages, lm)
	root.m.Unlock()
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
	}
}

type logMessage struct {
	message    string
	level      slog.Level
	stacktrace string
	// key is the slash-encoded group path to this value
	attrs map[string]slog.Value
	// this package deliberately ignores this, but passing
	// testing/slogtest requires us to store this
	time time.Time
}

func (lm *logMessage) print(w io.Writer) {
	msg := strings.Builder{}
	msg.WriteString("--------\nmessage:    ")
	msg.WriteString(lm.message)
	msg.WriteString("\nlevel:      ")
	msg.WriteString(lm.level.String())
	msg.WriteString("\nattributes:\n")
	keys := []string{}
	for attrKey := range lm.attrs {
		keys = append(keys, attrKey)
	}
	sort.Strings(keys)
	for _, key := range keys {
		val := lm.attrs[key]
		msg.WriteString("  ")
		msg.WriteString(key)
		msg.WriteString(" -> (")
		msg.WriteString(val.Kind().String())
		msg.WriteString(") ")
		msg.WriteString(fmt.Sprintf("%v", val.Any()))
		msg.WriteString("\n")
	}
	msg.WriteString("\nstack trace:\n")
	msg.WriteString(lm.stacktrace)
	msg.WriteString("\n")
	_, _ = w.Write([]byte(msg.String()))
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
