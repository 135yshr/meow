package meowrt

import (
	"strings"
	"testing"
)

// withFakeExit replaces the exit hook for one test, so a test that scrams does
// not take the test binary with it.
func withFakeExit(t *testing.T) *int {
	t.Helper()
	got := -1
	original := exit
	exit = func(code int) { got = code }
	t.Cleanup(func() { exit = original })
	return &got
}

func TestScram(t *testing.T) {
	tests := []struct {
		name string
		args []Value
		want int
	}{
		{"a status", []Value{NewInt(3)}, 3},
		{"success", []Value{NewInt(0)}, 0},
		// No argument means success, so a program with nothing to report need
		// not spell it out.
		{"no argument", nil, 0},
		{"the largest a process can report", []Value{NewInt(maxExitCode)}, maxExitCode},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := withFakeExit(t)

			Scram(tt.args...)

			if *got != tt.want {
				t.Errorf("exited with %d, want %d", *got, tt.want)
			}
		})
	}
}

// A shell sees the low eight bits of the status, so 256 would arrive as 0 — a
// program asking to fail would be read as having succeeded. Refused rather than
// quietly wrapped.
func TestScramRefusesAStatusAProcessCannotReport(t *testing.T) {
	tests := []struct {
		name string
		args []Value
	}{
		{"past the top", []Value{NewInt(maxExitCode + 1)}},
		{"negative", []Value{NewInt(-1)}},
		{"not a number", []Value{NewString("3")}},
		{"a furball", []Value{&Furball{Message: "boom"}}},
		{"too many arguments", []Value{NewInt(0), NewInt(1)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := withFakeExit(t)

			if _, ok := Scram(tt.args...).(*Furball); !ok {
				t.Error("expected a Furball")
			}
			if *got != -1 {
				t.Errorf("exited with %d, want no exit at all", *got)
			}
		})
	}
}

// A typed function returns a native Go type and cannot pass a Furball back, so
// a refused status is raised there instead — the same bridge hiss uses. A bare
// call would drop it and the program would carry on as if nothing was asked.
func TestScramOrHissRaisesARefusedStatus(t *testing.T) {
	got := withFakeExit(t)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected a panic")
		}
		if msg, ok := r.(string); !ok || !strings.Contains(msg, "0 to 255") {
			t.Errorf("panicked with %v, want the reason the status was refused", r)
		}
		if *got != -1 {
			t.Errorf("exited with %d, want no exit at all", *got)
		}
	}()

	ScramOrHiss(NewInt(300))
}

func TestScramOrHissEndsOnAGoodStatus(t *testing.T) {
	got := withFakeExit(t)

	ScramOrHiss(NewInt(3))

	if *got != 3 {
		t.Errorf("exited with %d, want 3", *got)
	}
}

// gag catches failures, and a program asking to end is not one. A compiled
// program could not catch os.Exit either, so letting the signal past is what
// keeps the playground in step with it.
func TestGagDoesNotCatchAProgramEnding(t *testing.T) {
	defer func() {
		r := recover()
		sig, ok := r.(ScramSignal)
		if !ok {
			t.Fatalf("got %v, want the scram signal back", r)
		}
		if sig.Code != 3 {
			t.Errorf("status %d, want 3", sig.Code)
		}
	}()

	Gag(NewFunc("thunk", func(...Value) Value {
		panic(ScramSignal{Code: 3})
	}))
}
