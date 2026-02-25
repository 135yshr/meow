package parser

import (
	"iter"
	"strconv"

	"github.com/135yshr/meow/pkg/ast"
	"github.com/135yshr/meow/pkg/token"
)

// Parser produces an AST from a token stream.
type Parser struct {
	next func() (token.Token, bool)
	stop func()
	cur  token.Token
	peek token.Token
	errs []*ParseError
}

// New creates a parser from an iter.Seq of tokens.
func New(tokens iter.Seq[token.Token]) *Parser {
	next, stop := iter.Pull(tokens)
	p := &Parser{next: next, stop: stop}
	p.advance()
	p.advance()
	return p
}

func (p *Parser) advance() token.Token {
	prev := p.cur
	p.cur = p.peek
	tok, ok := p.next()
	if ok {
		p.peek = tok
	} else {
		p.peek = token.Token{Type: token.EOF}
	}
	return prev
}

func (p *Parser) curIs(types ...token.TokenType) bool {
	for _, t := range types {
		if p.cur.Type == t {
			return true
		}
	}
	return false
}

func (p *Parser) expect(typ token.TokenType) token.Token {
	if p.cur.Type != typ {
		p.errs = append(p.errs, newError(p.cur.Pos, "expected %v but got %v (%q)", typ, p.cur.Type, p.cur.Literal))
	}
	return p.advance()
}

func (p *Parser) skipNewlines() {
	for p.cur.Type == token.NEWLINE || p.cur.Type == token.COMMENT {
		p.advance()
	}
}

// Parse parses the token stream into a Program AST.
func (p *Parser) Parse() (*ast.Program, []*ParseError) {
	defer p.stop()
	prog := &ast.Program{}
	p.skipNewlines()
	for p.cur.Type != token.EOF {
		stmt := p.parseStmt()
		if stmt != nil {
			prog.Stmts = append(prog.Stmts, stmt)
		}
		p.skipNewlines()
	}
	if len(p.errs) > 0 {
		return nil, p.errs
	}
	return prog, nil
}

// Errors returns parser errors.
func (p *Parser) Errors() []*ParseError {
	return p.errs
}

func (p *Parser) parseStmt() ast.Stmt {
	switch p.cur.Type {
	case token.NYAN:
		return p.parseVarStmt()
	case token.MEOW:
		return p.parseFuncStmt()
	case token.BRING:
		return p.parseReturnStmt()
	case token.SNIFF:
		return p.parseIfStmt()
	case token.PURR:
		return p.parsePurrStmt()
	case token.FETCH:
		return p.parseFetchStmt()
	default:
		return p.parseExprStmtOrAssign()
	}
}

func (p *Parser) parseVarStmt() *ast.VarStmt {
	tok := p.advance() // consume nyan
	name := p.expect(token.IDENT)
	var typeAnn ast.TypeExpr
	if p.isTypeToken() {
		typeAnn = p.parseTypeExpr()
	}
	p.expect(token.ASSIGN)
	value := p.parseExpr(0)
	p.consumeTerminator()
	return &ast.VarStmt{Token: tok, Name: name.Literal, TypeAnn: typeAnn, Value: value}
}

func (p *Parser) parseFuncStmt() *ast.FuncStmt {
	tok := p.advance() // consume meow
	name := p.expect(token.IDENT)
	p.expect(token.LPAREN)
	params := p.parseTypedParamList()
	p.expect(token.RPAREN)
	var returnType ast.TypeExpr
	if p.isTypeToken() {
		returnType = p.parseTypeExpr()
	}
	body := p.parseBlock()
	return &ast.FuncStmt{Token: tok, Name: name.Literal, Params: params, ReturnType: returnType, Body: body}
}

func (p *Parser) parseTypedParamList() []ast.Param {
	var params []ast.Param
	if p.cur.Type == token.RPAREN {
		return params
	}
	params = append(params, p.parseParam())
	for p.cur.Type == token.COMMA {
		p.advance()
		params = append(params, p.parseParam())
	}
	return params
}

func (p *Parser) parseParam() ast.Param {
	name := p.expect(token.IDENT)
	var typeAnn ast.TypeExpr
	if p.isTypeToken() {
		typeAnn = p.parseTypeExpr()
	}
	return ast.Param{Name: name.Literal, TypeAnn: typeAnn}
}

