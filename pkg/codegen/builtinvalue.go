package codegen

import (
	"fmt"
	"sort"
	"strings"
)

// builtinValue says how to reach a builtin's runtime function when the builtin
// is named rather than called.
//
// call is the Go expression that invokes it, and arity the number of arguments
// it takes — below zero for one that takes what it is given. A call has the Go
// compiler to check the count, since genCall splices the arguments straight
// into the runtime call; a value called at runtime has only what is recorded
// here, so the arity has to be carried alongside the name.
type builtinValue struct {
	call  string
	arity int
}

// seedCall stands where a runtime function would, for the one builtin that has
// none: a call to seed outside a fuzz run answers with catnap.
const seedCall = "__seed"

// builtinValues covers every name the checker accepts. A name missing from it
// would type-check, compile, and then be undefined Go, so
// TestEveryAcceptedBuiltinCanBeNamedAsAValue holds the two in step.
var builtinValues = map[string]builtinValue{
	"nya":   {"meow.Nya", -1},
	"hiss":  {"meow.Hiss", -1},
	"scram": {"meow.Scram", -1},
	// seed marks a fuzz corpus entry and is read off the AST before a body is
	// generated, so outside a fuzz run it is the no-op genCall also emits.
	"seed":       {seedCall, -1},
	"gag":        {"meow.Gag", 1},
	"is_furball": {"meow.IsFurball", 1},
	"len":        {"meow.Len", 1},
	"head":       {"meow.Head", 1},
	"tail":       {"meow.Tail", 1},
	"append":     {"meow.Append", 2},
	"lick":       {"meow.Lick", 2},
	"picky":      {"meow.Picky", 2},
	"curl":       {"meow.Curl", 3},
	"to_int":     {"meow.ToInt", 1},
	"to_float":   {"meow.ToFloat", 1},
	"to_string":  {"meow.ToString", 1},
	"to_bytes":   {"meow.ToBytes", 1},
	"to_runes":   {"meow.ToRunes", 1},
	"whiff":      {"meow.Whiff", 2},
	"track":      {"meow.Track", 2},
	"shred":      {"meow.Shred", 2},
	"tangle":     {"meow.Tangle", 2},
	"nibble":     {"meow.Nibble", 3},
	"upper":      {"meow.Upper", 1},
	"lower":      {"meow.Lower", 1},
	"trim":       {"meow.Trim", 1},
	"replace":    {"meow.Replace", 3},
	"pad":        {"meow.Pad", 2},
	"sort":       {"meow.Sort", 1},
	"reverse":    {"meow.Reverse", 1},
	"round":      {"meow.Round", 2},
	"judge":      {"meow_testing.Judge", -1},
	"expect":     {"meow_testing.Expect", -1},
	"refuse":     {"meow_testing.Refuse", -1},
}

// HasBuiltinValue reports whether a builtin of that name can be named as a
// value.
func HasBuiltinValue(name string) bool {
	_, ok := builtinValues[name]
	return ok
}

// BuiltinValueNames lists every builtin that can be named as a value, in order.
func BuiltinValueNames() []string {
	names := make([]string, 0, len(builtinValues))
	for name := range builtinValues {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// genBuiltinValue emits a builtin named rather than called, as the value it
// has to be for anything holding a function to take it.
func (g *Generator) genBuiltinValue(name string) (string, bool) {
	bv, ok := builtinValues[name]
	if !ok {
		return "", false
	}
	if strings.HasPrefix(bv.call, "meow_testing.") {
		g.ensureImport("testing")
	}

	var body string
	switch {
	case bv.call == seedCall:
		body = "meow.NewNil()"
	case bv.arity < 0:
		body = fmt.Sprintf("%s(__a...)", bv.call)
	default:
		args := make([]string, bv.arity)
		for i := range args {
			args[i] = fmt.Sprintf("__a[%d]", i)
		}
		body = fmt.Sprintf("%s(%s)", bv.call, strings.Join(args, ", "))
	}
	return fmt.Sprintf("meow.BuiltinFunc(%q, %d, func(__a ...meow.Value) meow.Value { return %s })",
		name, bv.arity, body), true
}
