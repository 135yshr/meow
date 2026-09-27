package meowrt

import "testing"

// atPosition sets the recorded position for one test and puts it back after, so
// tests do not report each other's positions.
func atPosition(t *testing.T, pos string) {
	t.Helper()
	original := Where()
	Here(pos)
	t.Cleanup(func() { Here(original) })
}

func TestLocated(t *testing.T) {
	atPosition(t, "probe.nyan:12:3")

	got := Located("Hiss! Cannot read \"x\" as an Int, nya~")

	want := "probe.nyan:12:3: Hiss! Cannot read \"x\" as an Int, nya~"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// A failure with nowhere to point at is left alone rather than given an empty
// prefix — nothing has run yet, so there is no line to blame.
func TestLocatedWithoutAPosition(t *testing.T) {
	atPosition(t, "")

	if got := Located("Hiss! boom, nya~"); got != "Hiss! boom, nya~" {
		t.Errorf("got %q, want the message unchanged", got)
	}
}

func TestHereIsWhatWhereReports(t *testing.T) {
	atPosition(t, "a.nyan:1:1")

	Here("b.nyan:2:2")

	if got := Where(); got != "b.nyan:2:2" {
		t.Errorf("got %q, want b.nyan:2:2", got)
	}
}

// A call that comes back goes back to where it was made from, whatever type its
// result is.
func TestReturningGoesBackToTheCaller(t *testing.T) {
	atPosition(t, "prog.nyan:3:3")

	if got := Returning("prog.nyan:5:1", NewInt(1)); got.Val != 1 {
		t.Errorf("got %v, want the value handed back unchanged", got)
	}
	if got := Where(); got != "prog.nyan:5:1" {
		t.Errorf("after a boxed result the program is at %q, want the caller's prog.nyan:5:1", got)
	}

	Here("prog.nyan:3:3")
	Returning("prog.nyan:5:1", "native")
	if got := Where(); got != "prog.nyan:5:1" {
		t.Errorf("after a native result the program is at %q, want the caller's prog.nyan:5:1", got)
	}
}

// A boxed body hands its failure back as a value. That is a call that failed,
// so the program stays where it failed instead of going back to the caller,
// which would blame the failure on the line that called the function (#163).
func TestReturningAFailureStaysWhereItFailed(t *testing.T) {
	atPosition(t, "prog.nyan:3:3")

	failure := NewFurball("Hiss! boom, nya~")
	if got := Returning[Value]("prog.nyan:5:1", failure); got != failure {
		t.Errorf("got %v, want the Furball handed back unchanged", got)
	}
	if got := Where(); got != "prog.nyan:3:3" {
		t.Errorf("after a failure the program is at %q, want where it failed, prog.nyan:3:3", got)
	}
}

// A failure that was caught is a value like any other, so the program goes back
// to the caller as it does for any result.
func TestReturningACaughtFailureGoesBackToTheCaller(t *testing.T) {
	atPosition(t, "prog.nyan:3:3")

	caught := NewFurball("Hiss! boom, nya~")
	caught.Handled = true
	Returning[Value]("prog.nyan:5:1", caught)
	if got := Where(); got != "prog.nyan:5:1" {
		t.Errorf("after a caught failure the program is at %q, want the caller's prog.nyan:5:1", got)
	}
}
