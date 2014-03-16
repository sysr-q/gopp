gopp - Go Pre-Processor
=======================

Why? I honestly could not tell you.

Go has something like this built in - compiler build flags and tags. They're
definitely a much more sane solution, so if at all possible, you should be
using those instead of something like gopp.

[Stack Overflow](http://stackoverflow.com/a/13889804) has more information, if
you're interested in that topic.

A lot of the key ideas and inspiration was taken from _bytbox_'s [goprep](https://github.com/bytbox/goprep)
project, which hasn't been updated in two years. Both projects are MIT Licensed.

-----

If, for some strange reason, you're still interested in gopp, there are a few
features (mainly modeled after the _c preprocessor_), which you can use:

_NB!_ As a general rule, gopp rules are _comments, prefixed with `//gopp:`!_
e.g., `//gopp:ifdef DEBUG`

| command   | description
|:----------|:-----------
| `ifdef`   | _ifdef_ allows you to test whether or not a certain macro is defined at the time of interpretation.
| `ifndef`  | _ifndef_ is the opposite of ifdef, only stepping in if a certain macro is _not_ defined.
| `else`    | _else_ works like an else block in a C-like programming language.
| `endif`   | _endif_ works like a closing bracket on an else block in a C-like programming language.
| `define`  | _define_ allows you to define arbitrary macros on the fly; e.g. `//gopp:define DEBUG true`
| `undef`   | _undef_ allows you to revoke currently-defined macros.

`if` and `elseif` are not supported: they'd make things needlessly complex, and
are more work than I'd like to put in - if you want something like this, you
should probably use Go's compiler flags/tags instead.

_Note:_ Currently, `define`s are _global_ - that is, they're not scoped to the
file currently being processed. This may change at a later date, be warned.

gopp
----

This is the main library for using the preprocessor. The import path is `github.com/sysr-q/gopp/gopp`.

You can view the [godocs](https://godoc.org/github.com/sysr-q/gopp/gopp) if
you're interested in using it.  
[![GoDoc](https://godoc.org/github.com/sysr-q/gopp/gopp?status.png)](https://godoc.org/github.com/sysr-q/gopp/gopp)

gppc
----

_gppc_ is a command line client that allows you to easily process Go code with
gopp, in a similar vein to the unix tool, `cpp`.

_NB!_ gppc is still _very_ much in development, and as such you should be weary
of anything it spits out. Values given with `-D` are added _as is_! That means
if you pass `-D VERSION=0.1`, everywhere `VERSION` is found, the literal `0.1`
will be subbed in (yes, a float!). This can cause hell if you're not careful.

To install: `go install github.com/sysr-q/gopp/gppc`, then check `gppc` and
`gppc prep --help`.
