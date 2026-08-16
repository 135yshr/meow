package interpreter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/135yshr/meow/pkg/checker"
	"github.com/135yshr/meow/pkg/lexer"
	"github.com/135yshr/meow/pkg/parser"
)

func runMeow(t *testing.T, source string) string {
	t.Helper()
	l := lexer.New(source, "test.nyan")
	p := parser.New(l.Tokens())
	prog, parseErrs := p.Parse()
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}

	c := checker.New()
	ti, checkErrs := c.Check(prog)
	if len(checkErrs) > 0 {
		t.Fatalf("checker errors: %v", checkErrs)
	}

	var buf bytes.Buffer
	interp := New(&buf)
	interp.SetTypeInfo(ti)
	if err := interp.RunSafe(prog); err != nil {
		t.Fatalf("runtime error: %v", err)
	}
	return buf.String()
}

func runMeowError(t *testing.T, source string) string {
	t.Helper()
	l := lexer.New(source, "test.nyan")
	p := parser.New(l.Tokens())
	prog, parseErrs := p.Parse()
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}

	c := checker.New()
	// Checker errors intentionally ignored — some error tests may have type issues
	ti, _ := c.Check(prog)

	var buf bytes.Buffer
	interp := New(&buf)
	interp.SetTypeInfo(ti)
	err := interp.RunSafe(prog)
	if err == nil {
		t.Fatal("expected error but got none")
	}
	return err.Error()
}

func TestHelloWorld(t *testing.T) {
	got := runMeow(t, `nya("Hello, World!")`)
	if got != "Hello, World!\n" {
		t.Errorf("got %q, want %q", got, "Hello, World!\n")
	}
}

func TestNyaMultipleArgs(t *testing.T) {
	got := runMeow(t, `nya("hello", "world")`)
	if got != "hello world\n" {
		t.Errorf("got %q, want %q", got, "hello world\n")
	}
}

func TestIntLiterals(t *testing.T) {
	got := runMeow(t, `nya(42)`)
	if got != "42\n" {
		t.Errorf("got %q, want %q", got, "42\n")
	}
}

func TestFloatLiterals(t *testing.T) {
	got := runMeow(t, `nya(3.14)`)
	if got != "3.14\n" {
		t.Errorf("got %q, want %q", got, "3.14\n")
	}
}

func TestBoolLiterals(t *testing.T) {
	got := runMeow(t, "nya(yarn)\nnya(hairball)")
	if got != "true\nfalse\n" {
		t.Errorf("got %q", got)
	}
}

func TestNilLiteral(t *testing.T) {
	got := runMeow(t, `nya(catnap)`)
	if got != "catnap\n" {
		t.Errorf("got %q, want %q", got, "catnap\n")
	}
}

func TestVariable(t *testing.T) {
	got := runMeow(t, `
nyan x = 10
nya(x)
`)
	if strings.TrimSpace(got) != "10" {
		t.Errorf("got %q", got)
	}
}

func TestArithmetic(t *testing.T) {
	got := runMeow(t, `
nyan a = 10
nyan b = 3
nya(a + b)
nya(a - b)
nya(a * b)
nya(a / b)
nya(a % b)
`)
	expected := "13\n7\n30\n3\n1\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestComparison(t *testing.T) {
	got := runMeow(t, `
nya(3 < 5)
nya(5 > 3)
nya(3 == 3)
nya(3 != 4)
nya(3 <= 3)
nya(3 >= 3)
`)
	expected := "true\ntrue\ntrue\ntrue\ntrue\ntrue\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestUnaryOps(t *testing.T) {
	got := runMeow(t, `
nya(-5)
nya(!yarn)
nya(!hairball)
`)
	expected := "-5\nfalse\ntrue\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestLogicalShortCircuit(t *testing.T) {
	got := runMeow(t, `
nya(yarn && hairball)
nya(hairball && yarn)
nya(yarn || hairball)
nya(hairball || yarn)
`)
	expected := "false\nfalse\ntrue\ntrue\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestIfStmt(t *testing.T) {
	got := runMeow(t, `
nyan x = 10
sniff (x > 5) {
    nya("big")
} scratch {
    nya("small")
}
`)
	if strings.TrimSpace(got) != "big" {
		t.Errorf("got %q", got)
	}
}

