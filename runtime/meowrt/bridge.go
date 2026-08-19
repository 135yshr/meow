package meowrt

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// The bridge between Meow values and Go ones.
//
// Reaching a Go library used to mean writing a package of wrappers by hand, in
// Go, one function at a time: read the arguments out of Meow values, call the
// thing, build a Map of the result. That is the same work every time, and it is
// work a reflection of the types can do instead.
//
// What is bridged is data — numbers, text, lists, records — and what is not is
// carried whole as an Opaque. A client, a connection, a handle goes back into
// the next call untouched, so a library does not have to be understood to be
// used.

// callTimeoutForBridge bounds a call that takes a context of its own making.
const callTimeoutForBridge = 30 * time.Second

var (
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
	timeType    = reflect.TypeOf(time.Time{})
	valueType   = reflect.TypeOf((*Value)(nil)).Elem()
)

// CallGo calls a Go function with Meow values, giving back a Meow value.
//
// what names the call in anything that goes wrong, so a reader is told which
// call it was rather than which line of reflection.
func CallGo(what string, fn any, args ...Value) Value {
	ctx, cancel := context.WithTimeout(context.Background(), callTimeoutForBridge)
	defer cancel()
	return CallGoContext(ctx, what, fn, args...)
}

// CallGoContext is CallGo with a context to hand to a function that takes one.
func CallGoContext(ctx context.Context, what string, fn any, args ...Value) Value {
	rv := reflect.ValueOf(fn)
	if !rv.IsValid() || rv.Kind() != reflect.Func {
		return NewFurball("Hiss! %s is not something that can be called, nya~", what)
	}
	return callReflected(ctx, what, rv, args)
}

// CallGoMethod calls a named method on a held Go value.
//
// This is what makes a library usable without wrapping it: hold whatever it
// hands back, then call the next thing on it by name.
func CallGoMethod(ctx context.Context, recv Value, name string, args ...Value) Value {
	o, ok := AsOpaque(recv)
	if !ok {
		return NewFurball("Hiss! Cannot call %s on a %s, nya~", name, recv.Type())
	}
	if o.V == nil {
		return NewFurball("Hiss! Cannot call %s on nothing, nya~", name)
	}
	m := reflect.ValueOf(o.V).MethodByName(name)
	if !m.IsValid() {
		return NewFurball("Hiss! %s has no %s, nya~", o.String(), name)
	}
	return callReflected(ctx, name, m, args)
}

func callReflected(ctx context.Context, what string, fn reflect.Value, args []Value) Value {
	t := fn.Type()

	// A function that asks for a context is given one rather than being asked
	// for it in Meow, where there is nothing to pass.
	in := make([]reflect.Value, 0, t.NumIn())
	next := 0
	if t.NumIn() > 0 && t.In(0) == contextType {
		in = append(in, reflect.ValueOf(ctx))
		next = 1
	}

	wanted := t.NumIn() - next
	if t.IsVariadic() {
		if len(args) < wanted-1 {
			return NewFurball("Hiss! %s wants at least %d arguments, got %d, nya~", what, wanted-1, len(args))
		}
	} else if len(args) != wanted {
		return NewFurball("Hiss! %s wants %d arguments, got %d, nya~", what, wanted, len(args))
	}

	for i, arg := range args {
		pos := next + i
		var pt reflect.Type
		switch {
		case t.IsVariadic() && pos >= t.NumIn()-1:
			pt = t.In(t.NumIn() - 1).Elem()
		default:
			pt = t.In(pos)
		}
		gv, err := toGo(arg, pt)
		if err != nil {
			return NewFurball("Hiss! %s argument %d: %s, nya~", what, i+1, err)
		}
		in = append(in, gv)
	}

	out := fn.Call(in)
	return resultOf(what, out)
}

// resultOf turns what a Go function returned into one Meow value.
//
// A trailing error is the failure itself rather than part of the answer, so it
// becomes a Furball and the rest is dropped.
func resultOf(what string, out []reflect.Value) Value {
	if n := len(out); n > 0 && out[n-1].Type() == errorType {
		if !out[n-1].IsNil() {
			return NewFurball("Hiss! %s: %s, nya~", what, out[n-1].Interface().(error))
		}
		out = out[:n-1]
	}
	switch len(out) {
	case 0:
		return NewNil()
	case 1:
		return fromGo(out[0])
	default:
		items := make([]Value, len(out))
		for i, o := range out {
			items[i] = fromGo(o)
		}
		return NewList(items...)
	}
}

// fromGo reads a Go value as a Meow one.
//
// Anything with no shape in Meow is carried whole rather than refused, which is
// what lets a client or a handle come back out of a call.
func fromGo(rv reflect.Value) Value {
	if !rv.IsValid() {
		return NewNil()
	}
	// A Meow value that made the round trip is already what it should be.
	if rv.Type().Implements(valueType) {
		if rv.IsNil() {
			return NewNil()
		}
		return rv.Interface().(Value)
	}
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface:
		if rv.IsNil() {
			return NewNil()
		}
		return fromGo(rv.Elem())
	case reflect.Bool:
		return NewBool(rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return NewInt(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return NewInt(int64(rv.Uint()))
	case reflect.Float32, reflect.Float64:
		return NewFloat(rv.Float())
	case reflect.String:
		return NewString(rv.String())
	case reflect.Slice, reflect.Array:
		if rv.Kind() == reflect.Slice && rv.IsNil() {
			return NewNil()
		}
		items := make([]Value, rv.Len())
		for i := range items {
			items[i] = fromGo(rv.Index(i))
		}
		return NewList(items...)
	case reflect.Map:
		if rv.IsNil() {
			return NewNil()
		}
		items := make(map[string]Value, rv.Len())
		for _, k := range rv.MapKeys() {
			items[fmt.Sprint(k.Interface())] = fromGo(rv.MapIndex(k))
		}
		return NewMap(items)
	case reflect.Struct:
		if rv.Type() == timeType {
			return NewString(rv.Interface().(time.Time).Format(time.RFC3339))
		}
		return structToMap(rv)
	}
	return NewOpaque(rv.Type().String(), rv.Interface())
}

// structToMap reads a record's exported fields as a basket, under the names a
// Meow program would write them with.
func structToMap(rv reflect.Value) Value {
	t := rv.Type()
	items := make(map[string]Value, t.NumField())
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		items[snake(f.Name)] = fromGo(rv.Field(i))
	}
	return NewMap(items)
}

