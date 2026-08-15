package meowrt

import "strings"

// requireString returns the value as a Go string, or a Furball if it is not a
// String. A Furball argument is returned unchanged for propagation.
func requireString(name, what string, v Value) (string, *Furball) {
	if f, ok := v.(*Furball); ok {
		return "", f
	}
	s, ok := v.(*String)
	if !ok {
		typeName := "catnap"
		if v != nil {
			typeName = v.Type()
		}
		return "", NewFurball("Hiss! %s requires a String %s, got %s, nya~", name, what, typeName)
	}
	return s.Val, nil
}

// Whiff reports whether s contains sub, as a Bool.
func Whiff(s, sub Value) Value {
	str, f := requireString("whiff", "haystack", s)
	if f != nil {
		return f
	}
	needle, f := requireString("whiff", "needle", sub)
	if f != nil {
		return f
	}
	return NewBool(strings.Contains(str, needle))
}

// Track returns the byte offset of the first occurrence of sub in s, or -1 when
// it does not occur.
func Track(s, sub Value) Value {
	str, f := requireString("track", "haystack", s)
	if f != nil {
		return f
	}
	needle, f := requireString("track", "needle", sub)
	if f != nil {
		return f
	}
	return NewInt(int64(strings.Index(str, needle)))
}

// Shred splits s around each occurrence of sep, returning a List of the pieces.
// An empty separator splits s into its individual characters.
func Shred(s, sep Value) Value {
	str, f := requireString("shred", "value", s)
	if f != nil {
		return f
	}
	separator, f := requireString("shred", "separator", sep)
	if f != nil {
		return f
	}
	var parts []string
	if separator == "" {
		parts = strings.Split(str, "")
	} else {
		parts = strings.Split(str, separator)
	}
	elems := make([]Value, len(parts))
	for i, p := range parts {
		elems[i] = NewString(p)
	}
	return NewList(elems...)
}

// Tangle concatenates the elements of a List into a single string, separated by
// sep. It is the inverse of Shred.
func Tangle(list, sep Value) Value {
	l, f := requireList("tangle", list)
	if f != nil {
		return f
	}
	separator, fb := requireString("tangle", "separator", sep)
	if fb != nil {
		return fb
	}
	parts := make([]string, 0, l.Len())
	for v := range l.Iter() {
		s, fb := requireString("tangle", "element", v)
		if fb != nil {
			return fb
		}
		parts = append(parts, s)
	}
	return NewString(strings.Join(parts, separator))
}

// Nibble returns the substring of s from start up to but not including end,
// counted in characters rather than bytes so that multi-byte text behaves the
// way it reads. A negative index counts back from the end; the range is
// clamped to the bounds of s, and an inverted range yields "".
func Nibble(s, start, end Value) Value {
	str, f := requireString("nibble", "value", s)
	if f != nil {
		return f
	}
	from, fb := TryAsInt(start)
	if fb != nil {
		return fb
	}
	to, fb := TryAsInt(end)
	if fb != nil {
		return fb
	}

	runes := []rune(str)
	n := int64(len(runes))
	from = clampIndex(from, n)
	to = clampIndex(to, n)
	if from >= to {
		return NewString("")
	}
	return NewString(string(runes[from:to]))
}

// clampIndex resolves a possibly negative index against a length of n and
// clamps it into [0, n].
func clampIndex(i, n int64) int64 {
	if i < 0 {
		i += n
	}
	if i < 0 {
		return 0
	}
	if i > n {
		return n
	}
	return i
}
