package checker

import (
	"fmt"

	"github.com/135yshr/meow/pkg/ast"
	"github.com/135yshr/meow/pkg/token"
	"github.com/135yshr/meow/pkg/types"
)

// TypeInfo stores resolved type information for AST nodes.
type TypeInfo struct {
	ExprTypes map[ast.Expr]types.Type
	VarTypes  map[string]types.Type
	FuncTypes map[string]types.FuncType
}

// NewTypeInfo creates an empty TypeInfo.
func NewTypeInfo() *TypeInfo {
	return &TypeInfo{
		ExprTypes: make(map[ast.Expr]types.Type),
		VarTypes:  make(map[string]types.Type),
		FuncTypes: make(map[string]types.FuncType),
	}
}

// TypeError represents a type checking error.
type TypeError struct {
	Pos     token.Position
	Message string
}

func (e *TypeError) Error() string {
	return fmt.Sprintf("Hiss! %s at %s, nya~", e.Message, e.Pos)
}

// Checker performs type checking on a Meow AST.
type Checker struct {
	info              *TypeInfo
	errors            []*TypeError
	scopes            []map[string]types.Type
	currentReturnType types.Type // return type of the function currently being checked
}

// New creates a new Checker.
func New() *Checker {
	c := &Checker{
		info: NewTypeInfo(),
	}
	c.pushScope()
	return c
}

func (c *Checker) pushScope() {
	c.scopes = append(c.scopes, make(map[string]types.Type))
}

func (c *Checker) popScope() {
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *Checker) define(name string, t types.Type) {
	c.scopes[len(c.scopes)-1][name] = t
}

func (c *Checker) lookup(name string) types.Type {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if t, ok := c.scopes[i][name]; ok {
			return t
		}
	}
	return types.AnyType{}
}

func (c *Checker) addError(pos token.Position, format string, args ...any) {
	c.errors = append(c.errors, &TypeError{
		Pos:     pos,
		Message: fmt.Sprintf(format, args...),
	})
}

// Check type-checks a program and returns type info and any errors.
func (c *Checker) Check(prog *ast.Program) (*TypeInfo, []*TypeError) {
	// First pass: register all function declarations
	for _, stmt := range prog.Stmts {
		if fn, ok := stmt.(*ast.FuncStmt); ok {
			ft := c.funcSignatureType(fn)
			c.info.FuncTypes[fn.Name] = ft
			c.define(fn.Name, ft)
		}
	}

	// Second pass: check all statements
	for _, stmt := range prog.Stmts {
		c.checkStmt(stmt)
	}

	if len(c.errors) > 0 {
		return c.info, c.errors
	}
	return c.info, nil
}

func (c *Checker) funcSignatureType(fn *ast.FuncStmt) types.FuncType {
	params := make([]types.Type, len(fn.Params))
	for i, p := range fn.Params {
		params[i] = c.resolveTypeExpr(p.TypeAnn)
	}
	ret := c.resolveTypeExpr(fn.ReturnType)
	return types.FuncType{Params: params, Return: ret}
}

func (c *Checker) resolveTypeExpr(te ast.TypeExpr) types.Type {
	if te == nil {
		return types.AnyType{}
	}
	bt, ok := te.(*ast.BasicType)
	if !ok {
		return types.AnyType{}
	}
	switch bt.Name {
	case "int":
		return types.IntType{}
	case "float":
		return types.FloatType{}
	case "string":
		return types.StringType{}
	case "bool":
		return types.BoolType{}
	case "furball":
		return types.FurballType{}
	case "list":
		return types.ListType{Elem: types.AnyType{}}
	default:
		return types.AnyType{}
	}
}

func (c *Checker) checkStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.VarStmt:
		c.checkVarStmt(s)
	case *ast.FuncStmt:
		c.checkFuncStmt(s)
	case *ast.ReturnStmt:
		c.checkReturnStmt(s)
	case *ast.IfStmt:
		c.checkIfStmt(s)
	case *ast.WhileStmt:
		c.checkWhileStmt(s)
	case *ast.ExprStmt:
		c.inferExpr(s.Expr)
	case *ast.AssignStmt:
		c.checkAssignStmt(s)
	case *ast.FetchStmt:
		// nothing to check
	}
}

func (c *Checker) checkVarStmt(s *ast.VarStmt) {
	valType := c.inferExpr(s.Value)
	declType := c.resolveTypeExpr(s.TypeAnn)

	if !types.IsAny(declType) && !types.IsAny(valType) {
		if !declType.Equals(valType) {
			c.addError(s.Token.Pos, "Variable %s declared as %s but assigned %s", s.Name, declType, valType)
		}
	}

	if !types.IsAny(declType) {
		c.define(s.Name, declType)
		c.info.VarTypes[s.Name] = declType
	} else {
		c.define(s.Name, valType)
		c.info.VarTypes[s.Name] = valType
	}
}

