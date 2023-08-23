package slogassert

import "testing"

func TestNullHandler(t *testing.T) {
	l := NullLogger()

	// Just verify this doesn't crash.
	l.With("x", "y").WithGroup("nope").Debug("no")
}
