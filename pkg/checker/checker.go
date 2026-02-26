package checker

import (
	"fmt"

	"github.com/135yshr/meow/pkg/ast"
	"github.com/135yshr/meow/pkg/token"
	"github.com/135yshr/meow/pkg/types"
)

// TypeInfo stores resolved type information for AST nodes.
type TypeInfo struct {
	ExprTypes   map[ast.Expr]types.Type
	VarTypes    map[string]types.Type
	FuncTypes   map[string]types.FuncType
	KittyTypes  map[string]types.KittyType
	AliasTypes  map[string]types.AliasType
	CollarTypes map[string]types.CollarType
}

// NewTypeInfo creates an empty TypeInfo.
func NewTypeInfo() *TypeInfo {
	return &TypeInfo{
		ExprTypes:   make(map[ast.Expr]types.Type),
		VarTypes:    make(map[string]types.Type),
		FuncTypes:   make(map[string]types.FuncType),
		KittyTypes:  make(map[string]types.KittyType),
		AliasTypes:  make(map[string]types.AliasType),
		CollarTypes: make(map[string]types.CollarType),
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
	// First pass: register type/function names as placeholders
	for _, stmt := range prog.Stmts {
		if bs, ok := stmt.(*ast.BreedStmt); ok {
			c.info.AliasTypes[bs.Name] = types.AliasType{Name: bs.Name, Underlying: types.AnyType{}}
		}
		if cs, ok := stmt.(*ast.CollarStmt); ok {
			c.info.CollarTypes[cs.Name] = types.CollarType{Name: cs.Name, Underlying: types.AnyType{}}
		}
		if ks, ok := stmt.(*ast.KittyStmt); ok {
			c.info.KittyTypes[ks.Name] = types.KittyType{Name: ks.Name}
		}
		if fn, ok := stmt.(*ast.FuncStmt); ok {
			ft := c.funcSignatureType(fn)
			c.info.FuncTypes[fn.Name] = ft
			c.define(fn.Name, ft)
		}
	}

	// Second pass: resolve underlying types (forward references now work)
	for _, stmt := range prog.Stmts {
		if bs, ok := stmt.(*ast.BreedStmt); ok {
			at := c.info.AliasTypes[bs.Name]
			at.Underlying = c.resolveTypeExpr(bs.Original)
			c.info.AliasTypes[bs.Name] = at
		}
		if cs, ok := stmt.(*ast.CollarStmt); ok {
			ct := c.info.CollarTypes[cs.Name]
			ct.Underlying = c.resolveTypeExpr(cs.Wrapped)
			c.info.CollarTypes[cs.Name] = ct
		}
		if ks, ok := stmt.(*ast.KittyStmt); ok {
			fields := make([]types.KittyFieldType, len(ks.Fields))
			for i, f := range ks.Fields {
				fields[i] = types.KittyFieldType{Name: f.Name, Type: c.resolveTypeExpr(f.TypeAnn)}
			}
			kt := c.info.KittyTypes[ks.Name]
			kt.Fields = fields
			c.info.KittyTypes[ks.Name] = kt
		}
	}

	// Fixup: refresh stale alias/collar references from forward declarations.
	// When breed A = B was resolved before breed B = int, A's underlying holds
	// a stale copy of B. Replace it with the latest from the map.
	for i := 0; i < len(c.info.AliasTypes); i++ {
		changed := false
		for name, at := range c.info.AliasTypes {
			if inner, ok := at.Underlying.(types.AliasType); ok {
				if latest, found := c.info.AliasTypes[inner.Name]; found {
					if !inner.Underlying.Equals(latest.Underlying) {
						at.Underlying = latest
						c.info.AliasTypes[name] = at
						changed = true
					}
				}
			}
		}
		if !changed {
			break
		}
	}

	// Third pass: check all statements
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
	switch t := te.(type) {
	case *ast.BasicType:
		switch t.Name {
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
	case *ast.NamedType:
		if at, ok := c.info.AliasTypes[t.Name]; ok {
			return at
		}
		if ct, ok := c.info.CollarTypes[t.Name]; ok {
			return ct
		}
		if kt, ok := c.info.KittyTypes[t.Name]; ok {
			return kt
		}
		c.addError(t.Token.Pos, "Unknown type %s", t.Name)
		return types.AnyType{}
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
	case *ast.RangeStmt:
		c.checkRangeStmt(s)
	case *ast.ExprStmt:
		c.inferExpr(s.Expr)
	case *ast.FetchStmt:
		// nothing to check
	case *ast.KittyStmt:
		// already registered in first pass
	case *ast.BreedStmt:
		// already registered in first pass
	case *ast.CollarStmt:
		// already registered in first pass
	}
}

func (c *Checker) checkVarStmt(s *ast.VarStmt) {
	valType := c.inferExpr(s.Value)
	declType := c.resolveTypeExpr(s.TypeAnn)

	// Reject same-scope redeclaration (shadowing in inner scopes is allowed)
	if s.Name != "_" {
		currentScope := c.scopes[len(c.scopes)-1]
		if _, exists := currentScope[s.Name]; exists {
			c.addError(s.Token.Pos, "Variable %s already declared in this scope", s.Name)
		}
	}

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

	// Typed functions must return on all paths
	if !types.IsAny(c.currentReturnType) && !blockAlwaysReturns(fn.Body) {
		c.addError(fn.Token.Pos, "Function %s declares return type %s but does not return on all paths",
			fn.Name, c.currentReturnType)
	}

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

// isPrimitiveType reports whether t is a simple scalar type (int, float, string, bool, nil).
func isPrimitiveType(t types.Type) bool {
	switch t.(type) {
	case types.IntType, types.FloatType, types.StringType, types.BoolType, types.NilType:
		return true
	default:
		return false
	}
}

// blockAlwaysReturns checks if all control-flow paths end with a return statement.
func blockAlwaysReturns(stmts []ast.Stmt) bool {
	if len(stmts) == 0 {
		return false
	}
	switch s := stmts[len(stmts)-1].(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.IfStmt:
		return blockAlwaysReturns(s.Body) && blockAlwaysReturns(s.ElseBody)
	default:
		return false
	}
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
		case *ast.RangeStmt:
			if hasReturnStmt(s.Body) {
				return true
			}
		}
	}
	return false
}

func (c *Checker) checkReturnStmt(s *ast.ReturnStmt) {
	if c.currentReturnType == nil {
		c.addError(s.Token.Pos, "bring used outside function")
		return
	}
	if s.Value == nil {
		if !types.IsAny(c.currentReturnType) {
			c.addError(s.Token.Pos, "Function requires a return value of type %s", c.currentReturnType)
		}
		return
	}
	valType := c.inferExpr(s.Value)
	if !types.IsAny(c.currentReturnType) && !types.IsAny(valType) {
		if !c.currentReturnType.Equals(valType) {
			c.addError(s.Token.Pos, "Return type mismatch: expected %s but got %s", c.currentReturnType, valType)
		}
	}
}

func (c *Checker) checkIfStmt(s *ast.IfStmt) {
	condType := types.Unwrap(c.inferExpr(s.Condition))
	if !types.IsAny(condType) {
		if _, ok := condType.(types.BoolType); !ok {
			c.addError(s.Token.Pos, "Condition must be bool, got %s", condType)
		}
	}
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

func (c *Checker) checkRangeStmt(s *ast.RangeStmt) {
	if s.Start != nil {
		startType := types.Unwrap(c.inferExpr(s.Start))
		if !types.IsAny(startType) {
			if _, ok := startType.(types.IntType); !ok {
				c.addError(s.Token.Pos, "Range start must be int, got %s", startType)
			}
		}
	}
	endType := types.Unwrap(c.inferExpr(s.End))
	if !types.IsAny(endType) {
		if _, ok := endType.(types.IntType); !ok {
			c.addError(s.Token.Pos, "Range end must be int, got %s", endType)
		}
	}
	c.pushScope()
	c.define(s.Var, types.IntType{})
	c.info.VarTypes[s.Var] = types.IntType{}
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
		return rightType
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
				continue
			}
			if !types.IsAny(armType) && !types.IsAny(t) && !armType.Equals(t) {
				c.addError(e.Token.Pos, "Match arms have inconsistent types: %s vs %s", armType, t)
				armType = types.AnyType{}
			}
		}
		if armType != nil {
			return armType
		}
		return types.AnyType{}
	case *ast.MemberExpr:
		objType := c.inferExpr(e.Object)
		if ct, ok := objType.(types.CollarType); ok {
			if e.Member == "value" {
				return ct.Underlying
			}
			c.addError(e.Token.Pos, "%s has no field %s", ct.Name, e.Member)
			return types.AnyType{}
		}
		if kt, ok := objType.(types.KittyType); ok {
			for _, f := range kt.Fields {
				if f.Name == e.Member {
					return f.Type
				}
			}
			c.addError(e.Token.Pos, "%s has no field %s", kt.Name, e.Member)
		}
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
		if types.IsNumeric(types.Unwrap(operand)) {
			return operand
		}
		c.addError(e.Token.Pos, "Cannot negate %s", operand)
		return types.AnyType{}
	case token.NOT:
		// NOT operates on truthiness, so it accepts any type.
		return types.BoolType{}
	}
	return types.AnyType{}
}

