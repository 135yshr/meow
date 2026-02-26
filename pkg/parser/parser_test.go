package parser_test

import (
	"testing"

	"github.com/135yshr/meow/pkg/ast"
	"github.com/135yshr/meow/pkg/lexer"
	"github.com/135yshr/meow/pkg/parser"
)

func parse(t *testing.T, input string) *ast.Program {
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
	return prog
}

func TestVarStmt(t *testing.T) {
	prog := parse(t, `nyan x = 42`)
	if len(prog.Stmts) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(prog.Stmts))
	}
	v, ok := prog.Stmts[0].(*ast.VarStmt)
	if !ok {
		t.Fatalf("expected VarStmt, got %T", prog.Stmts[0])
	}
	if v.Name != "x" {
		t.Errorf("expected name 'x', got %q", v.Name)
	}
	lit, ok := v.Value.(*ast.IntLit)
	if !ok {
		t.Fatalf("expected IntLit, got %T", v.Value)
	}
	if lit.Value != 42 {
		t.Errorf("expected 42, got %d", lit.Value)
	}
}

func TestFuncStmt(t *testing.T) {
	prog := parse(t, `meow greet(name) {
  bring "Hello, " + name
}`)
	if len(prog.Stmts) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(prog.Stmts))
	}
	fn, ok := prog.Stmts[0].(*ast.FuncStmt)
	if !ok {
		t.Fatalf("expected FuncStmt, got %T", prog.Stmts[0])
	}
	if fn.Name != "greet" {
		t.Errorf("expected name 'greet', got %q", fn.Name)
	}
	if len(fn.Params) != 1 || fn.Params[0].Name != "name" {
		t.Errorf("expected params [name], got %v", fn.Params)
	}
}

func TestIfStmt(t *testing.T) {
	prog := parse(t, `sniff (x > 0) {
  nya(x)
} scratch {
  nya(0)
}`)
	if len(prog.Stmts) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(prog.Stmts))
	}
	ifStmt, ok := prog.Stmts[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", prog.Stmts[0])
	}
	if len(ifStmt.Body) != 1 {
		t.Errorf("expected 1 body stmt, got %d", len(ifStmt.Body))
	}
	if len(ifStmt.ElseBody) != 1 {
		t.Errorf("expected 1 else stmt, got %d", len(ifStmt.ElseBody))
	}
}

func TestPurrCountForm(t *testing.T) {
	prog := parse(t, `purr i (10) {
  nya(i)
}`)
	if len(prog.Stmts) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(prog.Stmts))
	}
	rs, ok := prog.Stmts[0].(*ast.RangeStmt)
	if !ok {
		t.Fatalf("expected RangeStmt, got %T", prog.Stmts[0])
	}
	if rs.Var != "i" {
		t.Errorf("expected var 'i', got %q", rs.Var)
	}
	if rs.Start != nil {
		t.Errorf("expected nil Start for count form, got %T", rs.Start)
	}
	if rs.Inclusive {
		t.Error("expected Inclusive=false for count form")
	}
}

func TestPurrRangeForm(t *testing.T) {
	prog := parse(t, `purr i (1..20) {
  nya(i)
}`)
	if len(prog.Stmts) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(prog.Stmts))
	}
	rs, ok := prog.Stmts[0].(*ast.RangeStmt)
	if !ok {
		t.Fatalf("expected RangeStmt, got %T", prog.Stmts[0])
	}
	if rs.Var != "i" {
		t.Errorf("expected var 'i', got %q", rs.Var)
	}
	if rs.Start == nil {
		t.Fatal("expected non-nil Start for range form")
	}
	if !rs.Inclusive {
		t.Error("expected Inclusive=true for range form")
	}
}

