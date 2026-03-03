package coverage

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegisterAndHit(t *testing.T) {
	Reset()
	id0 := Register("test.nyan", 1, 1, 1, 20, 1)
	id1 := Register("test.nyan", 2, 1, 2, 20, 1)

	if id0 != 0 {
		t.Errorf("expected id 0, got %d", id0)
	}
	if id1 != 1 {
		t.Errorf("expected id 1, got %d", id1)
	}

	Hit(0)
	Hit(0)
	Hit(1)

	bs := Blocks()
	if len(bs) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(bs))
	}
	if bs[0].Count != 2 {
		t.Errorf("block 0 count: expected 2, got %d", bs[0].Count)
	}
	if bs[1].Count != 1 {
		t.Errorf("block 1 count: expected 1, got %d", bs[1].Count)
	}
}

func TestReportAllCovered(t *testing.T) {
	Reset()
	Register("test.nyan", 1, 1, 1, 10, 1)
	Register("test.nyan", 2, 1, 2, 10, 1)
	Hit(0)
	Hit(1)

	var buf bytes.Buffer
	Report(&buf)
	got := buf.String()
	if !strings.Contains(got, "100.0%") {
		t.Errorf("expected 100.0%%, got %q", got)
	}
	if !strings.Contains(got, "nya~") {
		t.Errorf("expected nya~ suffix, got %q", got)
	}
}

func TestReportPartialCoverage(t *testing.T) {
	Reset()
	Register("test.nyan", 1, 1, 1, 10, 1)
	Register("test.nyan", 2, 1, 2, 10, 1)
	Hit(0)

	var buf bytes.Buffer
	Report(&buf)
	got := buf.String()
	if !strings.Contains(got, "50.0%") {
		t.Errorf("expected 50.0%%, got %q", got)
	}
}

func TestReportNoBlocks(t *testing.T) {
	Reset()
	var buf bytes.Buffer
	Report(&buf)
	if buf.Len() != 0 {
		t.Errorf("expected empty output for no blocks, got %q", buf.String())
	}
}

func TestWriteProfile(t *testing.T) {
	Reset()
	Register("test.nyan", 1, 3, 1, 20, 1)
	Register("test.nyan", 2, 3, 2, 20, 1)
	Hit(0)

	dir := t.TempDir()
	path := filepath.Join(dir, "coverage.out")
	// Write header first (as CLI would do)
	if err := os.WriteFile(path, []byte("mode: set\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := WriteProfile(path); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.HasPrefix(content, "mode: set\n") {
		t.Errorf("expected mode: set header, got %q", content)
	}
	if !strings.Contains(content, "test.nyan:1.3,1.20 1 1") {
		t.Errorf("expected hit block line, got %q", content)
	}
	if !strings.Contains(content, "test.nyan:2.3,2.20 1 0") {
		t.Errorf("expected miss block line, got %q", content)
	}
}

func TestReset(t *testing.T) {
	Register("test.nyan", 1, 1, 1, 10, 1)
	Reset()
	if len(Blocks()) != 0 {
		t.Errorf("expected 0 blocks after reset, got %d", len(Blocks()))
	}
}