func TestIfElse(t *testing.T) {
	got := runMeow(t, `
nyan x = 2
sniff (x > 5) {
    nya("big")
} scratch {
    nya("small")
}
`)
	if strings.TrimSpace(got) != "small" {
		t.Errorf("got %q", got)
	}
}

func TestFunction(t *testing.T) {
	got := runMeow(t, `
meow add(a int, b int) int {
    bring a + b
}
nya(add(3, 4))
`)
	if strings.TrimSpace(got) != "7" {
		t.Errorf("got %q", got)
	}
}

func TestRecursion(t *testing.T) {
	got := runMeow(t, `
meow factorial(n int) int {
    sniff (n <= 1) {
        bring 1
    }
    bring n * factorial(n - 1)
}
nya(factorial(5))
`)
	if strings.TrimSpace(got) != "120" {
		t.Errorf("got %q", got)
	}
}

func TestFibonacci(t *testing.T) {
	got := runMeow(t, `
meow fib(n int) int {
    sniff (n <= 1) {
        bring n
    }
    bring fib(n - 1) + fib(n - 2)
}
nya(fib(10))
`)
	if strings.TrimSpace(got) != "55" {
		t.Errorf("got %q", got)
	}
}

func TestLambda(t *testing.T) {
	got := runMeow(t, `
nyan double = paw(x int) { x * 2 }
nya(double(5))
`)
	if strings.TrimSpace(got) != "10" {
		t.Errorf("got %q", got)
	}
}

func TestClosure(t *testing.T) {
	got := runMeow(t, `
nyan count = 0
nyan c = paw() { count + 1 }
nya(c())
`)
	if strings.TrimSpace(got) != "1" {
		t.Errorf("got %q", got)
	}
}

func TestList(t *testing.T) {
	got := runMeow(t, `
nyan xs = [1, 2, 3]
nya(xs)
nya(xs[0])
nya(xs[2])
nya(len(xs))
`)
	expected := "[1, 2, 3]\n1\n3\n3\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestListOps(t *testing.T) {
	got := runMeow(t, `
nyan xs = [1, 2, 3, 4, 5]
nyan doubled = lick(xs, paw(x int) { x * 2 })
nya(doubled)
nyan evens = picky(xs, paw(x int) { x % 2 == 0 })
nya(evens)
nyan sum = curl(xs, 0, paw(acc int, x int) { acc + x })
nya(sum)
`)
	expected := "[2, 4, 6, 8, 10]\n[2, 4]\n15\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestHeadTail(t *testing.T) {
	got := runMeow(t, `
nyan xs = [10, 20, 30]
nya(head(xs))
nya(tail(xs))
`)
	expected := "10\n[20, 30]\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestAppendList(t *testing.T) {
	got := runMeow(t, `
nyan xs = [1, 2]
nyan ys = append(xs, 3)
nya(ys)
`)
	if strings.TrimSpace(got) != "[1, 2, 3]" {
		t.Errorf("got %q", got)
	}
}

func TestRangeCount(t *testing.T) {
	got := runMeow(t, `
purr i (5) {
    nya(i)
}
`)
	expected := "0\n1\n2\n3\n4\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestRangeInclusive(t *testing.T) {
	got := runMeow(t, `
purr i (1..3) {
    nya(i)
}
`)
	expected := "1\n2\n3\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestRangeString(t *testing.T) {
	got := runMeow(t, `
purr ch (to_runes("meow")) {
    nya(ch)
}
`)
	expected := "m\ne\no\nw\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestRangeStringWithIndex(t *testing.T) {
	got := runMeow(t, `
purr i, ch (to_runes("nya")) {
    nya(i)
    nya(ch)
}
`)
	expected := "0\nn\n1\ny\n2\na\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestRangeStringMultibyte(t *testing.T) {
	got := runMeow(t, `
purr i, ch (to_runes("にゃん")) {
    nya(i)
    nya(ch)
}
`)
	expected := "0\nに\n1\nゃ\n2\nん\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestStringConcat(t *testing.T) {
	got := runMeow(t, `
nyan a = "hello"
nyan b = " world"
nya(a + b)
`)
	if strings.TrimSpace(got) != "hello world" {
		t.Errorf("got %q", got)
	}
}