func TestImplicitVarStmt(t *testing.T) {
	prog := parse(t, `x = 42`)
	if len(prog.Stmts) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(prog.Stmts))
	}
	v, ok := prog.Stmts[0].(*ast.VarStmt)
	if !ok {
		t.Fatalf("expected VarStmt, got %T", prog.Stmts[0])
	}
	if v.Name != "x" {
		t.Errorf("expected name 'x', got %q", v.Name)
	}
	lit, ok := v.Value.(*ast.IntLit)
	if !ok {
		t.Fatalf("expected IntLit, got %T", v.Value)
	}
	if lit.Value != 42 {
		t.Errorf("expected 42, got %d", lit.Value)
	}
}

func TestLambda(t *testing.T) {
	prog := parse(t, `nyan double = paw(x) { x * 2 }`)
	if len(prog.Stmts) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(prog.Stmts))
	}
	v := prog.Stmts[0].(*ast.VarStmt)
	_, ok := v.Value.(*ast.LambdaExpr)
	if !ok {
		t.Fatalf("expected LambdaExpr, got %T", v.Value)
	}
}

func TestListLit(t *testing.T) {
	prog := parse(t, `nyan xs = [1, 2, 3]`)
	v := prog.Stmts[0].(*ast.VarStmt)
	list, ok := v.Value.(*ast.ListLit)
	if !ok {
		t.Fatalf("expected ListLit, got %T", v.Value)
	}
	if len(list.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(list.Items))
	}
}

func TestPipeExpr(t *testing.T) {
	prog := parse(t, `xs |=| lick(double)`)
	stmt := prog.Stmts[0].(*ast.ExprStmt)
	_, ok := stmt.Expr.(*ast.PipeExpr)
	if !ok {
		t.Fatalf("expected PipeExpr, got %T", stmt.Expr)
	}
}

func TestArithmetic(t *testing.T) {
	prog := parse(t, `nyan x = 1 + 2 * 3`)
	v := prog.Stmts[0].(*ast.VarStmt)
	bin, ok := v.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", v.Value)
	}
	// Should be 1 + (2 * 3) due to precedence
	if _, ok := bin.Right.(*ast.BinaryExpr); !ok {
		t.Errorf("expected right side to be BinaryExpr (mul), got %T", bin.Right)
	}
}

func TestNyaCall(t *testing.T) {
	prog := parse(t, `nya("Hello")`)
	stmt := prog.Stmts[0].(*ast.ExprStmt)
	call, ok := stmt.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", stmt.Expr)
	}
	ident := call.Fn.(*ast.Ident)
	if ident.Name != "nya" {
		t.Errorf("expected 'nya', got %q", ident.Name)
	}
}

func TestFetchStmt(t *testing.T) {
	prog := parse(t, `fetch "file"`)
	if len(prog.Stmts) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(prog.Stmts))
	}
	f, ok := prog.Stmts[0].(*ast.FetchStmt)
	if !ok {
		t.Fatalf("expected FetchStmt, got %T", prog.Stmts[0])
	}
	if f.Path != "file" {
		t.Errorf("expected path 'file', got %q", f.Path)
	}
}

func TestMemberAccess(t *testing.T) {
	prog := parse(t, `nyan content = file.snoop("data.txt")`)
	if len(prog.Stmts) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(prog.Stmts))
	}
	v, ok := prog.Stmts[0].(*ast.VarStmt)
	if !ok {
		t.Fatalf("expected VarStmt, got %T", prog.Stmts[0])
	}
	call, ok := v.Value.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", v.Value)
	}
	member, ok := call.Fn.(*ast.MemberExpr)
	if !ok {
		t.Fatalf("expected MemberExpr, got %T", call.Fn)
	}
	ident, ok := member.Object.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident, got %T", member.Object)
	}
	if ident.Name != "file" {
		t.Errorf("expected object 'file', got %q", ident.Name)
	}
	if member.Member != "snoop" {
		t.Errorf("expected member 'snoop', got %q", member.Member)
	}
	if len(call.Args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(call.Args))
	}
}

