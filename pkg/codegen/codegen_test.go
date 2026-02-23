package codegen_test

import (
	"strings"
	"testing"

	"github.com/135yshr/meow/pkg/codegen"
	"github.com/135yshr/meow/pkg/lexer"
	"github.com/135yshr/meow/pkg/parser"
)

func generate(t *testing.T, input string) string {
	t.Helper()
	l := lexer.New(input, "test.nyan")
	p := parser.New(l.Tokens())
	prog, errs := p.Parse()
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e)
		}
		t.FailNow()
	}
	g := codegen.New()
	code, err := g.Generate(prog)
	if err != nil {
		t.Fatal(err)
	}
	return code
}

func TestHelloWorld(t *testing.T) {
	code := generate(t, `nyan name = "Tama"
nya(name)`)
	if !strings.Contains(code, `meow.NewString("Tama")`) {
		t.Error("expected meow.NewString")
	}
	if !strings.Contains(code, `meow.Nya(name)`) {
		t.Error("expected meow.Nya")
	}
}

func TestFuncGen(t *testing.T) {
	code := generate(t, `meow greet(who) {
  bring "Hello, " + who + "!"
}
nyan name = "Tama"
nya(greet(name))`)
	if !strings.Contains(code, "func greet(who meow.Value) meow.Value") {
		t.Error("expected function declaration")
	}
	if !strings.Contains(code, "meow.Add(") {
		t.Error("expected meow.Add")
	}
}

func TestArithmetic(t *testing.T) {
	code := generate(t, `nyan x = 1 + 2 * 3`)
	if !strings.Contains(code, "meow.Add(") {
		t.Error("expected meow.Add")
	}
	if !strings.Contains(code, "meow.Mul(") {
		t.Error("expected meow.Mul")
	}
}

func TestFetchAndMemberCall(t *testing.T) {
	code := generate(t, `fetch "file"
nyan content = file.snoop("data.txt")
nya(content)`)
	if !strings.Contains(code, `import meow_file "github.com/135yshr/meow/runtime/file"`) {
		t.Error("expected meow_file import")
	}
	if !strings.Contains(code, `meow_file.Snoop(meow.NewString("data.txt"))`) {
		t.Error("expected meow_file.Snoop call")
	}
	if !strings.Contains(code, `meow.Nya(content)`) {
		t.Error("expected meow.Nya")
	}
}

func TestIfElse(t *testing.T) {
	code := generate(t, `sniff (x > 0) {
  nya(x)
} scratch {
  nya(0)
}`)
	if !strings.Contains(code, "if (") {
		t.Error("expected if statement")
	}
	if !strings.Contains(code, "} else {") {
		t.Error("expected else clause")
	}
}
