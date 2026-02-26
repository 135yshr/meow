//go:build js && wasm

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/135yshr/meow/pkg/checker"
	"github.com/135yshr/meow/pkg/interpreter"
	"github.com/135yshr/meow/pkg/lexer"
	"github.com/135yshr/meow/pkg/parser"
)

type result struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

func runMeow(_ js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		b, _ := json.Marshal(result{Error: "no source provided"})
		return string(b)
	}
	source := args[0].String()

	l := lexer.New(source, "playground.nyan")
	p := parser.New(l.Tokens())
	prog, parseErrs := p.Parse()
	if len(parseErrs) > 0 {
		msgs := make([]string, len(parseErrs))
		for i, e := range parseErrs {
			msgs[i] = e.Error()
		}
		b, _ := json.Marshal(result{Error: fmt.Sprintf("Parse error:\n%s", join(msgs))})
		return string(b)
	}

	c := checker.New()
	ti, checkErrs := c.Check(prog)
	if len(checkErrs) > 0 {
		msgs := make([]string, len(checkErrs))
		for i, e := range checkErrs {
			msgs[i] = e.Error()
		}
		b, _ := json.Marshal(result{Error: fmt.Sprintf("Type error:\n%s", join(msgs))})
		return string(b)
	}

	var buf bytes.Buffer
	interp := interpreter.New(&buf)
	interp.SetTypeInfo(ti)
	interp.SetStepLimit(10_000_000)
	if err := interp.RunSafe(prog); err != nil {
		b, _ := json.Marshal(result{Output: buf.String(), Error: err.Error()})
		return string(b)
	}

	b, _ := json.Marshal(result{Output: buf.String()})
	return string(b)
}

func join(strs []string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += "\n"
		}
		result += s
	}
	return result
}

func main() {
	js.Global().Set("runMeow", js.FuncOf(runMeow))
	select {}
}
