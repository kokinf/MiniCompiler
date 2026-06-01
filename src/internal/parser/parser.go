package parser

import (
	"fmt"
	"strconv"

	"mikrocompiler/src/internal/ast"
	"mikrocompiler/src/internal/token"
)

type Parser struct {
	tokens   []token.Token
	position int
	current  token.Token

	errors []string
}

func NewParser(tokens []token.Token) *Parser {
	p := &Parser{
		tokens:   tokens,
		position: 0,
		errors:   []string{},
	}
	if len(tokens) > 0 {
		p.current = tokens[0]
	}
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.position++
	if p.position < len(p.tokens) {
		p.current = p.tokens[p.position]
	} else {
		p.current = token.Token{Type: token.EOF, Lexeme: "", Line: 0, Column: 0}
	}
}

func (p *Parser) peek() token.Token {
	if p.position+1 < len(p.tokens) {
		return p.tokens[p.position+1]
	}
	return token.Token{Type: token.EOF, Lexeme: "", Line: 0, Column: 0}
}

func (p *Parser) expect(expected token.TokenType) bool {
	if p.current.Type == expected {
		p.nextToken()
		return true
	}
	p.addError(fmt.Sprintf("ожидался токен %s, получен %s (строка %d, колонка %d)",
		expected, p.current.Type, p.current.Line, p.current.Column))
	return false
}

func (p *Parser) consume() token.Token {
	tok := p.current
	p.nextToken()
	return tok
}

func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, msg)
}

func (p *Parser) synchronize() {
	p.nextToken()
	for p.current.Type != token.EOF {
		if p.current.Type == token.SEMICOLON {
			p.nextToken()
			return
		}
		switch p.current.Type {
		case token.KW_FN, token.KW_STRUCT, token.KW_EXTERN,
			token.KW_IF, token.KW_WHILE, token.KW_FOR,
			token.KW_RETURN, token.RBRACE:
			return
		}
		p.nextToken()
	}
}

func (p *Parser) Parse() *ast.ProgramNode {
	program := &ast.ProgramNode{
		Declarations: []ast.DeclarationNode{},
	}

	for p.current.Type != token.EOF {
		decl := p.parseDeclaration()
		if decl != nil {
			program.Declarations = append(program.Declarations, decl)
			if program.LinePos == 0 {
				program.LinePos = decl.Line()
				program.ColumnPos = decl.Column()
			}
		} else {
			p.synchronize()
		}
	}

	return program
}

func (p *Parser) parseDeclaration() ast.DeclarationNode {
	switch p.current.Type {
	case token.KW_FN:
		return p.parseFunctionDecl()
	case token.KW_STRUCT:
		return p.parseStructDecl()
	case token.KW_EXTERN:
		return p.parseExternDecl()
	case token.KW_INT, token.KW_FLOAT, token.KW_BOOL, token.KW_STRING, token.KW_VOID, token.IDENTIFIER:
		return p.parseVarDeclOrFuncDecl()
	default:
		p.addError(fmt.Sprintf("неожиданный токен в объявлении: %s", p.current.Type))
		return nil
	}
}

// parseExternDecl парсит extern объявление
func (p *Parser) parseExternDecl() *ast.ExternFuncDeclNode {
	externToken := p.consume() // съедаем 'extern'

	// Парсим возвращаемый тип (может быть void, int, void*, etc.)
	returnType := p.parseType()
	if returnType == nil || returnType.Kind == "unknown" {
		p.addError("ожидался тип возврата после extern")
		return nil
	}

	// Имя функции
	if p.current.Type != token.IDENTIFIER {
		p.addError(fmt.Sprintf("ожидалось имя функции, получен %s", p.current.Type))
		return nil
	}

	name := &ast.IdentifierNode{
		Token: p.current,
		Value: p.current.Lexeme,
	}
	p.nextToken()

	// Открывающая скобка
	if !p.expect(token.LPAREN) {
		return nil
	}

	// Параметры
	params := p.parseParameters()

	// Закрывающая скобка
	if !p.expect(token.RPAREN) {
		return nil
	}

	// Точка с запятой
	if p.current.Type == token.SEMICOLON {
		p.nextToken()
	}

	return &ast.ExternFuncDeclNode{
		Token:      externToken,
		Name:       name,
		Parameters: params,
		ReturnType: returnType,
	}
}

