package interpreter

import (
	"fmt"
	"io"

	"github.com/135yshr/meow/pkg/ast"
	"github.com/135yshr/meow/pkg/checker"
	"github.com/135yshr/meow/pkg/token"
	"github.com/135yshr/meow/runtime/meowrt"
)

// returnSignal is used to implement bring (return) via panic/recover.
type returnSignal struct {
	Value meowrt.Value
}

// stepLimitExceeded signals that the step limit was reached.
type stepLimitExceeded struct{}

// Interpreter executes a Meow AST directly.
type Interpreter struct {
	globals    *Environment
	typeInfo   *checker.TypeInfo
	output     io.Writer
	kittyDefs  map[string]*ast.KittyStmt
	collarDefs map[string]*ast.CollarStmt
	funcDefs   map[string]*ast.FuncStmt
	stepCount  int64
	stepLimit  int64
}

// New creates a new Interpreter that writes output to w.
func New(w io.Writer) *Interpreter {
	return &Interpreter{
		globals:    NewEnvironment(),
		output:     w,
		kittyDefs:  make(map[string]*ast.KittyStmt),
		collarDefs: make(map[string]*ast.CollarStmt),
		funcDefs:   make(map[string]*ast.FuncStmt),
		stepLimit:  10_000_000,
	}
}

// SetTypeInfo sets the checker type information (optional).
func (interp *Interpreter) SetTypeInfo(ti *checker.TypeInfo) {
	interp.typeInfo = ti
}

// SetStepLimit sets the maximum number of evaluation steps.
func (interp *Interpreter) SetStepLimit(limit int64) {
	interp.stepLimit = limit
}

// RunSafe executes the program and returns any error (including panics).
func (interp *Interpreter) RunSafe(prog *ast.Program) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case stepLimitExceeded:
				err = fmt.Errorf("Hiss! step limit exceeded (%d steps), nya~", interp.stepLimit)
			default:
				if msg, ok := r.(string); ok {
					err = fmt.Errorf("%s", msg)
				} else {
					err = fmt.Errorf("internal error: %v", r)
				}
			}
		}
	}()
	interp.Run(prog)
	return nil
}

// Run executes the program. Panics propagate to the caller.
func (interp *Interpreter) Run(prog *ast.Program) {
	meowrt.ClearMethods()
	interp.stepCount = 0

	// Pass 1: collect declarations
	for _, stmt := range prog.Stmts {
		switch s := stmt.(type) {
		case *ast.KittyStmt:
			interp.kittyDefs[s.Name] = s
		case *ast.CollarStmt:
			interp.collarDefs[s.Name] = s
		case *ast.FuncStmt:
			interp.registerFunc(s, interp.globals)
		case *ast.LearnStmt:
			interp.registerLearnMethods(s)
		case *ast.BreedStmt, *ast.TrickStmt:
			// type-level declarations, nothing to do at runtime
		}
	}

	// Pass 2: execute top-level statements
	for _, stmt := range prog.Stmts {
		switch stmt.(type) {
		case *ast.KittyStmt, *ast.CollarStmt, *ast.FuncStmt,
			*ast.LearnStmt, *ast.BreedStmt, *ast.TrickStmt:
			continue
		}
		interp.execStmt(stmt, interp.globals)
	}
}

func (interp *Interpreter) checkStep() {
	interp.stepCount++
	if interp.stepCount > interp.stepLimit {
		panic(stepLimitExceeded{})
	}
}

// --- Statement Execution ---

