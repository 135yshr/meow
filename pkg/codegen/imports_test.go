package codegen_test

import (
	"strings"
	"testing"
)

// Go rejects an import nothing uses, so a leftover `nab` failed the build with
// a Go error naming a generated alias. Only what is called gets imported.
func TestUnusedNabIsNotImported(t *testing.T) {
	out := generate(t, `nab "env"
nab "clock"
nya(clock.now())`)

	if strings.Contains(out, "runtime/env") {
		t.Error("imported env, which the program never calls")
	}
	if !strings.Contains(out, "runtime/clock") {
		t.Error("dropped clock, which the program does call")
	}
}

func TestUsedNabIsImported(t *testing.T) {
	out := generate(t, `nab "env"
nya(env.hunt("HOME", ""))`)

	if !strings.Contains(out, "runtime/env") {
		t.Error("dropped env, which the program does call")
	}
}

// An aliased import is called by its alias but imported under the package name.
func TestAliasedNabIsImported(t *testing.T) {
	out := generate(t, `nab "clock" tag t
nya(t.now())`)

	if !strings.Contains(out, "runtime/clock") {
		t.Error("dropped clock, which the program calls through its tag")
	}
}
