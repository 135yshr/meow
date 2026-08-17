// Package env exposes the process environment to Meow programs.
//
// It exists so that values which should not be written down — tokens,
// passwords, endpoints that differ per deployment — can reach a program without
// being staged in a plain-text file first.
package env

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/135yshr/meow/runtime/meowrt"
)

// furball wraps an error as a Meow Furball value with the "Hiss! ... nya~" form.
func furball(format string, args ...any) meowrt.Value {
	return &meowrt.Furball{Message: fmt.Sprintf("Hiss! "+format+", nya~", args...)}
}

// expectName extracts the variable name from a Value.
func expectName(fn string, name meowrt.Value) (string, meowrt.Value) {
	if f, ok := name.(*meowrt.Furball); ok {
		return "", f
	}
	n, ok := name.(*meowrt.String)
	if !ok {
		typeName := "catnap"
		if name != nil {
			typeName = name.Type()
		}
		return "", furball("%s expects a String name, got %s", fn, typeName)
	}
	if n.Val == "" {
		return "", furball("%s expects a non-empty name", fn)
	}
	return n.Val, nil
}

// Hunt returns the value of the named environment variable.
//
// An unset variable reads as catnap, so a caller can tell "not set" apart from
// "set to the empty string" — which matters when the variable is the switch
// that decides whether to run at all. An optional second argument is returned
// in place of catnap when the variable is unset.
func Hunt(args ...meowrt.Value) meowrt.Value {
	if len(args) == 0 || len(args) > 2 {
		return furball("hunt expects 1 or 2 arguments, got %d", len(args))
	}
	name, fb := expectName("hunt", args[0])
	if fb != nil {
		return fb
	}
	if v, ok := os.LookupEnv(name); ok {
		return meowrt.NewString(v)
	}
	if len(args) == 2 {
		return args[1]
	}
	return meowrt.NewNil()
}

// Sniffed reports whether the named environment variable is set, including when
// it is set to the empty string.
//
// It is variadic so that a wrong argument count is reported as a Furball, the
// way every other error in this package is; a fixed arity would instead surface
// as a Go compile error from generated code.
func Sniffed(args ...meowrt.Value) meowrt.Value {
	if len(args) != 1 {
		return furball("sniffed expects 1 argument, got %d", len(args))
	}
	n, fb := expectName("sniffed", args[0])
	if fb != nil {
		return fb
	}
	_, ok := os.LookupEnv(n)
	return meowrt.NewBool(ok)
}

// Haul returns the arguments the program was started with, as a List of
// Strings.
//
// The program's own name is left out: a program wants what it was asked to do,
// and the path it happens to be installed at is not part of that. A program
// started with no arguments gets an empty litter, so len tells a caller whether
// it was given anything without a special case for "none".
//
// Variadic for the same reason as Sniffed.
func Haul(args ...meowrt.Value) meowrt.Value {
	if len(args) != 0 {
		return furball("haul expects no arguments, got %d", len(args))
	}
	// os.Args[0] is the program's own name, and os.Args is never empty in a
	// program the Go runtime started.
	given := os.Args[1:]
	values := make([]meowrt.Value, len(given))
	for i, a := range given {
		values[i] = meowrt.NewString(a)
	}
	return meowrt.NewList(values...)
}

// Prowl returns the names of every environment variable, sorted, as a List of
// Strings. Only the names are returned: listing the values would make it far
// too easy to print a secret by accident.
//
// Variadic for the same reason as Sniffed.
func Prowl(args ...meowrt.Value) meowrt.Value {
	if len(args) != 0 {
		return furball("prowl expects no arguments, got %d", len(args))
	}
	entries := os.Environ()
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		name, _, ok := strings.Cut(e, "=")
		if !ok || name == "" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	values := make([]meowrt.Value, len(names))
	for i, n := range names {
		values[i] = meowrt.NewString(n)
	}
	return meowrt.NewList(values...)
}
