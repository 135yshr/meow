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
