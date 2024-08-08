package slogassert_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/thejerf/slogassert"
)

// fakeTestingT is a testing.T used in the runnable example to demostrate usage
type fakeTestingT struct {
	*testing.T
}

func (ft *fakeTestingT) Run(_ string, f func(t *testing.T)) {
	f(ft.T)
}

var t = &fakeTestingT{
	T: &testing.T{},
}

// CodeUnderTest is an example function, used to demonstrate usage.
func CodeUnderTest() {
	slog.ErrorContext(context.Background(), "expected log message")
}

// --

func ExampleNewDefault() {
	t.Run("ensure correct slog message is written", func(t *testing.T) {
		// update the default logger, and then reset it at the end of the test
		st := slogassert.NewDefault(
			t,
			slogassert.WithLeveler(slog.LevelInfo), // only capture info and above
			slogassert.WithAssertEmpty(),           // ensure that all messages have been captured
		)

		// ...
		// run the test code

		CodeUnderTest()

		// ...

		// capture and assert that the logged message
		st.AssertMessage("expected log message")
	})

	// Output:
}

// --

func Test_NewDefault(t *testing.T) {
	t.Run("With default slog handler", func(t *testing.T) {
		defaultHandler := slogassert.NewDefault(t)

		slog.Info("This should be captured")
		slog.Info("This should be ignored")

		defaultHandler.AssertMessage("This should be captured")
	})

	t.Run("With custom slog handler", func(t *testing.T) {
		handler := slogassert.New(&slogassert.HandlerOptions{
			T:       t,
			Leveler: slog.LevelDebug,
		})
		slog.SetDefault(slog.New(handler))

		t.Run("subtest", func(t *testing.T) {
			defaultHandler := slogassert.NewDefault(t)
			slog.Info("This should be captured")
			slog.Info("This should be ignored")

			defaultHandler.AssertMessage("This should be captured")
		})

		slog.Info("This should also be captured")
		handler.AssertMessage("This should also be captured")
	})

	t.Run("With wrapped logger", func(t *testing.T) {
		handler := slogassert.New(&slogassert.HandlerOptions{
			T:       t,
			Leveler: slog.LevelDebug,
		})
		defaultHandler := slogassert.NewDefault(t, slogassert.WithWrapped(handler))

		slog.Info("This should be captured")

		handler.AssertMessage("This should be captured")
		defaultHandler.AssertMessage("This should be captured")
	})

	t.Run("With leveler", func(t *testing.T) {
		defaultHandler := slogassert.NewDefault(
			t,
			slogassert.WithLeveler(slog.LevelInfo),
			slogassert.WithAssertEmpty(),
		)

		slog.Debug("This should be ignored")
		slog.Info("This should be captured")

		defaultHandler.AssertMessage("This should be captured")
	})

}
