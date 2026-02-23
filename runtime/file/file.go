package file

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/135yshr/meow/runtime/meowrt"
)

// Snoop reads the entire contents of a file and returns it as a String.
func Snoop(path meowrt.Value) meowrt.Value {
	p, ok := path.(*meowrt.String)
	if !ok {
		panic(fmt.Sprintf("Hiss! snoop expects a String path, got %s, nya~", path.Type()))
	}
	data, err := os.ReadFile(p.Val)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	return meowrt.NewString(strings.TrimRight(string(data), "\n"))
}

// Stalk reads a file line by line and returns a List of Strings.
func Stalk(path meowrt.Value) meowrt.Value {
	p, ok := path.(*meowrt.String)
	if !ok {
		panic(fmt.Sprintf("Hiss! stalk expects a String path, got %s, nya~", path.Type()))
	}
	f, err := os.Open(p.Val)
	if err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	defer f.Close()

	var lines []meowrt.Value
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, meowrt.NewString(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		panic(fmt.Sprintf("Hiss! %s, nya~", err))
	}
	return meowrt.NewList(lines...)
}