func TestPipeToBareBuiltin(t *testing.T) {
	prog := parse(t, `xs |=| nya`)
	stmt := prog.Stmts[0].(*ast.ExprStmt)
	pipe, ok := stmt.Expr.(*ast.PipeExpr)
	if !ok {
		t.Fatalf("expected PipeExpr, got %T", stmt.Expr)
	}
	ident, ok := pipe.Right.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident on pipe RHS, got %T", pipe.Right)
	}
	if ident.Name != "nya" {
		t.Errorf("expected 'nya', got %q", ident.Name)
	}
}

func TestPipeToBareHiss(t *testing.T) {
	prog := parse(t, `xs |=| hiss`)
	stmt := prog.Stmts[0].(*ast.ExprStmt)
	pipe, ok := stmt.Expr.(*ast.PipeExpr)
	if !ok {
		t.Fatalf("expected PipeExpr, got %T", stmt.Expr)
	}
	ident, ok := pipe.Right.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident on pipe RHS, got %T", pipe.Right)
	}
	if ident.Name != "hiss" {
		t.Errorf("expected 'hiss', got %q", ident.Name)
	}
}

func TestMapLit(t *testing.T) {
	prog := parse(t, `nyan m = {"name": "Tama", "age": 3}`)
	v := prog.Stmts[0].(*ast.VarStmt)
	m, ok := v.Value.(*ast.MapLit)
	if !ok {
		t.Fatalf("expected MapLit, got %T", v.Value)
	}
	if len(m.Keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(m.Keys))
	}
	if len(m.Vals) != 2 {
		t.Errorf("expected 2 vals, got %d", len(m.Vals))
	}
}

func TestEmptyMapLit(t *testing.T) {
	prog := parse(t, `nyan m = {}`)
	v := prog.Stmts[0].(*ast.VarStmt)
	m, ok := v.Value.(*ast.MapLit)
	if !ok {
		t.Fatalf("expected MapLit, got %T", v.Value)
	}
	if len(m.Keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(m.Keys))
	}
}

func TestTypedVarStmt(t *testing.T) {
	prog := parse(t, `nyan x int = 42`)
	v := prog.Stmts[0].(*ast.VarStmt)
	if v.Name != "x" {
		t.Errorf("expected name 'x', got %q", v.Name)
	}
	if v.TypeAnn == nil {
		t.Fatal("expected type annotation, got nil")
	}
	bt, ok := v.TypeAnn.(*ast.BasicType)
	if !ok {
		t.Fatalf("expected BasicType, got %T", v.TypeAnn)
	}
	if bt.Name != "int" {
		t.Errorf("expected type 'int', got %q", bt.Name)
	}
}

func TestUntypedVarStmt(t *testing.T) {
	prog := parse(t, `nyan x = 42`)
	v := prog.Stmts[0].(*ast.VarStmt)
	if v.TypeAnn != nil {
		t.Errorf("expected no type annotation, got %v", v.TypeAnn)
	}
}

func TestTypedFuncStmt(t *testing.T) {
	prog := parse(t, `meow add(a int, b int) int {
  bring a + b
}`)
	fn := prog.Stmts[0].(*ast.FuncStmt)
	if fn.Name != "add" {
		t.Errorf("expected name 'add', got %q", fn.Name)
	}
	if len(fn.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(fn.Params))
	}
	if fn.Params[0].Name != "a" {
		t.Errorf("expected param 'a', got %q", fn.Params[0].Name)
	}
	if fn.Params[0].TypeAnn == nil {
		t.Fatal("expected type annotation on param a")
	}
	if fn.Params[0].TypeAnn.(*ast.BasicType).Name != "int" {
		t.Errorf("expected type 'int', got %q", fn.Params[0].TypeAnn.(*ast.BasicType).Name)
	}
	if fn.Params[1].Name != "b" {
		t.Errorf("expected param 'b', got %q", fn.Params[1].Name)
	}
	if fn.ReturnType == nil {
		t.Fatal("expected return type annotation")
	}
	if fn.ReturnType.(*ast.BasicType).Name != "int" {
		t.Errorf("expected return type 'int', got %q", fn.ReturnType.(*ast.BasicType).Name)
	}
}