func (c *Checker) checkAssignStmt(s *ast.AssignStmt) {
	valType := c.inferExpr(s.Value)
	existing := c.lookup(s.Name)
	if !types.IsAny(existing) && !types.IsAny(valType) {
		if !existing.Equals(valType) {
			c.addError(s.Token.Pos, "Cannot assign %s to variable %s of type %s", valType, s.Name, existing)
		}
	}
}

func (c *Checker) checkFuncStmt(fn *ast.FuncStmt) {
	// Enforce type annotations on all parameters
	for _, p := range fn.Params {
		if p.TypeAnn == nil {
			c.addError(fn.Token.Pos, "Parameter %q of function %s must have a type annotation", p.Name, fn.Name)
		}
	}

	// Enforce return type when function has bring statements
	if fn.ReturnType == nil && hasReturnStmt(fn.Body) {
		c.addError(fn.Token.Pos, "Function %s has bring statements but no return type annotation", fn.Name)
	}

	prevReturnType := c.currentReturnType
	c.currentReturnType = c.resolveTypeExpr(fn.ReturnType)

	c.pushScope()
	for _, p := range fn.Params {
		pt := c.resolveTypeExpr(p.TypeAnn)
		c.define(p.Name, pt)
	}
	for _, stmt := range fn.Body {
		c.checkStmt(stmt)
	}
	c.popScope()

	c.currentReturnType = prevReturnType
}

// hasReturnStmt checks whether a slice of statements contains any ReturnStmt (bring).
func hasReturnStmt(stmts []ast.Stmt) bool {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.ReturnStmt:
			return true
		case *ast.IfStmt:
			if hasReturnStmt(s.Body) || hasReturnStmt(s.ElseBody) {
				return true
			}
		case *ast.WhileStmt:
			if hasReturnStmt(s.Body) {
				return true
			}
		}
	}
	return false
}

func (c *Checker) checkReturnStmt(s *ast.ReturnStmt) {
	if s.Value == nil {
		return
	}
	valType := c.inferExpr(s.Value)
	if c.currentReturnType != nil && !types.IsAny(c.currentReturnType) && !types.IsAny(valType) {
		if !c.currentReturnType.Equals(valType) {
			c.addError(s.Token.Pos, "Return type mismatch: expected %s but got %s", c.currentReturnType, valType)
		}
	}
}

func (c *Checker) checkIfStmt(s *ast.IfStmt) {
	c.inferExpr(s.Condition)
	c.pushScope()
	for _, stmt := range s.Body {
		c.checkStmt(stmt)
	}
	c.popScope()
	if len(s.ElseBody) > 0 {
		c.pushScope()
		for _, stmt := range s.ElseBody {
			c.checkStmt(stmt)
		}
		c.popScope()
	}
}

func (c *Checker) checkWhileStmt(s *ast.WhileStmt) {
	c.inferExpr(s.Condition)
	c.pushScope()
	for _, stmt := range s.Body {
		c.checkStmt(stmt)
	}
	c.popScope()
}

func (c *Checker) inferExpr(expr ast.Expr) types.Type {
	t := c.inferExprInner(expr)
	c.info.ExprTypes[expr] = t
	return t
}

func (c *Checker) inferExprInner(expr ast.Expr) types.Type {
	switch e := expr.(type) {
	case *ast.IntLit:
		return types.IntType{}
	case *ast.FloatLit:
		return types.FloatType{}
	case *ast.StringLit:
		return types.StringType{}
	case *ast.BoolLit:
		return types.BoolType{}
	case *ast.NilLit:
		return types.NilType{}
	case *ast.Ident:
		return c.lookup(e.Name)
	case *ast.UnaryExpr:
		return c.inferUnary(e)
	case *ast.BinaryExpr:
		return c.inferBinary(e)
	case *ast.CallExpr:
		return c.inferCall(e)
	case *ast.LambdaExpr:
		return c.inferLambda(e)
	case *ast.ListLit:
		return c.inferList(e)
	case *ast.IndexExpr:
		leftType := c.inferExpr(e.Left)
		c.inferExpr(e.Index)
		if lt, ok := leftType.(types.ListType); ok {
			return lt.Elem
		}
		return types.AnyType{}
	case *ast.PipeExpr:
		c.inferExpr(e.Left)
		rightType := c.inferExpr(e.Right)
		if ft, ok := rightType.(types.FuncType); ok {
			return ft.Return
		}
		return types.AnyType{}
	case *ast.CatchExpr:
		leftType := c.inferExpr(e.Left)
		rightType := c.inferExpr(e.Right)
		if !types.IsAny(leftType) {
			return leftType
		}
		if ft, ok := rightType.(types.FuncType); ok {
			return ft.Return
		}
		return rightType
	case *ast.MapLit:
		for _, k := range e.Keys {
			c.inferExpr(k)
		}
		for _, v := range e.Vals {
			c.inferExpr(v)
		}
		return types.AnyType{}
	case *ast.MatchExpr:
		c.inferExpr(e.Subject)
		var armType types.Type
		for _, arm := range e.Arms {
			t := c.inferExpr(arm.Body)
			if armType == nil {
				armType = t
			}
		}
		if armType != nil {
			return armType
		}
		return types.AnyType{}
	case *ast.MemberExpr:
		c.inferExpr(e.Object)
		return types.AnyType{}
	default:
		return types.AnyType{}
	}
}