func (p *Parser) parseFunctionDecl() *ast.FunctionDeclNode {
	// Проверяем, есть ли уже тип возврата (из parseVarDeclOrFuncDecl)
	var returnType *ast.TypeNode

	// Если текущий токен - не 'fn', значит тип возврата уже спаршен
	if p.current.Type != token.KW_FN {
		returnType = p.parseType()
	}

	fnToken := p.current
	if fnToken.Type != token.KW_FN {
		// Это функция без 'fn', просто тип имя(...)
		fnToken = returnType.Token
	} else {
		p.nextToken() // съедаем 'fn'
	}

	if p.current.Type != token.IDENTIFIER {
		p.addError(fmt.Sprintf("ожидалось имя функции, получен %s (строка %d, колонка %d)",
			p.current.Type, p.current.Line, p.current.Column))
		return nil
	}

	name := &ast.IdentifierNode{
		Token: p.current,
		Value: p.current.Lexeme,
	}
	p.nextToken()

	if !p.expect(token.LPAREN) {
		return nil
	}

	params := p.parseParameters()

	if !p.expect(token.RPAREN) {
		return nil
	}

	// Проверяем стрелку -> для типа возврата
	if p.current.Type == token.ARROW {
		p.nextToken()
		returnType = p.parseType()
	}

	// Если тип возврата всё ещё не определён
	if returnType == nil {
		returnType = &ast.TypeNode{Kind: "void"}
	}

	// Только объявление (без тела)
	if p.current.Type == token.SEMICOLON {
		p.nextToken()
		return &ast.FunctionDeclNode{
			Token:      fnToken,
			Name:       name,
			Parameters: params,
			ReturnType: returnType,
			Body:       nil,
		}
	}

	// Тело функции
	if p.current.Type != token.LBRACE {
		p.addError(fmt.Sprintf("ожидался токен LBRACE или ';', получен %s (строка %d, колонка %d)",
			p.current.Type, p.current.Line, p.current.Column))
		return nil
	}

	body := p.parseBlockStmt()

	return &ast.FunctionDeclNode{
		Token:      fnToken,
		Name:       name,
		Parameters: params,
		ReturnType: returnType,
		Body:       body,
	}
}

func (p *Parser) parseParameters() []*ast.ParameterNode {
	params := []*ast.ParameterNode{}

	if p.current.Type == token.RPAREN {
		return params
	}

	param := p.parseParameter()
	if param != nil {
		params = append(params, param)
	}

	for p.current.Type == token.COMMA {
		p.nextToken()
		// Проверяем на variadic (...)
		if p.current.Type == token.ELLIPSIS {
			p.nextToken()
			break
		}
		param := p.parseParameter()
		if param != nil {
			params = append(params, param)
		}
	}

	return params
}

func (p *Parser) parseParameter() *ast.ParameterNode {
	// Проверяем на variadic
	if p.current.Type == token.ELLIPSIS {
		p.nextToken()
		return nil
	}

	// Парсим тип параметра
	paramType := p.parseType()
	if paramType == nil || paramType.Kind == "unknown" {
		p.addError(fmt.Sprintf("ожидался тип параметра, получен %s", p.current.Type))
		return nil
	}

	// Имя параметра
	if p.current.Type != token.IDENTIFIER {
		p.addError(fmt.Sprintf("ожидалось имя параметра, получен %s", p.current.Type))
		return nil
	}

	name := &ast.IdentifierNode{
		Token: p.current,
		Value: p.current.Lexeme,
	}
	p.nextToken()

	// Проверяем на массив (arr[])
	if p.current.Type == token.LBRACKET {
		p.nextToken()
		if !p.expect(token.RBRACKET) {
			return nil
		}
		// Превращаем тип в массив
		paramType = &ast.TypeNode{
			Token:     paramType.Token,
			Kind:      "array",
			BaseType:  paramType,
			ArraySize: -1,
		}
	}

	return &ast.ParameterNode{
		Token: paramType.Token,
		Type:  paramType,
		Name:  name,
	}
}