func TestToIntToFloat(t *testing.T) {
	got := runMeow(t, `
nya(to_int(3.14))
nya(to_float(42))
nya(to_string(123))
`)
	expected := "3\n42\n123\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestPipe(t *testing.T) {
	got := runMeow(t, `
nyan xs = [1, 2, 3]
nyan result = xs |=| lick(paw(x int) { x * 10 })
nya(result)
`)
	if strings.TrimSpace(got) != "[10, 20, 30]" {
		t.Errorf("got %q", got)
	}
}

func TestCatch(t *testing.T) {
	got := runMeow(t, `
nyan result = hiss("boom") ~> 42
nya(result)
`)
	if strings.TrimSpace(got) != "42" {
		t.Errorf("got %q", got)
	}
}

func TestCatchWithFunction(t *testing.T) {
	got := runMeow(t, `
nyan result = hiss("error") ~> paw(err string) { "caught" }
nya(result)
`)
	if strings.TrimSpace(got) != "caught" {
		t.Errorf("got %q", got)
	}
}

func TestIsFurball(t *testing.T) {
	got := runMeow(t, `
nyan result = gag(paw() { hiss("oops") })
nya(is_furball(result))
`)
	if strings.TrimSpace(got) != "true" {
		t.Errorf("got %q", got)
	}
}

func TestMatch(t *testing.T) {
	got := runMeow(t, `
nyan x = 3
nyan result = peek(x) {
    1 => "one"
    2 => "two"
    3 => "three"
    _ => "other"
}
nya(result)
`)
	if strings.TrimSpace(got) != "three" {
		t.Errorf("got %q", got)
	}
}

func TestMatchRange(t *testing.T) {
	got := runMeow(t, `
nyan x = 15
nyan result = peek(x) {
    1..10 => "small"
    11..20 => "medium"
    _ => "large"
}
nya(result)
`)
	if strings.TrimSpace(got) != "medium" {
		t.Errorf("got %q", got)
	}
}

func TestMatchWildcard(t *testing.T) {
	got := runMeow(t, `
nyan x = 999
nyan result = peek(x) {
    1 => "one"
    _ => "other"
}
nya(result)
`)
	if strings.TrimSpace(got) != "other" {
		t.Errorf("got %q", got)
	}
}

func TestKitty(t *testing.T) {
	got := runMeow(t, `
kitty Nyantyu {
    name: string
    age: int
}
nyan c = Nyantyu("Nyantyu", 3)
nya(c.name)
nya(c.age)
`)
	expected := "Nyantyu\n3\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestCollar(t *testing.T) {
	got := runMeow(t, `
collar UserId = int
nyan id = UserId(42)
nya(id.value)
`)
	if strings.TrimSpace(got) != "42" {
		t.Errorf("got %q", got)
	}
}

func TestGroom(t *testing.T) {
	got := runMeow(t, `
kitty Nyantyu {
    name: string
    age: int
}

groom Nyantyu {
    meow greet() string {
        bring "I am " + self.name
    }
}

nyan c = Nyantyu("Nyantyu", 3)
nya(c.greet())
`)
	if strings.TrimSpace(got) != "I am Nyantyu" {
		t.Errorf("got %q", got)
	}
}

func TestGroomWithParams(t *testing.T) {
	got := runMeow(t, `
kitty Counter {
    val: int
}

groom Counter {
    meow add(n int) int {
        bring self.val + n
    }
}

nyan c = Counter(10)
nya(c.add(5))
`)
	if strings.TrimSpace(got) != "15" {
		t.Errorf("got %q", got)
	}
}

func TestHissError(t *testing.T) {
	errMsg := runMeowError(t, `hiss("something went wrong")`)
	if !strings.Contains(errMsg, "Hiss!") {
		t.Errorf("expected Hiss! error, got %q", errMsg)
	}
}

