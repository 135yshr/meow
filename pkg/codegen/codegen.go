package codegen

import (
	"fmt"
	"maps"
	"sort"
	"strings"

	"github.com/135yshr/meow/pkg/ast"
	"github.com/135yshr/meow/pkg/checker"
	"github.com/135yshr/meow/pkg/mutation"
	"github.com/135yshr/meow/pkg/token"
	"github.com/135yshr/meow/pkg/types"
)

// Generator produces Go source code from a Meow AST.
type Generator struct {
	funcs             []string
	topLevel          []string
	imports           map[string]string // meow pkg name → Go import path
	testMode          bool
	testFuncs         []string // names of test_ prefixed functions
	catwalkFuncs      []string // names of catwalk_ prefixed functions
	catwalkOutput     CatwalkOutput
	mutations         map[ast.Expr][]mutation.MutationEntry
	coverEnabled      bool
	coverFilename     string
	coverBlocks       []coverBlock
	typeInfo          *checker.TypeInfo
	currentReturnType types.Type // return type of the function currently being generated
	kittyDefs         map[string]*ast.KittyStmt
	collarDefs        map[string]*ast.CollarStmt
	learnDefs         []*ast.LearnStmt
	inLearnMethod     bool              // true when generating a learn method body
	aliasToPackage    map[string]string // alias → real package name
	packageToAlias    map[string]string // real package name → alias
	// nativeVars holds the identifiers currently emitted as native Go values
	// (int64, string, ...) rather than as meow.Value. It is populated while
	// generating a fully-typed function body, and is what lets the untyped
	// generator box such an identifier when handing it to a runtime helper.
	nativeVars map[string]types.Type
}

// bindNativeVar records that name is currently held in a native Go variable of
// type t, or that it is not, when t is not a native type.
func (g *Generator) bindNativeVar(name string, t types.Type) {
	if g.nativeVars == nil {
		return
	}
	if t != nil && isNativeType(t) {
		g.nativeVars[name] = t
		return
	}
	delete(g.nativeVars, name)
}

// enterNativeScope starts tracking native variables, returning a function that
// restores the previous set.
func (g *Generator) enterNativeScope() func() {
	prev := g.nativeVars
	g.nativeVars = make(map[string]types.Type, len(prev))
	maps.Copy(g.nativeVars, prev)
	return func() { g.nativeVars = prev }
}

// heldAsValue reports whether name is a variable the enclosing typed function
// keeps in a meow.Value rather than a native Go type.
//
// It answers false outside a typed function body, where nothing is tracked and
// the question does not arise.
func (g *Generator) heldAsValue(name string) bool {
	if g.nativeVars == nil {
		return false
	}
	_, native := g.nativeVars[name]
	return !native
}

// enterBoxedScope opens a scope in which the given names are bound to
// meow.Value variables. Whatever those names meant outside is shadowed, so they
// must stop counting as native — otherwise genIdent would box a value that is
// already boxed. Returns a function that restores the previous set.
func (g *Generator) enterBoxedScope(names ...string) func() {
	restore := g.enterNativeScope()
	for _, name := range names {
		g.bindNativeVar(name, nil)
	}
	return restore
}

type coverBlock struct {
	startLine, startCol, endLine, endCol, numStmt int
}

var stdPackages = map[string]string{
	"aws":     "github.com/135yshr/meow/runtime/aws",
	"clock":   "github.com/135yshr/meow/runtime/clock",
	"env":     "github.com/135yshr/meow/runtime/env",
	"file":    "github.com/135yshr/meow/runtime/file",
	"http":    "github.com/135yshr/meow/runtime/http",
	"random":  "github.com/135yshr/meow/runtime/random",
	"testing": "github.com/135yshr/meow/runtime/testing",
}

// resolveImportName resolves name to a real package name, considering aliases.
// Returns (realPkg, true) if the name refers to an imported package,
// or ("", false) if it does not.
func (g *Generator) resolveImportName(name string) (string, bool) {
	if rp, isAlias := g.aliasToPackage[name]; isAlias {
		return rp, true
	}
	if _, imported := g.imports[name]; imported {
		if _, hasAlias := g.packageToAlias[name]; !hasAlias {
			return name, true
		}
		return "", false
	}
	return "", false
}

// capitalizeFirst returns s with its first byte uppercased.
// Returns s unchanged if s is empty.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// New creates a new code generator.
func New() *Generator {
	return &Generator{}
}

// NewTest creates a code generator in test mode.
func NewTest() *Generator {
	return &Generator{testMode: true}
}

// SetTypeInfo stores type checking results for use in code generation.
func (g *Generator) SetTypeInfo(ti *checker.TypeInfo) {
	g.typeInfo = ti
}

// SetCatwalkOutput sets the expected output map for catwalk_ functions.
func (g *Generator) SetCatwalkOutput(co CatwalkOutput) {
	g.catwalkOutput = co
}

// SetMutations sets the mutation schema for schemata-based mutation testing.
func (g *Generator) SetMutations(m map[ast.Expr][]mutation.MutationEntry) {
	g.mutations = m
}

// EnableCoverage activates statement coverage instrumentation.
func (g *Generator) EnableCoverage(filename string) {
	g.coverEnabled = true
	g.coverFilename = filename
}

// Generate produces Go source code from a Program AST.
func (g *Generator) Generate(prog *ast.Program) (string, error) {
	g.collectKittyDefs(prog)
	for _, stmt := range prog.Stmts {
		switch stmt.(type) {
		case *ast.KittyStmt, *ast.BreedStmt, *ast.CollarStmt, *ast.TrickStmt:
			continue
		}
		if ls, ok := stmt.(*ast.LearnStmt); ok {
			g.learnDefs = append(g.learnDefs, ls)
			for i := range ls.Methods {
				g.funcs = append(g.funcs, g.genLearnMethod(ls.TypeName, &ls.Methods[i]))
			}
			continue
		}
		if fn, ok := stmt.(*ast.FuncStmt); ok {
			g.funcs = append(g.funcs, g.genFuncDecl(fn))
		} else {
			code, err := g.genStmtOrError(stmt)
			if err != nil {
				return "", err
			}
			if code != "" {
				g.topLevel = append(g.topLevel, code)
			}
		}
	}
	return g.emit(), nil
}

// GenerateTest produces Go source code from a Program AST in test mode.
// It auto-imports the testing package and wraps test_ functions with Run/Report
// and catwalk_ functions with Catwalk.
func (g *Generator) GenerateTest(prog *ast.Program) (string, error) {
	g.collectKittyDefs(prog)
	for _, stmt := range prog.Stmts {
		switch stmt.(type) {
		case *ast.KittyStmt, *ast.BreedStmt, *ast.CollarStmt, *ast.TrickStmt:
			continue
		}
		if ls, ok := stmt.(*ast.LearnStmt); ok {
			g.learnDefs = append(g.learnDefs, ls)
			for i := range ls.Methods {
				g.funcs = append(g.funcs, g.genLearnMethod(ls.TypeName, &ls.Methods[i]))
			}
			continue
		}
		if fn, ok := stmt.(*ast.FuncStmt); ok {
			g.funcs = append(g.funcs, g.genFuncDecl(fn))
			if strings.HasPrefix(fn.Name, "test_") {
				if len(fn.Params) != 0 {
					return "", fmt.Errorf("test function %s must not take parameters", fn.Name)
				}
				g.testFuncs = append(g.testFuncs, fn.Name)
			} else if strings.HasPrefix(fn.Name, "catwalk_") {
				if len(fn.Params) != 0 {
					return "", fmt.Errorf("catwalk function %s must not take parameters", fn.Name)
				}
				if g.catwalkOutput != nil {
					if _, ok := g.catwalkOutput[fn.Name]; !ok {
						return "", fmt.Errorf("catwalk function %s has no # Output: block", fn.Name)
					}
				}
				g.catwalkFuncs = append(g.catwalkFuncs, fn.Name)
			}
		} else {
			code, err := g.genStmtOrError(stmt)
			if err != nil {
				return "", err
			}
			if code != "" {
				g.topLevel = append(g.topLevel, code)
			}
		}
	}
	g.ensureImport("testing")
	return g.emitTest(), nil
}

