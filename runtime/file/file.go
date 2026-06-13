package file

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/135yshr/meow/runtime/meowrt"
)

// furball wraps an error as a Meow Furball value with the "Hiss! ... nya~" form.
func furball(format string, args ...any) meowrt.Value {
	return &meowrt.Furball{Message: fmt.Sprintf("Hiss! "+format+", nya~", args...)}
}

// Snoop reads the entire contents of a file and returns it as a String.
func Snoop(path meowrt.Value) meowrt.Value {
	if f, ok := path.(*meowrt.Furball); ok {
		return f
	}
	p, ok := path.(*meowrt.String)
	if !ok {
		return furball("snoop expects a String path, got %s", path.Type())
	}
	data, err := os.ReadFile(p.Val)
	if err != nil {
		return furball("%s", err)
	}
	return meowrt.NewString(strings.TrimRight(string(data), "\r\n"))
}

// Stalk reads a file line by line and returns a List of Strings.
func Stalk(path meowrt.Value) meowrt.Value {
	if f, ok := path.(*meowrt.Furball); ok {
		return f
	}
	p, ok := path.(*meowrt.String)
	if !ok {
		return furball("stalk expects a String path, got %s", path.Type())
	}
	f, err := os.Open(p.Val)
	if err != nil {
		return furball("%s", err)
	}
	defer f.Close()

	var lines []meowrt.Value
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, meowrt.NewString(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return furball("%s", err)
	}
	return meowrt.NewList(lines...)
}