func TestFetchUnsupported(t *testing.T) {
	errMsg := runMeowError(t, `nab "file"`)
	if !strings.Contains(errMsg, "not supported") {
		t.Errorf("expected unsupported error, got %q", errMsg)
	}
}

func TestStepLimit(t *testing.T) {
	l := lexer.New(`
meow loop() {
    loop()
}
loop()
`, "test.nyan")
	p := parser.New(l.Tokens())
	prog, _ := p.Parse()

	c := checker.New()
	ti, _ := c.Check(prog)

	var buf bytes.Buffer
	interp := New(&buf)
	interp.SetTypeInfo(ti)
	interp.SetStepLimit(1000)
	err := interp.RunSafe(prog)
	if err == nil {
		t.Fatal("expected step limit error")
	}
	if !strings.Contains(err.Error(), "step limit") {
		t.Errorf("expected step limit error, got %q", err.Error())
	}
}

func TestFizzBuzz(t *testing.T) {
	got := runMeow(t, `
purr i (1..15) {
    sniff (i % 15 == 0) {
        nya("FizzBuzz")
    } scratch {
        sniff (i % 3 == 0) {
            nya("Fizz")
        } scratch {
            sniff (i % 5 == 0) {
                nya("Buzz")
            } scratch {
                nya(i)
            }
        }
    }
}
`)
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 15 {
		t.Fatalf("expected 15 lines, got %d: %q", len(lines), got)
	}
	if lines[0] != "1" {
		t.Errorf("line 1: got %q", lines[0])
	}
	if lines[2] != "Fizz" {
		t.Errorf("line 3: got %q", lines[2])
	}
	if lines[4] != "Buzz" {
		t.Errorf("line 5: got %q", lines[4])
	}
	if lines[14] != "FizzBuzz" {
		t.Errorf("line 15: got %q", lines[14])
	}
}

func TestMapLiteral(t *testing.T) {
	got := runMeow(t, `
nyan m = {"name": "Nyantyu", "color": "white"}
nya(m["name"])
nya(m["color"])
`)
	expected := "Nyantyu\nwhite\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestFirstClassFunction(t *testing.T) {
	got := runMeow(t, `
nyan double = paw(n int) { n * 2 }
nyan result = double(5)
nya(result)
`)
	if strings.TrimSpace(got) != "10" {
		t.Errorf("got %q", got)
	}
}

func TestNestedFunction(t *testing.T) {
	got := runMeow(t, `
meow outer(x int) int {
    meow inner(y int) int {
        bring x + y
    }
    bring inner(10)
}
nya(outer(5))
`)
	if strings.TrimSpace(got) != "15" {
		t.Errorf("got %q", got)
	}
}

func TestPipeChain(t *testing.T) {
	got := runMeow(t, `
nyan xs = [1, 2, 3, 4, 5]
nyan result = xs |=| picky(paw(x int) { x > 2 }) |=| lick(paw(x int) { x * 10 })
nya(result)
`)
	if strings.TrimSpace(got) != "[30, 40, 50]" {
		t.Errorf("got %q", got)
	}
}

// Indexing is shared with the transpiler through meowrt.Index, so the
// interpreter that backs the WASM playground must agree with compiled output.
func TestIndexLiteralSubscript(t *testing.T) {
	got := runMeow(t, `
nyan nums = [10, 20, 30]
nya(nums[0])
nya(nums[2])
nyan i = 1
nya(nums[i])
nya(nums[i + 1])
`)
	expected := "10\n30\n20\n30\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestIndexChained(t *testing.T) {
	got := runMeow(t, `
nyan data = {"items": [1, 2, 3], "inner": {"value": 7}}
nya(data["items"][1])
nya(data["inner"]["value"])
nyan grid = [[1, 2], [3, 4]]
nya(grid[1][0])
`)
	expected := "2\n7\n3\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestIndexMapMissingKey(t *testing.T) {
	got := runMeow(t, `
nyan m = {"name": "Nyantyu"}
nya(m["collar"])
`)
	if got != "catnap\n" {
		t.Errorf("got %q, want %q", got, "catnap\n")
	}
}

