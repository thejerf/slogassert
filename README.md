# slogassert


    import "github.com/thejerf/slogassert"

[![GoDoc](https://pkg.go.dev/badge/github.com/thejerf/slogassert)](https://pkg.go.dev/github.com/thejerf/slogassert)

## About Slogassert

slogassert implements a _testing handler_ for slog.

All important observable results of code should be tested. While
commonly ignored by people writing tests, log messages are often
important things to test for, for several reasons:

 * Because they are not tested for, it is suprisingly easy for them to
   break without anyone realizing.
 * Log messages can have security impact. Any log messages that may be
   used to reconstruct a security incident should be tested to ensure
   they contain the data they are supposed to contain.
 * Even when you don't otherwise deeply care about log
   messages, log messages can often still double as a way of asserting
   that certain code was actually reached, and the value of variables
   at the time it was reached is as expected.
   
While I wouldn't pull in slogassert just for that third point, if the
first two are in play for your code base, slog's strong focus on
structured attributes means that log messages can also double as
testing probe points within your code that may be otherwise
unreachable. I don't use this a lot but when you need it it's very
useful. (You may find it helpful to create a testing log level even
above Debug for this.)

This implements a handler for slog that more-or-less records the
incoming messages, then provides a mechanism for testing for the
presence and number of log messages at a variety of different detail
levels. As log messages are tested for and matched by assertions, they
are removed from the record. If an assertion is made and no log
messages match, a testing error will be thrown.

Once all the assertions are complete, an assertion should be made that
all log messages are accounted for. If there are unaccounted log
messages, the test will fail at the end.

Because this is a test logger, some things normally too expensive to
be done in normal logging are done, like taking a full stack trace at
the location of all log messages that are recorded. This assists in
diagnosing where any unasserted log messages are coming from.

See [usage in the godoc](https://pkg.go.dev/github.com/thejerf/slogassert).

This also includes a null logger, which I am surprised is not in the
library itself as I write this since a `nil` slog.Logger is invalid.

## Release Status

`slogassert` is **beta**. I consider the package as it nows stands to
be a 1.0 release candidate, but it is not yet officially 1.0.

The code is covered as much as possible, however it is not possible to
directly test covering the code that calls fatal errors, so that code
can not be covered or directly tested.

## Version Numbering

This repository will use semantic versioning.

I will be signing this repository with the ["jerf" keybase
account](https://keybase.io/jerf). If you are viewing this repository
through GitHub, you should see the commits as showing as "verified" in
the commit view.

(Bear in mind that due to the nature of how git commit signing works,
there may be runs of unverified commits; what matters is that the top
one is signed.)

## Version History

* v0.3.3:
  * Export LogMessageMatch.Matches for external use.
  * Add utility function for testing.
  * Take the `testing.TB` interface rather than a constant `*testing.T`.
* v0.3.2:
  * A LogValuer being used for an attribute match would fail to match
    because slogassert wouldn't resolve the value, but try to match
    the value against the slog.Value. If the LogValuer did something
    like change types or something, it would never match. Now values
    that implement LogValuer can be used directly in attribute
    matches. See the `ValueAsString` in `assertions_test.go` if you
    don't know what I mean.
* v0.3.1:
  * Annotate the internal .Assert\* functions as `t.Helper()`s to
    improve error messages when an assert fails.
* v0.3.0 more **BREAKING CHANGES**:
  * Significant API rewrite. This:
    * Exposes Assert directly, for functional-matching based
      assertions, which enables a lot of more complicated scenarios.
    * Now that the general power is available to end users, the
      package doesn't need to offer every marginal match method, so
      I'm removing the built-in message + level assertion. It's easy
      now to implement yourself if it's useful, and I haven't used it
      yet in my own code so I question its general utility, sandwiched
      between the very useful "please assert this message" (useful
      because it is agnostic about levels) and "please assert this
      exact match" methods.
    * If you wrap another slog Handler with this, we need to properly
      pass WithGroup and WithAttrs down to that wrapped handler too.
    
      I like slog overall but I will say writing a correct Handler
      wrapper is distinctly nontrivial.
* v0.2.0 **BREAKING CHANGE**:
  * If test code panics, and a `*testing.T.Cleanup` function itself
    has some sort of `.Fatal` call, the result is that the panic is
    eaten. Due to [Golang issue
    #49929](https://github.com/golang/go/issues/49929), there is no
    way to detect this in the cleanup function because the cleanup
    function is run in the wrong place to detect the panic and the
    `t.Failed()` method will return false.
    
    Practice has revealed that this is _way_ too confusing, so I'm
    removing the automatic cleanup function. Therefore, **any
    slogassert.New calls need to have a `defer handler.AssertEmpty()`
    manually added to retain the original behavior**. The
    `AssertEmpty` function has code added to see whether it is in the
    middle of a panic, and if so, it will not fatally error. (This is
    in practice good anyhow because panics frequently result in the
    log messages being logged but the assertions not running, so it
    frequently produced spurious and confusing messages anyhow.)
    
    This still mangles the panic a bit, unfortunately, but the
    necessary data is still there.
  * I was doing a lot of work on my work laptop and had my work email
    rather than my personal email, but my signing key is my personal
    email. This commit signs the top of the repository correctly.
    
    As I mention in the version numbering section, that's what
    matters; a signed commit at the top of a repository is essentially
    signing the whole thing, not just that commit. So it is not
    necessary to rewrite the whole repo to fix all the previous commits.
* v0.1.3:
  * Add a return value to the \*Some\* methods that return how many
    messages they consumed as being asserted.
* v0.1.2:
  * Allow use of ints to compare against Int64, Float64, and Uint64.
  
    This resolves an issue where you write an AssertPrecise and use a
    bare int in the source code, which the Go compiler decides is an
    `int`, and then that didn't match any of the numeric types. This
    adds the relevant clauses to the matchers.
* v0.1.1:
  * No code changes, just screwed up tagging.
* v0.1.0:
  * Fixed a major error: Somehow I completely overlooked adding the
    params on a sublogger added with `.With` to the resulting log
    messages. I guess I thought slog would do that for me. And this is
    why v0.0.9 was only a "release candidate".
    
    That said, I am advancing this up the semver chain to a release
    candidate. 
* v0.0.9:
  * Fix a locking issue in `Unasserted`, which should make this all
    completely thread-safe.
  * I consider this a v1.0.0. release candidate.
* v0.0.8:
  * Make Unasserted return a fully independent copy of the LogMessage
    so the user can't accidentally corrupt it.
* v0.0.7:
  * Add Unasserted call. I've resisted this because it's kind of a
    trap, but sometimes you just need it.
* v0.0.6:
  * *BREAKING RELEASE*: The ability to wrap a handler is added.
    This is useful for things like recording all the logs in a test
    into a wrapped handler, then if the test fails, printing out the
    logs as part of the test failure message. This changes the
    signature on `New` and `NewWithoutCleanup`.
    
    To recover previous behavior, add a `nil` on the end of all such calls.
* v0.0.5:
  * Add a NullHandler and NullLogger. This is not 100% on point for
    the package, but pretty useful for when you need a logger but
    don't need the logs, which comes up in testing a lot.
* v0.0.4:
  * Handlers are responsible for resolving LogValuer values.
    * This is why you don't promise no bugs.
* v0.0.3:
  * Bugs! Bugs everywhere! Fewer now, but still no promises.
* v0.0.2
  * README fixup.
* v0.0.1
  * Initial release.

