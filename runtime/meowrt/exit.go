package meowrt

import "os"

// maxExitCode is the largest status a process can report. A shell sees the low
// eight bits of whatever is passed, so 256 would arrive as 0 — a program asking
// to fail would be read as having succeeded. Anything outside the range is
// refused rather than quietly wrapped.
const maxExitCode = 255

// exit is os.Exit unless a test replaces it, since a test that really exited
// would take the test binary with it.
var exit = os.Exit

// Scram ends the program with the given status.
//
// A status is how a program tells the thing that started it — a shell, cron, a
// CI step — whether its answer was yes or no. Without one, a check that found
// an endpoint down could only say so in words nothing downstream reads.
//
// Called with no argument it reports success. Variadic so that a wrong argument
// count is a Furball, as every other error here is, rather than a Go compile
// error out of generated code.
func Scram(args ...Value) Value {
	code, fb := ScramCode(args...)
	if fb != nil {
		return fb
	}
	exit(code)
	// Reached only when a test has replaced exit.
	return NewNil()
}

// ScramOrHiss ends the program, or raises the reason the status was refused.
//
// It is what a fully typed function calls, because such a function returns a
// native Go type and has no way to hand back a Furball. Raising it is the same
// bridge hiss uses there: gag's deferred recover turns it back into a Furball
// at the boundary. Called as a bare statement the returned Furball would simply
// be dropped, and a refused status would vanish.
func ScramOrHiss(args ...Value) Value {
	v := Scram(args...)
	if f, ok := v.(*Furball); ok {
		panic(f.String())
	}
	return v
}

// ScramSignal is raised by the playground interpreter, which has no process to
// end, to unwind to the top of the run.
//
// It lives here beside Scram so that Gag can tell it apart from a failure and
// let it past: a program that asked to end is not an error to be caught, and a
// compiled program could not catch os.Exit either.
type ScramSignal struct {
	Code int
}

// ScramCode reads the status Scram was given, reporting a Furball if it is not
// one a process can report.
//
// It is separate from Scram so that the playground interpreter, which has no
// process to end, refuses exactly the same arguments as a compiled program
// rather than growing its own idea of what a status may be.
func ScramCode(args ...Value) (int, Value) {
	if len(args) > 1 {
		return 0, NewFurball("Hiss! scram expects 0 or 1 arguments, got %d, nya~", len(args))
	}
	code := int64(0)
	if len(args) == 1 {
		n, fb := TryAsInt(args[0])
		if fb != nil {
			return 0, fb
		}
		code = n
	}
	if code < 0 || code > maxExitCode {
		return 0, NewFurball("Hiss! scram expects a status of 0 to %d, got %d, nya~", maxExitCode, code)
	}
	return int(code), nil
}