// Lambda bodies may be statement blocks, not just single expressions.
func TestLambdaBlockBody(t *testing.T) {
	got := runMeow(t, `
nyan classify = paw(n) { sniff (n > 10) { bring "big" } scratch { bring "small" } }
nya(classify(5))
nya(classify(50))
`)
	if got != "small\nbig\n" {
		t.Errorf("got %q, want %q", got, "small\nbig\n")
	}
}

// A trailing expression statement is the result, the same way the
// single-expression form yields its value.
func TestLambdaBlockTrailingExpr(t *testing.T) {
	got := runMeow(t, `
nyan area = paw(w, h) {
  nyan a = w * h
  a + 1
}
nya(area(3, 4))
`)
	if strings.TrimSpace(got) != "13" {
		t.Errorf("got %q, want %q", got, "13")
	}
}

func TestLambdaSingleExpressionBodyUnchanged(t *testing.T) {
	got := runMeow(t, `
nyan double = paw(x int) { x * 2 }
nya(double(21))
`)
	if strings.TrimSpace(got) != "42" {
		t.Errorf("got %q, want %q", got, "42")
	}
}

// String builtins must behave the same on the interpreter that backs the
// WASM playground as they do in compiled output.
func TestStringBuiltins(t *testing.T) {
	got := runMeow(t, `
nyan s = "hello,world,meow"
nya(whiff(s, "world"))
nya(track(s, "world"))
nya(track(s, "dog"))
nya(shred(s, ","))
nya(tangle(shred(s, ","), " / "))
nya(nibble(s, 0, 5))
nya(nibble(s, -4, 16))
`)
	expected := "true\n6\n-1\n[hello, world, meow]\nhello / world / meow\nhello\nmeow\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestStringConversionsRoundTrip(t *testing.T) {
	got := runMeow(t, `
nya(to_string(to_bytes("hello")))
nya(tangle(to_runes("hello"), ""))
`)
	if got != "hello\nhello\n" {
		t.Errorf("got %q, want %q", got, "hello\nhello\n")
	}
}

// paw() {} has an empty block body, which must yield catnap rather than being
// mistaken for an expression body.
func TestLambdaEmptyBlockBody(t *testing.T) {
	got := runMeow(t, `
nyan nothing = paw() { }
nya(nothing())
`)
	if got != "catnap\n" {
		t.Errorf("got %q, want %q", got, "catnap\n")
	}
}

// A lambda parameter shadows an outer binding of the same name.
func TestLambdaParamShadowsOuter(t *testing.T) {
	got := runMeow(t, `
nyan x = 100
nyan bump = paw(x) { x + 1 }
nya(bump(5))
nya(x)
`)
	if got != "6\n100\n" {
		t.Errorf("got %q, want %q", got, "6\n100\n")
	}
}

// The body need not mention the loop variable. Compiled output used to fail
// with "declared and not used" for the count form.
func TestRangeUnusedLoopVariable(t *testing.T) {
	got := runMeow(t, `
purr unused (2) {
  nya("tick")
}
`)
	if got != "tick\ntick\n" {
		t.Errorf("got %q, want %q", got, "tick\ntick\n")
	}
}

// A list-form purr inside a fully-typed function. Compiled output used to fail
// to build, because the body was generated by the untyped statement generator
// whose Furball short-circuit returns a value the typed function cannot.
func TestTypedFunctionListRange(t *testing.T) {
	got := runMeow(t, `
meow each(n int) int {
  purr item (["a", "b"]) {
    nya(item)
  }
  bring n
}
nya(each(0))
`)
	if got != "a\nb\n0\n" {
		t.Errorf("got %q, want %q", got, "a\nb\n0\n")
	}
}

// The element is held as a meow.Value even where the checker knows its element
// type, so using it as a native must unbox rather than box it again.
func TestTypedFunctionListRangeElementAsNative(t *testing.T) {
	got := runMeow(t, `
meow suffix(sep string) string {
  purr word (["a", "b"]) {
    nya(word + sep)
  }
  bring sep
}
nya(suffix("-"))
`)
	if got != "a-\nb-\n-\n" {
		t.Errorf("got %q, want %q", got, "a-\nb-\n-\n")
	}
}

