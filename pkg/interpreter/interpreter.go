package interpreter

import (
	"fmt"
	"io"

	"github.com/135yshr/meow/pkg/ast"
	"github.com/135yshr/meow/pkg/checker"
	"github.com/135yshr/meow/pkg/token"
	"github.com/135yshr/meow/runtime/meowrt"
	meowtest "github.com/135yshr/meow/runtime/testing"
)

// propagateFurball panics with the Furball's message if v is an unhandled
// Furball, mirroring the codegen short-circuit. RunSafe's deferred recover
// surfaces this as the program's error result.
func propagateFurball(v meowrt.Value) {
	if f, ok := v.(*meowrt.Furball); ok && !f.Handled {
		panic(f.Message)
	}
}

// returnSignal is used to implement bring (return) via panic/recover.
type returnSignal struct {
	Value meowrt.Value
}

// stepLimitExceeded signals that the step limit was reached.
type stepLimitExceeded struct{}

// boltSignal signals that the loop should be left, and slinkSignal that this
// turn is over. They are raised where they are written and caught by the
// enclosing loop, the way bring is caught by the enclosing function.
type boltSignal struct{}

type slinkSignal struct{}

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
	exitCode   int
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
				err = fmt.Errorf("%s", meowrt.Located(
					fmt.Sprintf("Hiss! step limit exceeded (%d steps), nya~", interp.stepLimit)))
			default:
				// Prefixed with where the program was, the way a compiled one
				// reports a failure, so the same program reads the same either
				// side of the playground.
				if msg, ok := r.(string); ok {
					err = fmt.Errorf("%s", meowrt.Located(msg))
				} else {
					err = fmt.Errorf("internal error: %v", r)
				}
			}
		}
	}()
	interp.Run(prog)
	return nil
}

// ExitCode reports the status the last run asked to end with. A run that
// reached the end on its own reports 0, as a process that ran out of statements
// does.
func (interp *Interpreter) ExitCode() int {
	return interp.exitCode
}