func (p *Parser) parseStructDecl() *ast.StructDeclNode {
	structToken := p.consume()

	if p.current.Type != token.IDENTIFIER {
		p.addError(fmt.Sprintf("ожидалось имя структуры, получен %s (строка %d, колонка %d)",
			p.current.Type, p.current.Line, p.current.Column))
		return nil
	}
	name := &ast.IdentifierNode{
		Token: p.current,
		Value: p.current.Lexeme,
	}
	p.nextToken()

	if !p.expect(token.LBRACE) {
		return nil
	}
	fields := []*ast.VarDeclNode{}

	for p.current.Type != token.RBRACE && p.current.Type != token.EOF {
		if p.current.Type != token.KW_INT && p.current.Type != token.KW_FLOAT &&
			p.current.Type != token.KW_BOOL && p.current.Type != token.KW_STRING &&
			p.current.Type != token.IDENTIFIER {
			p.addError(fmt.Sprintf("ожидался тип поля, получен %s", p.current.Type))
			p.synchronize()
			continue
		}

		varType := p.parseType()

		if p.current.Type != token.IDENTIFIER {
			p.addError(fmt.Sprintf("ожидалось имя поля, получен %s", p.current.Type))
			p.synchronize()
			continue
		}

		fieldName := &ast.IdentifierNode{
			Token: p.current,
			Value: p.current.Lexeme,
		}
		p.nextToken()

		if p.current.Type == token.ASSIGN {
			p.addError("поля структуры не могут иметь инициализаторы")
			for p.current.Type != token.SEMICOLON && p.current.Type != token.RBRACE && p.current.Type != token.EOF {
				p.nextToken()
			}
		}

		field := &ast.VarDeclNode{
			Token: varType.Token,
			Type:  varType,
			Name:  fieldName,
		}
		fields = append(fields, field)

		if p.current.Type == token.SEMICOLON {
			p.nextToken()
		} else if p.current.Type != token.RBRACE {
			p.addError(fmt.Sprintf("ожидалась ';' или '}}', получен %s", p.current.Type))
			p.synchronize()
		}
	}

	if !p.expect(token.RBRACE) {
		return nil
	}

	return &ast.StructDeclNode{
		Token:  structToken,
		Name:   name,
		Fields: fields,
	}
}

// parseVarDeclOrFuncDecl различает объявление переменной и функции
func (p *Parser) parseVarDeclOrFuncDecl() ast.DeclarationNode {
	savedPos := p.position
	savedCur := p.current

	// Парсим тип (int, float, void, etc.)
	varType := p.parseType()
	if varType.Kind == "unknown" {
		return nil
	}

	if p.current.Type != token.IDENTIFIER {
		p.addError(fmt.Sprintf("ожидалось имя, получен %s", p.current.Type))
		return nil
	}

	name := &ast.IdentifierNode{
		Token: p.current,
		Value: p.current.Lexeme,
	}
	p.nextToken()

	// Проверяем на массив: int arr[5]
	if p.current.Type == token.LBRACKET {
		p.nextToken()
		if p.current.Type == token.INT_LITERAL {
			size, err := strconv.Atoi(p.current.Lexeme)
			if err == nil && size >= 0 {
				varType = &ast.TypeNode{
					Token:     varType.Token,
					Kind:      "array",
					BaseType:  varType,
					ArraySize: size,
				}
			}
			p.nextToken()
		}
		if !p.expect(token.RBRACKET) {
			return nil
		}
	}

	// Если следующая скобка - это функция
	if p.current.Type == token.LPAREN {
		// Восстанавливаем позицию и парсим как функцию с типом возврата
		p.position = savedPos
		p.current = savedCur
		return p.parseFunctionDecl()
	}

	// Иначе это переменная
	var initializer ast.ExpressionNode = nil
	if p.current.Type == token.ASSIGN {
		p.nextToken()
		if p.current.Type == token.LBRACE {
			initializer = p.parseArrayLiteral()
		} else {
			initializer = p.parseExpression()
		}
	}

	if p.current.Type != token.SEMICOLON {
		p.addError(fmt.Sprintf("ожидалась ';', получен %s (строка %d, колонка %d)",
			p.current.Type, p.current.Line, p.current.Column))
		p.synchronize()
		return nil
	}
	p.nextToken()

	return &ast.VarDeclNode{
		Token:       varType.Token,
		Type:        varType,
		Name:        name,
		Initializer: initializer,
	}
}