func TestUntypedFuncStmt(t *testing.T) {
	prog := parse(t, `meow greet(name) {
  bring name
}`)
	fn := prog.Stmts[0].(*ast.FuncStmt)
	if fn.Params[0].TypeAnn != nil {
		t.Errorf("expected no type annotation, got %v", fn.Params[0].TypeAnn)
	}
	if fn.ReturnType != nil {
		t.Errorf("expected no return type, got %v", fn.ReturnType)
	}
}

func TestTypedLambda(t *testing.T) {
	prog := parse(t, `nyan double = paw(x int) { x * 2 }`)
	v := prog.Stmts[0].(*ast.VarStmt)
	lambda := v.Value.(*ast.LambdaExpr)
	if len(lambda.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(lambda.Params))
	}
	if lambda.Params[0].Name != "x" {
		t.Errorf("expected param 'x', got %q", lambda.Params[0].Name)
	}
	if lambda.Params[0].TypeAnn == nil {
		t.Fatal("expected type annotation on lambda param")
	}
	if lambda.Params[0].TypeAnn.(*ast.BasicType).Name != "int" {
		t.Errorf("expected type 'int', got %q", lambda.Params[0].TypeAnn.(*ast.BasicType).Name)
	}
}

func TestGroupedParamType(t *testing.T) {
	prog := parse(t, `nyan sum = curl([1,2,3], 0, paw(acc, x int) { acc + x })`)
	v := prog.Stmts[0].(*ast.VarStmt)
	call := v.Value.(*ast.CallExpr)
	lambda := call.Args[2].(*ast.LambdaExpr)
	if len(lambda.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(lambda.Params))
	}
	for i, name := range []string{"acc", "x"} {
		if lambda.Params[i].Name != name {
			t.Errorf("param[%d]: expected name %q, got %q", i, name, lambda.Params[i].Name)
		}
		if lambda.Params[i].TypeAnn == nil {
			t.Fatalf("param[%d]: expected type annotation, got nil", i)
		}
		bt, ok := lambda.Params[i].TypeAnn.(*ast.BasicType)
		if !ok {
			t.Fatalf("param[%d]: expected BasicType, got %T", i, lambda.Params[i].TypeAnn)
		}
		if bt.Name != "int" {
			t.Errorf("param[%d]: expected type 'int', got %q", i, bt.Name)
		}
	}
}

func TestGroupedMultiType(t *testing.T) {
	prog := parse(t, `meow f(a, b int, c, d string) string {
  bring "ok"
}`)
	fn := prog.Stmts[0].(*ast.FuncStmt)
	if len(fn.Params) != 4 {
		t.Fatalf("expected 4 params, got %d", len(fn.Params))
	}
	expected := []struct {
		name     string
		typeName string
	}{
		{"a", "int"}, {"b", "int"}, {"c", "string"}, {"d", "string"},
	}
	for i, exp := range expected {
		if fn.Params[i].Name != exp.name {
			t.Errorf("param[%d]: expected name %q, got %q", i, exp.name, fn.Params[i].Name)
		}
		if fn.Params[i].TypeAnn == nil {
			t.Fatalf("param[%d]: expected type annotation, got nil", i)
		}
		bt, ok := fn.Params[i].TypeAnn.(*ast.BasicType)
		if !ok {
			t.Fatalf("param[%d]: expected BasicType, got %T", i, fn.Params[i].TypeAnn)
		}
		if bt.Name != exp.typeName {
			t.Errorf("param[%d]: expected type %q, got %q", i, exp.typeName, bt.Name)
		}
	}
}