// Run executes the program. Panics propagate to the caller, except the one
// scram raises: a program asking to end is not a failure, so the run stops
// where it asked to and keeps whatever it printed on the way.
func (interp *Interpreter) Run(prog *ast.Program) {
	meowrt.ClearMethods()
	interp.stepCount = 0
	interp.exitCode = 0
	// The playground runs one program after another in the same process, so a
	// position left over from the last one must not be reported against this.
	meowrt.Here("")

	defer func() {
		if r := recover(); r != nil {
			sig, ok := r.(meowrt.ScramSignal)
			if !ok {
				panic(r)
			}
			interp.exitCode = sig.Code
		}
	}()

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
	// Record where the program is, so that a failure raised while this
	// statement runs can say where. Generated code makes the same note in the
	// same place, so both backends report a failure identically.
	//
	// Noted before the step limit is checked, because reaching the limit is
	// itself a failure worth a position — a program that will not stop is
	// exactly the one whose reader needs to know which line it is going round.
	if pos := stmt.Pos(); pos.Line != 0 {
		meowrt.Here(pos.String())
	}
	interp.checkStep()
	switch s := stmt.(type) {
	case *ast.VarStmt:
		val := interp.evalExpr(s.Value, env)
		propagateFurball(val)
		env.Define(s.Name, val)
	case *ast.ExprStmt:
		propagateFurball(interp.evalExpr(s.Expr, env))
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
	case *ast.WhileStmt:
		interp.execWhile(s, env)
	case *ast.BoltStmt:
		panic(boltSignal{})
	case *ast.SlinkStmt:
		panic(slinkSignal{})
	case *ast.FuncStmt:
		// Nested function definition
		interp.registerFunc(s, env)
	case *ast.FetchStmt:
		// The playground has no Go toolchain, so an import of any kind — one of
		// Meow's own packages or, with `go`, a Go one — is out of reach here.
		spec := s.Path
		if s.Version != "" {
			// The pin was written inside the string, so it belongs back inside
			// it — an import said back without it is a different import.
			spec += "@" + s.Version
		}
		what := fmt.Sprintf("nab %q", spec)
		if s.Go {
			what = fmt.Sprintf("nab go %q", spec)
		}
		if s.Alias != "" {
			what += fmt.Sprintf(" tag %s", s.Alias)
		}
		panic(fmt.Sprintf("Hiss! %s is not supported in the playground, nya~", what))
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

// isWalkable reports whether a value is walked element by element rather than
// counted to.
func isWalkable(v meowrt.Value) bool {
	switch v.(type) {
	case *meowrt.List, *meowrt.Map:
		return true
	}
	return false
}

func (interp *Interpreter) execRange(s *ast.RangeStmt, env *Environment) {
	endVal := interp.evalExpr(s.End, env)

	// Elementwise iteration: a litter's elements, or a basket's keys. The same
	// runtime iterators the compiler emits are used here, so a program walks
	// them in the same order — a basket by sorted key — whichever backend runs.
	if s.Start == nil && !s.Inclusive && isWalkable(endVal) {
		if s.IndexVar != "" {
			for a, b := range meowrt.RangePair(endVal) {
				interp.checkStep()
				child := env.Child()
				child.Define(s.IndexVar, a)
				child.Define(s.Var, b)
				if interp.runLoopBody(s.Body, child) {
					break
				}
			}
			return
		}
		for elem := range meowrt.RangeSolo(endVal) {
			interp.checkStep()
			child := env.Child()
			child.Define(s.Var, elem)
			if interp.runLoopBody(s.Body, child) {
				break
			}
		}
		return
	}

	var start int64
	if s.Start != nil {
		start = meowrt.AsInt(interp.evalExpr(s.Start, env))
	}
	end := meowrt.AsInt(endVal)

	if s.Start == nil {
		// count form: purr i (n) → i = 0..n-1
		for i := int64(0); i < end; i++ {
			interp.checkStep()
			child := env.Child()
			child.Define(s.Var, meowrt.NewInt(i))
			if interp.runLoopBody(s.Body, child) {
				break
			}
		}
	} else if s.Inclusive {
		// range form inclusive: purr i (a..b) → i = a..b
		for i := start; i <= end; i++ {
			interp.checkStep()
			child := env.Child()
			child.Define(s.Var, meowrt.NewInt(i))
			if interp.runLoopBody(s.Body, child) {
				break
			}
		}
	} else {
		// range form exclusive
		for i := start; i < end; i++ {
			interp.checkStep()
			child := env.Child()
			child.Define(s.Var, meowrt.NewInt(i))
			if interp.runLoopBody(s.Body, child) {
				break
			}
		}
	}
}

// --- Function Registration ---

func (interp *Interpreter) registerFunc(fn *ast.FuncStmt, env *Environment) {
	interp.funcDefs[fn.Name] = fn
	captured := env
	fnVal := meowrt.NewFuncWithArity(fn.Name, len(fn.Params), func(args ...meowrt.Value) meowrt.Value {
		return interp.callUserFunc(fn, args, captured)
	})
	env.Define(fn.Name, fnVal)
}

func (interp *Interpreter) callUserFunc(fn *ast.FuncStmt, args []meowrt.Value, closure *Environment) meowrt.Value {
	if len(args) < len(fn.Params) {
		// Partial application: capture supplied args and return a new function
		captured := make([]meowrt.Value, len(args))
		copy(captured, args)
		remaining := len(fn.Params) - len(args)
		return meowrt.NewFuncWithArity(fn.Name, remaining, func(moreArgs ...meowrt.Value) meowrt.Value {
			allArgs := make([]meowrt.Value, 0, len(captured)+len(moreArgs))
			allArgs = append(allArgs, captured...)
			allArgs = append(allArgs, moreArgs...)
			return interp.callUserFunc(fn, allArgs, closure)
		})
	}

	child := closure.Child()
	for i, p := range fn.Params {
		if i < len(args) {
			child.Define(p.Name, args[i])
		} else {
			child.Define(p.Name, meowrt.NewNil())
		}
	}

	// Where the call was made from. A call that comes back leaves the program
	// here rather than inside the function it returned from, so a failure later
	// in the same statement is not blamed on the callee's last line. A call that
	// fails never reaches the restore, which is what keeps a failure reported
	// against the line it happened on.
	caller := meowrt.Where()

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

	// Only a call that succeeded goes back to where it was called from. One
	// that answers with a Furball has failed, and the line it failed on is the
	// one worth reporting — the same rule the compiled path follows, where a
	// failure raises before it can restore anything.
	if _, failed := meowrt.AsFurball(result); !failed {
		meowrt.Here(caller)
	}
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
	if len(args) != count {
		panic(fmt.Sprintf("Hiss! %s requires %d argument(s), got %d, nya~", name, count, len(args)))
	}
}

// builtinFn is what a builtin name does. The interpreter is handed along so
// that the few builtins with something to say — nya writes to the run's
// captured output — can reach it; the rest ignore it.
type builtinFn func(interp *Interpreter, args []meowrt.Value) meowrt.Value

// unary, binary and ternary wrap a runtime function of a fixed arity, checking
// the count before it is handed its arguments.
func unary(name string, fn func(meowrt.Value) meowrt.Value) builtinFn {
	return func(_ *Interpreter, args []meowrt.Value) meowrt.Value {
		requireArgs(name, args, 1)
		return fn(args[0])
	}
}

func binary(name string, fn func(a, b meowrt.Value) meowrt.Value) builtinFn {
	return func(_ *Interpreter, args []meowrt.Value) meowrt.Value {
		requireArgs(name, args, 2)
		return fn(args[0], args[1])
	}
}

func ternary(name string, fn func(a, b, c meowrt.Value) meowrt.Value) builtinFn {
	return func(_ *Interpreter, args []meowrt.Value) meowrt.Value {
		requireArgs(name, args, 3)
		return fn(args[0], args[1], args[2])
	}
}

// variadic wraps a runtime function that takes any number of arguments and
// decides for itself whether it was given enough.
func variadic(fn func(args ...meowrt.Value) meowrt.Value) builtinFn {
	return func(_ *Interpreter, args []meowrt.Value) meowrt.Value {
		return fn(args...)
	}
}

// builtins is every function a program can use without a nab, keyed by the
// name it is written under.
//
// A table rather than a switch, so that the set of names is something that can
// be read back: `pkg/checker` decides which names a program may write, this
// decides what they do, and nothing in the compiler links the two. A name in
// one and not the other type-checks and then dies at run time, which is how
// judge, expect, refuse and seed came to be accepted, compiled, and undefined
// here. The test holding these two tables to each other can only be honest
// because both are enumerable.
var builtins = map[string]builtinFn{
	"nya": func(interp *Interpreter, args []meowrt.Value) meowrt.Value {
		return interp.builtinNya(args)
	},
	"hiss": variadic(meowrt.Hiss),
	"scram": func(_ *Interpreter, args []meowrt.Value) meowrt.Value {
		code, fb := meowrt.ScramCode(args...)
		if fb != nil {
			return fb
		}
		panic(meowrt.ScramSignal{Code: code})
	},
	"len":        unary("len", meowrt.Len),
	"to_int":     unary("to_int", meowrt.ToInt),
	"to_float":   unary("to_float", meowrt.ToFloat),
	"to_string":  unary("to_string", meowrt.ToString),
	"to_bytes":   unary("to_bytes", meowrt.ToBytes),
	"to_runes":   unary("to_runes", meowrt.ToRunes),
	"whiff":      binary("whiff", meowrt.Whiff),
	"upper":      unary("upper", meowrt.Upper),
	"lower":      unary("lower", meowrt.Lower),
	"trim":       unary("trim", meowrt.Trim),
	"replace":    ternary("replace", meowrt.Replace),
	"pad":        binary("pad", meowrt.Pad),
	"sort":       unary("sort", meowrt.Sort),
	"reverse":    unary("reverse", meowrt.Reverse),
	"round":      binary("round", meowrt.Round),
	"track":      binary("track", meowrt.Track),
	"shred":      binary("shred", meowrt.Shred),
	"tangle":     binary("tangle", meowrt.Tangle),
	"nibble":     ternary("nibble", meowrt.Nibble),
	"gag":        unary("gag", meowrt.Gag),
	"is_furball": unary("is_furball", meowrt.IsFurball),
	"head":       unary("head", meowrt.Head),
	"tail":       unary("tail", meowrt.Tail),
	"append":     binary("append", meowrt.Append),
	"lick":       binary("lick", meowrt.Lick),
	"picky":      binary("picky", meowrt.Picky),
	"curl":       ternary("curl", meowrt.Curl),

	// The assertions are the very functions the compiled path calls, rather
	// than a second reading of what they ought to mean: an assertion that
	// fails answers with a Furball carrying the message `meow test` would
	// print after FAIL. The playground has no test harness to report into, so
	// that Furball is left to propagate like any other — the run stops on the
	// failing line and says why, which is what a reader of the playground can
	// act on, and it is the same wording and the same stopping point a
	// compiled program gives. An assertion that holds stays silent, as it does
	// under `meow test`, where only the enclosing test's PASS line is printed.
	"judge":  variadic(meowtest.Judge),
	"expect": variadic(meowtest.Expect),
	"refuse": variadic(meowtest.Refuse),

	// seed states one entry of a fuzz corpus. Only `meow test -fuzz` has
	// anything to do with one — it reads the calls out of the test's body
	// before the body is generated — and everywhere else codegen compiles a
	// seed call to catnap. The playground cannot fuzz, so everywhere else is
	// all there is here.
	"seed": func(_ *Interpreter, _ []meowrt.Value) meowrt.Value {
		return meowrt.NewNil()
	},
}

func (interp *Interpreter) dispatchBuiltin(name string, args []meowrt.Value) (meowrt.Value, bool) {
	fn, ok := builtins[name]
	if !ok {
		return nil, false
	}
	return fn(interp, args), true
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

	// A bare name here is resolved in one place, the same place a bare name
	// piped into is resolved. Two copies of that question is what let the
	// backends drift: the pipe's copy read the name as a variable and so could
	// not see a builtin at all.
	if ident, ok := e.Fn.(*ast.Ident); ok {
		return interp.evalCallByName(ident.Name, args, env)
	}

	// First-class function call (e.g. variable holding a Func)
	fnVal := interp.evalExpr(e.Fn, env)
	if fn, ok := fnVal.(*meowrt.Func); ok {
		return meowrt.Call(fn, args...)
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
			return meowrt.Call(fn, args...)
		}
		panic(fmt.Sprintf("Hiss! %s.%s is not callable, nya~", k.TypeName, member.Member))
	}

	panic(fmt.Sprintf("Hiss! cannot call method %s on %s, nya~", member.Member, obj.Type()))
}

// --- Lambda ---

func (interp *Interpreter) evalLambda(e *ast.LambdaExpr, env *Environment) meowrt.Value {
	captured := env
	arity := len(e.Params)
	evalWithArgs := func(args []meowrt.Value) meowrt.Value {
		child := captured.Child()
		for i, p := range e.Params {
			if i < len(args) {
				child.Define(p.Name, args[i])
			} else {
				child.Define(p.Name, meowrt.NewNil())
			}
		}
		if e.Block == nil {
			return interp.evalExpr(e.Body, child)
		}
		return interp.evalLambdaBlock(e.Block, child)
	}
	return meowrt.NewFuncWithArity("lambda", arity, func(args ...meowrt.Value) meowrt.Value {
		if len(args) < arity {
			return meowrt.PartialApply(
				meowrt.NewFuncWithArity("lambda", arity, func(allArgs ...meowrt.Value) meowrt.Value {
					return evalWithArgs(allArgs)
				}),
				args...,
			)
		}
		return evalWithArgs(args)
	})
}

// evalLambdaBlock runs a block-bodied lambda. A trailing expression statement
// is the result, mirroring the single-expression form; otherwise the value is
// whatever `bring` returned, or catnap.
func (interp *Interpreter) evalLambdaBlock(stmts []ast.Stmt, env *Environment) meowrt.Value {
	var result meowrt.Value
	func() {
		defer func() {
			if r := recover(); r != nil {
				sig, ok := r.(returnSignal)
				if !ok {
					panic(r)
				}
				result = sig.Value
			}
		}()
		for i, stmt := range stmts {
			if exprStmt, ok := stmt.(*ast.ExprStmt); ok && i == len(stmts)-1 {
				result = interp.evalExpr(exprStmt.Expr, env)
				return
			}
			interp.execStmt(stmt, env)
		}
	}()

	if result != nil {
		return result
	}
	return meowrt.NewNil()
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

	return meowrt.Index(left, index)
}

// --- Member Access ---

func (interp *Interpreter) evalMember(e *ast.MemberExpr, env *Environment) meowrt.Value {
	obj := interp.evalExpr(e.Object, env)
	// A member read rather than called is the same question in both backends,
	// so it is the same answer: what a kitty holds, or the method bound to it.
	return meowrt.GetMember(obj, e.Member)
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
			return meowrt.Call(fn, args...)
		}
		panic(fmt.Sprintf("Hiss! pipe target is not callable, nya~"))
	}

	// x |=| f → f(x).
	//
	// A bare name here is resolved the way a name in a call position is, not as
	// a variable read: a builtin is a builtin whether it is called or piped
	// into, and so is a kitty or collar constructor. Reading it as a variable
	// is what made `nums |=| nya` — the form the tutorial teaches throughout —
	// die with "undefined variable nya", because a builtin lives in the call
	// dispatch and never in the environment.
	if ident, ok := e.Right.(*ast.Ident); ok {
		return interp.evalCallByName(ident.Name, []meowrt.Value{left}, env)
	}

	fnVal := interp.evalExpr(e.Right, env)
	if fn, ok := fnVal.(*meowrt.Func); ok {
		return meowrt.Call(fn, left)
	}
	panic(fmt.Sprintf("Hiss! %s is not callable, nya~", fnVal.Type()))
}

