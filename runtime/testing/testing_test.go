package meowtest_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/135yshr/meow/runtime/meowrt"
	meowtest "github.com/135yshr/meow/runtime/testing"
)

func setup(t *testing.T) (*bytes.Buffer, *int) {
	t.Helper()
	var buf bytes.Buffer
	var exitCode int
	meowtest.Reset(&buf, func(code int) { exitCode = code })
	return &buf, &exitCode
}

func TestJudgePass(t *testing.T) {
	setup(t)
	result := meowtest.Judge(meowrt.NewBool(true))
	if _, ok := result.(*meowrt.NilValue); !ok {
		t.Errorf("expected NilValue, got %T", result)
	}
}

func TestJudgeTruthyInt(t *testing.T) {
	setup(t)
	result := meowtest.Judge(meowrt.NewInt(42))
	if _, ok := result.(*meowrt.NilValue); !ok {
		t.Errorf("expected NilValue, got %T", result)
	}
}

func TestJudgeFail(t *testing.T) {
	setup(t)
	v := meowtest.Judge(meowrt.NewBool(false))
	if _, ok := v.(*meowrt.Furball); !ok {
		t.Fatalf("expected Furball, got %T", v)
	}
}

func TestJudgeFailWithMessage(t *testing.T) {
	setup(t)
	v := meowtest.Judge(meowrt.NewBool(false), meowrt.NewString("custom message"))
	f, ok := v.(*meowrt.Furball)
	if !ok {
		t.Fatalf("expected Furball, got %T", v)
	}
	if !strings.Contains(f.Message, "custom message") {
		t.Errorf("expected custom message in Furball, got %q", f.Message)
	}
}

func TestJudgeNoArgs(t *testing.T) {
	setup(t)
	v := meowtest.Judge()
	f, ok := v.(*meowrt.Furball)
	if !ok {
		t.Fatalf("expected Furball, got %T", v)
	}
	if !strings.Contains(f.Message, "Hiss!") {
		t.Errorf("expected Hiss! in Furball, got: %s", f.Message)
	}
}

func TestExpectPass(t *testing.T) {
	setup(t)
	result := meowtest.Expect(meowrt.NewInt(42), meowrt.NewInt(42))
	if _, ok := result.(*meowrt.NilValue); !ok {
		t.Errorf("expected NilValue, got %T", result)
	}
}

func TestExpectPassString(t *testing.T) {
	setup(t)
	result := meowtest.Expect(meowrt.NewString("hello"), meowrt.NewString("hello"))
	if _, ok := result.(*meowrt.NilValue); !ok {
		t.Errorf("expected NilValue, got %T", result)
	}
}

func TestExpectFail(t *testing.T) {
	setup(t)
	v := meowtest.Expect(meowrt.NewInt(1), meowrt.NewInt(2))
	if _, ok := v.(*meowrt.Furball); !ok {
		t.Fatalf("expected Furball, got %T", v)
	}
}

func TestExpectFailWithMessage(t *testing.T) {
	setup(t)
	v := meowtest.Expect(meowrt.NewInt(1), meowrt.NewInt(2), meowrt.NewString("values differ"))
	f, ok := v.(*meowrt.Furball)
	if !ok {
		t.Fatalf("expected Furball, got %T", v)
	}
	if !strings.Contains(f.Message, "values differ") {
		t.Errorf("expected custom message, got %q", f.Message)
	}
}

func TestExpectNoArgs(t *testing.T) {
	setup(t)
	v := meowtest.Expect(meowrt.NewInt(1))
	f, ok := v.(*meowrt.Furball)
	if !ok {
		t.Fatalf("expected Furball, got %T", v)
	}
	if !strings.Contains(f.Message, "Hiss!") {
		t.Errorf("expected Hiss! in Furball, got: %s", f.Message)
	}
}

func TestRefusePass(t *testing.T) {
	setup(t)
	result := meowtest.Refuse(meowrt.NewBool(false))
	if _, ok := result.(*meowrt.NilValue); !ok {
		t.Errorf("expected NilValue, got %T", result)
	}
}

func TestRefusePassNil(t *testing.T) {
	setup(t)
	result := meowtest.Refuse(meowrt.NewNil())
	if _, ok := result.(*meowrt.NilValue); !ok {
		t.Errorf("expected NilValue, got %T", result)
	}
}

func TestRefuseFail(t *testing.T) {
	setup(t)
	v := meowtest.Refuse(meowrt.NewBool(true))
	if _, ok := v.(*meowrt.Furball); !ok {
		t.Fatalf("expected Furball, got %T", v)
	}
}

func TestRunPass(t *testing.T) {
	buf, _ := setup(t)
	fn := meowrt.NewFunc("test", func(args ...meowrt.Value) meowrt.Value {
		return meowrt.NewNil()
	})
	result := meowtest.Run(meowrt.NewString("test_example"), fn)
	b, ok := result.(*meowrt.Bool)
	if !ok || !b.Val {
		t.Errorf("expected true, got %v", result)
	}
	if !strings.Contains(buf.String(), "PASS: test_example") {
		t.Errorf("expected PASS output, got: %s", buf.String())
	}
}

func TestRunFailAssertion(t *testing.T) {
	buf, _ := setup(t)
	// In the new model, an assertion returns a Furball that the surrounding
	// (generated) function would short-circuit-return. The hand-written test
	// fn mirrors that by returning the assertion's Furball directly.
	fn := meowrt.NewFunc("test", func(args ...meowrt.Value) meowrt.Value {
		if f, ok := meowtest.Judge(meowrt.NewBool(false)).(*meowrt.Furball); ok {
			return f
		}
		return meowrt.NewNil()
	})
	result := meowtest.Run(meowrt.NewString("test_fail"), fn)
	b, ok := result.(*meowrt.Bool)
	if !ok || b.Val {
		t.Errorf("expected false, got %v", result)
	}
	if !strings.Contains(buf.String(), "FAIL: test_fail") {
		t.Errorf("expected FAIL output, got: %s", buf.String())
	}
}

