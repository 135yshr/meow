package codegen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/135yshr/meow/pkg/builtins"
)

// seedCall stands where a runtime function would, for the one builtin that has
// none: a call to seed outside a fuzz run answers with catnap.
const seedCall = "__seed"

// builtinCalls says how to reach a builtin's runtime function when the builtin
// is named rather than called: the Go expression that invokes it.
//
// How many arguments it takes is not written here. A call has the Go compiler
// to check the count, since genCall splices the arguments straight into the
// runtime call; a value called at run time has only what it was built with, and
// a program that writes the wrong count is now told so by the checker. Both
// read the one arity table in `pkg/builtins`, so there is no second number here
// to fall out of step with it.
//
// The names still have to cover every name the checker accepts: one missing
// would type-check, compile, and then be undefined Go, so
// TestEveryAcceptedBuiltinCanBeNamedAsAValue holds the two in step.
var builtinCalls = map[string]string{
	"nya":   "meow.Nya",
	"hiss":  "meow.Hiss",
	"scram": "meow.Scram",
	// seed marks a fuzz corpus entry and is read off the AST before a body is
	// generated, so outside a fuzz run it is the no-op genCall also emits.
	"seed":       seedCall,
	"gag":        "meow.Gag",
	"is_furball": "meow.IsFurball",
	"len":        "meow.Len",
	"head":       "meow.Head",
	"tail":       "meow.Tail",
	"append":     "meow.Append",
	"lick":       "meow.Lick",
	"picky":      "meow.Picky",
	"curl":       "meow.Curl",
	"to_int":     "meow.ToInt",
	"to_float":   "meow.ToFloat",
	"to_string":  "meow.ToString",
	"to_bytes":   "meow.ToBytes",
	"to_runes":   "meow.ToRunes",
	"whiff":      "meow.Whiff",
	"track":      "meow.Track",
	"shred":      "meow.Shred",
	"tangle":     "meow.Tangle",
	"nibble":     "meow.Nibble",
	"upper":      "meow.Upper",
	"lower":      "meow.Lower",
	"trim":       "meow.Trim",
	"replace":    "meow.Replace",
	"pad":        "meow.Pad",
	"sort":       "meow.Sort",
	"reverse":    "meow.Reverse",
	"round":      "meow.Round",
	"judge":      "meow_testing.Judge",
	"expect":     "meow_testing.Expect",
	"refuse":     "meow_testing.Refuse",
}

// HasBuiltinValue reports whether a builtin of that name can be named as a
// value.
func HasBuiltinValue(name string) bool {
	_, ok := builtinCalls[name]
	return ok
}

// BuiltinValueNames lists every builtin that can be named as a value, in order.
func BuiltinValueNames() []string {
	names := make([]string, 0, len(builtinCalls))
	for name := range builtinCalls {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// genBuiltinValue emits a builtin named rather than called, as the value it
// has to be for anything holding a function to take it.
func (g *Generator) genBuiltinValue(name string) (string, bool) {
	call, ok := builtinCalls[name]
	if !ok {
		return "", false
	}
	arity, ok := builtins.Arity(name)
	if !ok {
		return "", false
	}
	if strings.HasPrefix(call, "meow_testing.") {
		g.ensureImport("testing")
	}

	var body string
	switch {
	case call == seedCall:
		body = "meow.NewNil()"
	case arity == builtins.Variadic:
		body = fmt.Sprintf("%s(__a...)", call)
	default:
		args := make([]string, arity)
		for i := range args {
			args[i] = fmt.Sprintf("__a[%d]", i)
		}
		body = fmt.Sprintf("%s(%s)", call, strings.Join(args, ", "))
	}
	return fmt.Sprintf("meow.BuiltinFunc(%q, %d, func(__a ...meow.Value) meow.Value { return %s })",
		name, arity, body), true
}