func (p *Parser) parseTypeExpr() ast.TypeExpr {
	tok := p.advance()
	switch tok.Type {
	case token.TYPE_INT:
		return &ast.BasicType{Token: tok, Name: "int"}
	case token.TYPE_FLOAT:
		return &ast.BasicType{Token: tok, Name: "float"}
	case token.TYPE_STRING:
		return &ast.BasicType{Token: tok, Name: "string"}
	case token.TYPE_BOOL:
		return &ast.BasicType{Token: tok, Name: "bool"}
	case token.TYPE_FURBALL:
		return &ast.BasicType{Token: tok, Name: "furball"}
	case token.TYPE_LIST:
		return &ast.BasicType{Token: tok, Name: "list"}
	default:
		p.errs = append(p.errs, newError(tok.Pos, "expected type, got %v (%q)", tok.Type, tok.Literal))
		return &ast.BasicType{Token: tok, Name: tok.Literal}
	}
}

func (p *Parser) isTypeToken() bool {
	switch p.cur.Type {
	case token.TYPE_INT, token.TYPE_FLOAT, token.TYPE_STRING, token.TYPE_BOOL, token.TYPE_FURBALL, token.TYPE_LIST:
		return true
	}
	return false
}

func (p *Parser) parseBlock() []ast.Stmt {
	p.skipNewlines()
	p.expect(token.LBRACE)
	p.skipNewlines()
	var stmts []ast.Stmt
	for p.cur.Type != token.RBRACE && p.cur.Type != token.EOF {
		stmt := p.parseStmt()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
		p.skipNewlines()
	}
	p.expect(token.RBRACE)
	return stmts
}

func (p *Parser) parseReturnStmt() *ast.ReturnStmt {
	tok := p.advance() // consume bring
	var value ast.Expr
	if !p.curIs(token.NEWLINE, token.RBRACE, token.EOF) {
		value = p.parseExpr(0)
	}
	p.consumeTerminator()
	return &ast.ReturnStmt{Token: tok, Value: value}
}

func (p *Parser) parseIfStmt() *ast.IfStmt {
	tok := p.advance() // consume sniff
	p.expect(token.LPAREN)
	cond := p.parseExpr(0)
	p.expect(token.RPAREN)
	body := p.parseBlock()
	var elseBody []ast.Stmt
	p.skipNewlines()
	if p.cur.Type == token.SCRATCH {
		p.advance()
		if p.cur.Type == token.SNIFF {
			elseBody = []ast.Stmt{p.parseIfStmt()}
		} else {
			elseBody = p.parseBlock()
		}
	}
	return &ast.IfStmt{Token: tok, Condition: cond, Body: body, ElseBody: elseBody}
}

func (p *Parser) parsePurrStmt() *ast.RangeStmt {
	tok := p.advance() // consume purr
	varName := p.expect(token.IDENT)
	p.expect(token.LPAREN)
	first := p.parseExpr(0)
	if p.cur.Type == token.DOTDOT {
		// purr i (start..end) — inclusive range
		p.advance() // consume ..
		end := p.parseExpr(0)
		p.expect(token.RPAREN)
		body := p.parseBlock()
		return &ast.RangeStmt{Token: tok, Var: varName.Literal, Start: first, End: end, Inclusive: true, Body: body}
	}
	// purr i (count) — count form: 0 to count-1
	p.expect(token.RPAREN)
	body := p.parseBlock()
	return &ast.RangeStmt{Token: tok, Var: varName.Literal, Start: nil, End: first, Inclusive: false, Body: body}
}

func (p *Parser) parseFetchStmt() *ast.FetchStmt {
	tok := p.advance() // consume fetch
	path := p.expect(token.STRING)
	p.consumeTerminator()
	return &ast.FetchStmt{Token: tok, Path: path.Literal}
}

func (p *Parser) parseMemberAccess(object ast.Expr) ast.Expr {
	dot := p.advance() // consume .
	member := p.expect(token.IDENT)
	expr := &ast.MemberExpr{Token: dot, Object: object, Member: member.Literal}
	if p.cur.Type == token.LPAREN {
		return p.finishCall(expr)
	}
	return expr
}

func (p *Parser) parseExprStmtOrAssign() ast.Stmt {
	expr := p.parseExpr(0)
	if ident, ok := expr.(*ast.Ident); ok && p.cur.Type == token.ASSIGN {
		p.advance() // consume =
		value := p.parseExpr(0)
		p.consumeTerminator()
		// x = 42 is equivalent to nyan x = 42 (implicit variable declaration)
		return &ast.VarStmt{Token: ident.Token, Name: ident.Name, Value: value}
	}
	p.consumeTerminator()
	return &ast.ExprStmt{Token: expr.(ast.Node).Pos().AsToken(), Expr: expr}
}

func (p *Parser) consumeTerminator() {
	if p.curIs(token.NEWLINE, token.EOF, token.RBRACE) {
		if p.cur.Type == token.NEWLINE {
			p.advance()
		}
	}
}

// --- Expression Parsing (Pratt parser) ---

