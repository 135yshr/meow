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

func TestPurrStmt(t *testing.T) {
	prog := parse(t, `purr i (0, 10) {
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
