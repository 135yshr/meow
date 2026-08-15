package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeMeowModule lays out a directory that looks like the meow source tree.
func writeMeowModule(t *testing.T, dir, modulePath string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "runtime", "meowrt"), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "module " + modulePath + "\n\ngo 1.26.0\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestIsMeowModuleRoot(t *testing.T) {
	t.Run("meow source tree", func(t *testing.T) {
		dir := t.TempDir()
		writeMeowModule(t, dir, meowModulePath)
		if !isMeowModuleRoot(dir) {
			t.Error("expected the meow source tree to be recognised")
		}
	})

	// A different module must be rejected: pointing a `replace` at it produces a
	// go.mod that cannot resolve runtime/meowrt.
	t.Run("unrelated go module", func(t *testing.T) {
		dir := t.TempDir()
		writeMeowModule(t, dir, "example.com/other")
		if isMeowModuleRoot(dir) {
			t.Error("expected an unrelated module to be rejected")
		}
	})

	// The module path alone is not enough — a checkout without runtime/meowrt
	// cannot satisfy the generated import either.
	t.Run("meow module path without runtime", func(t *testing.T) {
		dir := t.TempDir()
		content := "module " + meowModulePath + "\n\ngo 1.26.0\n"
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if isMeowModuleRoot(dir) {
			t.Error("expected a checkout without runtime/meowrt to be rejected")
		}
	})

	t.Run("no go.mod", func(t *testing.T) {
		if isMeowModuleRoot(t.TempDir()) {
			t.Error("expected a directory without go.mod to be rejected")
		}
	})
}

func TestFindMeowModuleFrom(t *testing.T) {
	t.Run("walks up to the module root", func(t *testing.T) {
		root := t.TempDir()
		writeMeowModule(t, root, meowModulePath)
		nested := filepath.Join(root, "examples", "deep")
		if err := os.MkdirAll(nested, 0o755); err != nil {
			t.Fatal(err)
		}
		if got := findMeowModuleFrom(nested); got != root {
			t.Errorf("got %q, want %q", got, root)
		}
	})

	// Regression: previously any go.mod found while walking up was accepted and
	// used as a `replace` target, so running a .nyan file inside an unrelated Go
	// project emitted a go.mod that could not build.
	t.Run("ignores an unrelated module", func(t *testing.T) {
		root := t.TempDir()
		writeMeowModule(t, root, "example.com/other")
		if got := findMeowModuleFrom(root); got != "" {
			t.Errorf("got %q, want %q", got, "")
		}
	})
}

func TestModulePath(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{"simple", "module example.com/x\n\ngo 1.26.0\n", "example.com/x"},
		{"indented", "  module example.com/x  \n", "example.com/x"},
		{"missing", "go 1.26.0\n", ""},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := modulePath(tt.content); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildModContentInsideSourceTree(t *testing.T) {
	got, err := buildModContent("1.26", "/src/meow")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "replace "+meowModulePath+" => /src/meow") {
		t.Errorf("expected a replace directive, got:\n%s", got)
	}
}

// Outside the meow source tree there is nothing to replace with, so the
// published module is pinned by version instead. This is what lets a .nyan file
// anywhere on disk compile.
func TestBuildModContentOutsideSourceTree(t *testing.T) {
	original := Version
	t.Cleanup(func() { Version = original })
	Version = "0.10.6"

	got, err := buildModContent("1.26", "")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "replace") {
		t.Errorf("expected no replace directive, got:\n%s", got)
	}
	if !strings.Contains(got, "require "+meowModulePath+" v0.10.6") {
		t.Errorf("expected the version to be pinned, got:\n%s", got)
	}
}

func TestBuildModContentRejectsInvalidPath(t *testing.T) {
	if _, err := buildModContent("1.26", "/src/me\now"); err == nil {
		t.Error("expected an error for a module root containing a newline")
	}
}

func TestAsModuleVersion(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		want  string
		valid bool
	}{
		// Release tooling stamps the tag without its "v" prefix.
		{"bare version", "0.10.6", "v0.10.6", true},
		{"prefixed version", "v0.10.6", "v0.10.6", true},
		{"prerelease", "v1.0.0-rc.1", "v1.0.0-rc.1", true},
		{"padded", "  0.10.6  ", "v0.10.6", true},
		{"development default", "dev", "", false},
		{"go build placeholder", "(devel)", "", false},
		{"empty", "", "", false},
		{"not a version", "nyan", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := asModuleVersion(tt.in)
			if ok != tt.valid {
				t.Fatalf("got ok=%v, want %v", ok, tt.valid)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGoVersionFor(t *testing.T) {
	t.Run("reads the source tree go.mod", func(t *testing.T) {
		dir := t.TempDir()
		writeMeowModule(t, dir, meowModulePath)
		if got := goVersionFor(dir); got != "1.26.0" {
			t.Errorf("got %q, want %q", got, "1.26.0")
		}
	})

	t.Run("falls back outside the source tree", func(t *testing.T) {
		if got := goVersionFor(""); got != defaultGoVersion {
			t.Errorf("got %q, want %q", got, defaultGoVersion)
		}
	})
}