const (
	precNone  = iota
	precCatch // ~>
	precOr    // ||
	precAnd   // &&
	precEq    // == !=
	precCmp   // < > <= >=
	precPipe  // |=|
	precAdd   // + -
	precMul   // * / %
	precUnary // ! -
	precCall  // () []
)

func (p *Parser) prefixPrec(typ token.TokenType) int {
	switch typ {
	case token.MINUS, token.NOT:
		return precUnary
	default:
		return precNone
	}
}

func (p *Parser) infixPrec(typ token.TokenType) int {
	switch typ {
	case token.OR:
		return precOr
	case token.AND:
		return precAnd
	case token.EQ, token.NEQ:
		return precEq
	case token.LT, token.GT, token.LTE, token.GTE:
		return precCmp
	case token.TILDEARROW:
		return precCatch
	case token.PIPE:
		return precPipe
	case token.PLUS, token.MINUS:
		return precAdd
	case token.STAR, token.SLASH, token.PERCENT:
		return precMul
	default:
		return precNone
	}
}

func (p *Parser) parseExpr(minPrec int) ast.Expr {
	left := p.parsePrefix()
	for {
		prec := p.infixPrec(p.cur.Type)
		if prec <= minPrec {
			break
		}
		left = p.parseInfix(left, prec)
	}
	return left
}

func (p *Parser) parsePrefix() ast.Expr {
	switch p.cur.Type {
	case token.INT:
		return p.parseInt()
	case token.FLOAT:
		return p.parseFloat()
	case token.STRING:
		return p.parseString()
	case token.YARN:
		tok := p.advance()
		return &ast.BoolLit{Token: tok, Value: true}
	case token.HAIRBALL:
		tok := p.advance()
		return &ast.BoolLit{Token: tok, Value: false}
	case token.CATNAP:
		tok := p.advance()
		return &ast.NilLit{Token: tok}
	case token.IDENT:
		return p.parseIdentOrCall()
	case token.NYA:
		return p.parseNyaCall()
	case token.HISS:
		return p.parseBuiltinCall()
	case token.LICK, token.PICKY, token.CURL:
		return p.parseBuiltinCall()
	case token.LPAREN:
		return p.parseGrouped()
	case token.MINUS, token.NOT:
		tok := p.advance()
		right := p.parseExpr(precUnary)
		return &ast.UnaryExpr{Token: tok, Op: tok.Type, Right: right}
	case token.PAW:
		return p.parseLambda()
	case token.LBRACKET:
		return p.parseListLit()
	case token.LBRACE:
		return p.parseMapLit()
	case token.PEEK:
		return p.parseMatch()
	default:
		p.errs = append(p.errs, newError(p.cur.Pos, "unexpected token %v (%q)", p.cur.Type, p.cur.Literal))
		p.advance()
		return &ast.NilLit{Token: p.cur}
	}
}

func (p *Parser) parseInfix(left ast.Expr, prec int) ast.Expr {
	tok := p.advance()
	if tok.Type == token.PIPE {
		right := p.parseExpr(prec)
		return &ast.PipeExpr{Token: tok, Left: left, Right: right}
	}
	if tok.Type == token.TILDEARROW {
		right := p.parseExpr(prec)
		return &ast.CatchExpr{Token: tok, Left: left, Right: right}
	}
	right := p.parseExpr(prec)
	return &ast.BinaryExpr{Token: tok, Op: tok.Type, Left: left, Right: right}
}

func (p *Parser) parseInt() ast.Expr {
	tok := p.advance()
	val, err := strconv.ParseInt(tok.Literal, 10, 64)
	if err != nil {
		p.errs = append(p.errs, newError(tok.Pos, "invalid integer %q", tok.Literal))
	}
	return &ast.IntLit{Token: tok, Value: val}
}

func (p *Parser) parseFloat() ast.Expr {
	tok := p.advance()
	val, err := strconv.ParseFloat(tok.Literal, 64)
	if err != nil {
		p.errs = append(p.errs, newError(tok.Pos, "invalid float %q", tok.Literal))
	}
	return &ast.FloatLit{Token: tok, Value: val}
}

func (p *Parser) parseString() ast.Expr {
	tok := p.advance()
	return &ast.StringLit{Token: tok, Value: tok.Literal}
}

func (p *Parser) parseIdentOrCall() ast.Expr {
	tok := p.advance()
	ident := &ast.Ident{Token: tok, Name: tok.Literal}
	if p.cur.Type == token.DOT {
		return p.parseMemberAccess(ident)
	}
	if p.cur.Type == token.LPAREN {
		return p.finishCall(ident)
	}
	if p.cur.Type == token.LBRACKET {
		return p.parseIndex(ident)
	}
	return ident
}

