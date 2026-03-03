package meowtest

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/135yshr/meow/runtime/meowrt"
)

// testFailure is a panic value used by test assertions.
// Distinct from Hiss! panics so Run can differentiate.
type testFailure struct {
	message string
}

type testResult struct {
	name   string
	passed bool
	msg    string
}

var (
	output  io.Writer = os.Stdout
	results []testResult
	exitFn  = func(code int) { os.Exit(code) }
)

// Reset clears accumulated test results and reconfigures output/exit.
func Reset(w io.Writer, exit func(int)) {
	results = nil
	if w != nil {
		output = w
	} else {
		output = os.Stdout
	}
	if exit != nil {
		exitFn = exit
	} else {
		exitFn = func(code int) { os.Exit(code) }
	}
}

// Judge asserts that a condition is truthy.
// Usage: judge(condition) or judge(condition, "message")
func Judge(args ...meowrt.Value) meowrt.Value {
	if len(args) < 1 {
		panic("Hiss! judge expects at least 1 argument, nya~")
	}
	if !args[0].IsTruthy() {
		msg := "assertion failed: expected truthy value"
		if len(args) >= 2 {
			msg = args[1].String()
		}
		panic(testFailure{message: msg})
	}
	return meowrt.NewNil()
}

// Expect asserts that two values are equal (by String representation).
// Usage: expect(actual, expected) or expect(actual, expected, "message")
func Expect(args ...meowrt.Value) meowrt.Value {
	if len(args) < 2 {
		panic("Hiss! expect expects at least 2 arguments, nya~")
	}
	actual := args[0]
	expected := args[1]
	if actual.String() != expected.String() {
		msg := fmt.Sprintf("expected %s, got %s", expected.String(), actual.String())
		if len(args) >= 3 {
			msg = fmt.Sprintf("%s: expected %s, got %s", args[2].String(), expected.String(), actual.String())
		}
		panic(testFailure{message: msg})
	}
	return meowrt.NewNil()
}

// Refuse asserts that a condition is falsy.
// Usage: refuse(condition) or refuse(condition, "message")
func Refuse(args ...meowrt.Value) meowrt.Value {
	if len(args) < 1 {
		panic("Hiss! refuse expects at least 1 argument, nya~")
	}
	if args[0].IsTruthy() {
		msg := "assertion failed: expected falsy value"
		if len(args) >= 2 {
			msg = args[1].String()
		}
		panic(testFailure{message: msg})
	}
	return meowrt.NewNil()
}

// Run executes a named test function, catching panics, and records the result.
// Usage: run("test name", fn)
func Run(args ...meowrt.Value) meowrt.Value {
	if len(args) < 2 {
		panic("Hiss! run expects 2 arguments (name, fn), nya~")
	}
	name, ok := args[0].(*meowrt.String)
	if !ok {
		panic(fmt.Sprintf("Hiss! run expects a String name, got %s, nya~", args[0].Type()))
	}
	fn, ok := args[1].(*meowrt.Func)
	if !ok {
		panic(fmt.Sprintf("Hiss! run expects a Func, got %s, nya~", args[1].Type()))
	}

	passed := true
	var msg string
	func() {
		defer func() {
			if r := recover(); r != nil {
				passed = false
				if tf, ok := r.(testFailure); ok {
					msg = tf.message
				} else {
					msg = fmt.Sprintf("%v", r)
				}
			}
		}()
		fn.Call()
	}()

	status := "PASS"
	if !passed {
		status = "FAIL"
	}
	fmt.Fprintf(output, "  %s: %s", status, name.Val)
	if msg != "" {
		fmt.Fprintf(output, " - %s", msg)
	}
	fmt.Fprintln(output)

	results = append(results, testResult{name: name.Val, passed: passed, msg: msg})
	return meowrt.NewBool(passed)
}

// Catwalk executes a named function, captures its stdout output, and compares
// it with the expected output string. This is the Meow equivalent of Go's
// Example tests with // Output: comments.
// Usage: Catwalk(name, fn, expected)
func Catwalk(args ...meowrt.Value) meowrt.Value {
	if len(args) < 3 {
		panic("Hiss! catwalk expects 3 arguments (name, fn, expected), nya~")
	}
	name, ok := args[0].(*meowrt.String)
	if !ok {
		panic(fmt.Sprintf("Hiss! catwalk expects a String name, got %s, nya~", args[0].Type()))
	}
	fn, ok := args[1].(*meowrt.Func)
	if !ok {
		panic(fmt.Sprintf("Hiss! catwalk expects a Func, got %s, nya~", args[1].Type()))
	}
	expected, ok := args[2].(*meowrt.String)
	if !ok {
		panic(fmt.Sprintf("Hiss! catwalk expects a String expected, got %s, nya~", args[2].Type()))
	}

	// Capture stdout using os.Pipe.
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		panic(fmt.Sprintf("Hiss! cannot create pipe, nya~: %v", err))
	}
	os.Stdout = w

	// Read pipe output in a goroutine to prevent deadlock.
	captured := make(chan string)
	go func() {
		var buf bytes.Buffer
		buf.ReadFrom(r)
		captured <- buf.String()
	}()

	passed := true
	var msg string
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				passed = false
				if tf, ok := rec.(testFailure); ok {
					msg = tf.message
				} else {
					msg = fmt.Sprintf("panic: %v", rec)
				}
			}
		}()
		fn.Call()
	}()

	// Close writer and restore stdout.
	w.Close()
	os.Stdout = oldStdout
	got := <-captured
	r.Close()

	// Compare output if function didn't panic.
	if passed && got != expected.Val {
		passed = false
		msg = fmt.Sprintf("output mismatch:\ngot:\n%swant:\n%s", got, expected.Val)
	}

	status := "PASS"
	if !passed {
		status = "FAIL"
	}
	fmt.Fprintf(output, "  %s: %s", status, name.Val)
	if msg != "" {
		fmt.Fprintf(output, " - %s", msg)
	}
	fmt.Fprintln(output)

	results = append(results, testResult{name: name.Val, passed: passed, msg: msg})
	return meowrt.NewBool(passed)
}

// Report outputs the test summary. Calls os.Exit(1) if any test failed.
func Report(args ...meowrt.Value) meowrt.Value {
	passed := 0
	failed := 0
	for _, r := range results {
		if r.passed {
			passed++
		} else {
			failed++
		}
	}

	fmt.Fprintln(output)
	if failed == 0 {
		fmt.Fprintf(output, "All %d tests passed, nya~!\n", passed)
	} else {
		fmt.Fprintf(output, "%d passed, %d failed, nya~\n", passed, failed)
	}

	if failed > 0 {
		exitFn(1)
	}
	return meowrt.NewNil()
}