func (g *Generator) emitTest() string {
	var b strings.Builder
	b.WriteString("// Code generated by meow compiler. DO NOT EDIT.\n")
	b.WriteString("package main\n\n")
	b.WriteString("import meow \"github.com/135yshr/meow/runtime/meowrt\"\n")
	needsOS := g.coverEnabled || len(g.mutations) > 0
	if needsOS {
		b.WriteString("import \"os\"\n")
	}
	if len(g.mutations) > 0 {
		b.WriteString("import \"strconv\"\n")
	}
	if g.coverEnabled {
		b.WriteString("import meow_coverage \"github.com/135yshr/meow/runtime/coverage\"\n")
	}
	names := make([]string, 0, len(g.imports))
	for name := range g.imports {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(&b, "import meow_%s \"%s\"\n", name, g.imports[name])
	}
	b.WriteString("\n")

	if len(g.mutations) > 0 {
		b.WriteString("var __mutant int = -1\n\n")
	}

	needsInit := g.coverEnabled || len(g.mutations) > 0
	if needsInit {
		b.WriteString("func init() {\n")
		if len(g.mutations) > 0 {
			b.WriteString("\tif s := os.Getenv(\"MEOW_MUTANT\"); s != \"\" {\n")
			b.WriteString("\t\tif v, err := strconv.Atoi(s); err == nil {\n")
			b.WriteString("\t\t\t__mutant = v\n")
			b.WriteString("\t\t}\n")
			b.WriteString("\t}\n")
		}
		if g.coverEnabled {
			for i, cb := range g.coverBlocks {
				fmt.Fprintf(&b, "\tmeow_coverage.Register(%q, %d, %d, %d, %d, %d) // block %d\n",
					g.coverFilename, cb.startLine, cb.startCol, cb.endLine, cb.endCol, cb.numStmt, i)
			}
		}
		b.WriteString("}\n\n")
	}

	if initCode := g.genLearnInit(); initCode != "" {
		b.WriteString(initCode)
		b.WriteString("\n")
	}

	for _, fn := range g.funcs {
		b.WriteString(fn)
		b.WriteString("\n\n")
	}

	// Test-mode main wraps top-level work and emits Furball errors before
	// running test_/catwalk_ functions. RunMain handles both Furball returns
	// and internal panics so failures land as clean stderr messages.
	b.WriteString("func main() {\n")
	if len(g.topLevel) > 0 {
		b.WriteString("\tmeow.RunMain(func() meow.Value {\n")
		for _, line := range g.topLevel {
			b.WriteString("\t\t")
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\t\treturn meow.NewNil()\n")
		b.WriteString("\t})\n")
	}
	for _, name := range g.testFuncs {
		fmt.Fprintf(&b, "\tmeow_testing.Run(meow.NewString(%q), meow.NewFunc(%q, func(args ...meow.Value) meow.Value {\n", name, name)
		fmt.Fprintf(&b, "\t\treturn %s()\n", name)
		fmt.Fprintf(&b, "\t}))\n")
	}
	for _, name := range g.catwalkFuncs {
		expected := ""
		if g.catwalkOutput != nil {
			expected = g.catwalkOutput[name]
		}
		fmt.Fprintf(&b, "\tmeow_testing.Catwalk(meow.NewString(%q), meow.NewFunc(%q, func(args ...meow.Value) meow.Value {\n", name, name)
		fmt.Fprintf(&b, "\t\treturn %s()\n", name)
		fmt.Fprintf(&b, "\t}), meow.NewString(%q))\n", expected)
	}
	if g.coverEnabled {
		b.WriteString("\tmeow_coverage.Report(os.Stdout)\n")
		b.WriteString("\tif p := os.Getenv(\"MEOW_COVERPROFILE\"); p != \"\" {\n")
		b.WriteString("\t\tmeow_coverage.WriteProfile(p)\n")
		b.WriteString("\t}\n")
	}
	b.WriteString("\tmeow_testing.Report()\n")
	b.WriteString("}\n")
	return b.String()
}

func (g *Generator) ensureImport(name string) {
	if g.imports == nil {
		g.imports = make(map[string]string)
	}
	if _, ok := g.imports[name]; !ok {
		if path, ok := stdPackages[name]; ok {
			g.imports[name] = path
		}
	}
}

func (g *Generator) needsMeowImport() bool {
	if len(g.topLevel) > 0 {
		return true
	}
	for _, fn := range g.funcs {
		if strings.Contains(fn, "meow.") {
			return true
		}
	}
	return false
}

func (g *Generator) emit() string {
	var b strings.Builder
	b.WriteString("// Code generated by meow compiler. DO NOT EDIT.\n")
	b.WriteString("package main\n\n")
	if g.needsMeowImport() {
		b.WriteString("import meow \"github.com/135yshr/meow/runtime/meowrt\"\n")
	}
	if len(g.mutations) > 0 {
		b.WriteString("import \"os\"\n")
		b.WriteString("import \"strconv\"\n")
	}
	names := make([]string, 0, len(g.imports))
	for name := range g.imports {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(&b, "import meow_%s \"%s\"\n", name, g.imports[name])
	}
	b.WriteString("\n")

	if len(g.mutations) > 0 {
		b.WriteString("var __mutant int = -1\n\n")
		b.WriteString("func init() {\n")
		b.WriteString("\tif s := os.Getenv(\"MEOW_MUTANT\"); s != \"\" {\n")
		b.WriteString("\t\tif v, err := strconv.Atoi(s); err == nil {\n")
		b.WriteString("\t\t\t__mutant = v\n")
		b.WriteString("\t\t}\n")
		b.WriteString("\t}\n")
		b.WriteString("}\n\n")
	}

	if initCode := g.genLearnInit(); initCode != "" {
		b.WriteString(initCode)
		b.WriteString("\n")
	}

	for _, fn := range g.funcs {
		b.WriteString(fn)
		b.WriteString("\n\n")
	}

	// Wrap top-level statements in an inner function so the short-circuit
	// `return __f` pattern (injected by genStmtInner for Furball propagation)
	// is well-typed. main() then prints any surfaced Furball to stderr and
	// exits, replacing the old panic-based termination.
	if len(g.topLevel) > 0 && g.needsMeowImport() {
		b.WriteString("func __meow_main() meow.Value {\n")
		for _, line := range g.topLevel {
			b.WriteString("\t")
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\treturn meow.NewNil()\n")
		b.WriteString("}\n\n")
		b.WriteString("func main() {\n")
		// RunMain handles both Furball returns and internal As*/hiss panics
		// (from typed paths), producing a clean stderr message + exit 1.
		b.WriteString("\tmeow.RunMain(__meow_main)\n")
		b.WriteString("}\n")
	} else {
		b.WriteString("func main() {\n")
		for _, line := range g.topLevel {
			b.WriteString("\t")
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("}\n")
	}
	return b.String()
}

func (g *Generator) genFuncDecl(fn *ast.FuncStmt) string {
	if g.isFullyTypedFunc(fn) {
		return g.genTypedFuncDecl(fn)
	}
	var b strings.Builder
	params := make([]string, len(fn.Params))
	names := make([]string, len(fn.Params))
	for i, p := range fn.Params {
		params[i] = p.Name + " meow.Value"
		names[i] = p.Name
	}
	defer g.enterBoxedScope(names...)()
	fmt.Fprintf(&b, "func %s(%s) meow.Value {\n", fn.Name, strings.Join(params, ", "))
	for _, stmt := range fn.Body {
		b.WriteString("\t")
		b.WriteString(g.genStmt(stmt))
		b.WriteString("\n")
	}
	if !g.blockAlwaysReturns(fn.Body) {
		b.WriteString("\treturn meow.NewNil()\n")
	}
	b.WriteString("}")
	return b.String()
}

func (g *Generator) isFullyTypedFunc(fn *ast.FuncStmt) bool {
	if g.typeInfo == nil {
		return false
	}
	ft, ok := g.typeInfo.FuncTypes[fn.Name]
	if !ok {
		return false
	}
	return isFullyTypedFuncType(ft)
}

func (g *Generator) genTypedFuncDecl(fn *ast.FuncStmt) string {
	ft := g.typeInfo.FuncTypes[fn.Name]
	prevReturnType := g.currentReturnType
	g.currentReturnType = ft.Return
	defer func() { g.currentReturnType = prevReturnType }()
	defer g.enterNativeScope()()
	var b strings.Builder
	params := make([]string, len(fn.Params))
	for i, p := range fn.Params {
		params[i] = p.Name + " " + goTypeString(ft.Params[i])
		g.bindNativeVar(p.Name, ft.Params[i])
	}
	fmt.Fprintf(&b, "func %s(%s) %s {\n", fn.Name, strings.Join(params, ", "), goTypeString(ft.Return))
	for _, stmt := range fn.Body {
		b.WriteString("\t")
		b.WriteString(g.genTypedStmt(stmt))
		b.WriteString("\n")
	}
	b.WriteString("}")
	return b.String()
}

func goTypeString(t types.Type) string {
	switch t := t.(type) {
	case types.IntType:
		return "int64"
	case types.ByteType:
		return "byte"
	case types.FloatType:
		return "float64"
	case types.StringType:
		return "string"
	case types.BoolType:
		return "bool"
	case types.AliasType:
		return goTypeString(t.Underlying)
	default:
		return "meow.Value"
	}
}

func (g *Generator) genTypedStmt(stmt ast.Stmt) string {
	switch s := stmt.(type) {
	case *ast.VarStmt:
		return g.genTypedVarStmt(s)
	case *ast.ReturnStmt:
		return g.genTypedReturnStmt(s)
	case *ast.ExprStmt:
		return g.genTypedExprStmt(s)
	case *ast.IfStmt:
		return g.genTypedIf(s)
	case *ast.RangeStmt:
		return g.genTypedRange(s)
	default:
		return g.genStmt(stmt)
	}
}

func (g *Generator) genTypedVarStmt(s *ast.VarStmt) string {
	t := g.getExprType(s.Value)
	if t != nil && !types.IsAny(t) {
		// Generate the value before rebinding, so that `nyan x = x + 1` reads
		// the old binding.
		code := fmt.Sprintf("var %s %s = %s", s.Name, goTypeString(t), g.genTypedExpr(s.Value))
		g.bindNativeVar(s.Name, t)
		return code
	}
	code := fmt.Sprintf("var %s meow.Value = %s", s.Name, g.genExpr(s.Value))
	g.bindNativeVar(s.Name, nil)
	return code
}

func (g *Generator) genTypedReturnStmt(s *ast.ReturnStmt) string {
	if s.Value == nil {
		return "return"
	}
	t := g.getExprType(s.Value)
	if t != nil && !types.IsAny(t) {
		return fmt.Sprintf("return %s", g.genTypedExpr(s.Value))
	}
	// Expression is AnyType (e.g. match expression) but function has a concrete return type.
	// Generate the expression with typed boxing, then unbox the meow.Value result.
	exprCode := g.genTypedExpr(s.Value)
	if g.currentReturnType != nil && !types.IsAny(g.currentReturnType) {
		return fmt.Sprintf("return %s", unboxToNative(exprCode, g.currentReturnType))
	}
	return fmt.Sprintf("return %s", exprCode)
}

func (g *Generator) genTypedExprStmt(s *ast.ExprStmt) string {
	if call, ok := s.Expr.(*ast.CallExpr); ok {
		return g.genTypedCall(call)
	}
	return g.genExpr(s.Expr)
}

func (g *Generator) genTypedIf(s *ast.IfStmt) string {
	var b strings.Builder
	condType := g.getExprType(s.Condition)
	if condType != nil && !types.IsAny(condType) {
		fmt.Fprintf(&b, "if %s {\n", g.genTypedExpr(s.Condition))
	} else {
		fmt.Fprintf(&b, "if (%s).IsTruthy() {\n", g.genExpr(s.Condition))
	}
	for _, stmt := range s.Body {
		b.WriteString("\t")
		b.WriteString(g.genTypedStmt(stmt))
		b.WriteString("\n")
	}
	if len(s.ElseBody) > 0 {
		b.WriteString("} else {\n")
		for _, stmt := range s.ElseBody {
			b.WriteString("\t")
			b.WriteString(g.genTypedStmt(stmt))
			b.WriteString("\n")
		}
	}
	b.WriteString("}")
	return b.String()
}

func (g *Generator) genTypedRange(s *ast.RangeStmt) string {
	endType := g.getExprType(s.End)
	if endType != nil {
		endType = types.Unwrap(endType)
	}
	if _, ok := endType.(types.ListType); ok && s.Start == nil && !s.Inclusive {
		return g.genTypedListRange(s)
	}
	var b strings.Builder
	cmp := "<"
	if s.Inclusive {
		cmp = "<="
	}
	startCode := "int64(0)"
	if s.Start != nil {
		startType := g.getExprType(s.Start)
		if startType != nil && !types.IsAny(startType) {
			startCode = g.genTypedExpr(s.Start)
		} else {
			startCode = ""
		}
	}
	// The loop variable is an int64 in both forms below, so it needs boxing
	// wherever the body hands it to a runtime helper.
	defer g.enterNativeScope()()
	g.bindNativeVar(s.Var, types.IntType{})
	if startCode != "" && endType != nil && !types.IsAny(endType) {
		fmt.Fprintf(&b, "for %s := %s; %s %s %s; %s++ {\n",
			s.Var, startCode, s.Var, cmp, g.genTypedExpr(s.End), s.Var)
	} else {
		startExpr := "meow.NewInt(0)"
		if s.Start != nil {
			startExpr = g.boxValue(s.Start)
		}
		fmt.Fprintf(&b, "for __i := meow.AsInt(%s); __i %s meow.AsInt(%s); __i++ {\n",
			startExpr, cmp, g.boxValue(s.End))
		fmt.Fprintf(&b, "\tvar %s int64 = __i\n", s.Var)
		fmt.Fprintf(&b, "\t_ = %s\n", s.Var)
	}
	for _, stmt := range s.Body {
		b.WriteString("\t")
		b.WriteString(g.genTypedStmt(stmt))
		b.WriteString("\n")
	}
	b.WriteString("}")
	return b.String()
}

func (g *Generator) getExprType(expr ast.Expr) types.Type {
	if g.typeInfo == nil {
		return nil
	}
	return g.typeInfo.ExprTypes[expr]
}

func (g *Generator) genTypedExpr(expr ast.Expr) string {
	t := g.getExprType(expr)
	// Handle match expressions specially: they always produce meow.Value
	// but may contain typed variables that need boxing.
	if _, isMatch := expr.(*ast.MatchExpr); isMatch {
		matchCode := g.genTypedMatch(expr.(*ast.MatchExpr))
		if t != nil && isNativeType(t) {
			return unboxToNative(matchCode, t)
		}
		return matchCode
	}
	if t == nil || types.IsAny(t) {
		return g.genExprBoxed(expr)
	}
	switch e := expr.(type) {
	case *ast.IntLit:
		return fmt.Sprintf("int64(%d)", e.Value)
	case *ast.FloatLit:
		return fmt.Sprintf("float64(%g)", e.Value)
	case *ast.StringLit:
		return fmt.Sprintf("%q", e.Value)
	case *ast.BoolLit:
		if e.Value {
			return "true"
		}
		return "false"
	case *ast.Ident:
		// Wanted as a native Go value, but stored boxed — unwrap it.
		if g.heldAsValue(e.Name) && isNativeType(t) {
			return unboxToNative(e.Name, t)
		}
		return e.Name
	case *ast.UnaryExpr:
		return g.genTypedUnary(e)
	case *ast.BinaryExpr:
		return g.genTypedBinary(e)
	case *ast.CallExpr:
		return g.genTypedCall(e)
	default:
		return g.genExprBoxed(expr)
	}
}

// genExprBoxed generates an expression in untyped (meow.Value) mode. Native Go
// variables referenced anywhere within it are boxed by genIdent, which knows
// which identifiers are actually held natively — the inferred type alone does
// not say that, since a typed function can hold an int-typed local in a
// meow.Value.
func (g *Generator) genExprBoxed(expr ast.Expr) string {
	return g.genExpr(expr)
}

// boxNative wraps a native Go value in its meow.Value constructor.
func (g *Generator) boxNative(name string, t types.Type) string {
	switch types.Unwrap(t).(type) {
	case types.IntType:
		return fmt.Sprintf("meow.NewInt(%s)", name)
	case types.ByteType:
		return fmt.Sprintf("meow.NewByte(%s)", name)
	case types.FloatType:
		return fmt.Sprintf("meow.NewFloat(%s)", name)
	case types.StringType:
		return fmt.Sprintf("meow.NewString(%s)", name)
	case types.BoolType:
		return fmt.Sprintf("meow.NewBool(%s)", name)
	default:
		return name
	}
}

// genTypedMatch generates a match expression where the subject may be a native type.
// It boxes the subject to meow.Value and returns the match result as meow.Value.
func (g *Generator) genTypedMatch(e *ast.MatchExpr) string {
	var b strings.Builder
	subject := g.boxValue(e.Subject)
	b.WriteString(fmt.Sprintf("func() meow.Value {\n\t__subject := %s\n", subject))
	hasCondArm := false
	for i, arm := range e.Arms {
		if _, ok := arm.Pattern.(*ast.WildcardPattern); ok {
			if hasCondArm {
				b.WriteString("\t} else {\n")
				b.WriteString(fmt.Sprintf("\t\treturn %s\n", g.boxValue(arm.Body)))
				b.WriteString("\t}\n")
			} else {
				b.WriteString(fmt.Sprintf("\treturn %s\n", g.boxValue(arm.Body)))
			}
			b.WriteString("\treturn meow.NewNil()\n}()")
			return b.String()
		}
		keyword := "if"
		if i > 0 {
			keyword = "} else if"
		}
		hasCondArm = true
		b.WriteString(fmt.Sprintf("\t%s %s {\n", keyword, g.genPatternCond("__subject", arm.Pattern)))
		b.WriteString(fmt.Sprintf("\t\treturn %s\n", g.boxValue(arm.Body)))
	}
	if hasCondArm {
		b.WriteString("\t}\n")
	}
	b.WriteString("\treturn meow.NewNil()\n}()")
	return b.String()
}

func (g *Generator) genTypedUnary(e *ast.UnaryExpr) string {
	switch e.Op {
	case token.MINUS:
		return fmt.Sprintf("(-%s)", g.genTypedExpr(e.Right))
	case token.NOT:
		if _, ok := g.getExprType(e.Right).(types.BoolType); ok {
			return fmt.Sprintf("(!%s)", g.genTypedExpr(e.Right))
		}
		return fmt.Sprintf("(!(%s).IsTruthy())", g.boxValue(e.Right))
	}
	return g.genUnary(e)
}

// nilOperand reports whether an operand is catnap — the literal, or a value
// the checker typed as nil. Such a comparison cannot use Go's ==: NewNil
// allocates, so comparing boxed values would compare pointers and always be
// false. It has to go through the runtime instead.
func (g *Generator) nilOperand(e ast.Expr) bool {
	if _, ok := e.(*ast.NilLit); ok {
		return true
	}
	t := g.getExprType(e)
	if t == nil {
		return false
	}
	_, ok := types.Unwrap(t).(types.NilType)
	return ok
}

func (g *Generator) genTypedBinary(e *ast.BinaryExpr) string {
	left := g.genTypedExpr(e.Left)
	right := g.genTypedExpr(e.Right)
	switch e.Op {
	case token.PLUS:
		return fmt.Sprintf("(%s + %s)", left, right)
	case token.MINUS:
		return fmt.Sprintf("(%s - %s)", left, right)
	case token.STAR:
		return fmt.Sprintf("(%s * %s)", left, right)
	case token.SLASH:
		return fmt.Sprintf("(%s / %s)", left, right)
	case token.PERCENT:
		return fmt.Sprintf("(%s %% %s)", left, right)
	case token.EQ:
		if g.nilOperand(e.Left) || g.nilOperand(e.Right) {
			return fmt.Sprintf("meow.Equal(%s, %s).IsTruthy()", g.boxValue(e.Left), g.boxValue(e.Right))
		}
		return fmt.Sprintf("(%s == %s)", left, right)
	case token.NEQ:
		if g.nilOperand(e.Left) || g.nilOperand(e.Right) {
			return fmt.Sprintf("meow.NotEqual(%s, %s).IsTruthy()", g.boxValue(e.Left), g.boxValue(e.Right))
		}
		return fmt.Sprintf("(%s != %s)", left, right)
	case token.LT:
		return fmt.Sprintf("(%s < %s)", left, right)
	case token.GT:
		return fmt.Sprintf("(%s > %s)", left, right)
	case token.LTE:
		return fmt.Sprintf("(%s <= %s)", left, right)
	case token.GTE:
		return fmt.Sprintf("(%s >= %s)", left, right)
	case token.AND:
		return fmt.Sprintf("(%s && %s)", left, right)
	case token.OR:
		return fmt.Sprintf("(%s || %s)", left, right)
	}
	return g.genBinary(e)
}

func (g *Generator) genTypedCall(e *ast.CallExpr) string {
	// Handle method calls on learn types: unbox dispatch result if typed
	if member, ok := e.Fn.(*ast.MemberExpr); ok {
		call := g.genMemberCall(member, e.Args)
		if t := g.getExprType(e); t != nil && !types.IsAny(t) && isNativeType(t) {
			return unboxToNative(call, t)
		}
		return call
	}

	ident, isIdent := e.Fn.(*ast.Ident)
	if !isIdent {
		return g.genCall(e)
	}

	// Builtin functions that need boxing
	switch ident.Name {
	case "nya":
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = g.boxValue(a)
		}
		return fmt.Sprintf("meow.Nya(%s)", strings.Join(args, ", "))
	case "hiss":
		// In typed contexts a function returns a native Go type (int64, etc.)
		// and cannot return a Furball value. Panic so that `gag`'s deferred
		// recover converts the failure into a Furball at the boundary —
		// this is the typed-path bridge to the value-propagation model.
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = g.boxValue(a)
		}
		return fmt.Sprintf("panic(meow.Hiss(%s).String())", strings.Join(args, ", "))
	case "judge", "expect", "refuse":
		g.ensureImport("testing")
		fn := capitalizeFirst(ident.Name)
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = g.boxValue(a)
		}
		return fmt.Sprintf("meow_testing.%s(%s)", fn, strings.Join(args, ", "))
	case "to_string", "to_int", "to_float", "to_bytes", "to_runes", "is_furball", "gag", "len",
		"head", "tail", "append", "lick", "picky", "curl",
		"whiff", "track", "shred", "tangle", "nibble":
		builtinNames := map[string]string{
			"to_string":  "ToString",
			"to_int":     "ToInt",
			"to_float":   "ToFloat",
			"to_bytes":   "ToBytes",
			"to_runes":   "ToRunes",
			"is_furball": "IsFurball",
			"gag":        "Gag",
			"len":        "Len",
			"head":       "Head",
			"tail":       "Tail",
			"append":     "Append",
			"lick":       "Lick",
			"picky":      "Picky",
			"curl":       "Curl",
			"whiff":      "Whiff",
			"track":      "Track",
			"shred":      "Shred",
			"tangle":     "Tangle",
			"nibble":     "Nibble",
		}
		// Known return types for builtins that produce typed results
		builtinRetTypes := map[string]types.Type{
			"to_string":  types.StringType{},
			"to_int":     types.IntType{},
			"to_float":   types.FloatType{},
			"is_furball": types.BoolType{},
			"len":        types.IntType{},
			"whiff":      types.BoolType{},
			"track":      types.IntType{},
			"tangle":     types.StringType{},
			"nibble":     types.StringType{},
		}
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = g.boxValue(a)
		}
		call := fmt.Sprintf("meow.%s(%s)", builtinNames[ident.Name], strings.Join(args, ", "))
		// If the builtin has a known return type and we're in a typed context, unbox
		if retType, ok := builtinRetTypes[ident.Name]; ok {
			return unboxToNative(call, retType)
		}
		return call
	}

	// Collar constructors
	if _, ok := g.collarDefs[ident.Name]; ok {
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = g.boxValue(a)
		}
		return fmt.Sprintf("meow.NewKitty(%q, []string{\"value\"}, %s)",
			ident.Name, strings.Join(args, ", "))
	}

	// Typed user-defined functions (only for fully native signatures)
	if ft, ok := g.typeInfo.FuncTypes[ident.Name]; ok {
		if len(e.Args) < len(ft.Params) {
			return g.genPartialCall(ident.Name, ft, e.Args)
		}
		if isFullyTypedFuncType(ft) {
			args := make([]string, len(e.Args))
			for i, a := range e.Args {
				args[i] = g.genTypedExpr(a)
			}
			return fmt.Sprintf("%s(%s)", ident.Name, strings.Join(args, ", "))
		}
	}

	return g.genCall(e)
}