func (c *Checker) inferBinary(e *ast.BinaryExpr) types.Type {
	left := c.inferExpr(e.Left)
	right := c.inferExpr(e.Right)

	// Unwrap aliases for transparent type checking
	uleft := types.Unwrap(left)
	uright := types.Unwrap(right)

	// If either side is AnyType, skip checking
	if types.IsAny(uleft) || types.IsAny(uright) {
		switch e.Op {
		case token.EQ, token.NEQ, token.LT, token.GT, token.LTE, token.GTE, token.AND, token.OR:
			return types.BoolType{}
		default:
			return types.AnyType{}
		}
	}

	switch e.Op {
	case token.PLUS:
		if uleft.Equals(uright) {
			switch uleft.(type) {
			case types.IntType, types.FloatType, types.StringType:
				return left
			}
		}
		c.addError(e.Token.Pos, "Cannot add %s and %s", left, right)
		return types.AnyType{}

	case token.MINUS, token.STAR, token.SLASH:
		if uleft.Equals(uright) && types.IsNumeric(uleft) {
			return left
		}
		op := map[token.TokenType]string{token.MINUS: "subtract", token.STAR: "multiply", token.SLASH: "divide"}
		c.addError(e.Token.Pos, "Cannot %s %s and %s", op[e.Op], left, right)
		return types.AnyType{}

	case token.PERCENT:
		if uleft.Equals(uright) {
			if _, ok := uleft.(types.IntType); ok {
				return types.IntType{}
			}
		}
		c.addError(e.Token.Pos, "Cannot modulo %s and %s", left, right)
		return types.AnyType{}

	case token.EQ, token.NEQ:
		// Allow comparison between different collar types (runtime handles it)
		_, leftIsCollar := uleft.(types.CollarType)
		_, rightIsCollar := uright.(types.CollarType)
		if !(leftIsCollar && rightIsCollar) && !uleft.Equals(uright) {
			c.addError(e.Token.Pos, "Cannot compare %s and %s", left, right)
		}
		return types.BoolType{}

	case token.LT, token.GT, token.LTE, token.GTE:
		if uleft.Equals(uright) && types.IsNumeric(uleft) {
			return types.BoolType{}
		}
		c.addError(e.Token.Pos, "Cannot compare %s and %s", left, right)
		return types.BoolType{}

	case token.AND, token.OR:
		_, lok := uleft.(types.BoolType)
		_, rok := uright.(types.BoolType)
		if !lok || !rok {
			c.addError(e.Token.Pos, "Logical operator requires bool operands, got %s and %s", left, right)
		}
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

		// Check collar constructors
		if ct, ok := c.info.CollarTypes[ident.Name]; ok {
			if len(e.Args) != 1 {
				c.addError(e.Token.Pos, "%s expects 1 argument but got %d",
					ident.Name, len(e.Args))
			} else {
				argType := c.info.ExprTypes[e.Args[0]]
				if argType != nil && !types.IsAny(argType) && !types.IsAny(ct.Underlying) {
					if !ct.Underlying.Equals(argType) {
						c.addError(e.Token.Pos, "%s expects %s but got %s",
							ident.Name, ct.Underlying, argType)
					}
				}
			}
			return ct
		}

		// Check kitty constructors
		if kt, ok := c.info.KittyTypes[ident.Name]; ok {
			if len(e.Args) != len(kt.Fields) {
				c.addError(e.Token.Pos, "%s expects %d fields but got %d",
					ident.Name, len(kt.Fields), len(e.Args))
			}
			return kt
		}

		// Check user-defined functions
		if ft, ok := c.info.FuncTypes[ident.Name]; ok {
			if len(e.Args) != len(ft.Params) {
				c.addError(e.Token.Pos, "Function %s expects %d arguments but got %d",
					ident.Name, len(ft.Params), len(e.Args))
				return ft.Return
			}
			// Validate argument types
			for i, arg := range e.Args {
				argType := c.info.ExprTypes[arg]
				if argType != nil && !types.IsAny(argType) && !types.IsAny(ft.Params[i]) {
					if !ft.Params[i].Equals(argType) {
						c.addError(e.Token.Pos, "Argument %d: expected %s but got %s", i+1, ft.Params[i], argType)
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
		if p.TypeAnn == nil {
			c.addError(e.Token.Pos, "Lambda parameter %q must have a type annotation", p.Name)
		}
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
		t := c.inferExpr(item)
		if !types.IsAny(elemType) && !types.IsAny(t) && !elemType.Equals(t) {
			if isPrimitiveType(elemType) && isPrimitiveType(t) {
				c.addError(e.Token.Pos, "List elements must have consistent types: %s vs %s", elemType, t)
			}
			elemType = types.AnyType{}
		}
	}
	return types.ListType{Elem: elemType}
}