func (interp *Interpreter) evalCallByName(name string, args []meowrt.Value, env *Environment) meowrt.Value {
	// A builtin is consulted before the environment, which is the order the
	// compiled path resolves a call in: codegen answers a builtin name from
	// its own table before it looks at what the program bound. Keeping the
	// order means a program that shadows a builtin's name reads the same
	// either side of the playground.
	if val, ok := interp.dispatchBuiltin(name, args); ok {
		return val
	}

	if env.Has(name) {
		fnVal := env.Get(name)
		if fn, ok := fnVal.(*meowrt.Func); ok {
			return meowrt.Call(fn, args...)
		}
		panic(fmt.Sprintf("Hiss! %s is not callable, nya~", fnVal.Type()))
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
	// A failure is not something to print. Printing it would put the message in
	// the program's output and let the program run on, where a compiled one
	// stops — meow.Nya hands the Furball back for exactly that reason.
	for _, a := range args {
		if f, ok := meowrt.AsFurball(a); ok {
			return f
		}
	}
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

// execWhile runs the conditional form of purr.
//
// The condition is checked for failure rather than only for truthiness: read as
// a plain truthiness test a Furball is false, which would end the loop quietly
// and let the program carry on as though the condition had stopped holding.
func (interp *Interpreter) execWhile(s *ast.WhileStmt, env *Environment) {
	for {
		interp.checkStep()
		cond := interp.evalExpr(s.Cond, env)
		propagateFurball(cond)
		if !cond.IsTruthy() {
			return
		}
		if interp.runLoopBody(s.Body, env.Child()) {
			return
		}
	}
}

// runLoopBody runs one turn of a loop, reporting whether the loop should stop.
//
// bolt and slink are raised where they are written, so this is where they are
// caught — the same shape as bring, which the enclosing function catches.
func (interp *Interpreter) runLoopBody(stmts []ast.Stmt, env *Environment) (stop bool) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		switch r.(type) {
		case boltSignal:
			stop = true
		case slinkSignal:
			// This turn is over; the next one starts as usual.
		default:
			panic(r)
		}
	}()
	interp.execBlock(stmts, env)
	return false
}
