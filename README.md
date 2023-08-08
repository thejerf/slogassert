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

## Release Status

`slogassert` is **super alpha**. I reserve the right to
change _everything_ as I continue to work out how to correctly write
slog handlers.

In fact, as you read this, it doesn't even quite work yet.

Please do try it out and give it a test.

## Normal Usage

Normal usage looks like this:

    func TestSomething(t *testing.T) {
        // t is a *testing.T, slogtest uses this to register a
        // t.Cleanup function to assert that the log is empty
        st := slogtest.New(t, slog.LevelWarn)
        logger := slog.New(st)

        // inject the logger into your test code and run it

        // Now start asserting things:
        
If you have something like a slog handler that consumes information
from a context to decorate log messages, you may need to compose that
into the slogtest handler. The slogtest handler should be at the
"bottom" of any stack of loggers, in place of the final output
handler. You can also choose to simply leave it out, as it will
simplify testing, if you have it tested elsewhere.

## Usage Hints

I think you'll find you want the textual portion of your log messages
to be constants in the relevant packages.

## Version Numbering

This package will not go to 1.0 until `slog` does. That 1.0 will be
pointed at the final standard library version as its dependency.

Until then, this package by necessity can't be more stable than `slog`
itself. The latest version will track `slog`'s latest version. If it
happens to diverge in a way that makes this stop working, please file
an Issue.

I will be signing this repository with the ["jerf" keybase
account](https://keybase.io/jerf). If you are viewing this repository
through GitHub, you should see the commits as showing as "verified" in
the commit view.

(Bear in mind that due to the nature of how git commit signing works,
there may be runs of unverified commits; what matters is that the top
one is signed.)

## Version History

* v0.0.1
  * Initial release.

