package slogassert

import (
	"log"
	"log/slog"
	"os"
	"testing"
)

// config is used to configure the default handler, and allow the
// use of functional options.
type config struct {
	level       slog.Leveler
	assertEmpty bool
	wrapped     slog.Handler
	detectDupes bool
}

// An Option allows for configuration of the default handler created
// by [NewDefault].
type Option func(*config)

// WithLeveler is a functional option for [NewDefault] that sets the
// minimum log level of the default handler.
// All messages below this level will be ignored.
func WithLeveler(level slog.Leveler) Option {
	return func(c *config) {
		c.level = level
	}
}

// WithAssertEmpty is a functional option for [NewDefault] that configures
// the handler to validated that all messages have been captured and
// asserted.
func WithAssertEmpty() Option {
	return func(c *config) {
		c.assertEmpty = true
	}
}

// WithWrapped is a functional option for [NewDefault] that wraps the
// generated default handler with another handler.
// If set then handle calls will be passed down to that
// handler as well.
func WithWrapped(wrapped slog.Handler) Option {
	return func(c *config) {
		c.wrapped = wrapped
	}
}

// WithDetectDupes is a functional option for [NewDefault] that configures the
// handler to panic on detection of duplicate slog.Attr keys.
func WithDetectDupes() Option {
	return func(c *config) {
		c.detectDupes = true
	}
}

// NewDefault is a helper function for tests that creates a slogassert [Handler]
// and sets it as the default slog handler
// Once the test is complete it will attempt to restore the previous handler.
//
// It accepts options to customize the handler, including
//   - [WithLeveler] to set the log level
//   - [WithAssertEmpty] to assert that the handler is empty at the end of the test
//   - [WithWrapped] to wrap the handler with another handler
//
// Example:
//
//	func TestExample(t *testing.T) {
//		handler := NewDefault(t, WithLeveler(slog.LevelError))
//
//	 	CodeUnderTest()
//
//	 	handler.AssertMessage("expected log message")
//	}
//
// This function MUST NOT be used with t.Parallel(). Doing so will cause unexpected
// results.
func NewDefault(t testing.TB, opts ...Option) *Handler {
	// config used to allow for functional options
	c := config{
		level:       slog.LevelDebug,
		assertEmpty: false,
		wrapped:     nil,
		detectDupes: false,
	}

	for _, opt := range opts {
		opt(&c)
	}
	handler := New(&HandlerOptions{
		T:           t,
		Leveler:     c.level,
		Wrapped:     c.wrapped,
		DetectDupes: c.detectDupes,
	})

	// take a copy of the original logger and flags so that we can restore
	// once the test is complete
	origLogger := slog.Default()
	origFlags := log.Flags()

	t.Cleanup(func() {
		// if WithAssertEmpty was set, then ensure that all messages
		// have been captured.
		if c.assertEmpty {
			handler.AssertEmpty()
		}

		// if the original logger was a default logger, then slog does not
		// allow full restoration as this would cause a race condition.
		// Instead, we attempt to manually  restore.
		// This may cause unexpected results if the output of the default
		// logged had been set to something other than stdout.
		// This is unlikely to have happened during test flows, and would
		// also be unlikely to cause problems in tests if it has happened.
		log.SetOutput(os.Stdout)
		log.SetFlags(origFlags)

		// if the original logger was not a default logger, then we can
		// safely restore it.
		slog.SetDefault(origLogger)
	})

	// create a new logger with the slogassert handler, and set it as the
	// default logger
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// the slogassert handler is returned to allow for tests to validate log
	//messages
	return handler
}
