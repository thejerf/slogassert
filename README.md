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

