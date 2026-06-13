package meowrt

import (
	"fmt"
	"os"
)

// NewFurball constructs a Furball value with the "Hiss! ... nya~" wrapping.
// If the message already starts with "Hiss!", it is left as-is to allow
// nested wrapping without redundancy.
func NewFurball(format string, args ...any) *Furball {
	msg := fmt.Sprintf(format, args...)
	return &Furball{Message: msg}
}

// AsFurball returns the value as *Furball if it is an UNHANDLED Furball,
// along with true. This is the propagation gate: handled Furballs (those
// already surfaced to user code via gag/~>) return false so they don't
// re-propagate. Use rawAsFurball or a direct type assertion to inspect any
// Furball regardless of handled state.
func AsFurball(v Value) (*Furball, bool) {
	f, ok := v.(*Furball)
	if !ok || f.Handled {
		return nil, false
	}
	return f, true
}

// Recover implements the `~>` operator without panic/recover.
// If left is a Furball, it returns fallback (called with the Furball when
// fallback is a Func, otherwise returned as-is). Otherwise left is returned.
func Recover(left, fallback Value) Value {
	if f, ok := left.(*Furball); ok {
		if fn, fok := fallback.(*Func); fok {
			return fn.Call(f)
		}
		return fallback
	}
	return left
}

// ExitOnFurball prints v's message to stderr and exits with code 1 if v is
// a Furball. Used by generated main() to surface unhandled top-level errors,
// replacing the previous panic-based termination.
func ExitOnFurball(v Value) {
	if f, ok := v.(*Furball); ok {
		fmt.Fprintln(os.Stderr, f.Message)
		os.Exit(1)
	}
}

// RunMain invokes fn (the generated __meow_main) and surfaces any failure
// — a returned Furball OR an internal panic from typed As*/hiss paths — as
// a clean stderr message + exit 1, without exposing Go's runtime traceback.
func RunMain(fn func() Value) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, r)
			os.Exit(1)
		}
	}()
	ExitOnFurball(fn())
}