func (p *Parser) parseVarDecl() ast.DeclarationNode {
	return p.parseVarDeclOrFuncDecl()
}

func (p *Parser) parseArrayLiteral() *ast.ArrayLiteralExprNode {
	tok := p.current
	p.nextToken()

	elements := []ast.ExpressionNode{}

	if p.current.Type != token.RBRACE {
		elements = append(elements, p.parseExpression())

		for p.current.Type == token.COMMA {
			p.nextToken()
			if p.current.Type == token.RBRACE {
				break
			}
			elements = append(elements, p.parseExpression())
		}
	}

	if !p.expect(token.RBRACE) {
		return nil
	}

	return &ast.ArrayLiteralExprNode{
		Token:    tok,
		Elements: elements,
	}
}

func (p *Parser) parseStatement() ast.StatementNode {
	switch p.current.Type {
	case token.LBRACE:
		return p.parseBlockStmt()
	case token.KW_IF:
		return p.parseIfStmt()
	case token.KW_WHILE:
		return p.parseWhileStmt()
	case token.KW_FOR:
		return p.parseForStmt()
	case token.KW_RETURN:
		return p.parseReturnStmt()
	case token.KW_INT, token.KW_FLOAT, token.KW_BOOL, token.KW_STRING:
		decl := p.parseVarDecl()
		if decl != nil {
			if vd, ok := decl.(*ast.VarDeclNode); ok {
				return vd
			}
		}
		return nil
	case token.IDENTIFIER:
		return p.parseExprStmt()
	case token.SEMICOLON:
		p.nextToken()
		return nil
	default:
		return p.parseExprStmt()
	}
}

func (p *Parser) parseBlockStmt() *ast.BlockStmtNode {
	if p.current.Type != token.LBRACE {
		p.addError(fmt.Sprintf("ожидался токен LBRACE, получен %s (строка %d, колонка %d)",
			p.current.Type, p.current.Line, p.current.Column))
		return nil
	}

	block := &ast.BlockStmtNode{
		Token:      p.current,
		Statements: []ast.StatementNode{},
	}
	p.nextToken()

	for p.current.Type != token.RBRACE && p.current.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		} else {
			p.synchronize()
		}
	}

	if !p.expect(token.RBRACE) {
		return nil
	}

	return block
}

func (p *Parser) parseIfStmt() *ast.IfStmtNode {
	ifStmt := &ast.IfStmtNode{
		Token: p.current,
	}
	p.nextToken()

	if !p.expect(token.LPAREN) {
		return nil
	}
	ifStmt.Condition = p.parseExpression()
	if !p.expect(token.RPAREN) {
		return nil
	}

	if p.current.Type == token.LBRACE {
		ifStmt.Consequence = p.parseBlockStmt()
	} else {
		stmt := p.parseStatement()
		if stmt != nil {
			ifStmt.Consequence = &ast.BlockStmtNode{
				Token:      token.Token{Type: token.LBRACE, Lexeme: "{", Line: stmt.Line(), Column: stmt.Column()},
				Statements: []ast.StatementNode{stmt},
			}
		}
	}

	if p.current.Type == token.KW_ELSE {
		p.nextToken()

		if p.current.Type == token.LBRACE {
			ifStmt.Alternative = p.parseBlockStmt()
		} else if p.current.Type == token.KW_IF {
			ifStmt.Alternative = p.parseIfStmt()
		} else {
			stmt := p.parseStatement()
			if stmt != nil {
				ifStmt.Alternative = &ast.BlockStmtNode{
					Token:      token.Token{Type: token.LBRACE, Lexeme: "{", Line: stmt.Line(), Column: stmt.Column()},
					Statements: []ast.StatementNode{stmt},
				}
			}
		}
	}

	return ifStmt
}