func TestTypedFunctionListRangeWithIndex(t *testing.T) {
	got := runMeow(t, `
meow indexed(n int) int {
  purr i, word (["x", "y"]) {
    nya(i)
    nya(word)
  }
  bring n
}
nya(indexed(7))
`)
	if got != "0\nx\n1\ny\n7\n" {
		t.Errorf("got %q, want %q", got, "0\nx\n1\ny\n7\n")
	}
}

// The interpreter has always decided a purr's form at run time. These lock that
// in alongside the compiler, which used to decide it statically and guess wrong
// whenever the subject's type was unknown.
func TestRangeOverUntypedListExpressions(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{"append", `purr w (append(["a"], "b")) { nya(w) }`, "a\nb\n"},
		{"lick", `purr w (lick(["a"], paw(x) { x + x })) { nya(w) }`, "aa\n"},
		{"picky", `purr w (picky(["a", "bb"], paw(x) { len(x) > 1 })) { nya(w) }`, "bb\n"},
		{"tail", `purr w (tail(["x", "y"])) { nya(w) }`, "y\n"},
		{"litter inside a map", `nyan c = {"h": ["p", "q"]}
purr w (c["h"]) { nya(w) }`, "p\nq\n"},
		{"number inside a map", `nyan c = {"n": 2}
purr i (c["n"]) { nya(i) }`, "0\n1\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runMeow(t, tt.src); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// Escapes are decoded by the parser, so both backends see the same string.
func TestStringEscapesAreDecoded(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{"newline", `nya("a\nb")`, "a\nb\n"},
		{"tab length", `nya(len("\t"))`, "1\n"},
		{"double quote", `nya("{\"n\": 1}")`, "{\"n\": 1}\n"},
		{"backslash", `nya("C:\\tmp")`, "C:\\tmp\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runMeow(t, tt.src); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// Top-level bindings are visible inside functions. The interpreter has always
// put them in the global environment; the compiler used to make them locals of
// the wrapper that runs the program, so a function could not see them at all.
func TestGlobalBindingsVisibleInFunctions(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{"read a global", `nyan limit = 5
meow under(n int) bool { bring n < limit }
nya(under(3))`, "true\n"},
		{"parameter shadows the global", `nyan limit = 5
meow shadow(limit int) int { bring limit }
nya(shadow(99))`, "99\n"},
		{"walk a global litter", `nyan hosts = ["h1", "h2"]
meow show() int {
  purr h (hosts) { nya(h) }
  bring len(hosts)
}
nya(show())`, "h1\nh2\n2\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runMeow(t, tt.src); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// Reaching a top-level binding from a function that runs before the binding
// does is a mistake, and both backends have to name it the same way.
func TestGlobalBindingBeforeItIsBound(t *testing.T) {
	got := runMeowError(t, `meow f() int { bring to_int(limit) }
nya(f())
nyan limit = 5`)
	if !strings.Contains(got, "undefined variable limit") {
		t.Errorf("got %q, want it to name the unbound variable", got)
	}
}

// `&&` and `||` weigh their operands by truthiness and yield one of them, so an
// operand whose type is unknown has to stay boxed. Reading it as the other
// side's Go type demanded a bool of it, and the compiler answered
// "expected bool but got String" where the interpreter answered the flag.
func TestLogicalOperandsKeepTruthiness(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{"string on the left of and", `meow either(flag bool) bool {
  nyan m = {"x": "yes"}
  bring m["x"] && flag
}
nya(either(yarn))`, "true\n"},
		// `&&` yields the left operand when it is falsy, whatever its type, so
		// this is asserted where no declared return type is in the way.
		{"and yields a falsy left operand", `nyan m = {"x": ""}
nya(m["x"] && yarn)`, "\n"},
		{"empty string on the left of or", `meow fallback(flag bool) bool {
  nyan m = {"x": ""}
  bring m["x"] || flag
}
nya(fallback(yarn))`, "true\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runMeow(t, tt.src); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
