package compiler_test

import "testing"

// A failure inside a lambda has to be reported at the line that failed, as it
// is inside a named function. The compiled program handed a boxed result back
// through meow.Returning, which put the program back at the caller's line
// whatever came back — so a Furball returned by a lambda, or by any function
// whose result is boxed, was blamed on the line that called it, and a lambda
// passed to lick was blamed on the whole call (#163). The playground reported
// the line that failed.
func TestAFailureInsideALambdaIsBlamedOnItsOwnLine(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		output  string
		failure string
	}{
		{
			"a lambda called by name",
			"nyan g = paw() { bring to_float(\"bad\") }\nnya(g())\n",
			"",
			"prog.nyan:1:18: Hiss! Cannot read \"bad\" as a Float, nya~",
		},
		{
			"a lambda passed to lick",
			"nyan xs = [1, 2]\nnya(lick(xs, paw(x) {\n  bring to_float(\"bad\")\n}))\n",
			"",
			"prog.nyan:3:3: Hiss! Cannot read \"bad\" as a Float, nya~",
		},
		{
			"the last line of a longer lambda",
			"nyan g = paw() {\n  nya(\"in\")\n  bring to_float(\"bad\")\n}\nnya(g())\n",
			"in\n",
			"prog.nyan:3:3: Hiss! Cannot read \"bad\" as a Float, nya~",
		},
		{
			"a lambda called inside a typed function",
			"meow f() float {\n  nyan h = paw() { bring to_float(\"bad\") }\n  bring h()\n}\nnya(to_string(f()))\n",
			"",
			"prog.nyan:2:20: Hiss! Cannot read \"bad\" as a Float, nya~",
		},
		{
			// Not a lambda, but the same return: a function whose result is
			// boxed hands a failure back as a value too.
			"a named function returning a kitty",
			"kitty P { x: float }\nmeow f() P {\n  nya(\"in f\")\n  bring P(to_float(\"bad\"))\n}\nnya(f())\n",
			"in f\n",
			"prog.nyan:4:3: Hiss! Cannot read \"bad\" as a Float, nya~",
		},
		{
			// A lambda that succeeds still leaves the program where it was
			// called, so a failure later in the same statement is blamed on
			// that statement and not on the lambda's last line.
			"a failure after a lambda that succeeded",
			"nyan g = paw() {\n  bring to_float(\"1.5\")\n}\nnya(g(), to_float(\"bad\"))\n",
			"",
			"prog.nyan:4:1: Hiss! Cannot read \"bad\" as a Float, nya~",
		},
		{
			// A failure the lambda caught is a value it returned, not a
			// failure, so the program goes back to the caller as for any
			// other result.
			"a failure after a lambda that caught its own",
			"nyan g = paw() {\n  bring to_float(\"bad\") ~> 0.0\n}\nnya(g())\nnya(to_float(\"x\"))\n",
			"0\n",
			"prog.nyan:5:1: Hiss! Cannot read \"x\" as a Float, nya~",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, backend := range []struct {
				name string
				run  func(*testing.T, string) (string, string)
			}{
				{"compiled", runCompiled},
				{"interpreted", runInterpreted},
			} {
				output, failure := backend.run(t, tt.source)
				if output != tt.output {
					t.Errorf("%s printed %q, want %q", backend.name, output, tt.output)
				}
				if failure != tt.failure {
					t.Errorf("%s failed with %q, want %q", backend.name, failure, tt.failure)
				}
			}
		})
	}
}
