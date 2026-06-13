package meowtest

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/135yshr/meow/runtime/meowrt"
)

// testFailure was previously used as a panic sentinel for assertion failures.
// After the panic-to-Furball migration, assertions return *meowrt.Furball
// values that propagate via the codegen's short-circuit. The type is kept
// for any remaining recover() compatibility but is no longer raised.
type testFailure struct {
	message string
}

// assertionFailure constructs a Furball that represents a test assertion
// failure. Generated code's statement-level short-circuit returns it from
// the enclosing test function, where Run inspects it.
func assertionFailure(msg string) meowrt.Value {
	return &meowrt.Furball{Message: msg}
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

// Judge asserts that a condition is truthy. On failure it returns a Furball
// whose propagation through the enclosing test function signals the failure
// to Run/Catwalk. Returns NewNil on success.
func Judge(args ...meowrt.Value) meowrt.Value {
	if len(args) < 1 {
		return &meowrt.Furball{Message: "Hiss! judge expects at least 1 argument, nya~"}
	}
	if f, ok := args[0].(*meowrt.Furball); ok && !f.Handled {
		return f
	}
	if !args[0].IsTruthy() {
		msg := "assertion failed: expected truthy value"
		if len(args) >= 2 {
			msg = args[1].String()
		}
		return assertionFailure(msg)
	}
	return meowrt.NewNil()
}

// Expect asserts that two values are equal (by String representation).
func Expect(args ...meowrt.Value) meowrt.Value {
	if len(args) < 2 {
		return &meowrt.Furball{Message: "Hiss! expect expects at least 2 arguments, nya~"}
	}
	if f, ok := args[0].(*meowrt.Furball); ok && !f.Handled {
		return f
	}
	if f, ok := args[1].(*meowrt.Furball); ok && !f.Handled {
		return f
	}
	actual := args[0]
	expected := args[1]
	if actual.String() != expected.String() {
		msg := fmt.Sprintf("expected %s, got %s", expected.String(), actual.String())
		if len(args) >= 3 {
			msg = fmt.Sprintf("%s: expected %s, got %s", args[2].String(), expected.String(), actual.String())
		}
		return assertionFailure(msg)
	}
	return meowrt.NewNil()
}

// Refuse asserts that a condition is falsy.
func Refuse(args ...meowrt.Value) meowrt.Value {
	if len(args) < 1 {
		return &meowrt.Furball{Message: "Hiss! refuse expects at least 1 argument, nya~"}
	}
	if f, ok := args[0].(*meowrt.Furball); ok && !f.Handled {
		return f
	}
	if args[0].IsTruthy() {
		msg := "assertion failed: expected falsy value"
		if len(args) >= 2 {
			msg = args[1].String()
		}
		return assertionFailure(msg)
	}
	return meowrt.NewNil()
}

// Run executes a named test function, recording the result.
// A returned *Furball (from a failed assertion that propagated via short-circuit)
// or any panic is treated as a failure.
func Run(args ...meowrt.Value) meowrt.Value {
	if len(args) < 2 {
		return &meowrt.Furball{Message: "Hiss! run expects 2 arguments (name, fn), nya~"}
	}
	name, ok := args[0].(*meowrt.String)
	if !ok {
		return &meowrt.Furball{Message: fmt.Sprintf("Hiss! run expects a String name, got %s, nya~", args[0].Type())}
	}
	fn, ok := args[1].(*meowrt.Func)
	if !ok {
		return &meowrt.Furball{Message: fmt.Sprintf("Hiss! run expects a Func, got %s, nya~", args[1].Type())}
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
		ret := fn.Call()
		if f, ok := ret.(*meowrt.Furball); ok {
			passed = false
			msg = f.Message
		}
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

// Catwalk executes a named function, captures stdout, and compares it with
// the expected output. A Furball return (from a failed assertion) or a panic
// inside the function counts as a failure.
func Catwalk(args ...meowrt.Value) meowrt.Value {
	if len(args) < 3 {
		return &meowrt.Furball{Message: "Hiss! catwalk expects 3 arguments (name, fn, expected), nya~"}
	}
	name, ok := args[0].(*meowrt.String)
	if !ok {
		return &meowrt.Furball{Message: fmt.Sprintf("Hiss! catwalk expects a String name, got %s, nya~", args[0].Type())}
	}
	fn, ok := args[1].(*meowrt.Func)
	if !ok {
		return &meowrt.Furball{Message: fmt.Sprintf("Hiss! catwalk expects a Func, got %s, nya~", args[1].Type())}
	}
	expected, ok := args[2].(*meowrt.String)
	if !ok {
		return &meowrt.Furball{Message: fmt.Sprintf("Hiss! catwalk expects a String expected, got %s, nya~", args[2].Type())}
	}

	// Capture stdout using os.Pipe.
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return &meowrt.Furball{Message: fmt.Sprintf("Hiss! cannot create pipe, nya~: %v", err)}
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
		ret := fn.Call()
		if f, ok := ret.(*meowrt.Furball); ok {
			passed = false
			msg = f.Message
		}
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