func (g *Generator) boxValue(expr ast.Expr) string {
	// The inferred type says what a value is, not how it is stored. A `purr`
	// element is a meow.Value even where the checker knows its element type, so
	// boxing on the type alone would wrap it a second time.
	if ident, ok := expr.(*ast.Ident); ok && g.heldAsValue(ident.Name) {
		return ident.Name
	}
	t := g.getExprType(expr)
	if t == nil || types.IsAny(t) {
		return g.genExpr(expr)
	}
	typed := g.genTypedExpr(expr)
	switch types.Unwrap(t).(type) {
	case types.IntType:
		return fmt.Sprintf("meow.NewInt(%s)", typed)
	case types.ByteType:
		return fmt.Sprintf("meow.NewByte(%s)", typed)
	case types.FloatType:
		return fmt.Sprintf("meow.NewFloat(%s)", typed)
	case types.StringType:
		return fmt.Sprintf("meow.NewString(%s)", typed)
	case types.BoolType:
		return fmt.Sprintf("meow.NewBool(%s)", typed)
	default:
		return g.genExpr(expr)
	}
}

// isNativeType reports whether t maps to a native Go type (int64, float64, string, bool).
// ListType, FurballType, and AnyType are NOT native types; they use meow.Value.
func isNativeType(t types.Type) bool {
	switch t.(type) {
	case types.IntType, types.ByteType, types.FloatType, types.StringType, types.BoolType:
		return true
	case types.AliasType:
		return isNativeType(types.Unwrap(t))
	}
	return false
}