func (interp *Interpreter) execStmt(stmt ast.Stmt, env *Environment) {
	interp.checkStep()
	switch s := stmt.(type) {
	case *ast.VarStmt:
		val := interp.evalExpr(s.Value, env)
		env.Define(s.Name, val)
	case *ast.ExprStmt:
		interp.evalExpr(s.Expr, env)
	case *ast.ReturnStmt:
		var val meowrt.Value
		if s.Value != nil {
			val = interp.evalExpr(s.Value, env)
		} else {
			val = meowrt.NewNil()
		}
		panic(returnSignal{Value: val})
	case *ast.IfStmt:
		interp.execIf(s, env)
	case *ast.RangeStmt:
		interp.execRange(s, env)
	case *ast.FuncStmt:
		// Nested function definition
		interp.registerFunc(s, env)
	case *ast.FetchStmt:
		panic(fmt.Sprintf("Hiss! nab %q is not supported in the playground, nya~", s.Path))
	default:
		// KittyStmt, CollarStmt, etc. already handled in Pass 1
	}
}

func (interp *Interpreter) execBlock(stmts []ast.Stmt, env *Environment) {
	for _, stmt := range stmts {
		interp.execStmt(stmt, env)
	}
}

func (interp *Interpreter) execIf(s *ast.IfStmt, env *Environment) {
	cond := interp.evalExpr(s.Condition, env)
	if cond.IsTruthy() {
		child := env.Child()
		interp.execBlock(s.Body, child)
	} else if len(s.ElseBody) > 0 {
		child := env.Child()
		interp.execBlock(s.ElseBody, child)
	}
}

func (interp *Interpreter) execRange(s *ast.RangeStmt, env *Environment) {
	var start, end int64
	if s.Start != nil {
		start = meowrt.AsInt(interp.evalExpr(s.Start, env))
	}
	end = meowrt.AsInt(interp.evalExpr(s.End, env))

	if s.Start == nil {
		// count form: purr i (n) → i = 0..n-1
		for i := int64(0); i < end; i++ {
			interp.checkStep()
			child := env.Child()
			child.Define(s.Var, meowrt.NewInt(i))
			interp.execBlock(s.Body, child)
		}
	} else if s.Inclusive {
		// range form inclusive: purr i (a..b) → i = a..b
		for i := start; i <= end; i++ {
			interp.checkStep()
			child := env.Child()
			child.Define(s.Var, meowrt.NewInt(i))
			interp.execBlock(s.Body, child)
		}
	} else {
		// range form exclusive
		for i := start; i < end; i++ {
			interp.checkStep()
			child := env.Child()
			child.Define(s.Var, meowrt.NewInt(i))
			interp.execBlock(s.Body, child)
		}
	}
}

// --- Function Registration ---

func (interp *Interpreter) registerFunc(fn *ast.FuncStmt, env *Environment) {
	interp.funcDefs[fn.Name] = fn
	captured := env
	fnVal := meowrt.NewFunc(fn.Name, func(args ...meowrt.Value) meowrt.Value {
		return interp.callUserFunc(fn, args, captured)
	})
	env.Define(fn.Name, fnVal)
}

func (interp *Interpreter) callUserFunc(fn *ast.FuncStmt, args []meowrt.Value, closure *Environment) meowrt.Value {
	child := closure.Child()
	for i, p := range fn.Params {
		if i < len(args) {
			child.Define(p.Name, args[i])
		} else {
			child.Define(p.Name, meowrt.NewNil())
		}
	}

	var result meowrt.Value
	func() {
		defer func() {
			if r := recover(); r != nil {
				if sig, ok := r.(returnSignal); ok {
					result = sig.Value
				} else {
					panic(r)
				}
			}
		}()
		interp.execBlock(fn.Body, child)
	}()

	if result != nil {
		return result
	}
	return meowrt.NewNil()
}