func TestRunCatchesHiss(t *testing.T) {
	buf, _ := setup(t)
	fn := meowrt.NewFunc("test", func(args ...meowrt.Value) meowrt.Value {
		panic("Hiss! something broke, nya~")
	})
	result := meowtest.Run(meowrt.NewString("test_hiss"), fn)
	b, ok := result.(*meowrt.Bool)
	if !ok || b.Val {
		t.Errorf("expected false, got %v", result)
	}
	out := buf.String()
	if !strings.Contains(out, "FAIL: test_hiss") {
		t.Errorf("expected FAIL output, got: %s", out)
	}
	if !strings.Contains(out, "Hiss!") {
		t.Errorf("expected Hiss! in failure message, got: %s", out)
	}
}

func TestRunMultipleAccumulates(t *testing.T) {
	buf, _ := setup(t)
	for i := 0; i < 3; i++ {
		fn := meowrt.NewFunc("test", func(args ...meowrt.Value) meowrt.Value {
			return meowrt.NewNil()
		})
		meowtest.Run(meowrt.NewString(fmt.Sprintf("test_%d", i)), fn)
	}
	out := buf.String()
	if strings.Count(out, "PASS") != 3 {
		t.Errorf("expected 3 PASS lines, got: %s", out)
	}
}

func TestReportAllPass(t *testing.T) {
	buf, exitCode := setup(t)
	fn := meowrt.NewFunc("test", func(args ...meowrt.Value) meowrt.Value {
		return meowrt.NewNil()
	})
	meowtest.Run(meowrt.NewString("test_a"), fn)
	meowtest.Run(meowrt.NewString("test_b"), fn)
	meowtest.Report()
	if *exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", *exitCode)
	}
	out := buf.String()
	if !strings.Contains(out, "All 2 tests passed, nya~!") {
		t.Errorf("expected all-pass message, got: %s", out)
	}
}

func TestReportWithFailures(t *testing.T) {
	buf, exitCode := setup(t)
	passFn := meowrt.NewFunc("test", func(args ...meowrt.Value) meowrt.Value {
		return meowrt.NewNil()
	})
	failFn := meowrt.NewFunc("test", func(args ...meowrt.Value) meowrt.Value {
		if f, ok := meowtest.Judge(meowrt.NewBool(false)).(*meowrt.Furball); ok {
			return f
		}
		return meowrt.NewNil()
	})
	meowtest.Run(meowrt.NewString("test_pass"), passFn)
	meowtest.Run(meowrt.NewString("test_fail"), failFn)
	meowtest.Report()
	if *exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", *exitCode)
	}
	out := buf.String()
	if !strings.Contains(out, "1 passed, 1 failed, nya~") {
		t.Errorf("expected failure summary, got: %s", out)
	}
}

func TestRunBadArgs(t *testing.T) {
	setup(t)
	v := meowtest.Run(meowrt.NewInt(1), meowrt.NewInt(2))
	f, ok := v.(*meowrt.Furball)
	if !ok {
		t.Fatalf("expected Furball, got %T", v)
	}
	if !strings.Contains(f.Message, "Hiss!") {
		t.Errorf("expected Hiss! in Furball, got: %s", f.Message)
	}
}

func TestCatwalkPass(t *testing.T) {
	buf, _ := setup(t)
	fn := meowrt.NewFunc("test", func(args ...meowrt.Value) meowrt.Value {
		fmt.Print("Hello, Nyantyu!\n")
		return meowrt.NewNil()
	})
	result := meowtest.Catwalk(
		meowrt.NewString("catwalk_hello"),
		fn,
		meowrt.NewString("Hello, Nyantyu!\n"),
	)
	b, ok := result.(*meowrt.Bool)
	if !ok || !b.Val {
		t.Errorf("expected true, got %v", result)
	}
	if !strings.Contains(buf.String(), "PASS: catwalk_hello") {
		t.Errorf("expected PASS output, got: %s", buf.String())
	}
}

func TestCatwalkFail(t *testing.T) {
	buf, _ := setup(t)
	fn := meowrt.NewFunc("test", func(args ...meowrt.Value) meowrt.Value {
		fmt.Print("wrong output\n")
		return meowrt.NewNil()
	})
	result := meowtest.Catwalk(
		meowrt.NewString("catwalk_mismatch"),
		fn,
		meowrt.NewString("expected output\n"),
	)
	b, ok := result.(*meowrt.Bool)
	if !ok || b.Val {
		t.Errorf("expected false, got %v", result)
	}
	out := buf.String()
	if !strings.Contains(out, "FAIL: catwalk_mismatch") {
		t.Errorf("expected FAIL output, got: %s", out)
	}
	if !strings.Contains(out, "output mismatch") {
		t.Errorf("expected output mismatch message, got: %s", out)
	}
}

func TestCatwalkPanic(t *testing.T) {
	buf, _ := setup(t)
	fn := meowrt.NewFunc("test", func(args ...meowrt.Value) meowrt.Value {
		panic("Hiss! something broke, nya~")
	})
	result := meowtest.Catwalk(
		meowrt.NewString("catwalk_panic"),
		fn,
		meowrt.NewString("anything\n"),
	)
	b, ok := result.(*meowrt.Bool)
	if !ok || b.Val {
		t.Errorf("expected false, got %v", result)
	}
	out := buf.String()
	if !strings.Contains(out, "FAIL: catwalk_panic") {
		t.Errorf("expected FAIL output, got: %s", out)
	}
}
