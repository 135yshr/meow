package meowrt

// Index evaluates the `container[key]` expression for any indexable value.
//
// It is the single implementation shared by the transpiler (pkg/codegen) and
// the tree-walking interpreter (pkg/interpreter), so `nums[0]` and `m["a"]`
// mean the same thing on both backends.
//
// Dispatch is on the runtime type of container, not on the static shape of the
// index expression:
//
//	List — key must be an Int; out-of-range yields a Furball (see List.Get).
//	Map  — key must be a String; a missing key yields Nil.
//
// Anything else yields a Furball rather than panicking, so `~>` can recover.
func Index(container, key Value) Value {
	if f := propagate(container, key); f != nil {
		return f
	}
	switch obj := container.(type) {
	case *List:
		i, f := TryAsInt(key)
		if f != nil {
			return f
		}
		return obj.Get(int(i))
	case *Map:
		k, f := TryAsString(key)
		if f != nil {
			return f
		}
		if v, ok := obj.Get(k); ok {
			return v
		}
		return NewNil()
	case nil:
		return NewFurball("Hiss! cannot index nil, nya~")
	default:
		return NewFurball("Hiss! cannot index %s, nya~", obj.Type())
	}
}