func (interp *Interpreter) registerLearnMethods(ls *ast.LearnStmt) {
	for i := range ls.Methods {
		m := &ls.Methods[i]
		typeName := ls.TypeName
		method := m
		meowrt.RegisterMethod(typeName, method.Name, func(args ...meowrt.Value) meowrt.Value {
			child := interp.globals.Child()
			if len(args) > 0 {
				child.Define("self", args[0])
			}
			for j, p := range method.Params {
				if j+1 < len(args) {
					child.Define(p.Name, args[j+1])
				} else {
					child.Define(p.Name, meowrt.NewNil())
				}
			}

			var result meowrt.Value
			func() {
				defer func() {
					if r := recover(); r != nil {
						if sig, ok := r.(returnSignal); ok {
							result = sig.Value
						} else {
							panic(r)
						}
					}
				}()
				interp.execBlock(method.Body, child)
			}()

			if result != nil {
				return result
			}
			return meowrt.NewNil()
		})
	}
}

// --- Expression Evaluation ---

func (interp *Interpreter) evalExpr(expr ast.Expr, env *Environment) meowrt.Value {
	interp.checkStep()
	switch e := expr.(type) {
	case *ast.IntLit:
		return meowrt.NewInt(e.Value)
	case *ast.FloatLit:
		return meowrt.NewFloat(e.Value)
	case *ast.StringLit:
		return meowrt.NewString(e.Value)
	case *ast.BoolLit:
		return meowrt.NewBool(e.Value)
	case *ast.NilLit:
		return meowrt.NewNil()
	case *ast.Ident:
		return env.Get(e.Name)
	case *ast.SelfExpr:
		return env.Get("self")
	case *ast.UnaryExpr:
		return interp.evalUnary(e, env)
	case *ast.BinaryExpr:
		return interp.evalBinary(e, env)
	case *ast.CallExpr:
		return interp.evalCall(e, env)
	case *ast.LambdaExpr:
		return interp.evalLambda(e, env)
	case *ast.ListLit:
		return interp.evalList(e, env)
	case *ast.MapLit:
		return interp.evalMap(e, env)
	case *ast.IndexExpr:
		return interp.evalIndex(e, env)
	case *ast.MemberExpr:
		return interp.evalMember(e, env)
	case *ast.PipeExpr:
		return interp.evalPipe(e, env)
	case *ast.CatchExpr:
		return interp.evalCatch(e, env)
	case *ast.MatchExpr:
		return interp.evalMatch(e, env)
	default:
		panic(fmt.Sprintf("Hiss! unsupported expression: %T, nya~", expr))
	}
}

func (interp *Interpreter) evalUnary(e *ast.UnaryExpr, env *Environment) meowrt.Value {
	right := interp.evalExpr(e.Right, env)
	switch e.Op {
	case token.MINUS:
		return meowrt.Negate(right)
	case token.NOT:
		return meowrt.Not(right)
	default:
		panic(fmt.Sprintf("Hiss! unsupported unary operator: %v, nya~", e.Op))
	}
}

func (interp *Interpreter) evalBinary(e *ast.BinaryExpr, env *Environment) meowrt.Value {
	// Short-circuit for AND/OR
	if e.Op == token.AND {
		left := interp.evalExpr(e.Left, env)
		if !left.IsTruthy() {
			return left
		}
		return interp.evalExpr(e.Right, env)
	}
	if e.Op == token.OR {
		left := interp.evalExpr(e.Left, env)
		if left.IsTruthy() {
			return left
		}
		return interp.evalExpr(e.Right, env)
	}

	left := interp.evalExpr(e.Left, env)
	right := interp.evalExpr(e.Right, env)

	switch e.Op {
	case token.PLUS:
		return meowrt.Add(left, right)
	case token.MINUS:
		return meowrt.Sub(left, right)
	case token.STAR:
		return meowrt.Mul(left, right)
	case token.SLASH:
		return meowrt.Div(left, right)
	case token.PERCENT:
		return meowrt.Mod(left, right)
	case token.EQ:
		return meowrt.Equal(left, right)
	case token.NEQ:
		return meowrt.NotEqual(left, right)
	case token.LT:
		return meowrt.LessThan(left, right)
	case token.GT:
		return meowrt.GreaterThan(left, right)
	case token.LTE:
		return meowrt.LessEqual(left, right)
	case token.GTE:
		return meowrt.GreaterEqual(left, right)
	default:
		panic(fmt.Sprintf("Hiss! unsupported binary operator: %v, nya~", e.Op))
	}
}

