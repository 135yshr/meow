// Package clock exposes wall-clock time and sleeping to Meow programs.
//
// Times are reported as plain integers and strings rather than as an opaque
// timestamp type, because Meow has no such type: a value that cannot be
// printed, compared or written to a file would be of little use to a program.
package clock

import (
	"fmt"
	"math"
	"time"

	"github.com/135yshr/meow/runtime/meowrt"
)

// furball wraps an error as a Meow Furball value with the "Hiss! ... nya~" form.
func furball(format string, args ...any) meowrt.Value {
	return &meowrt.Furball{Message: fmt.Sprintf("Hiss! "+format+", nya~", args...)}
}

// now is swapped out in tests. Production code always reads the wall clock.
var now = time.Now

// sleep is swapped out in tests so they need not actually wait.
var sleep = time.Sleep

// expectNoArgs reports a Furball when a no-argument function is given some.
//
// These functions are variadic so that a wrong argument count is reported as a
// Furball, the way every other error is; a fixed arity would instead surface as
// a Go compile error from generated code.
func expectNoArgs(fn string, args []meowrt.Value) meowrt.Value {
	if len(args) != 0 {
		return furball("%s expects no arguments, got %d", fn, len(args))
	}
	return nil
}

// Now returns the current time as whole seconds since the Unix epoch.
func Now(args ...meowrt.Value) meowrt.Value {
	if fb := expectNoArgs("now", args); fb != nil {
		return fb
	}
	return meowrt.NewInt(now().Unix())
}

// Nanos returns the current time as nanoseconds since the Unix epoch.
//
// Wire formats that carry timestamps — OpenTelemetry among them — ask for
// nanoseconds, which is why this is offered alongside Now rather than left to
// the caller to multiply out.
func Nanos(args ...meowrt.Value) meowrt.Value {
	if fb := expectNoArgs("nanos", args); fb != nil {
		return fb
	}
	return meowrt.NewInt(now().UnixNano())
}

// Stamp returns the current UTC time as an RFC 3339 string.
func Stamp(args ...meowrt.Value) meowrt.Value {
	if fb := expectNoArgs("stamp", args); fb != nil {
		return fb
	}
	return meowrt.NewString(now().UTC().Format(time.RFC3339))
}

// maxNapMillis is the largest delay that fits in a time.Duration, which counts
// nanoseconds. Beyond it the conversion silently wraps negative and the sleep
// returns at once — the opposite of what was asked for.
const maxNapMillis = int64(math.MaxInt64) / int64(time.Millisecond)

// Nap pauses for the given number of milliseconds.
//
// A negative duration is an error rather than a silent no-op, since it almost
// always means the caller computed the delay wrongly.
func Nap(args ...meowrt.Value) meowrt.Value {
	if len(args) != 1 {
		return furball("nap expects 1 argument, got %d", len(args))
	}
	ms, fb := meowrt.TryAsInt(args[0])
	if fb != nil {
		return fb
	}
	if ms < 0 {
		return furball("nap expects a non-negative number of milliseconds, got %d", ms)
	}
	if ms > maxNapMillis {
		return furball("nap expects at most %d milliseconds, got %d", maxNapMillis, ms)
	}
	sleep(time.Duration(ms) * time.Millisecond)
	return meowrt.NewNil()
}
