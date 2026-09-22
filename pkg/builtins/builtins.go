// Package builtins names the functions a Meow program may call without a
// `nab`, and says how many arguments each of them takes.
//
// The name and the count are written down once, here, because three components
// need both and each of them used to keep its own copy: `pkg/checker` decides
// which names a program may write, `pkg/codegen` turns a name into a call on a
// `runtime/*` function, and `pkg/interpreter` runs one for the playground. A
// name in one table and not another type-checks and then dies at run time —
// that is what #140 and #141 were — and an arity in one and not another is the
// same failure a step later, which is what #155 was.
package builtins

import (
	"maps"
	"slices"
)

// Variadic is the arity of a builtin that takes what it is given and decides
// for itself whether that was enough. `nya` prints however many values it is
// handed; `judge` takes a condition and an optional message.
const Variadic = -1

// arities is the one table. Every count below is checked against the real Go
// signature by `testdata/builtin_value_arity`, which names each builtin as a
// value and lets `go build` disagree if a number here is wrong.
var arities = map[string]int{
	"nya":   Variadic,
	"hiss":  Variadic,
	"scram": Variadic,
	// seed marks one entry of a fuzz corpus. `meow test -fuzz` reads the calls
	// off the AST before a body is generated and everywhere else compiles one
	// to catnap, so nothing here ever counts its arguments.
	"seed": Variadic,
	// The assertions take a condition (or a pair of values) and an optional
	// message, so the count is theirs to judge rather than ours.
	"judge":  Variadic,
	"expect": Variadic,
	"refuse": Variadic,

	"gag":        1,
	"is_furball": 1,
	"len":        1,
	"head":       1,
	"tail":       1,
	"to_int":     1,
	"to_float":   1,
	"to_string":  1,
	"to_bytes":   1,
	"to_runes":   1,
	"upper":      1,
	"lower":      1,
	"trim":       1,
	"sort":       1,
	"reverse":    1,

	"append": 2,
	"lick":   2,
	"picky":  2,
	"whiff":  2,
	"track":  2,
	"shred":  2,
	"tangle": 2,
	"pad":    2,
	"round":  2,

	"curl":    3,
	"nibble":  3,
	"replace": 3,
}

// Known reports whether name is a builtin.
func Known(name string) bool {
	_, ok := arities[name]
	return ok
}

// Arity answers how many arguments the builtin of that name takes, and whether
// there is such a builtin at all. A [Variadic] count means the builtin decides
// for itself.
func Arity(name string) (int, bool) {
	n, ok := arities[name]
	return n, ok
}

// Names lists every builtin, in order.
func Names() []string {
	return slices.Sorted(maps.Keys(arities))
}

// Wrong reports whether a call of the builtin `name` with `got` arguments is
// the wrong number, and what the right number is. A name that is not a builtin
// and a builtin that counts for itself are both reported as fine, since
// neither is this table's to refuse.
func Wrong(name string, got int) (want int, wrong bool) {
	n, ok := arities[name]
	if !ok || n == Variadic || n == got {
		return n, false
	}
	return n, true
}
