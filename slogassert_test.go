package slogassert

import (
	"log/slog"
	"testing"
	"time"
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
