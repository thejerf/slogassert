package slogassert

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"testing"
	"time"
)

type MatchTest struct {
	Value     slog.Value
	GoodVal   func(slog.Value) bool
	BadVal    func(slog.Value) bool
	GoodFunc  any
	BadFunc   any
	GoodMatch any
	BadMatch  any
}

func TestMatchers(t *testing.T) {
	now := time.Now()
	notNow := now.Add(time.Minute)

	for idx, test := range []MatchTest{
		// Any, here played by a slice of string
		{
			Value: slog.AnyValue([]string{}),
			GoodVal: func(val slog.Value) bool {
				return len(val.Any().([]string)) == 0
			},
			BadVal: func(val slog.Value) bool {
				return len(val.Any().([]string)) == 1
			},
			GoodFunc: func(v any) bool {
				return len(v.([]string)) == 0
			},
			BadFunc: func(v any) bool {
				return len(v.([]string)) == 1
			},
			GoodMatch: []string{},
			BadMatch:  []string{"nope"},
		},

		// KindBool
		{
			Value: slog.BoolValue(true),
			GoodVal: func(val slog.Value) bool {
				return val.Bool()
			},
			BadVal: func(val slog.Value) bool {
				return !val.Bool()
			},
			GoodFunc: func(b bool) bool {
				return b
			},
			BadFunc: func(b bool) bool {
				return !b
			},
			GoodMatch: true,
			BadMatch:  false,
		},

		// KindDuration
		{
			Value: slog.DurationValue(time.Second),
			GoodVal: func(val slog.Value) bool {
				return val.Duration() == time.Second
			},
			BadVal: func(val slog.Value) bool {
				return val.Duration() == 0
			},
			GoodFunc: func(d time.Duration) bool {
				return d == time.Second
			},
			BadFunc: func(d time.Duration) bool {
				return d == 0
			},
			GoodMatch: time.Second,
			BadMatch:  time.Minute,
		},

		// KindFloat64
		{
			Value: slog.Float64Value(1),
			GoodVal: func(val slog.Value) bool {
				return val.Float64() == 1
			},
			BadVal: func(val slog.Value) bool {
				return val.Float64() == 0
			},
			GoodFunc: func(f float64) bool {
				return f == 1
			},
			BadFunc: func(f float64) bool {
				return f == 0
			},
			GoodMatch: float64(1),
			BadMatch:  float64(0),
		},

		// KindInt64
		{
			Value: slog.Int64Value(1),
			GoodVal: func(val slog.Value) bool {
				return val.Int64() == 1
			},
			BadVal: func(val slog.Value) bool {
				return val.Int64() == 0
			},
			GoodFunc: func(f int64) bool {
				return f == 1
			},
			BadFunc: func(f int64) bool {
				return f == 0
			},
			GoodMatch: int64(1),
			BadMatch:  int64(0),
		},

		// KindString
		{
			Value: slog.StringValue("hi!"),
			GoodVal: func(val slog.Value) bool {
				return val.String() == "hi!"
			},
			BadVal: func(val slog.Value) bool {
				return val.String() == ""
			},
			GoodFunc: func(s string) bool {
				return s == "hi!"
			},
			BadFunc: func(s string) bool {
				return s == ""
			},
			GoodMatch: "hi!",
			BadMatch:  "",
		},

		// KindTime
		{
			Value: slog.TimeValue(now),
			GoodVal: func(val slog.Value) bool {
				return val.Time().Equal(now)
			},
			BadVal: func(val slog.Value) bool {
				return val.Time().Equal(notNow)
			},
			GoodFunc: func(t time.Time) bool {
				return t.Equal(now)
			},
			BadFunc: func(t time.Time) bool {
				return t.Equal(notNow)
			},
			GoodMatch: now,
			BadMatch:  notNow,
		},

		// KindUint64
		{
			Value: slog.Uint64Value(1),
			GoodVal: func(val slog.Value) bool {
				return val.Uint64() == 1
			},
			BadVal: func(val slog.Value) bool {
				return val.Uint64() == 0
			},
			GoodFunc: func(f uint64) bool {
				return f == 1
			},
			BadFunc: func(f uint64) bool {
				return f == 0
			},
			GoodMatch: uint64(1),
			BadMatch:  uint64(0),
		},
	} {
		if matchAttr(test.GoodVal, test.Value) != nil {
			t.Fatalf("test %d failed", idx)
		}
		if matchAttr(test.BadVal, test.Value) != errNoMatch {
			t.Fatalf("test %d failed", idx)
		}
		if matchAttr(test.GoodFunc, test.Value) != nil {
			fmt.Println(matchAttr(test.GoodFunc, test.Value))
			t.Fatalf("test %d failed", idx)
		}
		if matchAttr(test.BadFunc, test.Value) != errNoMatch {
			t.Fatalf("test %d failed", idx)
		}
		if matchAttr(test.GoodMatch, test.Value) != nil {
			t.Fatalf("test %d failed", idx)
		}
		if matchAttr(test.BadMatch, test.Value) != errNoMatch {
			t.Fatalf("test %d failed", idx)
		}

		// idx 0 is a special case; since it's the KindAny
		// case, there is no value that can be passed to the
		// first parameter of matchAttr that can't be
		// reflect.DeepEqual'd against the test.Value to get a
		// "invalid type" error. Everything else has invalid
		// type possibilities.
		if idx != 0 {
			err := matchAttr([]int{}, test.Value)
			if err == nil || !strings.Contains(err.Error(), "invalid type") {
				t.Fatalf("incorrect bad type error: %v", err)
			}
		}
	}

	if matchAttr(ValueAsString(2), slog.StringValue("2")) != nil {
		t.Fatal("can't handle LogValue matchers")
	}
	if matchAttr(ValueAsString(2), slog.StringValue("3")) == nil {
		t.Fatal("can't handle LogValue matchers")
	}
}

// A ValueAsString is an int that is also a LogValuer that returns
// itself as a string for slog. This is used for testing that the
// matching value is correct.
type ValueAsString int

func (vas ValueAsString) LogValue() slog.Value {
	return slog.StringValue(strconv.Itoa(int(vas)))
}