func isFullyTypedFuncType(ft types.FuncType) bool {
	if !isNativeType(ft.Return) {
		return false
	}
	for _, p := range ft.Params {
		if !isNativeType(p) {
			return false
		}
	}
	return true
}

func isLiteralExpr(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.IntLit, *ast.FloatLit, *ast.StringLit, *ast.BoolLit:
		return true
	case *ast.UnaryExpr:
		return isLiteralExpr(e.Right)
	default:
		return false
	}
}

func unboxToNative(boxedExpr string, targetType types.Type) string {
	switch types.Unwrap(targetType).(type) {
	case types.IntType:
		return fmt.Sprintf("meow.AsInt(%s)", boxedExpr)
	case types.ByteType:
		return fmt.Sprintf("meow.AsByte(%s)", boxedExpr)
	case types.FloatType:
		return fmt.Sprintf("meow.AsFloat(%s)", boxedExpr)
	case types.StringType:
		return fmt.Sprintf("meow.AsString(%s)", boxedExpr)
	case types.BoolType:
		return fmt.Sprintf("meow.AsBool(%s)", boxedExpr)
	default:
		return boxedExpr
	}
}

func boxNativeCall(call string, retType types.Type) string {
	switch types.Unwrap(retType).(type) {
	case types.IntType:
		return fmt.Sprintf("meow.NewInt(%s)", call)
	case types.ByteType:
		return fmt.Sprintf("meow.NewByte(%s)", call)
	case types.FloatType:
		return fmt.Sprintf("meow.NewFloat(%s)", call)
	case types.StringType:
		return fmt.Sprintf("meow.NewString(%s)", call)
	case types.BoolType:
		return fmt.Sprintf("meow.NewBool(%s)", call)
	default:
		return call
	}
}

