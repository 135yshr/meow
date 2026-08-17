// Package json turns JSON text into Meow values and back.
//
// It exists because a program that talks to an HTTP API can otherwise only
// match its replies as text. Deciding "did the record arrive" by searching a
// response body for a word is fragile in a way that matters: the word may
// appear somewhere else in the payload, and the check quietly answers yes.
package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/135yshr/meow/runtime/meowrt"
)

// furball wraps an error as a Meow Furball value with the "Hiss! ... nya~" form.
func furball(format string, args ...any) meowrt.Value {
	return &meowrt.Furball{Message: fmt.Sprintf("Hiss! "+format+", nya~", args...)}
}

// maxDepth bounds how deeply nested a document may be. Beyond it the
// conversion would recurse as deeply as the input asks, which a hostile or
// merely broken source could push until the stack runs out.
const maxDepth = 200

// Unravel reads JSON text and returns it as Meow values: an object becomes a
// basket, an array a litter, and null catnap.
//
// Malformed text is a Furball rather than a panic, so a reply that is not JSON
// at all — an HTML error page from a proxy, say — can be recovered from with
// `~>` like any other failure.
func Unravel(args ...meowrt.Value) meowrt.Value {
	if len(args) != 1 {
		return furball("unravel expects 1 argument, got %d", len(args))
	}
	if f, ok := args[0].(*meowrt.Furball); ok {
		return f
	}
	s, ok := args[0].(*meowrt.String)
	if !ok {
		return furball("unravel expects a String, got %s", typeName(args[0]))
	}
	// UseNumber keeps each number as the text that was written. Decoding into
	// `any` would route every one through float64, which cannot hold an int64
	// exactly past 2^53 — an id of 9007199254740993 would come back as
	// ...992, and one of MaxInt64 would come back negative. Silently wrong ids
	// are worse than no ids.
	dec := json.NewDecoder(bytes.NewReader([]byte(s.Val)))
	dec.UseNumber()
	var decoded any
	if err := dec.Decode(&decoded); err != nil {
		return furball("cannot read that as JSON: %s", err)
	}
	// Decode stops at the end of the first value, so trailing text would go
	// unnoticed; `{"a":1} nonsense` is not a document.
	if dec.More() {
		return furball("cannot read that as JSON: unexpected text after the value")
	}
	return toMeow(decoded, 0)
}

// Wind renders a Meow value as JSON text.
func Wind(args ...meowrt.Value) meowrt.Value {
	if len(args) != 1 {
		return furball("wind expects 1 argument, got %d", len(args))
	}
	if f, ok := args[0].(*meowrt.Furball); ok {
		return f
	}
	plain, fb := toGo(args[0], 0)
	if fb != nil {
		return fb
	}
	out, err := json.Marshal(plain)
	if err != nil {
		return furball("cannot write that as JSON: %s", err)
	}
	return meowrt.NewString(string(out))
}

// typeName names a value for an error message, allowing for a missing one.
func typeName(v meowrt.Value) string {
	if v == nil {
		return "catnap"
	}
	return v.Type()
}

// toMeow converts a decoded JSON document into Meow values.
//
// JSON has one number type and Meow has two, so a number that is a whole value
// becomes an Int and anything else a Float. Without that, an id or a count read
// back from an API would arrive as 1.0 and print that way.
func toMeow(v any, depth int) meowrt.Value {
	if depth > maxDepth {
		return furball("JSON nested deeper than %d levels", maxDepth)
	}
	switch v := v.(type) {
	case nil:
		return meowrt.NewNil()
	case bool:
		return meowrt.NewBool(v)
	case string:
		return meowrt.NewString(v)
	case json.Number:
		// A whole value becomes an Int, so an id or a count does not come back
		// reading 42.0. Parsing the text rather than a float keeps every int64
		// exact.
		if n, err := strconv.ParseInt(v.String(), 10, 64); err == nil {
			return meowrt.NewInt(n)
		}
		f, err := v.Float64()
		if err != nil {
			return furball("cannot read %s as a number", v.String())
		}
		return meowrt.NewFloat(f)
	case []any:
		items := make([]meowrt.Value, 0, len(v))
		for _, item := range v {
			converted := toMeow(item, depth+1)
			if f, ok := converted.(*meowrt.Furball); ok {
				return f
			}
			items = append(items, converted)
		}
		return meowrt.NewList(items...)
	case map[string]any:
		items := make(map[string]meowrt.Value, len(v))
		for key, item := range v {
			converted := toMeow(item, depth+1)
			if f, ok := converted.(*meowrt.Furball); ok {
				return f
			}
			items[key] = converted
		}
		return meowrt.NewMap(items)
	default:
		return furball("cannot read %T out of JSON", v)
	}
}

// toGo converts a Meow value into something encoding/json can render.
func toGo(v meowrt.Value, depth int) (any, meowrt.Value) {
	if depth > maxDepth {
		return nil, furball("value nested deeper than %d levels", maxDepth)
	}
	switch v := v.(type) {
	case nil:
		return nil, nil
	case *meowrt.NilValue:
		return nil, nil
	case *meowrt.Bool:
		return v.Val, nil
	case *meowrt.Int:
		return v.Val, nil
	case *meowrt.Float:
		return v.Val, nil
	case *meowrt.String:
		return v.Val, nil
	case *meowrt.List:
		items := make([]any, 0, len(v.Items))
		for _, item := range v.Items {
			converted, fb := toGo(item, depth+1)
			if fb != nil {
				return nil, fb
			}
			items = append(items, converted)
		}
		return items, nil
	case *meowrt.Map:
		items := make(map[string]any, len(v.Items))
		for key, item := range v.Items {
			converted, fb := toGo(item, depth+1)
			if fb != nil {
				return nil, fb
			}
			items[key] = converted
		}
		return items, nil
	default:
		// A Furball, a kitty, a function — nothing JSON has a shape for. Saying
		// so beats writing something that reads back as a different value.
		return nil, furball("cannot write a %s as JSON", typeName(v))
	}
}
