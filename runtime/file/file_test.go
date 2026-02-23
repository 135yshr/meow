package file_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/135yshr/meow/runtime/file"
	"github.com/135yshr/meow/runtime/meowrt"
)

func TestSnoop(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("hello world\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result := file.Snoop(meowrt.NewString(path))
	s, ok := result.(*meowrt.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if s.Val != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", s.Val)
	}
}

func TestSnoopNotFound(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
	}()
	file.Snoop(meowrt.NewString("/nonexistent/path.txt"))
}

func TestSnoopNonString(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
	}()
	file.Snoop(meowrt.NewInt(42))
}

func TestStalk(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lines.txt")
	if err := os.WriteFile(path, []byte("alpha\nbeta\ngamma\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result := file.Stalk(meowrt.NewString(path))
	lst, ok := result.(*meowrt.List)
	if !ok {
		t.Fatalf("expected List, got %T", result)
	}
	if lst.Len() != 3 {
		t.Fatalf("expected 3 lines, got %d", lst.Len())
	}
	expected := []string{"alpha", "beta", "gamma"}
	for i, want := range expected {
		got := lst.Get(i).(*meowrt.String).Val
		if got != want {
			t.Errorf("line[%d]: expected %q, got %q", i, want, got)
		}
	}
}

func TestStalkEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	result := file.Stalk(meowrt.NewString(path))
	lst, ok := result.(*meowrt.List)
	if !ok {
		t.Fatalf("expected List, got %T", result)
	}
	if lst.Len() != 0 {
		t.Errorf("expected 0 lines, got %d", lst.Len())
	}
}

func TestStalkNotFound(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	file.Stalk(meowrt.NewString("/nonexistent/path.txt"))
}

func TestStalkNonString(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	file.Stalk(meowrt.NewInt(42))
}