func (g *Generator) blockAlwaysReturns(stmts []ast.Stmt) bool {
	if len(stmts) == 0 {
		return false
	}
	switch s := stmts[len(stmts)-1].(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.IfStmt:
		return g.blockAlwaysReturns(s.Body) && g.blockAlwaysReturns(s.ElseBody)
	default:
		return false
	}
}

func (g *Generator) genStmt(stmt ast.Stmt) string {
	code := g.genStmtInner(stmt)
	if !g.coverEnabled {
		return code
	}
	pos := stmt.Pos()
	endLine, endCol := g.estimateEndPos(stmt)
	id := len(g.coverBlocks)
	g.coverBlocks = append(g.coverBlocks, coverBlock{pos.Line, pos.Column, endLine, endCol, 1})
	return fmt.Sprintf("meow_coverage.Hit(%d)\n%s", id, code)
}

func (g *Generator) genStmtInner(stmt ast.Stmt) string {
	switch s := stmt.(type) {
	case *ast.VarStmt:
		// Bind, then short-circuit if the value is a Furball — the untyped
		// path's error-propagation point.
		code := fmt.Sprintf(
			"var %s meow.Value = %s\n\tif __f, __ok := meow.AsFurball(%s); __ok { return __f }",
			s.Name, g.genExpr(s.Value), s.Name)
		g.bindNativeVar(s.Name, nil)
		return code
	case *ast.ReturnStmt:
		if s.Value != nil {
			return fmt.Sprintf("return %s", g.genExpr(s.Value))
		}
		return "return meow.NewNil()"
	case *ast.ExprStmt:
		// Evaluate the expression. If the result is a Furball (e.g. from
		// hiss(...) or a failed runtime helper), short-circuit by returning it.
		return fmt.Sprintf(
			"if __f, __ok := meow.AsFurball(%s); __ok { return __f }",
			g.genExpr(s.Expr))
	case *ast.IfStmt:
		return g.genIf(s)
	case *ast.RangeStmt:
		return g.genRange(s)
	default:
		return fmt.Sprintf("/* unsupported stmt: %T */", stmt)
	}
}

func (g *Generator) estimateEndPos(stmt ast.Stmt) (int, int) {
	pos := stmt.Pos()
	switch s := stmt.(type) {
	case *ast.IfStmt:
		body := s.ElseBody
		if len(body) == 0 {
			body = s.Body
		}
		if len(body) > 0 {
			last := body[len(body)-1]
			endLine, _ := g.estimateEndPos(last)
			return endLine + 1, 1
		}
		return pos.Line + 1, 1
	case *ast.RangeStmt:
		if len(s.Body) > 0 {
			last := s.Body[len(s.Body)-1]
			endLine, _ := g.estimateEndPos(last)
			return endLine + 1, 1
		}
		return pos.Line + 1, 1
	default:
		return pos.Line, pos.Column + 1
	}
}

func (g *Generator) collectKittyDefs(prog *ast.Program) {
	g.kittyDefs = make(map[string]*ast.KittyStmt)
	g.collarDefs = make(map[string]*ast.CollarStmt)
	for _, stmt := range prog.Stmts {
		if ks, ok := stmt.(*ast.KittyStmt); ok {
			g.kittyDefs[ks.Name] = ks
		}
		if cs, ok := stmt.(*ast.CollarStmt); ok {
			g.collarDefs[cs.Name] = cs
		}
	}
}

func (g *Generator) genStmtOrError(stmt ast.Stmt) (string, error) {
	switch stmt.(type) {
	case *ast.KittyStmt, *ast.BreedStmt, *ast.CollarStmt, *ast.TrickStmt, *ast.LearnStmt:
		return "", nil
	}
	if s, ok := stmt.(*ast.FetchStmt); ok {
		path, ok := stdPackages[s.Path]
		if !ok {
			return "", fmt.Errorf("unknown package: %s", s.Path)
		}
		if g.imports == nil {
			g.imports = make(map[string]string)
		}
		g.imports[s.Path] = path
		if s.Alias != "" {
			if g.aliasToPackage == nil {
				g.aliasToPackage = make(map[string]string)
			}
			if g.packageToAlias == nil {
				g.packageToAlias = make(map[string]string)
			}
			g.aliasToPackage[s.Alias] = s.Path
			g.packageToAlias[s.Path] = s.Alias
		}
		return "", nil
	}
	return g.genStmt(stmt), nil
}

func (g *Generator) genIf(s *ast.IfStmt) string {
	var b strings.Builder
	fmt.Fprintf(&b, "if (%s).IsTruthy() {\n", g.genExpr(s.Condition))
	for _, stmt := range s.Body {
		b.WriteString("\t")
		b.WriteString(g.genStmt(stmt))
		b.WriteString("\n")
	}
	if len(s.ElseBody) > 0 {
		b.WriteString("} else {\n")
		for _, stmt := range s.ElseBody {
			b.WriteString("\t")
			b.WriteString(g.genStmt(stmt))
			b.WriteString("\n")
		}
	}
	b.WriteString("}")
	return b.String()
}

func (g *Generator) genRange(s *ast.RangeStmt) string {
	endType := g.getExprType(s.End)
	if endType != nil {
		endType = types.Unwrap(endType)
	}
	if _, ok := endType.(types.ListType); ok && s.Start == nil && !s.Inclusive {
		return g.genListRange(s)
	}
	var b strings.Builder
	cmp := "<"
	if s.Inclusive {
		cmp = "<="
	}
	startExpr := "meow.NewInt(0)"
	if s.Start != nil {
		startExpr = g.genExpr(s.Start)
	}
	fmt.Fprintf(&b, "for __i := meow.AsInt(%s); __i %s meow.AsInt(%s); __i++ {\n",
		startExpr, cmp, g.genExpr(s.End))
	fmt.Fprintf(&b, "\tvar %s meow.Value = meow.NewInt(__i)\n", s.Var)
	// The body need not mention the loop variable, and Go rejects an unused
	// one. The list form already guards this the same way.
	fmt.Fprintf(&b, "\t_ = %s\n", s.Var)
	defer g.enterBoxedScope(s.Var)()
	for _, stmt := range s.Body {
		b.WriteString("\t")
		b.WriteString(g.genStmt(stmt))
		b.WriteString("\n")
	}
	b.WriteString("}")
	return b.String()
}

