package slogassert

import (
	"log/slog"
	"testing"
	"time"

	"testing/slogtest"
)

const (
	testWarning = "test warning"
)

func TestVeryBasicFunctionality(t *testing.T) {
	handler := New(t, slog.LevelWarn)
	log := slog.New(handler)

	sublog := log.WithGroup("test").With(
		slog.String("constant", "hello"),
	)

	sublog.Warn(testWarning,
		slog.Group("req",
			slog.String("method", "GET"),
			slog.Group("notreq",
				slog.String("innotgroup", "can nest"),
			),
			slog.String("url", "/url")),
		slog.Int("status", 200),
		slog.Duration("duration", time.Second),
	)

	handler.AssertSomeOf(testWarning)
}

func TestSlogHandler(t *testing.T) {
	handler := New(t, slog.LevelDebug)
	err := slogtest.TestHandler(handler, func() []map[string]any {
		results := []map[string]any{}

		for _, lm := range handler.logMessages {
			handler.logMessages = handler.logMessages[:0]

			result := map[string]any{
				slog.TimeKey:    lm.time,
				slog.LevelKey:   lm.level,
				slog.MessageKey: lm.message,
			}
			results = append(results, result)
		}

		// empty out the results for next time
		handler.logMessages = []logMessage{}
		return results
	})
	if err != nil {
		t.Fatalf("incorrect handler behavior: %v", err)
	}
}
