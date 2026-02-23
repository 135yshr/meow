// Package meowrt provides the runtime support for compiled Meow programs.
// It defines the dynamic [Value] interface and concrete types used for all
// values at runtime, along with built-in functions and operators.
//
// # Value Types
//
// Every Meow value implements the [Value] interface:
//
//   - [Int]       integer (int64)
//   - [Float]     floating-point (float64)
//   - [String]    string
//   - [Bool]      boolean
//   - [NilValue]  nil / catnap
//   - [Func]      callable function
//   - [List]      ordered collection of values
//
// # Constructors
//
//   - [NewInt]     creates an Int
//   - [NewFloat]   creates a Float
//   - [NewString]  creates a String
//   - [NewBool]    creates a Bool
//   - [NewNil]     creates a NilValue
//   - [NewFunc]    creates a Func
//   - [NewList]    creates a List
//
// # Built-in Functions
//
//   - [Nya]       print values to stdout (like fmt.Println)
//   - [Call]      invoke a Func value
//   - [Len]       return the length of a String or List
//   - [ToInt]     convert to Int
//   - [ToFloat]   convert to Float
//   - [ToString]  convert to String
//
// # List Higher-Order Functions
//
//   - [Lick]     map a function over a list
//   - [Picky]    filter a list by predicate
//   - [Curl]     reduce (fold) a list
//   - [Append]   append a value to a list (returns new list)
//   - [Head]     return the first element
//   - [Tail]     return all elements except the first
//
// # Operators
//
//   - Arithmetic: [Add], [Sub], [Mul], [Div], [Mod], [Negate]
//   - Equality:   [Equal], [NotEqual]
//   - Comparison: [LessThan], [GreaterThan], [LessEqual], [GreaterEqual]
//   - Logical:    [And], [Or], [Not]
//
// # Pattern Matching
//
//   - [MatchValue]  check if two values are equal
//   - [MatchRange]  check if a value is within an integer range [low, high]
package meowrt