func (p *Parser) parseWhileStmt() *ast.WhileStmtNode {
	whileStmt := &ast.WhileStmtNode{
		Token: p.current,
	}
	p.nextToken()

	if !p.expect(token.LPAREN) {
		return nil
	}
	whileStmt.Condition = p.parseExpression()
	if !p.expect(token.RPAREN) {
		return nil
	}

	if p.current.Type == token.LBRACE {
		whileStmt.Body = p.parseBlockStmt()
	} else {
		stmt := p.parseStatement()
		if stmt != nil {
			whileStmt.Body = &ast.BlockStmtNode{
				Token:      token.Token{Type: token.LBRACE, Lexeme: "{", Line: stmt.Line(), Column: stmt.Column()},
				Statements: []ast.StatementNode{stmt},
			}
		}
	}

	return whileStmt
}

func (p *Parser) parseForStmt() *ast.ForStmtNode {
	forStmt := &ast.ForStmtNode{
		Token: p.current,
	}
	p.nextToken()

	if !p.expect(token.LPAREN) {
		return nil
	}

	if p.current.Type != token.SEMICOLON {
		if p.current.Type == token.KW_INT || p.current.Type == token.KW_FLOAT ||
			p.current.Type == token.KW_BOOL || p.current.Type == token.KW_STRING {
			decl := p.parseVarDecl()
			if decl != nil {
				if vd, ok := decl.(*ast.VarDeclNode); ok {
					forStmt.Init = vd
				}
			}
		} else {
			forStmt.Init = p.parseExprStmt()
		}
	} else {
		p.nextToken()
	}

	if p.current.Type != token.SEMICOLON {
		forStmt.Condition = p.parseExpression()
	}
	if !p.expect(token.SEMICOLON) {
		return nil
	}

	if p.current.Type != token.RPAREN {
		forStmt.Update = p.parseExpression()
	}
	if !p.expect(token.RPAREN) {
		return nil
	}

	if p.current.Type == token.LBRACE {
		forStmt.Body = p.parseBlockStmt()
	} else {
		stmt := p.parseStatement()
		if stmt != nil {
			forStmt.Body = &ast.BlockStmtNode{
				Token:      token.Token{Type: token.LBRACE, Lexeme: "{", Line: stmt.Line(), Column: stmt.Column()},
				Statements: []ast.StatementNode{stmt},
			}
		}
	}

	return forStmt
}

func (p *Parser) parseReturnStmt() *ast.ReturnStmtNode {
	returnStmt := &ast.ReturnStmtNode{
		Token: p.current,
	}
	p.nextToken()

	if p.current.Type != token.SEMICOLON {
		returnStmt.RetValue = p.parseExpression()
	}

	if !p.expect(token.SEMICOLON) {
		return nil
	}

	return returnStmt
}

func (p *Parser) parseExprStmt() *ast.ExprStmtNode {
	exprStmt := &ast.ExprStmtNode{
		Token:      p.current,
		Expression: p.parseExpression(),
	}

	if !p.expect(token.SEMICOLON) {
		return nil
	}

	return exprStmt
}

func (p *Parser) parseExpression() ast.ExpressionNode {
	if p.current.Type == token.ILLEGAL {
		p.addError(p.current.Lexeme)
		p.nextToken()
		return nil
	}
	return p.parseAssignment()
}