// toGo reads a Meow value as the Go type a call is asking for.
func toGo(v Value, t reflect.Type) (reflect.Value, error) {
	// Something being carried goes back exactly as it came.
	if o, ok := AsOpaque(v); ok {
		ov := reflect.ValueOf(o.V)
		if o.V != nil && ov.Type().AssignableTo(t) {
			return ov, nil
		}
		return reflect.Value{}, fmt.Errorf("cannot read %s as a %s", o.String(), t)
	}
	// A function asking for a Meow value is handed it untouched.
	if t == valueType {
		return reflect.ValueOf(v), nil
	}
	if _, isNil := v.(*NilValue); isNil {
		switch t.Kind() {
		case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map:
			return reflect.Zero(t), nil
		}
		return reflect.Value{}, fmt.Errorf("cannot read nothing as a %s", t)
	}
	if t.Kind() == reflect.Pointer {
		inner, err := toGo(v, t.Elem())
		if err != nil {
			return reflect.Value{}, err
		}
		p := reflect.New(t.Elem())
		p.Elem().Set(inner)
		return p, nil
	}

	switch t.Kind() {
	case reflect.Bool:
		return reflect.ValueOf(v.IsTruthy()).Convert(t), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, ok := v.(*Int)
		if !ok {
			return reflect.Value{}, fmt.Errorf("cannot read a %s as a %s", v.Type(), t)
		}
		return reflect.ValueOf(i.Val).Convert(t), nil
	case reflect.Float32, reflect.Float64:
		switch n := v.(type) {
		case *Float:
			return reflect.ValueOf(n.Val).Convert(t), nil
		case *Int:
			return reflect.ValueOf(float64(n.Val)).Convert(t), nil
		}
		return reflect.Value{}, fmt.Errorf("cannot read a %s as a %s", v.Type(), t)
	case reflect.String:
		s, ok := v.(*String)
		if !ok {
			return reflect.Value{}, fmt.Errorf("cannot read a %s as a %s", v.Type(), t)
		}
		return reflect.ValueOf(s.Val).Convert(t), nil
	case reflect.Slice:
		l, ok := v.(*List)
		if !ok {
			return reflect.Value{}, fmt.Errorf("cannot read a %s as a %s", v.Type(), t)
		}
		out := reflect.MakeSlice(t, len(l.Items), len(l.Items))
		for i, item := range l.Items {
			gv, err := toGo(item, t.Elem())
			if err != nil {
				return reflect.Value{}, fmt.Errorf("element %d: %w", i+1, err)
			}
			out.Index(i).Set(gv)
		}
		return out, nil
	case reflect.Map:
		m, ok := v.(*Map)
		if !ok {
			return reflect.Value{}, fmt.Errorf("cannot read a %s as a %s", v.Type(), t)
		}
		if t.Key().Kind() != reflect.String {
			return reflect.Value{}, fmt.Errorf("cannot read a basket as a %s", t)
		}
		out := reflect.MakeMapWithSize(t, len(m.Items))
		for k, item := range m.Items {
			gv, err := toGo(item, t.Elem())
			if err != nil {
				return reflect.Value{}, fmt.Errorf("%q: %w", k, err)
			}
			out.SetMapIndex(reflect.ValueOf(k).Convert(t.Key()), gv)
		}
		return out, nil
	case reflect.Struct:
		return mapToStruct(v, t)
	}
	return reflect.Value{}, fmt.Errorf("cannot read a %s as a %s", v.Type(), t)
}

// mapToStruct fills a record's exported fields from a basket, by the names a
// Meow program writes. A field the basket does not mention is left as it is.
func mapToStruct(v Value, t reflect.Type) (reflect.Value, error) {
	m, ok := v.(*Map)
	if !ok {
		return reflect.Value{}, fmt.Errorf("cannot read a %s as a %s", v.Type(), t)
	}
	out := reflect.New(t).Elem()
	fields := make(map[string]int, t.NumField())
	for i := range t.NumField() {
		if f := t.Field(i); f.IsExported() {
			fields[snake(f.Name)] = i
		}
	}
	for k, item := range m.Items {
		i, known := fields[k]
		if !known {
			return reflect.Value{}, fmt.Errorf("%s has no %q", t, k)
		}
		gv, err := toGo(item, t.Field(i).Type)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("%q: %w", k, err)
		}
		out.Field(i).Set(gv)
	}
	return out, nil
}

// snake writes a Go field name the way a Meow program reads it: UserId becomes
// user_id, ARN becomes arn.
func snake(name string) string {
	var b strings.Builder
	runes := []rune(name)
	for i, r := range runes {
		upper := r >= 'A' && r <= 'Z'
		if upper && i > 0 {
			prev := runes[i-1]
			prevLower := (prev >= 'a' && prev <= 'z') || (prev >= '0' && prev <= '9')
			nextLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'
			if prevLower || nextLower {
				b.WriteByte('_')
			}
		}
		if upper {
			b.WriteRune(r - 'A' + 'a')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