func (c *Checker) inferUnary(e *ast.UnaryExpr) types.Type {
	operand := c.inferExpr(e.Right)
	switch e.Op {
	case token.MINUS:
		if types.IsAny(operand) {
			return types.AnyType{}
		}
		if types.IsNumeric(operand) {
			return operand
		}
		c.addError(e.Token.Pos, "Cannot negate %s", operand)
		return types.AnyType{}
	case token.NOT:
		return types.BoolType{}
	}
	return types.AnyType{}
}

func (c *Checker) inferBinary(e *ast.BinaryExpr) types.Type {
	left := c.inferExpr(e.Left)
	right := c.inferExpr(e.Right)

	// If either side is AnyType, skip checking
	if types.IsAny(left) || types.IsAny(right) {
		switch e.Op {
		case token.EQ, token.NEQ, token.LT, token.GT, token.LTE, token.GTE, token.AND, token.OR:
			return types.BoolType{}
		default:
			return types.AnyType{}
		}
	}

	switch e.Op {
	case token.PLUS:
		if left.Equals(right) {
			switch left.(type) {
			case types.IntType, types.FloatType, types.StringType:
				return left
			}
		}
		c.addError(e.Token.Pos, "Cannot add %s and %s", left, right)
		return types.AnyType{}

	case token.MINUS, token.STAR, token.SLASH:
		if left.Equals(right) && types.IsNumeric(left) {
			return left
		}
		op := map[token.TokenType]string{token.MINUS: "subtract", token.STAR: "multiply", token.SLASH: "divide"}
		c.addError(e.Token.Pos, "Cannot %s %s and %s", op[e.Op], left, right)
		return types.AnyType{}

	case token.PERCENT:
		if left.Equals(right) {
			if _, ok := left.(types.IntType); ok {
				return types.IntType{}
			}
		}
		c.addError(e.Token.Pos, "Cannot modulo %s and %s", left, right)
		return types.AnyType{}

	case token.EQ, token.NEQ:
		if !left.Equals(right) {
			c.addError(e.Token.Pos, "Cannot compare %s and %s", left, right)
		}
		return types.BoolType{}

	case token.LT, token.GT, token.LTE, token.GTE:
		if left.Equals(right) && types.IsNumeric(left) {
			return types.BoolType{}
		}
		c.addError(e.Token.Pos, "Cannot compare %s and %s", left, right)
		return types.BoolType{}

	case token.AND, token.OR:
		return types.BoolType{}
	}

	return types.AnyType{}
}

func (c *Checker) inferCall(e *ast.CallExpr) types.Type {
	for _, arg := range e.Args {
		c.inferExpr(arg)
	}

	if ident, ok := e.Fn.(*ast.Ident); ok {
		// Check built-in functions
		switch ident.Name {
		case "to_int":
			return types.IntType{}
		case "to_float":
			return types.FloatType{}
		case "to_string":
			return types.StringType{}
		case "is_furball":
			return types.BoolType{}
		case "len":
			return types.IntType{}
		case "nya", "hiss", "gag":
			return types.AnyType{}
		case "head":
			return types.AnyType{}
		case "tail", "append":
			return types.AnyType{}
		case "lick", "picky", "curl":
			return types.AnyType{}
		case "judge", "expect", "refuse":
			return types.AnyType{}
		}

		// Check user-defined functions
		if ft, ok := c.info.FuncTypes[ident.Name]; ok {
			// Validate argument types
			if len(e.Args) == len(ft.Params) {
				for i, arg := range e.Args {
					argType := c.info.ExprTypes[arg]
					if argType != nil && !types.IsAny(argType) && !types.IsAny(ft.Params[i]) {
						if !ft.Params[i].Equals(argType) {
							c.addError(e.Token.Pos, "Argument %d: expected %s but got %s", i+1, ft.Params[i], argType)
						}
					}
				}
			}
			return ft.Return
		}
	}

	c.inferExpr(e.Fn)
	return types.AnyType{}
}

func (c *Checker) inferLambda(e *ast.LambdaExpr) types.Type {
	c.pushScope()
	paramTypes := make([]types.Type, len(e.Params))
	for i, p := range e.Params {
		pt := c.resolveTypeExpr(p.TypeAnn)
		paramTypes[i] = pt
		c.define(p.Name, pt)
	}
	retType := c.inferExpr(e.Body)
	c.popScope()
	return types.FuncType{Params: paramTypes, Return: retType}
}

func (c *Checker) inferList(e *ast.ListLit) types.Type {
	if len(e.Items) == 0 {
		return types.ListType{Elem: types.AnyType{}}
	}
	elemType := c.inferExpr(e.Items[0])
	for _, item := range e.Items[1:] {
		c.inferExpr(item)
	}
	return types.ListType{Elem: elemType}
}