// --- Builtin Helpers ---

func requireArgs(name string, args []meowrt.Value, count int) {
	if len(args) < count {
		panic(fmt.Sprintf("Hiss! %s requires %d argument(s), got %d, nya~", name, count, len(args)))
	}
}

func (interp *Interpreter) dispatchBuiltin(name string, args []meowrt.Value) (meowrt.Value, bool) {
	switch name {
	case "nya":
		return interp.builtinNya(args), true
	case "hiss":
		return meowrt.Hiss(args...), true
	case "len":
		requireArgs("len", args, 1)
		return meowrt.Len(args[0]), true
	case "to_int":
		requireArgs("to_int", args, 1)
		return meowrt.ToInt(args[0]), true
	case "to_float":
		requireArgs("to_float", args, 1)
		return meowrt.ToFloat(args[0]), true
	case "to_string":
		requireArgs("to_string", args, 1)
		return meowrt.ToString(args[0]), true
	case "gag":
		requireArgs("gag", args, 1)
		return meowrt.Gag(args[0]), true
	case "is_furball":
		requireArgs("is_furball", args, 1)
		return meowrt.IsFurball(args[0]), true
	case "head":
		requireArgs("head", args, 1)
		return meowrt.Head(args[0]), true
	case "tail":
		requireArgs("tail", args, 1)
		return meowrt.Tail(args[0]), true
	case "append":
		requireArgs("append", args, 2)
		return meowrt.Append(args[0], args[1]), true
	case "lick":
		requireArgs("lick", args, 2)
		return meowrt.Lick(args[0], args[1]), true
	case "picky":
		requireArgs("picky", args, 2)
		return meowrt.Picky(args[0], args[1]), true
	case "curl":
		requireArgs("curl", args, 3)
		return meowrt.Curl(args[0], args[1], args[2]), true
	default:
		return nil, false
	}
}

// --- Call Expression ---

func (interp *Interpreter) evalCall(e *ast.CallExpr, env *Environment) meowrt.Value {
	// Handle member calls (method dispatch and stdlib)
	if member, ok := e.Fn.(*ast.MemberExpr); ok {
		return interp.evalMemberCall(member, e.Args, env)
	}

	// Evaluate arguments
	args := make([]meowrt.Value, len(e.Args))
	for i, a := range e.Args {
		args[i] = interp.evalExpr(a, env)
	}

	// Identifier-based calls
	if ident, ok := e.Fn.(*ast.Ident); ok {
		// Builtins
		if val, ok := interp.dispatchBuiltin(ident.Name, args); ok {
			return val
		}

		// Kitty constructor
		if ks, ok := interp.kittyDefs[ident.Name]; ok {
			fieldNames := make([]string, len(ks.Fields))
			for i, f := range ks.Fields {
				fieldNames[i] = f.Name
			}
			return meowrt.NewKitty(ident.Name, fieldNames, args...)
		}

		// Collar constructor
		if _, ok := interp.collarDefs[ident.Name]; ok {
			return meowrt.NewKitty(ident.Name, []string{"value"}, args...)
		}

		// User-defined function (looked up from environment)
		if env.Has(ident.Name) {
			fnVal := env.Get(ident.Name)
			if fn, ok := fnVal.(*meowrt.Func); ok {
				return fn.Call(args...)
			}
			panic(fmt.Sprintf("Hiss! %s is not callable, nya~", ident.Name))
		}

		panic(fmt.Sprintf("Hiss! undefined function %s, nya~", ident.Name))
	}

	// First-class function call (e.g. variable holding a Func)
	fnVal := interp.evalExpr(e.Fn, env)
	if fn, ok := fnVal.(*meowrt.Func); ok {
		return fn.Call(args...)
	}
	panic(fmt.Sprintf("Hiss! %s is not callable, nya~", fnVal.Type()))
}

