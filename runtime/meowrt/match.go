package meowrt

// MatchRange checks if a value is within an integer range [low, high].
func MatchRange(v Value, low, high int64) bool {
	switch v := v.(type) {
	case *Int:
		return v.Val >= low && v.Val <= high
	case *Float:
		return v.Val >= float64(low) && v.Val <= float64(high)
	default:
		return false
	}
}

// MatchValue checks if two values are equal.
func MatchValue(v, pattern Value) bool {
	return Equal(v, pattern).IsTruthy()
}
