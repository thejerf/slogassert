package slogassert

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"testing/slogtest"
)

const (
	testWarning = "test warning"
	test2       = "test2"
	test3       = "test3"
)

func TestVeryBasicFunctionality(t *testing.T) {
	handler := New(t, slog.LevelWarn, nil)
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

	handler.AssertSomeMessage(testWarning)
}

func TestSlogHandler(t *testing.T) {
	handler := New(t, slog.LevelDebug, nil)
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

func TestWithoutCleanup(t *testing.T) {
	handler := NewWithoutCleanup(t, slog.LevelWarn, nil)
	log := slog.New(handler)
	log.Warn(testWarning)
	// this test doesn't fail at the end, because we didn't
	// register a cleanup function
}

func TestAssertSomeMessage(t *testing.T) {
	handler := New(t, slog.LevelWarn, nil)
	log := slog.New(handler)
	log.Warn(testWarning)
	log.Warn(testWarning)
	log.Warn(testWarning)

	handler.AssertSomeMessage(testWarning)
	// does not crash because they are all consumed
}

func TestAssertMessage(t *testing.T) {
	handler := New(t, slog.LevelWarn, nil)
	log := slog.New(handler)
	log.Warn(testWarning)

	handler.AssertMessage(testWarning)
	// does not crash because the message is consumed
}

func TestAssertSomeMessageLevel(t *testing.T) {
	handler := New(t, slog.LevelWarn, nil)
	log := slog.New(handler)
	log.Warn(testWarning)
	log.Warn(testWarning)
	log.Warn(testWarning)

	handler.AssertSomeMessageLevel(testWarning, slog.LevelWarn)
	// does not crash because the messages are consumed
}

func TestAssertPrecise(t *testing.T) {
	handler := New(t, slog.LevelWarn, nil)
	log := slog.New(handler)
	log.Warn(testWarning, "key", "val")

	handler.AssertPrecise(LogMessageMatch{
		Message: testWarning,
		Level:   slog.LevelWarn,
		Attrs: map[string]any{
			"key": "val",
		},
	})

	// does not crash because the message is consumed
}

func TestReset(t *testing.T) {
	handler := New(t, slog.LevelWarn, nil)
	log := slog.New(handler)
	log.Warn(testWarning, "key", "val")

	handler.Reset()
	// does not crash because the message is consumed
}

func TestFiltering(t *testing.T) {
	handler := New(t, slog.LevelWarn, nil)
	log := slog.New(handler)

	log.Warn(testWarning)
	log.Warn(test2)
	log.Warn(test3)

	handler.AssertMessage(test2)
	handler.AssertMessage(test3)
	handler.AssertMessage(testWarning)

	// does not crash because the messages were properly consumed
}

func TestAssertSomePrecise(t *testing.T) {
	handler := New(t, slog.LevelWarn, nil)
	log := slog.New(handler)
	log.Warn(testWarning, "key", "val")
	log.Warn(testWarning, "key", "val")
	log.Warn(testWarning, "key", "val")

	handler.AssertSomePrecise(LogMessageMatch{
		Message: testWarning,
		Level:   slog.LevelWarn,
		Attrs: map[string]any{
			"key": "val",
		},
	})

	// does not crash because the message is consumed
}

func TestAssertMessageLevel(t *testing.T) {
	handler := New(t, slog.LevelWarn, nil)
	log := slog.New(handler)
	log.Warn(testWarning)

	handler.AssertMessageLevel(testWarning, slog.LevelWarn)
	// does not crash because the message is consumed
}

func TestLogValuer(t *testing.T) {
	handler := New(t, slog.LevelWarn, nil)
	log := slog.New(handler)
	log.Warn(testWarning,
		"valuer", testLogValuer{},
		slog.Group("group", "subgroup", testLogValuer{}),
	)

	handler.AssertPrecise(LogMessageMatch{
		Message: testWarning,
		Level:   slog.LevelWarn,
		Attrs: map[string]any{
			"valuer.a":         "a",
			"valuer.b":         "b",
			"group.subgroup.a": "a",
			"group.subgroup.b": "b",
		},
		AllAttrsMatch: true,
	})
}

func TestWrapping(t *testing.T) {
	buf := &strings.Builder{}
	wrappedHandler := slog.NewTextHandler(buf, nil)
	handler := New(t, slog.LevelWarn, wrappedHandler)
	logger := slog.New(handler)

	logger.WithGroup("test").Warn(testWarning,
		"a", "b")

	handler.AssertPrecise(LogMessageMatch{
		Message: testWarning,
		Level:   slog.LevelWarn,
		Attrs: map[string]any{
			"test.a": "b",
		},
		AllAttrsMatch: true,
	})

	logged := buf.String()
	// hack off the timestamp because it is always different
	_, remainder, _ := strings.Cut(logged, " ")
	if strings.TrimSpace(remainder) != `level=WARN msg="test warning" a=b` {
		t.Fatal("did not get expected log result")
	}
}

type testLogValuer struct{}

func (t testLogValuer) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("a", "a"),
		slog.String("b", "b"),
	)
}

// various little assertions to cover the code
func TestCoverage(t *testing.T) {
	panics(t, "New with nil", func() { New(nil, slog.LevelWarn, nil) })
	panics(t, "NewWithoutCleanup",
		func() { NewWithoutCleanup(nil, slog.LevelWarn, nil) })
}

func panics(t *testing.T, name string, f func()) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("%s failed to panic", name)
		}
	}()

	f()
}
