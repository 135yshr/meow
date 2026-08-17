package meowrt

import (
	"math"
	"sort"
	"strings"
)

// Upper returns s with every letter in upper case.
func Upper(s Value) Value {
	str, f := requireString("upper", "argument", s)
	if f != nil {
		return f
	}
	return NewString(strings.ToUpper(str))
}

// Lower returns s with every letter in lower case.
func Lower(s Value) Value {
	str, f := requireString("lower", "argument", s)
	if f != nil {
		return f
	}
	return NewString(strings.ToLower(str))
}

// Trim returns s without leading or trailing whitespace.
//
// This is what a line read from a file or an environment variable tends to
// carry, and comparing against it without trimming quietly answers no.
func Trim(s Value) Value {
	str, f := requireString("trim", "argument", s)
	if f != nil {
		return f
	}
	return NewString(strings.TrimSpace(str))
}

// Replace returns s with every occurrence of search replaced by replacement.
//
// Every occurrence rather than the first, because replacing only the first is
// the rarer intent and the one a program can still express by other means.
func Replace(s, search, replacement Value) Value {
	str, f := requireString("replace", "argument", s)
	if f != nil {
		return f
	}
	from, f := requireString("replace", "search string", search)
	if f != nil {
		return f
	}
	to, f := requireString("replace", "replacement", replacement)
	if f != nil {
		return f
	}
	// An empty search string would have strings.ReplaceAll insert the
	// replacement between every character, which is never what was meant.
	if from == "" {
		return NewFurball("Hiss! replace needs something to search for, nya~")
	}
	return NewString(strings.ReplaceAll(str, from, to))
}

// maxPad bounds how wide a padded string may be asked to become, so a width
// read from somewhere unchecked cannot ask for a string the machine has no
// room for.
const maxPad = 1 << 20

// Pad returns s widened to width characters with spaces.
//
// A positive width pads on the right, lining a column up on its left edge; a
// negative one pads on the left, which is what a column of numbers wants. A
// string already that wide is returned unchanged rather than cut, since losing
// text to make a table line up is the worse trade.
func Pad(s, width Value) Value {
	str, f := requireString("pad", "argument", s)
	if f != nil {
		return f
	}
	n, fb := TryAsInt(width)
	if fb != nil {
		return fb
	}
	right := n >= 0
	if n < 0 {
		n = -n
	}
	if n > maxPad {
		return NewFurball("Hiss! pad expects a width of at most %d, got %d, nya~", maxPad, n)
	}
	// Counted in characters, as nibble and track are, so a padded column lines
	// up for text that is not all ASCII.
	missing := int(n) - len([]rune(str))
	if missing <= 0 {
		return NewString(str)
	}
	spaces := strings.Repeat(" ", missing)
	if right {
		return NewString(str + spaces)
	}
	return NewString(spaces + str)
}

// Reverse returns a litter with its elements in the opposite order.
func Reverse(v Value) Value {
	if f, ok := v.(*Furball); ok {
		return f
	}
	list, ok := v.(*List)
	if !ok {
		return NewFurball("Hiss! reverse requires a litter, got %s, nya~", typeNameOf(v))
	}
	items := make([]Value, len(list.Items))
	for i, item := range list.Items {
		items[len(list.Items)-1-i] = item
	}
	return NewList(items...)
}

// Sort returns a litter with its elements in ascending order.
//
// A litter holding more than one kind of value is a Furball rather than an
// arbitrary order: there is no answer to "is 1 before "a"", and inventing one
// would put a program's output at the mercy of which happened to come first.
func Sort(v Value) Value {
	if f, ok := v.(*Furball); ok {
		return f
	}
	list, ok := v.(*List)
	if !ok {
		return NewFurball("Hiss! sort requires a litter, got %s, nya~", typeNameOf(v))
	}
	items := make([]Value, len(list.Items))
	copy(items, list.Items)
	if len(items) < 2 {
		return NewList(items...)
	}
	kind, fb := sortableKind(items)
	if fb != nil {
		return fb
	}
	switch kind {
	case "String":
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].(*String).Val < items[j].(*String).Val
		})
	default:
		sort.SliceStable(items, func(i, j int) bool {
			return numberOf(items[i]) < numberOf(items[j])
		})
	}
	return NewList(items...)
}

// sortableKind reports the one kind of value a litter holds, so mixtures and
// values with no order can be refused before any of them is compared.
func sortableKind(items []Value) (string, *Furball) {
	kind := ""
	for _, item := range items {
		var this string
		switch item.(type) {
		case *String:
			this = "String"
		case *Int, *Float, *Byte:
			this = "number"
		default:
			return "", NewFurball("Hiss! sort cannot order a %s, nya~", typeNameOf(item))
		}
		if kind == "" {
			kind = this
			continue
		}
		if kind != this {
			return "", NewFurball("Hiss! sort cannot order a litter holding both %s and %s, nya~", kind, this)
		}
	}
	return kind, nil
}

// numberOf reads any of the numeric kinds as a float for comparison.
func numberOf(v Value) float64 {
	switch v := v.(type) {
	case *Int:
		return float64(v.Val)
	case *Byte:
		return float64(v.Val)
	case *Float:
		return v.Val
	}
	return 0
}

// maxRoundPlaces bounds the decimal places Round accepts. Beyond about this
// many a float64 has no digits left to give, and the scaling factor overflows.
const maxRoundPlaces = 15

// Round returns x rounded to the given number of decimal places.
//
// Rounding is half away from zero, the arithmetic convention, rather than
// Go's half-to-even: a program printing 2.5 as 3 is what a reader expects.
func Round(x, places Value) Value {
	if f, ok := x.(*Furball); ok {
		return f
	}
	n, fb := TryAsInt(places)
	if fb != nil {
		return fb
	}
	if n < 0 || n > maxRoundPlaces {
		return NewFurball("Hiss! round expects 0 to %d decimal places, got %d, nya~", maxRoundPlaces, n)
	}
	var val float64
	switch x := x.(type) {
	case *Float:
		val = x.Val
	case *Int:
		// Already whole; rounding it changes nothing, and answering with the
		// Int keeps it printing as one.
		return x
	case *Byte:
		return NewInt(int64(x.Val))
	default:
		return NewFurball("Hiss! round requires a number, got %s, nya~", typeNameOf(x))
	}
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return NewFloat(val)
	}
	factor := math.Pow(10, float64(n))
	return NewFloat(math.Round(val*factor) / factor)
}

// typeNameOf names a value for an error message, allowing for a missing one.
func typeNameOf(v Value) string {
	if v == nil {
		return "catnap"
	}
	return v.Type()
}
