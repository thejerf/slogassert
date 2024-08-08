package slogassert

import (
	"log/slog"
	"reflect"
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
	handler := New(&HandlerOptions{
		T:       t,
		Leveler: slog.LevelWarn,
	})
	defer handler.AssertEmpty()
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

	if handler.AssertSomeMessage(testWarning) != 1 {
		t.Fatal("incorrect number return from AssertSomeMessage")
	}
}

func TestSlogHandler(t *testing.T) {
	handler := New(&HandlerOptions{
		T:       t,
		Leveler: slog.LevelDebug,
	})
	defer handler.AssertEmpty()
	err := slogtest.TestHandler(handler, func() []map[string]any {
		results := []map[string]any{}

		for _, lm := range handler.logMessages {
			handler.logMessages = handler.logMessages[:0]

			result := map[string]any{
				slog.TimeKey:    lm.Time,
				slog.LevelKey:   lm.Level,
				slog.MessageKey: lm.Message,
			}
			results = append(results, result)
		}

		// empty out the results for next time
		handler.logMessages = []LogMessage{}
		return results
	})
	if err != nil {
		t.Fatalf("incorrect handler behavior: %v", err)
	}
}

func TestAssertSomeMessage(t *testing.T) {
	handler := New(&HandlerOptions{
		T:       t,
		Leveler: slog.LevelWarn,
	})
	defer handler.AssertEmpty()
	log := slog.New(handler)
	log.Warn(testWarning)

	msgs := handler.Unasserted()
	if len(msgs) != 1 {
		t.Fatal("incorrect number of unasserted messages")
	}
	msgs[0].Time = time.Time{}
	msgs[0].Stacktrace = ""
	if !reflect.DeepEqual(msgs, []LogMessage{
		{
			Message: testWarning,
			Level:   slog.LevelWarn,
			Attrs:   map[string]slog.Value{},
		},
	}) {
		t.Fatal("incorrect return for Unasserted")
	}

	log.Warn(testWarning)
	log.Warn(testWarning)

	if handler.AssertSomeMessage(testWarning) != 3 {
		t.Fatal("incorrect number returned by AssertSomeMessage")
	}
	// does not crash because they are all consumed
}

func TestAssertMessage(t *testing.T) {
	handler := New(&HandlerOptions{
		T:       t,
		Leveler: slog.LevelWarn,
	})
	defer handler.AssertEmpty()
	log := slog.New(handler)
	log.Warn(testWarning)

	handler.AssertMessage(testWarning)
	// does not crash because the message is consumed
}

func TestAssertPrecise(t *testing.T) {
	handler := New(&HandlerOptions{
		T:       t,
		Leveler: slog.LevelWarn,
	})
	defer handler.AssertEmpty()
	log := slog.New(handler)
	log.Warn(testWarning, "key", "val")

	handler.AssertPrecise(LogMessageMatch{
		Message: testWarning,
		Level:   slog.LevelWarn,
		Attrs: map[string]any{
			"key": "val",
		},
	})

	// slog upgrades ints to int64s, and then when we pass a base
	// "1" in the source code to test for them the type doesn't
	// match. This is quite annoying. Test that we fixed that,
	// which is to say, in the Attrs map, we don't need to cast to
	// int64:
	log.Warn(testWarning,
		"int", 1,
		"float", float64(2),
		"uint", uint64(3),
	)
	handler.AssertPrecise(LogMessageMatch{
		Message: testWarning,
		Level:   slog.LevelWarn,
		Attrs: map[string]any{
			"int":   1,
			"float": 2,
			"uint":  3,
		},
	})

	// does not crash because the message is consumed
}

func TestReset(t *testing.T) {
	handler := New(&HandlerOptions{
		T:       t,
		Leveler: slog.LevelWarn,
	})
	defer handler.AssertEmpty()
	log := slog.New(handler)
	log.Warn(testWarning, "key", "val")

	handler.Reset()
	// does not crash because the message is consumed
}

func TestFiltering(t *testing.T) {
	handler := New(&HandlerOptions{
		T:       t,
		Leveler: slog.LevelWarn,
	})
	defer handler.AssertEmpty()
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
	handler := New(&HandlerOptions{
		T:       t,
		Leveler: slog.LevelWarn,
	})
	defer handler.AssertEmpty()
	log := slog.New(handler)
	log.Warn(testWarning, "key", "val")
	log.Warn(testWarning, "key", "val")
	log.Warn(testWarning, "key", "val")

	if handler.AssertSomePrecise(LogMessageMatch{
		Message: testWarning,
		Level:   slog.LevelWarn,
		Attrs: map[string]any{
			"key": "val",
		},
	}) != 3 {
		t.Fatal("incorrect number returned from AssertSomePrecise")
	}

	// does not crash because the message is consumed
}

func TestLogValuer(t *testing.T) {
	handler := New(&HandlerOptions{
		T:       t,
		Leveler: slog.LevelWarn,
	})
	defer handler.AssertEmpty()
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
	handler := New(&HandlerOptions{
		T:       t,
		Leveler: slog.LevelWarn,
		Wrapped: wrappedHandler,
	})
	defer handler.AssertEmpty()
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
	if strings.TrimSpace(remainder) != `level=WARN msg="test warning" test.a=b` {
		t.Fatal("did not get expected log result")
	}
}

func TestWith(t *testing.T) {
	handler := New(&HandlerOptions{
		T:       t,
		Leveler: slog.LevelWarn,
	})
	defer handler.AssertEmpty()
	log := slog.New(handler)
	subLog := log.With("test_attr", "value").WithGroup("group").With("test2", "value2")
	subLog.Error(testWarning)
	handler.AssertPrecise(LogMessageMatch{
		Message: testWarning,
		Level:   slog.LevelError,
		Attrs: map[string]any{
			"test_attr":   "value",
			"group.test2": "value2",
		},
		AllAttrsMatch: true,
	})
}

func TestDetectDupes(t *testing.T) {
	handler := New(&HandlerOptions{
		T:           t,
		Leveler:     slog.LevelWarn,
		DetectDupes: true,
	})
	defer handler.AssertEmpty()
	log := slog.New(handler)
	subLog := log.With(slog.String("test_attr", "a"))
	panics(t, "Sub log with duplicate attr", func() { subLog.With(slog.String("test_attr", "b")) })
	panics(t, "Simultaneously registered duplicate attr", func() { log.With(slog.String("test_attr", "a"), slog.String("test_attr", "b")) })
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
	panics(t, "New with nil", func() { New(nil) })
	panics(t, "New with empty handler options", func() { New(&HandlerOptions{}) })
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
