# slogassert


    import "github.com/thejerf/slogassert"

[![GoDoc](https://pkg.go.dev/badge/github.com/thejerf/slogassert)](https://pkg.go.dev/github.com/thejerf/slogassert)

## About Slogassert

slogassert implements a _testing handler_ for slog.

All important observable results of code should be tested. While
commonly ignored by people writing tests, log messages are often
important things to test for, for several reasons:

 * Because they are not tested for, it is suprisingly easy for them to
   break without people realizing.
 * Log messages can have security impact. Any log messages that may be
   used to reconstruct a security incident should be tested.
 * Log messages often implicitly test that some other code paths were
   covered, or provide a mechanism for asserting coverage of code
   paths that would otherwise be unasserted.

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

## Release Status

`slogassert` is **super alpha**. I reserve the right to
change _everything_ as I continue to work out how to correctly write
slog handlers.

As you read this, I'm basically committing stuff with enough version
information that I can pull it into some repos I'll be working with
to test this out.

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

* v0.0.3:
  * Bugs! Bugs everywhere! Fewer now, but still no promises.
* v0.0.2
  * README fixup.
* v0.0.1
  * Initial release.