func (interp *Interpreter) evalMemberCall(member *ast.MemberExpr, rawArgs []ast.Expr, env *Environment) meowrt.Value {
	args := make([]meowrt.Value, len(rawArgs))
	for i, a := range rawArgs {
		args[i] = interp.evalExpr(a, env)
	}

	obj := interp.evalExpr(member.Object, env)

	// Method dispatch via registry
	if k, ok := obj.(*meowrt.Kitty); ok {
		if _, found := meowrt.LookupMethod(k.TypeName, member.Member); found {
			return meowrt.DispatchMethod(obj, member.Member, args...)
		}
		// Kitty field that is a function
		field := k.GetField(member.Member)
		if fn, ok := field.(*meowrt.Func); ok {
			return fn.Call(args...)
		}
		panic(fmt.Sprintf("Hiss! %s.%s is not callable, nya~", k.TypeName, member.Member))
	}

	panic(fmt.Sprintf("Hiss! cannot call method %s on %s, nya~", member.Member, obj.Type()))
}

// --- Lambda ---

func (interp *Interpreter) evalLambda(e *ast.LambdaExpr, env *Environment) meowrt.Value {
	captured := env
	return meowrt.NewFunc("lambda", func(args ...meowrt.Value) meowrt.Value {
		child := captured.Child()
		for i, p := range e.Params {
			if i < len(args) {
				child.Define(p.Name, args[i])
			} else {
				child.Define(p.Name, meowrt.NewNil())
			}
		}
		return interp.evalExpr(e.Body, child)
	})
}

// --- Collections ---

func (interp *Interpreter) evalList(e *ast.ListLit, env *Environment) meowrt.Value {
	items := make([]meowrt.Value, len(e.Items))
	for i, item := range e.Items {
		items[i] = interp.evalExpr(item, env)
	}
	return meowrt.NewList(items...)
}

func (interp *Interpreter) evalMap(e *ast.MapLit, env *Environment) meowrt.Value {
	items := make(map[string]meowrt.Value, len(e.Keys))
	for i := range e.Keys {
		key := interp.evalExpr(e.Keys[i], env)
		val := interp.evalExpr(e.Vals[i], env)
		items[meowrt.AsString(key)] = val
	}
	return meowrt.NewMap(items)
}

func (interp *Interpreter) evalIndex(e *ast.IndexExpr, env *Environment) meowrt.Value {
	left := interp.evalExpr(e.Left, env)
	index := interp.evalExpr(e.Index, env)

	switch obj := left.(type) {
	case *meowrt.List:
		return obj.Get(int(meowrt.AsInt(index)))
	case *meowrt.Map:
		key := meowrt.AsString(index)
		if v, ok := obj.Get(key); ok {
			return v
		}
		return meowrt.NewNil()
	default:
		panic(fmt.Sprintf("Hiss! cannot index %s, nya~", left.Type()))
	}
}

// --- Member Access ---

func (interp *Interpreter) evalMember(e *ast.MemberExpr, env *Environment) meowrt.Value {
	obj := interp.evalExpr(e.Object, env)
	if k, ok := obj.(*meowrt.Kitty); ok {
		return k.GetField(e.Member)
	}
	panic(fmt.Sprintf("Hiss! cannot access field %s on %s, nya~", e.Member, obj.Type()))
}

// --- Pipe ---

