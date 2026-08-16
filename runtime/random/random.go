// Package random exposes random values to Meow programs.
//
// Roll, Drift and Pick draw from math/rand/v2, which is fine for sampling and
// jitter. Tuft draws from crypto/rand instead, because it exists to label
// things — a marker that collides, or that an observer can predict, defeats the
// purpose of having one.
package random

import (
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"math/rand/v2"

	"github.com/135yshr/meow/runtime/meowrt"
)

// maxTuftBytes bounds Tuft so a mistyped length cannot ask for a huge string.
const maxTuftBytes = 1024

// furball wraps an error as a Meow Furball value with the "Hiss! ... nya~" form.
func furball(format string, args ...any) meowrt.Value {
	return &meowrt.Furball{Message: fmt.Sprintf("Hiss! "+format+", nya~", args...)}
}

// randomInt is swapped out in tests to make the outcome predictable.
var randomInt = rand.Int64N

// randomFloat is swapped out in tests to make the outcome predictable.
var randomFloat = rand.Float64

// randomBytes is swapped out in tests to make the outcome predictable.
var randomBytes = crand.Read

// Roll returns a random integer in [0, n).
func Roll(args ...meowrt.Value) meowrt.Value {
	if len(args) != 1 {
		return furball("roll expects 1 argument, got %d", len(args))
	}
	n, fb := meowrt.TryAsInt(args[0])
	if fb != nil {
		return fb
	}
	if n <= 0 {
		return furball("roll expects a positive bound, got %d", n)
	}
	return meowrt.NewInt(randomInt(n))
}

// Drift returns a random float in [0, 1).
func Drift(args ...meowrt.Value) meowrt.Value {
	if len(args) != 0 {
		return furball("drift expects no arguments, got %d", len(args))
	}
	return meowrt.NewFloat(randomFloat())
}

// Pick returns a random element of a litter.
func Pick(args ...meowrt.Value) meowrt.Value {
	if len(args) != 1 {
		return furball("pick expects 1 argument, got %d", len(args))
	}
	l, fb := meowrt.TryAsList(args[0])
	if fb != nil {
		return fb
	}
	if l.Len() == 0 {
		return furball("pick cannot choose from an empty litter")
	}
	return l.Get(int(randomInt(int64(l.Len()))))
}

// Tuft returns n random bytes as a lowercase hex string, so the result is
// 2n characters long. It is drawn from a cryptographic source, which makes it
// suitable for the markers and correlation IDs it is meant for.
func Tuft(args ...meowrt.Value) meowrt.Value {
	if len(args) != 1 {
		return furball("tuft expects 1 argument, got %d", len(args))
	}
	n, fb := meowrt.TryAsInt(args[0])
	if fb != nil {
		return fb
	}
	if n <= 0 {
		return furball("tuft expects a positive length, got %d", n)
	}
	if n > maxTuftBytes {
		return furball("tuft expects at most %d bytes, got %d", maxTuftBytes, n)
	}
	buf := make([]byte, n)
	if _, err := randomBytes(buf); err != nil {
		return furball("%s", err)
	}
	return meowrt.NewString(hex.EncodeToString(buf))
}