func (p *Parser) parseAssignment() ast.ExpressionNode {
	expr := p.parseLogicalOr()

	if p.current.Type == token.ASSIGN ||
		p.current.Type == token.PLUS_ASSIGN ||
		p.current.Type == token.MINUS_ASSIGN ||
		p.current.Type == token.MULTIPLY_ASSIGN ||
		p.current.Type == token.DIVIDE_ASSIGN {

		switch expr.(type) {
		case *ast.IdentifierNode, *ast.IndexExprNode:
			// Допустимо
		default:
			p.addError(fmt.Sprintf("левая часть присваивания должна быть идентификатором (строка %d, колонка %d)",
				p.current.Line, p.current.Column))
			p.nextToken()
			p.parseAssignment()
			return expr
		}

		tok := p.current
		operator := tok.Lexeme
		p.nextToken()
		right := p.parseAssignment()

		if operator != "=" {
			left := expr
			return &ast.AssignmentExprNode{
				Token:    tok,
				Left:     left,
				Operator: "=",
				Right: &ast.BinaryExprNode{
					Token:    tok,
					Left:     left,
					Operator: operator[:len(operator)-1],
					Right:    right,
				},
			}
		}

		return &ast.AssignmentExprNode{
			Token:    tok,
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) parseLogicalOr() ast.ExpressionNode {
	expr := p.parseLogicalAnd()

	for p.current.Type == token.OR {
		tok := p.current
		p.nextToken()
		right := p.parseLogicalAnd()
		expr = &ast.BinaryExprNode{
			Token:    tok,
			Left:     expr,
			Operator: tok.Lexeme,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) parseLogicalAnd() ast.ExpressionNode {
	expr := p.parseEquality()

	for p.current.Type == token.AND {
		tok := p.current
		p.nextToken()
		right := p.parseEquality()
		expr = &ast.BinaryExprNode{
			Token:    tok,
			Left:     expr,
			Operator: tok.Lexeme,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) parseEquality() ast.ExpressionNode {
	expr := p.parseRelational()

	for p.current.Type == token.EQ || p.current.Type == token.NOT_EQ {
		tok := p.current
		p.nextToken()
		right := p.parseRelational()
		expr = &ast.BinaryExprNode{
			Token:    tok,
			Left:     expr,
			Operator: tok.Lexeme,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) parseRelational() ast.ExpressionNode {
	expr := p.parseAdditive()

	for p.current.Type == token.LT || p.current.Type == token.LT_EQ ||
		p.current.Type == token.GT || p.current.Type == token.GT_EQ {
		tok := p.current
		p.nextToken()
		right := p.parseAdditive()
		expr = &ast.BinaryExprNode{
			Token:    tok,
			Left:     expr,
			Operator: tok.Lexeme,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) parseAdditive() ast.ExpressionNode {
	expr := p.parseMultiplicative()

	for p.current.Type == token.PLUS || p.current.Type == token.MINUS {
		tok := p.current
		p.nextToken()
		right := p.parseMultiplicative()
		expr = &ast.BinaryExprNode{
			Token:    tok,
			Left:     expr,
			Operator: tok.Lexeme,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) parseMultiplicative() ast.ExpressionNode {
	expr := p.parseUnary()

	for p.current.Type == token.MULTIPLY || p.current.Type == token.DIVIDE ||
		p.current.Type == token.MODULO {
		tok := p.current
		p.nextToken()
		right := p.parseUnary()
		expr = &ast.BinaryExprNode{
			Token:    tok,
			Left:     expr,
			Operator: tok.Lexeme,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) parseUnary() ast.ExpressionNode {
	if p.current.Type == token.MINUS {
		tok := p.current
		operator := tok.Lexeme
		p.nextToken()
		right := p.parseUnary()
		return &ast.UnaryExprNode{
			Token:    tok,
			Operator: operator,
			Right:    right,
		}
	}

	if p.current.Type == token.NOT {
		tok := p.current
		p.nextToken()
		right := p.parseUnary()
		return &ast.UnaryExprNode{
			Token:    tok,
			Operator: "!",
			Right:    right,
		}
	}

	return p.parsePostfix()
}

func (p *Parser) parsePostfix() ast.ExpressionNode {
	expr := p.parsePrimary()

	for {
		switch p.current.Type {
		case token.LPAREN:
			tok := p.current
			p.nextToken()
			args := []ast.ExpressionNode{}

			if p.current.Type != token.RPAREN {
				args = append(args, p.parseExpression())

				for p.current.Type == token.COMMA {
					p.nextToken()
					args = append(args, p.parseExpression())
				}
			}

			if !p.expect(token.RPAREN) {
				return expr
			}

			expr = &ast.CallExprNode{
				Token:     tok,
				Function:  expr,
				Arguments: args,
			}

		case token.LBRACKET:
			tok := p.current
			p.nextToken()
			index := p.parseExpression()
			if !p.expect(token.RBRACKET) {
				return expr
			}
			expr = &ast.IndexExprNode{
				Token: tok,
				Array: expr,
				Index: index,
			}

		case token.DOT:
			p.nextToken()
			if p.current.Type == token.IDENTIFIER {
				field := &ast.IdentifierNode{
					Token: p.current,
					Value: p.current.Lexeme,
				}
				p.nextToken()
				expr = field
			}
		default:
			return expr
		}
	}
}

func (p *Parser) parsePrimary() ast.ExpressionNode {
	switch p.current.Type {
	case token.IDENTIFIER:
		ident := &ast.IdentifierNode{
			Token: p.current,
			Value: p.current.Lexeme,
		}
		p.nextToken()
		return ident
	case token.INT_LITERAL:
		lit := &ast.LiteralExprNode{
			Token:    p.current,
			TypeName: "int",
			IntValue: p.current.Literal.IntValue,
		}
		p.nextToken()
		return lit
	case token.FLOAT_LITERAL:
		lit := &ast.LiteralExprNode{
			Token:      p.current,
			TypeName:   "float",
			FloatValue: p.current.Literal.FloatValue,
		}
		p.nextToken()
		return lit
	case token.STRING_LITERAL:
		lit := &ast.LiteralExprNode{
			Token:       p.current,
			TypeName:    "string",
			StringValue: p.current.Literal.StringValue,
		}
		p.nextToken()
		return lit
	case token.KW_TRUE:
		lit := &ast.LiteralExprNode{
			Token:     p.current,
			TypeName:  "bool",
			BoolValue: true,
		}
		p.nextToken()
		return lit
	case token.KW_FALSE:
		lit := &ast.LiteralExprNode{
			Token:     p.current,
			TypeName:  "bool",
			BoolValue: false,
		}
		p.nextToken()
		return lit
	case token.LBRACE:
		return p.parseArrayLiteral()
	case token.ILLEGAL:
		p.addError(p.current.Lexeme)
		p.nextToken()
		return nil
	case token.LPAREN:
		p.nextToken()
		expr := p.parseExpression()
		if !p.expect(token.RPAREN) {
			return nil
		}
		return expr
	default:
		p.addError(fmt.Sprintf("неожиданный токен в выражении: %s (строка %d, колонка %d)",
			p.current.Type, p.current.Line, p.current.Column))
		return nil
	}
}

func (p *Parser) parseType() *ast.TypeNode {
	var baseType *ast.TypeNode

	switch p.current.Type {
	case token.KW_INT:
		baseType = &ast.TypeNode{Token: p.current, Kind: "int"}
		p.nextToken()
	case token.KW_FLOAT:
		baseType = &ast.TypeNode{Token: p.current, Kind: "float"}
		p.nextToken()
	case token.KW_BOOL:
		baseType = &ast.TypeNode{Token: p.current, Kind: "bool"}
		p.nextToken()
	case token.KW_VOID:
		baseType = &ast.TypeNode{Token: p.current, Kind: "void"}
		p.nextToken()
	case token.KW_STRING:
		baseType = &ast.TypeNode{Token: p.current, Kind: "string"}
		p.nextToken()
	case token.IDENTIFIER:
		baseType = &ast.TypeNode{Token: p.current, Kind: "identifier", Name: p.current.Lexeme}
		p.nextToken()
	default:
		p.addError(fmt.Sprintf("ожидался тип, получен %s (строка %d, колонка %d)",
			p.current.Type, p.current.Line, p.current.Column))
		return &ast.TypeNode{Kind: "unknown"}
	}

	// Проверяем на указатель (*)
	if p.current.Type == token.MULTIPLY {
		p.nextToken()
		baseType = &ast.TypeNode{
			Token:    baseType.Token,
			Kind:     "pointer",
			BaseType: baseType,
		}
	}

	// Проверяем на массив ([])
	for p.current.Type == token.LBRACKET {
		p.nextToken()

		arrayType := &ast.TypeNode{
			Token:     baseType.Token,
			Kind:      "array",
			BaseType:  baseType,
			ArraySize: -1,
		}

		if p.current.Type == token.INT_LITERAL {
			size, err := strconv.Atoi(p.current.Lexeme)
			if err == nil && size >= 0 {
				arrayType.ArraySize = size
			}
			p.nextToken()
		}

		if !p.expect(token.RBRACKET) {
			return &ast.TypeNode{Kind: "unknown"}
		}

		baseType = arrayType
	}

	return baseType
}
