package meowrt

// here is where the program is: the source position of the last statement to
// start running.
//
// A failure knows what went wrong and not where. Everything a program reads
// from outside itself arrives as text, so "Cannot read "3 " as an Int" is a
// message a real program produces — and without a position, finding which of
// two hundred lines asked for that number is the reader's problem.
//
// It is a single variable rather than something carried on each Furball because
// a Furball is built in a hundred places, and because the answer wanted is the
// same either way: a failure propagates without running further statements, so
// the last statement to start is the innermost one that was running.
var here string

// Here records where the program is. Generated code calls it before each
// statement, and the interpreter calls it as it walks them, so both report the
// same position for the same program.
func Here(pos string) { here = pos }

// Where reports the position last recorded by Here, or "" before any statement
// has run.
func Where() string { return here }

// Located prefixes a message with where the program was, in the form the
// compiler's own errors use — file:line:column. A failure with nowhere to point
// at is left alone rather than given an empty prefix.
func Located(message string) string {
	if here == "" {
		return message
	}
	return here + ": " + message
}

// Returning restores the position a call was made from and hands back v.
//
// Generated code returns through it so that a call which succeeds does not
// leave the program pointing inside the function it just came back from: a
// failure later in the same statement would be blamed on the callee's last
// line, which is code that worked.
//
// It restores only on the way back from a call that returned. A call that fails
// never reaches it, so the innermost position — where the failure actually
// happened — is the one that survives.
func Returning[T any](pos string, v T) T {
	Here(pos)
	return v
}