func (p *Parser) parseNyaCall() ast.Expr {
	tok := p.advance() // consume nya
	ident := &ast.Ident{Token: tok, Name: "nya"}
	if p.cur.Type == token.LPAREN {
		return p.finishCall(ident)
	}
	return ident
}

func (p *Parser) parseBuiltinCall() ast.Expr {
	tok := p.advance()
	ident := &ast.Ident{Token: tok, Name: tok.Literal}
	if p.cur.Type == token.LPAREN {
		return p.finishCall(ident)
	}
	return ident
}

func (p *Parser) finishCall(fn ast.Expr) ast.Expr {
	tok := p.expect(token.LPAREN)
	args := p.parseArgList()
	p.expect(token.RPAREN)
	return &ast.CallExpr{Token: tok, Fn: fn, Args: args}
}

func (p *Parser) parseArgList() []ast.Expr {
	var args []ast.Expr
	if p.cur.Type == token.RPAREN {
		return args
	}
	args = append(args, p.parseExpr(0))
	for p.cur.Type == token.COMMA {
		p.advance()
		args = append(args, p.parseExpr(0))
	}
	return args
}

func (p *Parser) parseGrouped() ast.Expr {
	p.advance() // consume (
	expr := p.parseExpr(0)
	p.expect(token.RPAREN)
	return expr
}

func (p *Parser) parseLambda() ast.Expr {
	tok := p.advance() // consume paw
	p.expect(token.LPAREN)
	params := p.parseTypedParamList()
	p.expect(token.RPAREN)
	p.expect(token.LBRACE)
	body := p.parseExpr(0)
	p.expect(token.RBRACE)
	return &ast.LambdaExpr{Token: tok, Params: params, Body: body}
}

func (p *Parser) parseListLit() ast.Expr {
	tok := p.advance() // consume [
	var items []ast.Expr
	p.skipNewlines()
	if p.cur.Type != token.RBRACKET {
		items = append(items, p.parseExpr(0))
		for p.cur.Type == token.COMMA {
			p.advance()
			p.skipNewlines()
			if p.cur.Type == token.RBRACKET {
				break
			}
			items = append(items, p.parseExpr(0))
		}
	}
	p.skipNewlines()
	p.expect(token.RBRACKET)
	return &ast.ListLit{Token: tok, Items: items}
}

func (p *Parser) parseMapLit() ast.Expr {
	tok := p.advance() // consume {
	var keys, vals []ast.Expr
	p.skipNewlines()
	if p.cur.Type != token.RBRACE {
		key := p.parseExpr(0)
		p.expect(token.COLON)
		val := p.parseExpr(0)
		keys = append(keys, key)
		vals = append(vals, val)
		for p.cur.Type == token.COMMA {
			p.advance()
			p.skipNewlines()
			if p.cur.Type == token.RBRACE {
				break
			}
			key = p.parseExpr(0)
			p.expect(token.COLON)
			val = p.parseExpr(0)
			keys = append(keys, key)
			vals = append(vals, val)
		}
	}
	p.skipNewlines()
	p.expect(token.RBRACE)
	return &ast.MapLit{Token: tok, Keys: keys, Vals: vals}
}

func (p *Parser) parseIndex(left ast.Expr) ast.Expr {
	tok := p.advance() // consume [
	index := p.parseExpr(0)
	p.expect(token.RBRACKET)
	return &ast.IndexExpr{Token: tok, Left: left, Index: index}
}

func (p *Parser) parseMatch() ast.Expr {
	tok := p.advance() // consume peek
	p.expect(token.LPAREN)
	subject := p.parseExpr(0)
	p.expect(token.RPAREN)
	p.skipNewlines()
	p.expect(token.LBRACE)
	p.skipNewlines()
	var arms []ast.MatchArm
	for p.cur.Type != token.RBRACE && p.cur.Type != token.EOF {
		pattern := p.parsePattern()
		p.expect(token.ARROW)
		body := p.parseExpr(0)
		arms = append(arms, ast.MatchArm{Pattern: pattern, Body: body})
		p.skipNewlines()
		if p.cur.Type == token.COMMA {
			p.advance()
			p.skipNewlines()
		}
	}
	p.expect(token.RBRACE)
	return &ast.MatchExpr{Token: tok, Subject: subject, Arms: arms}
}

func (p *Parser) parsePattern() ast.Pattern {
	if p.cur.Type == token.IDENT && p.cur.Literal == "_" {
		tok := p.advance()
		return &ast.WildcardPattern{Token: tok}
	}
	expr := p.parsePrefix()
	if p.cur.Type == token.DOTDOT {
		tok := p.advance()
		high := p.parsePrefix()
		return &ast.RangePattern{Token: tok, Low: expr, High: high}
	}
	return &ast.LiteralPattern{Token: expr.(ast.Node).Pos().AsToken(), Value: expr}
}
