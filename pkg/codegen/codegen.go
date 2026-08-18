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
	// globalVars names the bindings written at the top level of the program, in
	// source order. They become package-level Go variables so that functions can
	// read them; see genStmtInner's VarStmt case.
	globalVars []string
	// usedPackages records the `nab` packages a selector was actually emitted
	// for, so an import nothing calls can be left out. It is recorded where the
	// selector is written rather than read back out of the rendered source,
	// which would also match a package name appearing inside a string literal.
	usedPackages map[string]bool
	// hoistedVar is the top-level binding currently being generated, if any.
	// Holding the node itself rather than a depth counter keeps a `nyan` nested
	// inside a top-level `sniff` or `purr` block a local, which is what it is.
	hoistedVar *ast.VarStmt
	// nestedFuncs names the `meow`s declared inside the body being generated.
	// They are emitted as closures held in meow.Value variables rather than as
	// package-level Go functions, so that they can read the enclosing scope the
	// way the playground interpreter lets them. A call therefore dispatches
	// through meow.Call, and a name here shadows a top-level function of the
	// same name, as it does for the checker.
	nestedFuncs map[string]bool
}

// enterNestedScope starts tracking nested function names, returning a function
// that restores the previous set.
func (g *Generator) enterNestedScope() func() {
	prev := g.nestedFuncs
	g.nestedFuncs = make(map[string]bool, len(prev))
	maps.Copy(g.nestedFuncs, prev)
	return func() { g.nestedFuncs = prev }
}

// isNestedFunc reports whether name is a `meow` declared inside the body being
// generated, and so held as a value rather than emitted as a Go function.
func (g *Generator) isNestedFunc(name string) bool {
	return g.nestedFuncs[name]
}

// hoistNestedFuncs declares every `meow` written in body up front, so that one
// can call another regardless of which comes first — which is what the
// interpreter does, since it registers them all before running anything.
//
// The declaration is separate from the assignment because a closure that refers
// to itself, or to a sibling declared later, needs the name to exist before the
// function value does.
func (g *Generator) hoistNestedFuncs(body []ast.Stmt) string {
	var b strings.Builder
	for _, stmt := range body {
		fn, ok := stmt.(*ast.FuncStmt)
		if !ok {
			continue
		}
		g.nestedFuncs[fn.Name] = true
		// It starts as the Furball a caller would deserve, so that calling one
		// before its declaration has run says so — the interpreter answers
		// "undefined function" there, where a nil meow.Value would be a bare Go
		// nil dereference. _ = name because Go rejects a variable that is only
		// ever assigned, and a nested function nothing calls is still valid.
		fmt.Fprintf(&b, "\tvar %s meow.Value = meow.NewFurball(\"Hiss! undefined function %s, nya~\")\n\t_ = %s\n",
			fn.Name, fn.Name, fn.Name)
	}
	return b.String()
}