func (g *Generator) genListRange(s *ast.RangeStmt) string {
	return g.genListRangeWith(s, g.genExpr(s.End), g.genStmt)
}

// genTypedListRange emits a list-form purr inside a fully-typed function.
//
// It exists because the body must be generated by the typed statement
// generator. The untyped one ends each statement with the Furball
// short-circuit `return __f`, and a typed function returns a native Go type —
// int64, string — so that return does not compile. Typed bodies report failure
// by panicking instead, which gag's deferred recover turns back into a Furball
// at the boundary.
func (g *Generator) genTypedListRange(s *ast.RangeStmt) string {
	return g.genListRangeWith(s, g.boxValue(s.End), g.genTypedStmt)
}

// genListRangeWith emits the loop shared by both range forms. list is the
// already-generated expression yielding the litter, and genStmt generates the
// body in whichever mode the enclosing function is being written in.
func (g *Generator) genListRangeWith(
	s *ast.RangeStmt,
	list string,
	genStmt func(ast.Stmt) string,
) string {
	var b strings.Builder
	fmt.Fprintf(&b, "for __idx, __elem := range meow.AsList(%s).Items {\n", list)
	if s.IndexVar != "" {
		fmt.Fprintf(&b, "\tvar %s meow.Value = meow.NewInt(int64(__idx))\n", s.IndexVar)
		fmt.Fprintf(&b, "\t_ = %s\n", s.IndexVar)
	} else {
		b.WriteString("\t_ = __idx\n")
	}
	fmt.Fprintf(&b, "\tvar %s meow.Value = __elem\n", s.Var)
	fmt.Fprintf(&b, "\t_ = %s\n", s.Var)
	// Both variables hold meow.Value, whatever the enclosing function does.
	defer g.enterBoxedScope(s.Var, s.IndexVar)()
	for _, stmt := range s.Body {
		b.WriteString("\t")
		b.WriteString(genStmt(stmt))
		b.WriteString("\n")
	}
	b.WriteString("}")
	return b.String()
}

func (g *Generator) genExpr(expr ast.Expr) string {
	if entries, ok := g.mutations[expr]; ok && len(entries) > 0 {
		return g.genMutatedExpr(expr, entries)
	}
	switch e := expr.(type) {
	case *ast.IntLit:
		return fmt.Sprintf("meow.NewInt(%d)", e.Value)
	case *ast.FloatLit:
		return fmt.Sprintf("meow.NewFloat(%g)", e.Value)
	case *ast.StringLit:
		return fmt.Sprintf("meow.NewString(%q)", e.Value)
	case *ast.BoolLit:
		if e.Value {
			return "meow.NewBool(true)"
		}
		return "meow.NewBool(false)"
	case *ast.NilLit:
		return "meow.NewNil()"
	case *ast.Ident:
		return g.genIdent(e)
	case *ast.UnaryExpr:
		return g.genUnary(e)
	case *ast.BinaryExpr:
		return g.genBinary(e)
	case *ast.CallExpr:
		return g.genCall(e)
	case *ast.LambdaExpr:
		return g.genLambda(e)
	case *ast.ListLit:
		return g.genList(e)
	case *ast.IndexExpr:
		return g.genIndex(e)
	case *ast.PipeExpr:
		return g.genPipe(e)
	case *ast.CatchExpr:
		return g.genCatch(e)
	case *ast.MapLit:
		return g.genMap(e)
	case *ast.MatchExpr:
		return g.genMatch(e)
	case *ast.SelfExpr:
		return "self"
	case *ast.MemberExpr:
		// Check if this is a self.field access in a learn method
		if _, isSelf := e.Object.(*ast.SelfExpr); isSelf {
			if e.Member == "value" {
				return "self.(*meow.Kitty).GetField(\"value\")"
			}
			return fmt.Sprintf("self.(*meow.Kitty).GetField(%q)", e.Member)
		}
		obj, ok := e.Object.(*ast.Ident)
		if ok {
			if realPkg, ok := g.resolveImportName(obj.Name); ok {
				return fmt.Sprintf("meow_%s.%s", realPkg, capitalizeFirst(e.Member))
			}
			return fmt.Sprintf("%s.(*meow.Kitty).GetField(%q)", obj.Name, e.Member)
		}
		return fmt.Sprintf("(%s).(*meow.Kitty).GetField(%q)", g.genExpr(e.Object), e.Member)
	default:
		return fmt.Sprintf("/* unsupported expr: %T */", expr)
	}
}

// genIdent emits an identifier in untyped (meow.Value) mode. Inside a
// fully-typed function body the variable may be held in a native Go type, in
// which case it is boxed so it can be used where a meow.Value is expected —
// passing a `s string` parameter to a runtime helper, for instance.
func (g *Generator) genIdent(e *ast.Ident) string {
	if t, ok := g.nativeVars[e.Name]; ok {
		return g.boxNative(e.Name, t)
	}
	return e.Name
}

func (g *Generator) genUnary(e *ast.UnaryExpr) string {
	switch e.Op {
	case token.MINUS:
		return fmt.Sprintf("meow.Negate(%s)", g.genExpr(e.Right))
	case token.NOT:
		return fmt.Sprintf("meow.Not(%s)", g.genExpr(e.Right))
	default:
		return fmt.Sprintf("/* unsupported unary: %v */", e.Op)
	}
}

func (g *Generator) genBinary(e *ast.BinaryExpr) string {
	l := g.genExpr(e.Left)
	r := g.genExpr(e.Right)
	switch e.Op {
	case token.PLUS:
		return fmt.Sprintf("meow.Add(%s, %s)", l, r)
	case token.MINUS:
		return fmt.Sprintf("meow.Sub(%s, %s)", l, r)
	case token.STAR:
		return fmt.Sprintf("meow.Mul(%s, %s)", l, r)
	case token.SLASH:
		return fmt.Sprintf("meow.Div(%s, %s)", l, r)
	case token.PERCENT:
		return fmt.Sprintf("meow.Mod(%s, %s)", l, r)
	case token.EQ:
		return fmt.Sprintf("meow.Equal(%s, %s)", l, r)
	case token.NEQ:
		return fmt.Sprintf("meow.NotEqual(%s, %s)", l, r)
	case token.LT:
		return fmt.Sprintf("meow.LessThan(%s, %s)", l, r)
	case token.GT:
		return fmt.Sprintf("meow.GreaterThan(%s, %s)", l, r)
	case token.LTE:
		return fmt.Sprintf("meow.LessEqual(%s, %s)", l, r)
	case token.GTE:
		return fmt.Sprintf("meow.GreaterEqual(%s, %s)", l, r)
	case token.AND:
		return fmt.Sprintf("meow.And(%s, %s)", l, r)
	case token.OR:
		return fmt.Sprintf("meow.Or(%s, %s)", l, r)
	default:
		return fmt.Sprintf("/* unsupported op: %v */", e.Op)
	}
}

