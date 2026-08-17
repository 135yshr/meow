package meowrt

import "testing"

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