// genNestedFunc emits a `meow` written inside another as a closure bound to the
// name hoisted for it. The body is generated boxed, like a lambda's, so it
// reads an enclosing typed function's parameters through the same boxing an
// ordinary runtime call gets.
func (g *Generator) genNestedFunc(fn *ast.FuncStmt) string {
	names := make([]string, len(fn.Params))
	for i, p := range fn.Params {
		names[i] = p.Name
	}
	defer g.enterBoxedScope(names...)()
	defer g.enterNestedScope()()

	var b strings.Builder
	fmt.Fprintf(&b, "%s = meow.NewFuncWithArity(%q, %d, func(args ...meow.Value) meow.Value {\n\t%s\n",
		fn.Name, fn.Name, len(fn.Params), g.genLambdaParamBindings(fn.Params))
	b.WriteString(g.callerPrologue())
	b.WriteString(g.hoistNestedFuncs(fn.Body))
	for _, stmt := range fn.Body {
		b.WriteString("\t")
		b.WriteString(g.genStmt(stmt))
		b.WriteString("\n")
	}
	if !g.blockAlwaysReturns(fn.Body) {
		b.WriteString("\treturn meow.NewNil()\n")
	}
	b.WriteString("})")
	return b.String()
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
	"json":    "github.com/135yshr/meow/runtime/json",
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
			code, err := g.genTopLevelStmt(stmt)
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

// genTopLevelStmt generates one statement written at the top level of the
// program, hoisting a binding there to package scope.
//
// Top-level bindings used to be generated as locals of the wrapper that runs
// the program, so a function referring to one failed to build with Go's
// "undefined: name" — a constant declared once and read from a function, the
// most ordinary shape there is, did not work. The playground interpreter puts
// them in the global environment and has always resolved them, so this is what
// makes the two backends agree.
func (g *Generator) genTopLevelStmt(stmt ast.Stmt) (string, error) {
	if v, ok := stmt.(*ast.VarStmt); ok && !isGeneratedName(v.Name) {
		g.globalVars = append(g.globalVars, v.Name)
		g.hoistedVar = v
		defer func() { g.hoistedVar = nil }()
	}
	return g.genStmtOrError(stmt)
}

// isGeneratedName reports whether a name is one the generator emits itself.
//
// Hoisting such a binding to package scope would redeclare that identifier —
// `nyan main = 5` against the program's own entry point — so it stays a local
// of the wrapper, exactly where it was before hoisting existed. It is then
// invisible to functions, but a binding named after the entry point always was.
func isGeneratedName(name string) bool {
	switch name {
	case "main", "__meow_main", "meow", "init", "__mutant":
		return true
	}
	return strings.HasPrefix(name, "meow_") || strings.HasPrefix(name, "__")
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
			code, err := g.genTopLevelStmt(stmt)
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
	for _, name := range g.usedImports(g.testWrapperImports()...) {
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

	b.WriteString(g.genGlobalDecls())

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

// usedImports names the `nab` packages the generated code actually calls, in a
// stable order.
//
// Go rejects an import nothing uses, so `nab "env"` left over from an edit
// failed the build with a Go error naming a generated alias — the abstraction
// leaking again, over something Meow itself has no objection to. Importing only
// what is called keeps that a non-event.
// testWrapperImports names what the test wrapper always calls.
func (g *Generator) testWrapperImports() []string {
	names := []string{"testing"}
	if g.coverEnabled {
		names = append(names, "coverage")
	}
	return names
}

// alsoUsed names packages the wrapper written after the imports calls, which no
// scan of the generated bodies would find.
func (g *Generator) usedImports(alsoUsed ...string) []string {
	if len(g.imports) == 0 {
		return nil
	}
	keep := make(map[string]bool, len(alsoUsed))
	for _, name := range alsoUsed {
		keep[name] = true
	}
	names := make([]string, 0, len(g.imports))
	for name := range g.imports {
		if keep[name] || g.usedPackages[name] {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// markPackageUsed records that a selector into a `nab` package was emitted.
func (g *Generator) markPackageUsed(name string) {
	if g.usedPackages == nil {
		g.usedPackages = make(map[string]bool)
	}
	g.usedPackages[name] = true
}

// genGlobalDecls declares the program's top-level bindings at package scope.
// They are assigned, in source order, where they were written; this only puts
// the names somewhere a function body can see them.
func (g *Generator) genGlobalDecls() string {
	if len(g.globalVars) == 0 {
		return ""
	}
	var b strings.Builder
	for _, name := range g.globalVars {
		// Reaching one from a function that runs before the binding does is a
		// real mistake, and a bare declaration would make it a nil dereference
		// with nothing to read. Starting it as the same Furball the playground
		// interpreter raises keeps the two backends saying the same thing.
		fmt.Fprintf(&b,
			"var %s meow.Value = meow.NewFurball(\"Hiss! undefined variable %s, nya~\")\n",
			name, name)
	}
	b.WriteString("\n")
	return b.String()
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
	for _, name := range g.usedImports() {
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

	b.WriteString(g.genGlobalDecls())

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
	defer g.enterNestedScope()()
	fmt.Fprintf(&b, "func %s(%s) meow.Value {\n", fn.Name, strings.Join(params, ", "))
	b.WriteString(g.callerPrologue())
	b.WriteString(g.hoistNestedFuncs(fn.Body))
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
	defer g.enterNestedScope()()
	var b strings.Builder
	params := make([]string, len(fn.Params))
	for i, p := range fn.Params {
		params[i] = p.Name + " " + goTypeString(ft.Params[i])
		g.bindNativeVar(p.Name, ft.Params[i])
	}
	fmt.Fprintf(&b, "func %s(%s) %s {\n", fn.Name, strings.Join(params, ", "), goTypeString(ft.Return))
	b.WriteString(g.callerPrologue())
	b.WriteString(g.hoistNestedFuncs(fn.Body))
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
	return g.located(stmt, g.genTypedStmtInner(stmt))
}

func (g *Generator) genTypedStmtInner(stmt ast.Stmt) string {
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
	case *ast.WhileStmt:
		return g.genTypedWhile(s)
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
		return "meow.Here(__caller)\nreturn"
	}
	t := g.getExprType(s.Value)
	if t != nil && !types.IsAny(t) {
		return fmt.Sprintf("return meow.Returning(__caller, %s)", g.genTypedExpr(s.Value))
	}
	// Expression is AnyType (e.g. match expression) but function has a concrete return type.
	// Generate the expression with typed boxing, then unbox the meow.Value result.
	exprCode := g.genTypedExpr(s.Value)
	if g.currentReturnType != nil && !types.IsAny(g.currentReturnType) {
		return fmt.Sprintf("return meow.Returning(__caller, %s)", unboxToNative(exprCode, g.currentReturnType))
	}
	return fmt.Sprintf("return meow.Returning(__caller, %s)", exprCode)
}

func (g *Generator) genTypedExprStmt(s *ast.ExprStmt) string {
	if call, ok := s.Expr.(*ast.CallExpr); ok {
		code := g.genTypedCall(call)
		// A statement's value is discarded, so a call that answers with a
		// Furball would fail in silence — the function would carry on and
		// report success. Where the result is a native Go type there is nothing
		// to check: the unboxing, or the callee itself, already raised it.
		//
		// hiss is left alone because it raises rather than answering: the typed
		// path emits it as a Go panic, which is a statement and has no value to
		// pass through anything.
		if t := g.getExprType(call); (t == nil || !isNativeType(t)) && !isCallTo(call, "hiss") {
			return fmt.Sprintf("meow.Propagate(%s)", code)
		}
		return code
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
	b.WriteString(g.genBlockStmts(s.Body, g.genTypedStmt))
	if len(s.ElseBody) > 0 {
		b.WriteString("} else {\n")
		b.WriteString(g.genBlockStmts(s.ElseBody, g.genTypedStmt))
	}
	b.WriteString("}")
	return b.String()
}

func (g *Generator) genTypedRange(s *ast.RangeStmt) string {
	endType := g.getExprType(s.End)
	if endType != nil {
		endType = types.Unwrap(endType)
	}
	if isElementRange(s, endType) {
		return g.genTypedElementRange(s)
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
	b.WriteString(g.genBlockStmts(s.Body, g.genTypedStmt))
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
// the checker typed as nil.
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

// operandKind reports an operand's type with aliases resolved, and whether it
// is one the typed path can name a Go type for.
func (g *Generator) operandKind(expr ast.Expr) (types.Type, bool) {
	t := g.getExprType(expr)
	if t != nil {
		t = types.Unwrap(t)
	}
	return t, t != nil && !types.IsAny(t) && isNativeType(t)
}

// genTypedOperands generates both sides of a binary expression as native Go
// values.
//
// A `nab` package call reports `any`, having no declared type, so asking the
// typed generator for it yields a meow.Value — and `clock.nanos() > deadline`
// reached Go as `meow.Value > int64`, which does not build. Where one side is
// native the other is read as that same type; where neither is, the operator is
// applied boxed and the result unboxed once.
func (g *Generator) genTypedOperands(e *ast.BinaryExpr) (left, right string, native bool) {
	// `&&` and `||` are not bool operators here: they weigh their operands by
	// truthiness and yield one of them, so `m["x"] && flag` with a string on the
	// left answers flag. Reading such an operand as the other side's Go type
	// would demand a bool of it and fail on anything else.
	if e.Op == token.AND || e.Op == token.OR {
		if _, lok := g.operandKind(e.Left); !lok {
			return "", "", false
		}
		if _, rok := g.operandKind(e.Right); !rok {
			return "", "", false
		}
	}
	lt, lok := g.operandKind(e.Left)
	rt, rok := g.operandKind(e.Right)
	switch {
	case lok && rok:
		return g.genTypedExpr(e.Left), g.genTypedExpr(e.Right), true
	case lok && !rok:
		return g.genTypedExpr(e.Left), unboxToNative(g.genExpr(e.Right), lt), true
	case rok && !lok:
		return unboxToNative(g.genExpr(e.Left), rt), g.genTypedExpr(e.Right), true
	default:
		return "", "", false
	}
}

func (g *Generator) genTypedBinary(e *ast.BinaryExpr) string {
	// catnap never compares natively. NewNil allocates, so Go's == would weigh
	// two pointers and always answer false; and reading catnap as the other
	// side's Go type would demand a string of it. Either side being catnap goes
	// to the runtime, which knows catnap equals only itself.
	if e.Op == token.EQ || e.Op == token.NEQ {
		if g.nilOperand(e.Left) || g.nilOperand(e.Right) {
			fn := "Equal"
			if e.Op == token.NEQ {
				fn = "NotEqual"
			}
			return fmt.Sprintf("meow.%s(%s, %s).IsTruthy()", fn, g.boxValue(e.Left), g.boxValue(e.Right))
		}
	}
	left, right, native := g.genTypedOperands(e)
	if !native {
		// Neither side names a Go type. The boxed operators still know what to
		// do, so use them and unbox once for whatever the result feeds into.
		boxed := g.genBinary(e)
		if t, ok := g.operandKind(e); ok {
			return unboxToNative(boxed, t)
		}
		return boxed
	}
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
		return fmt.Sprintf("(%s == %s)", left, right)
	case token.NEQ:
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
	// A status a process can report never comes back, but a refused one is a
	// Furball — and a typed function returns a native Go type, so it has no way
	// to pass one on. ScramOrHiss raises it instead, the same bridge hiss uses
	// here; emitting a bare call would drop it and carry on as if the program
	// had asked for nothing.
	case "scram":
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = g.boxValue(a)
		}
		return fmt.Sprintf("meow.ScramOrHiss(%s)", strings.Join(args, ", "))
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
		"whiff", "track", "shred", "tangle", "nibble",
		"upper", "lower", "trim", "replace", "pad", "sort", "reverse", "round":
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
			"upper":      "Upper",
			"lower":      "Lower",
			"trim":       "Trim",
			"replace":    "Replace",
			"pad":        "Pad",
			"sort":       "Sort",
			"reverse":    "Reverse",
			"round":      "Round",
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
			"upper":      types.StringType{},
			"lower":      types.StringType{},
			"trim":       types.StringType{},
			"replace":    types.StringType{},
			"pad":        types.StringType{},
			// round is deliberately absent: it hands back the kind of number it
			// was given, so there is no one type to unbox it to.
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

	// Typed user-defined functions (only for fully native signatures). A nested
	// `meow` is held as a value and shadows a top-level one of the same name, so
	// it is asked about first — FuncTypes only ever names the top-level ones.
	if ft, ok := g.typeInfo.FuncTypes[ident.Name]; ok && !g.isNestedFunc(ident.Name) {
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

	// Anything left dispatches through meow.Call and answers with a boxed
	// value: a nested `meow`, a lambda, a partial application. Where the
	// checker knows what type that value has, a typed context wants it
	// unboxed — the same thing the builtin table above does.
	boxed := g.genCall(e)
	if t := g.getExprType(e); t != nil && !types.IsAny(t) {
		return unboxToNative(boxed, t)
	}
	return boxed
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
	code := g.located(stmt, g.genStmtInner(stmt))
	if !g.coverEnabled {
		return code
	}
	pos := stmt.Pos()
	endLine, endCol := g.estimateEndPos(stmt)
	id := len(g.coverBlocks)
	g.coverBlocks = append(g.coverBlocks, coverBlock{pos.Line, pos.Column, endLine, endCol, 1})
	g.markPackageUsed("coverage")
	return fmt.Sprintf("meow_coverage.Hit(%d)\n%s", id, code)
}

func (g *Generator) genStmtInner(stmt ast.Stmt) string {
	switch s := stmt.(type) {
	case *ast.VarStmt:
		// Bind, then short-circuit if the value is a Furball — the untyped
		// path's error-propagation point.
		//
		// A top-level binding is declared at package level instead, so that
		// functions can read it; only the assignment stays here, keeping both
		// the source order of the program and this Furball check.
		bind := "var %s meow.Value = %s"
		if s == g.hoistedVar {
			bind = "%s = %s"
		}
		code := fmt.Sprintf(
			bind+"\n\tif __f, __ok := meow.AsFurball(%s); __ok { return __f }",
			s.Name, g.genExpr(s.Value), s.Name)
		g.bindNativeVar(s.Name, nil)
		return code
	case *ast.ReturnStmt:
		if s.Value != nil {
			return fmt.Sprintf("return meow.Returning(__caller, %s)", g.genExpr(s.Value))
		}
		return "return meow.Returning(__caller, meow.NewNil())"
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
	case *ast.FuncStmt:
		// A `meow` inside another. Left as an unsupported-statement comment,
		// every call to it referred to a name that was never emitted, so a whole
		// construct the playground runs would not compile at all.
		return g.genNestedFunc(s)
	case *ast.WhileStmt:
		return g.genWhile(s)
	case *ast.BoltStmt:
		return "break"
	case *ast.SlinkStmt:
		return "continue"
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
	b.WriteString(g.genBlockStmts(s.Body, g.genStmt))
	if len(s.ElseBody) > 0 {
		b.WriteString("} else {\n")
		b.WriteString(g.genBlockStmts(s.ElseBody, g.genStmt))
	}
	b.WriteString("}")
	return b.String()
}

func (g *Generator) genRange(s *ast.RangeStmt) string {
	endType := g.getExprType(s.End)
	if endType != nil {
		endType = types.Unwrap(endType)
	}
	if isElementRange(s, endType) {
		return g.genElementRange(s)
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
	b.WriteString(g.genBlockStmts(s.Body, g.genStmt))
	b.WriteString("}")
	return b.String()
}

// genElementRange emits the elementwise `purr` — over a litter's elements, a
// basket's keys, or, when the subject turns out to be a number, the count.
//
// Which of those it is cannot always be decided here: the subject may be a
// call's result or a basket lookup, carrying no static type. Deciding it
// statically is what made `purr w (append(xs, y))` compile to a counting loop
// and fail at run time with "expected int but got List". The runtime decides
// instead, which is what the playground interpreter has always done, so the two
// backends agree.
func (g *Generator) genElementRange(s *ast.RangeStmt) string {
	return g.genElementRangeWith(s, g.genExpr(s.End), g.genStmt)
}

// genTypedElementRange is genElementRange inside a fully-typed function.
//
// It exists because the body must be generated by the typed statement
// generator. The untyped one ends each statement with the Furball
// short-circuit `return __f`, and a typed function returns a native Go type —
// int64, string — so that return does not compile. Typed bodies report failure
// by panicking instead, which gag's deferred recover turns back into a Furball
// at the boundary.
func (g *Generator) genTypedElementRange(s *ast.RangeStmt) string {
	return g.genElementRangeWith(s, g.boxValue(s.End), g.genTypedStmt)
}

// isElementRange reports whether a `purr` is written in the elementwise form
// and so has to walk its subject rather than count to it.
//
// The two forms are written alike, so anything but a subject the checker knows
// to be a number is walked; a number reached that way is still counted, by the
// runtime.
func isElementRange(s *ast.RangeStmt, endType types.Type) bool {
	if s.Start != nil || s.Inclusive {
		return false
	}
	switch endType.(type) {
	case types.ListType, types.MapType:
		return true
	}
	return endType == nil || types.IsAny(endType)
}

// genElementRangeWith emits the elementwise loop. subject is the
// already-generated expression to walk, and genStmt generates the body in
// whichever mode the enclosing function is being written in.
func (g *Generator) genElementRangeWith(
	s *ast.RangeStmt,
	subject string,
	genStmt func(ast.Stmt) string,
) string {
	var b strings.Builder
	if s.IndexVar != "" {
		// `purr a, b (x)` binds a litter's index and element, or a basket's key
		// and value.
		fmt.Fprintf(&b, "for __a, __b := range meow.RangePair(%s) {\n", subject)
		fmt.Fprintf(&b, "\tvar %s meow.Value = __a\n", s.IndexVar)
		fmt.Fprintf(&b, "\t_ = %s\n", s.IndexVar)
		fmt.Fprintf(&b, "\tvar %s meow.Value = __b\n", s.Var)
	} else {
		fmt.Fprintf(&b, "for __elem := range meow.RangeSolo(%s) {\n", subject)
		fmt.Fprintf(&b, "\tvar %s meow.Value = __elem\n", s.Var)
	}
	fmt.Fprintf(&b, "\t_ = %s\n", s.Var)
	// Both variables hold meow.Value, whatever the enclosing function does.
	defer g.enterBoxedScope(s.Var, s.IndexVar)()
	b.WriteString(g.genBlockStmts(s.Body, genStmt))
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
				g.markPackageUsed(realPkg)
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
		case "scram":
			return fmt.Sprintf("meow.Scram(%s)", argStr)
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
		case "upper":
			return fmt.Sprintf("meow.Upper(%s)", argStr)
		case "lower":
			return fmt.Sprintf("meow.Lower(%s)", argStr)
		case "trim":
			return fmt.Sprintf("meow.Trim(%s)", argStr)
		case "replace":
			return fmt.Sprintf("meow.Replace(%s)", argStr)
		case "pad":
			return fmt.Sprintf("meow.Pad(%s)", argStr)
		case "sort":
			return fmt.Sprintf("meow.Sort(%s)", argStr)
		case "reverse":
			return fmt.Sprintf("meow.Reverse(%s)", argStr)
		case "round":
			return fmt.Sprintf("meow.Round(%s)", argStr)
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
				if ft, ok := g.typeInfo.FuncTypes[ident.Name]; ok && !g.isNestedFunc(ident.Name) {
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
		g.markPackageUsed(realPkg)
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
	defer g.enterNestedScope()()
	return fmt.Sprintf("meow.NewFuncWithArity(\"lambda\", %d, func(args ...meow.Value) meow.Value {\n"+
		"\t%s\n"+
		"%s"+
		"%s"+
		"})", len(e.Params), g.genLambdaParamBindings(e.Params), g.callerPrologue(), g.genLambdaBody(e))
}

// genLambdaBody emits the closure body. The lambda closure has the same shape
// as an untyped function body — func(...) meow.Value — so a block body reuses
// the ordinary statement generator.
func (g *Generator) genLambdaBody(e *ast.LambdaExpr) string {
	if e.Block == nil {
		return fmt.Sprintf("\treturn %s\n", g.genExpr(e.Body))
	}
	var b strings.Builder
	b.WriteString(g.hoistNestedFuncs(e.Block))
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
	b.WriteString(g.callerPrologue())
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

// isCallTo reports whether e calls the named builtin directly.
func isCallTo(e *ast.CallExpr, name string) bool {
	ident, ok := e.Fn.(*ast.Ident)
	return ok && ident.Name == name
}

// genBlockStmts emits the statements of a block, declaring any `meow` written
// in it first.
//
// Every block gets its own pass, because a nested function is visible where it
// was written and nowhere else — the interpreter gives a `sniff` or `purr` body
// its own scope too. Without one, a nested function inside such a block was
// assigned to a variable that had never been declared and the build failed.
func (g *Generator) genBlockStmts(stmts []ast.Stmt, gen func(ast.Stmt) string) string {
	defer g.enterNestedScope()()
	var b strings.Builder
	b.WriteString(g.hoistNestedFuncs(stmts))
	for _, stmt := range stmts {
		b.WriteString("\t")
		b.WriteString(gen(stmt))
		b.WriteString("\n")
	}
	return b.String()
}

// located prefixes a statement with a note of where it came from, so that a
// failure raised while it runs can say where the program was.
//
// A statement with no position of its own — one the generator made up rather
// than read — is left alone, so it does not claim the file's first line.
func (g *Generator) located(stmt ast.Stmt, code string) string {
	pos := stmt.Pos()
	if pos.Line == 0 {
		return code
	}
	return fmt.Sprintf("meow.Here(%q)\n%s", pos.String(), code)
}

// callerPrologue opens a callable body by remembering where it was called from.
//
// Every return hands that position back, so a call which succeeds leaves the
// program where the call was made rather than inside the function it came back
// from. A body that never returns normally — one that fails — never restores it,
// which is what keeps a failure reported against the line it happened on.
func (g *Generator) callerPrologue() string {
	return "\t__caller := meow.Where()\n\t_ = __caller\n"
}

// genWhile emits the conditional purr.
//
// The condition is tested inside the loop rather than in the for clause so that
// a failure while working it out is propagated. Read as a plain truthiness test
// a Furball is false, which would end the loop quietly and let the program carry
// on as though the condition had simply stopped holding.
func (g *Generator) genWhile(s *ast.WhileStmt) string {
	var b strings.Builder
	b.WriteString("for {\n")
	fmt.Fprintf(&b, "\t__cond := %s\n", g.genExpr(s.Cond))
	b.WriteString("\tif __f, __ok := meow.AsFurball(__cond); __ok {\n\t\treturn __f\n\t}\n")
	b.WriteString("\tif !__cond.IsTruthy() {\n\t\tbreak\n\t}\n")
	b.WriteString(g.genBlockStmts(s.Body, g.genStmt))
	b.WriteString("}")
	return b.String()
}

// genTypedWhile emits the conditional purr inside a fully typed function, where
// a condition the checker knows to be a bool needs no truthiness test — and
// where a failure working it out raises rather than answering.
func (g *Generator) genTypedWhile(s *ast.WhileStmt) string {
	var b strings.Builder
	condType := g.getExprType(s.Cond)
	if condType != nil && !types.IsAny(condType) {
		fmt.Fprintf(&b, "for %s {\n", g.genTypedExpr(s.Cond))
	} else {
		fmt.Fprintf(&b, "for (%s).IsTruthy() {\n", g.genExpr(s.Cond))
	}
	b.WriteString(g.genBlockStmts(s.Body, g.genTypedStmt))
	b.WriteString("}")
	return b.String()
}