func TestGroupedNoType(t *testing.T) {
	prog := parse(t, `nyan f = paw(a, b) { a + b }`)
	v := prog.Stmts[0].(*ast.VarStmt)
	lambda := v.Value.(*ast.LambdaExpr)
	if len(lambda.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(lambda.Params))
	}
	for i, name := range []string{"a", "b"} {
		if lambda.Params[i].Name != name {
			t.Errorf("param[%d]: expected name %q, got %q", i, name, lambda.Params[i].Name)
		}
		if lambda.Params[i].TypeAnn != nil {
			t.Errorf("param[%d]: expected no type annotation, got %v", i, lambda.Params[i].TypeAnn)
		}
	}
}

func TestParseTrickStmt(t *testing.T) {
	prog := parse(t, `trick Showable {
    meow show() string
    meow greet(name string) string
}`)
	if len(prog.Stmts) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(prog.Stmts))
	}
	trick, ok := prog.Stmts[0].(*ast.TrickStmt)
	if !ok {
		t.Fatalf("expected TrickStmt, got %T", prog.Stmts[0])
	}
	if trick.Name != "Showable" {
		t.Errorf("expected name 'Showable', got %q", trick.Name)
	}
	if len(trick.Methods) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(trick.Methods))
	}
	if trick.Methods[0].Name != "show" {
		t.Errorf("expected method 'show', got %q", trick.Methods[0].Name)
	}
	if len(trick.Methods[0].Params) != 0 {
		t.Errorf("expected 0 params for show, got %d", len(trick.Methods[0].Params))
	}
	if trick.Methods[0].ReturnType == nil {
		t.Fatal("expected return type for show")
	}
	if trick.Methods[1].Name != "greet" {
		t.Errorf("expected method 'greet', got %q", trick.Methods[1].Name)
	}
	if len(trick.Methods[1].Params) != 1 {
		t.Errorf("expected 1 param for greet, got %d", len(trick.Methods[1].Params))
	}
}

func TestParseLearnStmt(t *testing.T) {
	prog := parse(t, `learn Cat {
    meow show() string {
        bring "hello"
    }
    meow greet(name string) string {
        bring name
    }
}`)
	if len(prog.Stmts) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(prog.Stmts))
	}
	learn, ok := prog.Stmts[0].(*ast.LearnStmt)
	if !ok {
		t.Fatalf("expected LearnStmt, got %T", prog.Stmts[0])
	}
	if learn.TypeName != "Cat" {
		t.Errorf("expected type 'Cat', got %q", learn.TypeName)
	}
	if len(learn.Methods) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(learn.Methods))
	}
	if learn.Methods[0].Name != "show" {
		t.Errorf("expected method 'show', got %q", learn.Methods[0].Name)
	}
	if learn.Methods[1].Name != "greet" {
		t.Errorf("expected method 'greet', got %q", learn.Methods[1].Name)
	}
}

func TestParseSelfExpr(t *testing.T) {
	prog := parse(t, `learn Cat {
    meow show() string {
        bring self.name
    }
}`)
	learn := prog.Stmts[0].(*ast.LearnStmt)
	ret := learn.Methods[0].Body[0].(*ast.ReturnStmt)
	member, ok := ret.Value.(*ast.MemberExpr)
	if !ok {
		t.Fatalf("expected MemberExpr, got %T", ret.Value)
	}
	if _, ok := member.Object.(*ast.SelfExpr); !ok {
		t.Fatalf("expected SelfExpr as object, got %T", member.Object)
	}
	if member.Member != "name" {
		t.Errorf("expected member 'name', got %q", member.Member)
	}
}

func TestMatchExpr(t *testing.T) {
	prog := parse(t, `nyan result = peek(x) {
  0 => "zero",
  1..10 => "small",
  _ => "big"
}`)
	v := prog.Stmts[0].(*ast.VarStmt)
	m, ok := v.Value.(*ast.MatchExpr)
	if !ok {
		t.Fatalf("expected MatchExpr, got %T", v.Value)
	}
	if len(m.Arms) != 3 {
		t.Errorf("expected 3 arms, got %d", len(m.Arms))
	}
}
