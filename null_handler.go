package slogassert

import (
	"context"
	"log/slog"
)

type nullHandler struct{}

func (nh nullHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}

func (nh nullHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

func (nh nullHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return nh
}

func (nh nullHandler) WithGroup(_ string) slog.Handler {
	return nh
}

// NullHandler returns a slog.Handler that does nothing.
func NullHandler() slog.Handler { return nullHandler{} }

// NullLogger returns a *slog.Logger pointed at a NullHandler.
func NullLogger() *slog.Logger {
	return slog.New(NullHandler())
}