func (g *Generator) genCall(e *ast.CallExpr) string {
	if member, ok := e.Fn.(*ast.MemberExpr); ok {
		return g.genMemberCall(member, e.Args)
	}
	ident, isIdent := e.Fn.(*ast.Ident)
	args := make([]string, len(e.Args))
	for i, a := range e.Args {
		args[i] = g.genExpr(a)
	}
	argStr := strings.Join(args, ", ")

	if isIdent {
		switch ident.Name {
		case "nya":
			return fmt.Sprintf("meow.Nya(%s)", argStr)
		case "hiss":
			return fmt.Sprintf("meow.Hiss(%s)", argStr)
		case "lick":
			return fmt.Sprintf("meow.Lick(%s)", argStr)
		case "picky":
			return fmt.Sprintf("meow.Picky(%s)", argStr)
		case "curl":
			return fmt.Sprintf("meow.Curl(%s)", argStr)
		case "len":
			return fmt.Sprintf("meow.Len(%s)", argStr)
		case "head":
			return fmt.Sprintf("meow.Head(%s)", argStr)
		case "tail":
			return fmt.Sprintf("meow.Tail(%s)", argStr)
		case "append":
			return fmt.Sprintf("meow.Append(%s)", argStr)
		case "to_int":
			return fmt.Sprintf("meow.ToInt(%s)", argStr)
		case "to_float":
			return fmt.Sprintf("meow.ToFloat(%s)", argStr)
		case "to_string":
			return fmt.Sprintf("meow.ToString(%s)", argStr)
		case "to_bytes":
			return fmt.Sprintf("meow.ToBytes(%s)", argStr)
		case "to_runes":
			return fmt.Sprintf("meow.ToRunes(%s)", argStr)
		case "whiff":
			return fmt.Sprintf("meow.Whiff(%s)", argStr)
		case "track":
			return fmt.Sprintf("meow.Track(%s)", argStr)
		case "shred":
			return fmt.Sprintf("meow.Shred(%s)", argStr)
		case "tangle":
			return fmt.Sprintf("meow.Tangle(%s)", argStr)
		case "nibble":
			return fmt.Sprintf("meow.Nibble(%s)", argStr)
		case "gag":
			return fmt.Sprintf("meow.Gag(%s)", argStr)
		case "is_furball":
			return fmt.Sprintf("meow.IsFurball(%s)", argStr)
		case "judge":
			g.ensureImport("testing")
			return fmt.Sprintf("meow_testing.Judge(%s)", argStr)
		case "expect":
			g.ensureImport("testing")
			return fmt.Sprintf("meow_testing.Expect(%s)", argStr)
		case "refuse":
			g.ensureImport("testing")
			return fmt.Sprintf("meow_testing.Refuse(%s)", argStr)
		case "seed":
			return "meow.NewNil()"
		default:
			if ks, ok := g.kittyDefs[ident.Name]; ok {
				fieldNames := make([]string, len(ks.Fields))
				for i, f := range ks.Fields {
					fieldNames[i] = fmt.Sprintf("%q", f.Name)
				}
				return fmt.Sprintf("meow.NewKitty(%q, []string{%s}, %s)",
					ident.Name, strings.Join(fieldNames, ", "), argStr)
			}
			if _, ok := g.collarDefs[ident.Name]; ok {
				return fmt.Sprintf("meow.NewKitty(%q, []string{\"value\"}, %s)",
					ident.Name, argStr)
			}
			if g.typeInfo != nil {
				if ft, ok := g.typeInfo.FuncTypes[ident.Name]; ok {
					if len(e.Args) < len(ft.Params) {
						return g.genPartialCall(ident.Name, ft, e.Args)
					}
					if len(e.Args) == len(ft.Params) && isFullyTypedFuncType(ft) {
						nativeArgs := make([]string, len(e.Args))
						for i, a := range e.Args {
							if isLiteralExpr(a) {
								nativeArgs[i] = g.genTypedExpr(a)
							} else {
								nativeArgs[i] = unboxToNative(args[i], ft.Params[i])
							}
						}
						call := fmt.Sprintf("%s(%s)", ident.Name, strings.Join(nativeArgs, ", "))
						return boxNativeCall(call, ft.Return)
					}
				} else {
					// Not in FuncTypes → must be a variable holding a function
					// value (e.g. partial application result or lambda).
					// The checker populates FuncTypes for all `meow` function
					// declarations, so any identifier absent from FuncTypes is
					// a runtime value that requires meow.Call for dispatch.
					if argStr != "" {
						return fmt.Sprintf("meow.Call(%s, %s)", ident.Name, argStr)
					}
					return fmt.Sprintf("meow.Call(%s)", ident.Name)
				}
			}
			return fmt.Sprintf("%s(%s)", ident.Name, argStr)
		}
	}
	return fmt.Sprintf("meow.Call(%s, %s)", g.genExpr(e.Fn), argStr)
}

func (g *Generator) genMemberCall(member *ast.MemberExpr, rawArgs []ast.Expr) string {
	args := make([]string, len(rawArgs))
	for i, a := range rawArgs {
		args[i] = g.genExpr(a)
	}
	argStr := strings.Join(args, ", ")

	// Check if this is a method call on an object that has learn impls
	if g.typeInfo != nil {
		objCode := g.genExpr(member.Object)
		typeName := g.resolveTypeName(member.Object)
		if typeName != "" {
			if methods, ok := g.typeInfo.LearnImpls[typeName]; ok {
				if _, hasMethod := methods[member.Member]; hasMethod {
					if argStr != "" {
						return fmt.Sprintf("meow.DispatchMethod(%s, %q, %s)", objCode, member.Member, argStr)
					}
					return fmt.Sprintf("meow.DispatchMethod(%s, %q)", objCode, member.Member)
				}
			}
		}
	}

	obj, ok := member.Object.(*ast.Ident)
	if !ok {
		if argStr == "" {
			return fmt.Sprintf("meow.Call((%s).(*meow.Kitty).GetField(%q))",
				g.genExpr(member.Object), member.Member)
		}
		return fmt.Sprintf("meow.Call((%s).(*meow.Kitty).GetField(%q), %s)",
			g.genExpr(member.Object), member.Member, argStr)
	}
	if realPkg, ok := g.resolveImportName(obj.Name); ok {
		return fmt.Sprintf("meow_%s.%s(%s)", realPkg, capitalizeFirst(member.Member), argStr)
	}
	if argStr == "" {
		return fmt.Sprintf("meow.Call(%s.(*meow.Kitty).GetField(%q))",
			obj.Name, member.Member)
	}
	return fmt.Sprintf("meow.Call(%s.(*meow.Kitty).GetField(%q), %s)",
		obj.Name, member.Member, argStr)
}

// resolveTypeName tries to determine the type name for a given expression.
func (g *Generator) resolveTypeName(expr ast.Expr) string {
	if g.typeInfo == nil {
		return ""
	}
	t := g.typeInfo.ExprTypes[expr]
	if t == nil {
		return ""
	}
	t = types.Unwrap(t)
	switch tt := t.(type) {
	case types.KittyType:
		return tt.Name
	case types.CollarType:
		return tt.Name
	}
	return ""
}

func (g *Generator) genPartialCall(fnName string, ft types.FuncType, suppliedArgs []ast.Expr) string {
	remaining := len(ft.Params) - len(suppliedArgs)

	if isFullyTypedFuncType(ft) {
		// Typed function: generate native-typed partial application
		var captureLines []string
		for i, a := range suppliedArgs {
			captureLines = append(captureLines,
				fmt.Sprintf("__c%d := %s", i, g.genTypedExpr(a)))
		}
		var callArgs []string
		for i := range suppliedArgs {
			callArgs = append(callArgs, fmt.Sprintf("__c%d", i))
		}
		for i := range remaining {
			callArgs = append(callArgs,
				fmt.Sprintf("%s", unboxToNative(fmt.Sprintf("args[%d]", i), ft.Params[len(suppliedArgs)+i])))
		}
		call := fmt.Sprintf("%s(%s)", fnName, strings.Join(callArgs, ", "))
		boxed := boxNativeCall(call, ft.Return)

		capture := strings.Join(captureLines, "\n\t")
		if capture != "" {
			capture += "\n\t"
		}
		return fmt.Sprintf("func() meow.Value {\n"+
			"\t%sreturn meow.NewFuncWithArity(%q, %d, func(args ...meow.Value) meow.Value {\n"+
			"\t\treturn %s\n"+
			"\t})\n"+
			"}()", capture, fnName, remaining, boxed)
	}

	// Untyped function: box arguments to ensure meow.Value
	var captureLines []string
	for i, a := range suppliedArgs {
		captureLines = append(captureLines,
			fmt.Sprintf("__c%d := %s", i, g.boxValue(a)))
	}
	var callArgs []string
	for i := range suppliedArgs {
		callArgs = append(callArgs, fmt.Sprintf("__c%d", i))
	}
	for i := range remaining {
		callArgs = append(callArgs, fmt.Sprintf("args[%d]", i))
	}
	call := fmt.Sprintf("%s(%s)", fnName, strings.Join(callArgs, ", "))

	capture := strings.Join(captureLines, "\n\t")
	if capture != "" {
		capture += "\n\t"
	}
	return fmt.Sprintf("func() meow.Value {\n"+
		"\t%sreturn meow.NewFuncWithArity(%q, %d, func(args ...meow.Value) meow.Value {\n"+
		"\t\treturn %s\n"+
		"\t})\n"+
		"}()", capture, fnName, remaining, call)
}

func (g *Generator) genLambda(e *ast.LambdaExpr) string {
	names := make([]string, len(e.Params))
	for i, p := range e.Params {
		names[i] = p.Name
	}
	defer g.enterBoxedScope(names...)()
	return fmt.Sprintf("meow.NewFuncWithArity(\"lambda\", %d, func(args ...meow.Value) meow.Value {\n"+
		"\t%s\n"+
		"%s"+
		"})", len(e.Params), g.genLambdaParamBindings(e.Params), g.genLambdaBody(e))
}