func (interp *Interpreter) evalPipe(e *ast.PipeExpr, env *Environment) meowrt.Value {
	left := interp.evalExpr(e.Left, env)

	// x |=| f(y) → f(x, y)
	if call, ok := e.Right.(*ast.CallExpr); ok {
		args := make([]meowrt.Value, 0, len(call.Args)+1)
		args = append(args, left)
		for _, a := range call.Args {
			args = append(args, interp.evalExpr(a, env))
		}

		// Handle member call
		if member, ok := call.Fn.(*ast.MemberExpr); ok {
			obj := interp.evalExpr(member.Object, env)
			if k, ok := obj.(*meowrt.Kitty); ok {
				if _, found := meowrt.LookupMethod(k.TypeName, member.Member); found {
					return meowrt.DispatchMethod(obj, member.Member, args...)
				}
			}
		}

		// Handle ident call
		if ident, ok := call.Fn.(*ast.Ident); ok {
			return interp.evalCallByName(ident.Name, args, env)
		}

		fnVal := interp.evalExpr(call.Fn, env)
		if fn, ok := fnVal.(*meowrt.Func); ok {
			return fn.Call(args...)
		}
		panic(fmt.Sprintf("Hiss! pipe target is not callable, nya~"))
	}

	// x |=| f → f(x)
	fnVal := interp.evalExpr(e.Right, env)
	if fn, ok := fnVal.(*meowrt.Func); ok {
		return fn.Call(left)
	}
	panic(fmt.Sprintf("Hiss! pipe target is not callable, nya~"))
}

func (interp *Interpreter) evalCallByName(name string, args []meowrt.Value, env *Environment) meowrt.Value {
	if val, ok := interp.dispatchBuiltin(name, args); ok {
		return val
	}

	if env.Has(name) {
		fnVal := env.Get(name)
		if fn, ok := fnVal.(*meowrt.Func); ok {
			return fn.Call(args...)
		}
	}

	// Kitty constructor
	if ks, ok := interp.kittyDefs[name]; ok {
		fieldNames := make([]string, len(ks.Fields))
		for i, f := range ks.Fields {
			fieldNames[i] = f.Name
		}
		return meowrt.NewKitty(name, fieldNames, args...)
	}

	// Collar constructor
	if _, ok := interp.collarDefs[name]; ok {
		return meowrt.NewKitty(name, []string{"value"}, args...)
	}

	panic(fmt.Sprintf("Hiss! undefined function %s, nya~", name))
}

// --- Catch ---

func (interp *Interpreter) evalCatch(e *ast.CatchExpr, env *Environment) meowrt.Value {
	// Wrap left side in a thunk
	thunk := meowrt.NewFunc("~>", func(args ...meowrt.Value) meowrt.Value {
		return interp.evalExpr(e.Left, env)
	})
	fallback := interp.evalExpr(e.Right, env)
	return meowrt.GagOr(thunk, fallback)
}

// --- Pattern Match ---

func (interp *Interpreter) evalMatch(e *ast.MatchExpr, env *Environment) meowrt.Value {
	subject := interp.evalExpr(e.Subject, env)
	for _, arm := range e.Arms {
		if interp.matchPattern(subject, arm.Pattern, env) {
			return interp.evalExpr(arm.Body, env)
		}
	}
	return meowrt.NewNil()
}

func (interp *Interpreter) matchPattern(subject meowrt.Value, pattern ast.Pattern, env *Environment) bool {
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return true
	case *ast.LiteralPattern:
		patternVal := interp.evalExpr(p.Value, env)
		return meowrt.MatchValue(subject, patternVal)
	case *ast.RangePattern:
		lowLit, lowOk := p.Low.(*ast.IntLit)
		highLit, highOk := p.High.(*ast.IntLit)
		if !lowOk || !highOk {
			return false
		}
		return meowrt.MatchRange(subject, lowLit.Value, highLit.Value)
	default:
		return false
	}
}

// --- Builtin nya (output capture) ---

func (interp *Interpreter) builtinNya(args []meowrt.Value) meowrt.Value {
	parts := make([]string, len(args))
	for i, v := range args {
		parts[i] = v.String()
	}
	for i, p := range parts {
		if i > 0 {
			fmt.Fprint(interp.output, " ")
		}
		fmt.Fprint(interp.output, p)
	}
	fmt.Fprintln(interp.output)
	return meowrt.NewNil()
}