// genLambdaBody emits the closure body. The lambda closure has the same shape
// as an untyped function body — func(...) meow.Value — so a block body reuses
// the ordinary statement generator.
func (g *Generator) genLambdaBody(e *ast.LambdaExpr) string {
	if e.Block == nil {
		return fmt.Sprintf("\treturn %s\n", g.genExpr(e.Body))
	}
	var b strings.Builder
	for i, stmt := range e.Block {
		b.WriteString("\t")
		// A trailing expression statement is the lambda's result, mirroring the
		// single-expression form; genStmt would otherwise discard its value.
		if exprStmt, ok := stmt.(*ast.ExprStmt); ok && i == len(e.Block)-1 {
			fmt.Fprintf(&b, "return %s\n", g.genExpr(exprStmt.Expr))
			return b.String()
		}
		b.WriteString(g.genStmt(stmt))
		b.WriteString("\n")
	}
	if !g.blockAlwaysReturns(e.Block) {
		b.WriteString("\treturn meow.NewNil()\n")
	}
	return b.String()
}

func (g *Generator) genLambdaParamBindings(params []ast.Param) string {
	var lines []string
	for i, p := range params {
		lines = append(lines, fmt.Sprintf("%s := args[%d]", p.Name, i))
		lines = append(lines, fmt.Sprintf("_ = %s", p.Name))
	}
	return strings.Join(lines, "\n\t")
}

func (g *Generator) genList(e *ast.ListLit) string {
	items := make([]string, len(e.Items))
	for i, item := range e.Items {
		items[i] = g.genExpr(item)
	}
	return fmt.Sprintf("meow.NewList(%s)", strings.Join(items, ", "))
}

func (g *Generator) genMap(e *ast.MapLit) string {
	if len(e.Keys) == 0 {
		return "meow.NewMap(map[string]meow.Value{})"
	}
	entries := make([]string, len(e.Keys))
	for i := range e.Keys {
		key, ok := e.Keys[i].(*ast.StringLit)
		if !ok {
			entries[i] = fmt.Sprintf("/* unsupported map key: %T */", e.Keys[i])
			continue
		}
		entries[i] = fmt.Sprintf("%q: %s", key.Value, g.genExpr(e.Vals[i]))
	}
	return fmt.Sprintf("meow.NewMap(map[string]meow.Value{%s})", strings.Join(entries, ", "))
}

func (g *Generator) genIndex(e *ast.IndexExpr) string {
	return fmt.Sprintf("meow.Index(%s, %s)", g.genExpr(e.Left), g.genExpr(e.Index))
}

func (g *Generator) genPipe(e *ast.PipeExpr) string {
	var fn ast.Expr
	var args []ast.Expr

	if call, ok := e.Right.(*ast.CallExpr); ok {
		fn = call.Fn
		args = make([]ast.Expr, 0, len(call.Args)+1)
		args = append(args, e.Left)
		args = append(args, call.Args...)
	} else {
		fn = e.Right
		args = []ast.Expr{e.Left}
	}

	return g.genCall(&ast.CallExpr{Token: e.Token, Fn: fn, Args: args})
}

func (g *Generator) genCatch(e *ast.CatchExpr) string {
	// Left side is wrapped in a thunk so GagOr can both (a) propagate Furball
	// values from the untyped path, and (b) recover from panics raised by
	// typed function bodies — without this, evaluating a typed expression
	// inline would bypass ~>'s recovery.
	left := g.genExpr(e.Left)
	right := g.genExpr(e.Right)
	return fmt.Sprintf(
		"meow.GagOr(meow.NewFunc(\"~>\", func(args ...meow.Value) meow.Value {\n"+
			"\treturn %s\n"+
			"}), %s)", left, right)
}

func (g *Generator) genMatch(e *ast.MatchExpr) string {
	var b strings.Builder
	subject := g.genExpr(e.Subject)
	b.WriteString(fmt.Sprintf("func() meow.Value {\n\t__subject := %s\n", subject))
	hasCondArm := false
	for i, arm := range e.Arms {
		if _, ok := arm.Pattern.(*ast.WildcardPattern); ok {
			if hasCondArm {
				b.WriteString("\t} else {\n")
				b.WriteString(fmt.Sprintf("\t\treturn %s\n", g.genExpr(arm.Body)))
				b.WriteString("\t}\n")
			} else {
				b.WriteString(fmt.Sprintf("\treturn %s\n", g.genExpr(arm.Body)))
			}
			b.WriteString("\treturn meow.NewNil()\n}()")
			return b.String()
		}
		keyword := "if"
		if i > 0 {
			keyword = "} else if"
		}
		hasCondArm = true
		b.WriteString(fmt.Sprintf("\t%s %s {\n", keyword, g.genPatternCond("__subject", arm.Pattern)))
		b.WriteString(fmt.Sprintf("\t\treturn %s\n", g.genExpr(arm.Body)))
	}
	if hasCondArm {
		b.WriteString("\t}\n")
	}
	b.WriteString("\treturn meow.NewNil()\n}()")
	return b.String()
}

func (g *Generator) genMutatedExpr(original ast.Expr, entries []mutation.MutationEntry) string {
	// Remove this expression's entries to avoid self-recursion
	delete(g.mutations, original)

	var b strings.Builder
	b.WriteString("func() meow.Value {\n")

	// Mutation branches: generate without child mutations
	// (each branch represents only its specific mutation, not combinations)
	saved := g.mutations
	g.mutations = nil
	for _, entry := range entries {
		fmt.Fprintf(&b, "\t\tif __mutant == %d { return %s }\n", entry.ID, g.genExpr(entry.Expr))
	}
	g.mutations = saved

	// Default branch: child mutations remain active
	fmt.Fprintf(&b, "\t\treturn %s\n", g.genExpr(original))
	b.WriteString("\t}()")

	// Restore this entry
	g.mutations[original] = entries
	return b.String()
}

func (g *Generator) genLearnMethod(typeName string, fn *ast.FuncStmt) string {
	names := make([]string, 0, len(fn.Params)+1)
	names = append(names, "self")
	for _, p := range fn.Params {
		names = append(names, p.Name)
	}
	defer g.enterBoxedScope(names...)()

	var b strings.Builder
	methodFuncName := fmt.Sprintf("meow_method_%s_%s", typeName, fn.Name)

	fmt.Fprintf(&b, "func %s(args ...meow.Value) meow.Value {\n", methodFuncName)
	// Arity guard: self + params
	fmt.Fprintf(&b, "\tif len(args) < %d {\n", 1+len(fn.Params))
	b.WriteString("\t\treturn meow.NewNil()\n")
	b.WriteString("\t}\n")
	// Extract self from first argument
	b.WriteString("\tself := args[0]\n")
	b.WriteString("\t_ = self\n")
	// Extract additional params
	for i, p := range fn.Params {
		fmt.Fprintf(&b, "\t%s := args[%d]\n", p.Name, i+1)
		fmt.Fprintf(&b, "\t_ = %s\n", p.Name)
	}

	prevInLearn := g.inLearnMethod
	g.inLearnMethod = true
	for _, stmt := range fn.Body {
		b.WriteString("\t")
		b.WriteString(g.genStmt(stmt))
		b.WriteString("\n")
	}
	g.inLearnMethod = prevInLearn

	if !g.blockAlwaysReturns(fn.Body) {
		b.WriteString("\treturn meow.NewNil()\n")
	}
	b.WriteString("}")
	return b.String()
}

func (g *Generator) genLearnInit() string {
	if len(g.learnDefs) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("func init() {\n")
	for _, ls := range g.learnDefs {
		for i := range ls.Methods {
			m := &ls.Methods[i]
			fmt.Fprintf(&b, "\tmeow.RegisterMethod(%q, %q, meow_method_%s_%s)\n",
				ls.TypeName, m.Name, ls.TypeName, m.Name)
		}
	}
	b.WriteString("}\n")
	return b.String()
}

func (g *Generator) genPatternCond(subject string, pattern ast.Pattern) string {
	switch p := pattern.(type) {
	case *ast.LiteralPattern:
		return fmt.Sprintf("meow.MatchValue(%s, %s)", subject, g.genExpr(p.Value))
	case *ast.RangePattern:
		low := p.Low.(*ast.IntLit).Value
		high := p.High.(*ast.IntLit).Value
		return fmt.Sprintf("meow.MatchRange(%s, %d, %d)", subject, low, high)
	case *ast.WildcardPattern:
		return "true"
	default:
		return "true"
	}
}
