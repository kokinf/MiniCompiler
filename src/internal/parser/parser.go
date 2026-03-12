package parser

import (
	"fmt"

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

func (p *Parser) peekNext() token.Token {
	if p.position+2 < len(p.tokens) {
		return p.tokens[p.position+2]
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

func (p *Parser) expectPeek(expected token.TokenType) bool {
	if p.peek().Type == expected {
		p.nextToken()
		return true
	}
	p.addError(fmt.Sprintf("ожидался токен %s, получен %s (строка %d, колонка %d)",
		expected, p.peek().Type, p.peek().Line, p.peek().Column))
	return false
}

func (p *Parser) consume() token.Token {
	tok := p.current
	p.nextToken()
	return tok
}

func (p *Parser) consumeIf(typ token.TokenType) (token.Token, bool) {
	if p.current.Type == typ {
		tok := p.current
		p.nextToken()
		return tok, true
	}
	return token.Token{}, false
}

func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, msg)
}

func (p *Parser) isAtEnd() bool {
	return p.current.Type == token.EOF
}

// Parse запускает парсинг программы
func (p *Parser) Parse() *ast.Program {
	program := &ast.Program{
		Declarations: []ast.Declaration{},
	}

	for !p.isAtEnd() {
		decl := p.parseDeclaration()
		if decl != nil {
			program.Declarations = append(program.Declarations, decl)
		} else {
			p.nextToken()
		}
	}

	return program
}

// parseDeclaration парсит объявление
func (p *Parser) parseDeclaration() ast.Declaration {
	switch p.current.Type {
	case token.KW_FN:
		return p.parseFunctionDecl()
	case token.KW_STRUCT:
		return p.parseStructDecl()
	case token.KW_INT, token.KW_FLOAT, token.KW_BOOL, token.IDENTIFIER:
		return p.parseVarDecl()
	default:
		return nil
	}
}

// parseFunctionDecl парсит объявление функции
func (p *Parser) parseFunctionDecl() *ast.FunctionDecl {
	fnToken := p.consume()

	if p.current.Type != token.IDENTIFIER {
		p.addError(fmt.Sprintf("ожидалось имя функции, получен %s (строка %d, колонка %d)",
			p.current.Type, p.current.Line, p.current.Column))
		return nil
	}
	name := &ast.Identifier{
		Token: p.current,
		Value: p.current.Lexeme,
	}
	p.nextToken()

	// Параметры
	if !p.expect(token.LPAREN) {
		return nil
	}
	params := p.parseParameters()
	if !p.expect(token.RPAREN) {
		return nil
	}

	returnType := ast.Type{Kind: "void"}
	if p.current.Type == token.ARROW {
		p.nextToken()
		returnType = p.parseType()
	}

	if p.current.Type == token.SEMICOLON {
		p.nextToken()
		return &ast.FunctionDecl{
			Token:      fnToken,
			Name:       name,
			Parameters: params,
			ReturnType: returnType,
			Body:       nil,
		}
	}

	if p.current.Type != token.LBRACE {
		p.addError(fmt.Sprintf("ожидался токен LBRACE или ';', получен %s (строка %d, колонка %d)",
			p.current.Type, p.current.Line, p.current.Column))
		return nil
	}

	body := p.parseBlockStmt()

	return &ast.FunctionDecl{
		Token:      fnToken,
		Name:       name,
		Parameters: params,
		ReturnType: returnType,
		Body:       body,
	}
}

// parseParameters парсит параметры функции
func (p *Parser) parseParameters() []*ast.Parameter {
	params := []*ast.Parameter{}

	if p.current.Type == token.RPAREN {
		return params
	}

	param := p.parseParameter()
	if param != nil {
		params = append(params, param)
	}

	for p.current.Type == token.COMMA {
		p.nextToken() // потребляем ','
		param := p.parseParameter()
		if param != nil {
			params = append(params, param)
		}
	}

	return params
}

// parseParameter парсит один параметр
func (p *Parser) parseParameter() *ast.Parameter {
	if p.current.Type != token.IDENTIFIER {
		p.addError(fmt.Sprintf("ожидалось имя параметра, получен %s (строка %d, колонка %d)",
			p.current.Type, p.current.Line, p.current.Column))
		return nil
	}

	name := &ast.Identifier{
		Token: p.current,
		Value: p.current.Lexeme,
	}
	p.nextToken()

	// После имени должен идти тип
	paramType := p.parseType()

	return &ast.Parameter{
		Token: paramType.Token,
		Type:  paramType,
		Name:  name,
	}
}

// parseStructDecl парсит объявление структуры
func (p *Parser) parseStructDecl() *ast.StructDecl {
	structToken := p.consume()

	// Имя структуры
	if p.current.Type != token.IDENTIFIER {
		p.addError(fmt.Sprintf("ожидалось имя структуры, получен %s (строка %d, колонка %d)",
			p.current.Type, p.current.Line, p.current.Column))
		return nil
	}
	name := &ast.Identifier{
		Token: p.current,
		Value: p.current.Lexeme,
	}
	p.nextToken()

	// Поля структуры
	if !p.expect(token.LBRACE) {
		return nil
	}
	fields := []*ast.VarDecl{}

	for p.current.Type != token.RBRACE && !p.isAtEnd() {

		if p.current.Type == token.KW_INT || p.current.Type == token.KW_FLOAT ||
			p.current.Type == token.KW_BOOL || p.current.Type == token.IDENTIFIER {

			varType := p.parseType()

			if p.current.Type != token.IDENTIFIER {
				p.addError(fmt.Sprintf("ожидалось имя поля, получен %s (строка %d, колонка %d)",
					p.current.Type, p.current.Line, p.current.Column))
				for p.current.Type != token.SEMICOLON && p.current.Type != token.RBRACE && !p.isAtEnd() {
					p.nextToken()
				}
				if p.current.Type == token.SEMICOLON {
					p.nextToken()
				}
				continue
			}

			fieldName := &ast.Identifier{
				Token: p.current,
				Value: p.current.Lexeme,
			}
			p.nextToken()

			// Инициализатор не допускается в полях структуры
			if p.current.Type == token.ASSIGN {
				p.addError("поля структуры не могут иметь инициализаторы")
				for p.current.Type != token.SEMICOLON && p.current.Type != token.RBRACE && !p.isAtEnd() {
					p.nextToken()
				}
			}

			field := &ast.VarDecl{
				Token: varType.Token,
				Type:  varType,
				Name:  fieldName,
			}
			fields = append(fields, field)

			// Ожидаем точку с запятой или закрывающую скобку
			if p.current.Type != token.SEMICOLON && p.current.Type != token.RBRACE {
				p.addError(fmt.Sprintf("ожидалась ';' или '}}', получен %s (строка %d, колонка %d)",
					p.current.Type, p.current.Line, p.current.Column))
				for p.current.Type != token.SEMICOLON && p.current.Type != token.RBRACE && !p.isAtEnd() {
					p.nextToken()
				}
			}
			if p.current.Type == token.SEMICOLON {
				p.nextToken()
			}
		} else {
			// Если встретили неожиданный токен
			if p.current.Type == token.RBRACE {
				break
			}
			if p.current.Type == token.LBRACKET {
				p.nextToken()
				for p.current.Type != token.RBRACKET && !p.isAtEnd() {
					p.nextToken()
				}
				if p.current.Type == token.RBRACKET {
					p.nextToken()
				}
				if p.current.Type == token.IDENTIFIER {
					p.nextToken()
				}
				if p.current.Type == token.SEMICOLON {
					p.nextToken()
				}
				continue
			}

			p.addError(fmt.Sprintf("неожиданный токен в объявлении структуры: %s (строка %d, колонка %d)",
				p.current.Type, p.current.Line, p.current.Column))
			p.nextToken()
		}
	}

	if !p.expect(token.RBRACE) {
		return nil
	}

	return &ast.StructDecl{
		Token:  structToken,
		Name:   name,
		Fields: fields,
	}
}

// parseVarDecl парсит объявление переменной
func (p *Parser) parseVarDecl() ast.Declaration {
	startPos := p.position
	startToken := p.current

	varType := p.parseType()
	if varType.Kind == "unknown" {
		return nil
	}

	if p.current.Type != token.IDENTIFIER {
		p.addError(fmt.Sprintf("ожидалось имя переменной, получен %s (строка %d, колонка %d)",
			p.current.Type, p.current.Line, p.current.Column))
		p.position = startPos
		p.current = startToken
		return nil
	}

	name := &ast.Identifier{
		Token: p.current,
		Value: p.current.Lexeme,
	}
	p.nextToken()

	// Опциональная инициализация
	var initializer ast.Expression = nil
	if p.current.Type == token.ASSIGN {
		p.nextToken()
		initializer = p.parseExpression()
	}

	// Ожидаем точку с запятой
	if p.current.Type != token.SEMICOLON {
		p.addError(fmt.Sprintf("ожидалась ';', получен %s (строка %d, колонка %d)",
			p.current.Type, p.current.Line, p.current.Column))
		for p.current.Type != token.SEMICOLON && p.current.Type != token.RBRACE && !p.isAtEnd() {
			p.nextToken()
		}
		if p.current.Type == token.SEMICOLON {
			p.nextToken()
		}
		return nil
	}
	p.nextToken() // потребляем ';'

	return &ast.VarDecl{
		Token:       varType.Token,
		Type:        varType,
		Name:        name,
		Initializer: initializer,
	}
}

// parseStatement парсит инструкцию
func (p *Parser) parseStatement() ast.Statement {
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
	case token.KW_INT, token.KW_FLOAT, token.KW_BOOL:
		// Объявление переменной с базовым типом
		decl := p.parseVarDecl()
		if decl != nil {
			return decl.(*ast.VarDecl)
		}
		return nil
	case token.IDENTIFIER:

		// Проверяем, является ли следующий токен идентификатором
		next := p.peek()
		if next.Type == token.IDENTIFIER {
			decl := p.parseVarDecl()
			if decl != nil {
				return decl.(*ast.VarDecl)
			}
			return nil
		}
		// Иначе это выражение
		return p.parseExprStmt()
	case token.SEMICOLON:
		// Пустая инструкция
		p.nextToken()
		return nil
	default:
		return p.parseExprStmt()
	}
}

// parseBlockStmt парсит блок инструкций { ... }
func (p *Parser) parseBlockStmt() *ast.BlockStmt {
	// Проверяем, что текущий токен - '{'
	if p.current.Type != token.LBRACE {
		p.addError(fmt.Sprintf("ожидался токен LBRACE, получен %s (строка %d, колонка %d)",
			p.current.Type, p.current.Line, p.current.Column))
		return nil
	}

	block := &ast.BlockStmt{
		Token:      p.current,
		Statements: []ast.Statement{},
	}
	p.nextToken() // потребляем '{'

	for p.current.Type != token.RBRACE && !p.isAtEnd() {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		} else {
			// Если не удалось распарсить инструкцию, пропускаем токен
			p.nextToken()
		}
	}

	if !p.expect(token.RBRACE) {
		return nil
	}

	return block
}

// parseIfStmt парсит if-инструкцию
func (p *Parser) parseIfStmt() *ast.IfStmt {
	ifStmt := &ast.IfStmt{
		Token: p.current,
	}
	p.nextToken() // потребляем 'if'

	if !p.expect(token.LPAREN) {
		return nil
	}
	ifStmt.Condition = p.parseExpression()
	if !p.expect(token.RPAREN) {
		return nil
	}

	// Проверяем, является ли следующая инструкция блоком
	if p.current.Type == token.LBRACE {
		ifStmt.Consequence = p.parseBlockStmt()
	} else {
		// Оборачиваем одиночную инструкцию в блок
		stmt := p.parseStatement()
		if stmt != nil {
			block := &ast.BlockStmt{
				Token:      token.Token{Type: token.LBRACE, Lexeme: "{", Line: stmt.Line(), Column: stmt.Column()},
				Statements: []ast.Statement{stmt},
			}
			ifStmt.Consequence = block
		}
	}

	// Обработка else (опционально)
	if p.current.Type == token.KW_ELSE {
		p.nextToken() // потребляем 'else'
		if p.current.Type == token.LBRACE {
			ifStmt.Alternative = p.parseBlockStmt()
		} else if p.current.Type == token.KW_IF {
			// else if обрабатывается как if внутри else
			ifStmt.Alternative = p.parseIfStmt()
		} else {
			stmt := p.parseStatement()
			if stmt != nil {
				block := &ast.BlockStmt{
					Token:      token.Token{Type: token.LBRACE, Lexeme: "{", Line: stmt.Line(), Column: stmt.Column()},
					Statements: []ast.Statement{stmt},
				}
				ifStmt.Alternative = block
			}
		}
	}

	return ifStmt
}

// parseWhileStmt парсит while-инструкцию
func (p *Parser) parseWhileStmt() *ast.WhileStmt {
	whileStmt := &ast.WhileStmt{
		Token: p.current,
	}
	p.nextToken() // потребляем 'while'

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
			whileStmt.Body = &ast.BlockStmt{
				Token:      token.Token{Type: token.LBRACE, Lexeme: "{", Line: stmt.Line(), Column: stmt.Column()},
				Statements: []ast.Statement{stmt},
			}
		}
	}

	return whileStmt
}

// parseForStmt парсит for-инструкцию
func (p *Parser) parseForStmt() *ast.ForStmt {
	forStmt := &ast.ForStmt{
		Token: p.current,
	}
	p.nextToken() // потребляем 'for'

	if !p.expect(token.LPAREN) {
		return nil
	}

	// Инициализация (может быть пустой)
	if p.current.Type != token.SEMICOLON {
		forStmt.Init = p.parseStatement()
	} else {
		// Пустая инициализация
		forStmt.Init = nil
		p.nextToken() // потребляем ';'
	}

	// Условие (может быть пустым)
	if p.current.Type != token.SEMICOLON {
		forStmt.Condition = p.parseExpression()
	}
	if !p.expect(token.SEMICOLON) {
		return nil
	}

	// Обновление (может быть пустым)
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
			forStmt.Body = &ast.BlockStmt{
				Token:      token.Token{Type: token.LBRACE, Lexeme: "{", Line: stmt.Line(), Column: stmt.Column()},
				Statements: []ast.Statement{stmt},
			}
		}
	}

	return forStmt
}

// parseReturnStmt парсит return-инструкцию
func (p *Parser) parseReturnStmt() *ast.ReturnStmt {
	returnStmt := &ast.ReturnStmt{
		Token: p.current,
	}
	p.nextToken() // потребляем 'return'

	if p.current.Type != token.SEMICOLON {
		returnStmt.RetValue = p.parseExpression()
	}

	if !p.expect(token.SEMICOLON) {
		return nil
	}

	return returnStmt
}

// parseExprStmt парсит выражение как инструкцию
func (p *Parser) parseExprStmt() *ast.ExprStmt {
	exprStmt := &ast.ExprStmt{
		Token:      p.current,
		Expression: p.parseExpression(),
	}

	if !p.expect(token.SEMICOLON) {
		return nil
	}

	return exprStmt
}

// parseExpression парсит выражение (начинаем с самого низкого приоритета)
func (p *Parser) parseExpression() ast.Expression {
	return p.parseAssignment()
}

// parseAssignment парсит присваивание (право-ассоциативное)
func (p *Parser) parseAssignment() ast.Expression {
	expr := p.parseLogicalOr()

	if p.current.Type == token.ASSIGN ||
		p.current.Type == token.PLUS_ASSIGN ||
		p.current.Type == token.MINUS_ASSIGN ||
		p.current.Type == token.MULTIPLY_ASSIGN ||
		p.current.Type == token.DIVIDE_ASSIGN {

		tok := p.current
		operator := tok.Lexeme
		p.nextToken()
		right := p.parseAssignment() // право-ассоциативно

		// Для составных операторов создаем бинарное выражение
		if operator != "=" {
			// a += b  превращается в a = a + b
			left := expr
			return &ast.AssignmentExpr{
				Token:    tok,
				Left:     left,
				Operator: "=",
				Right: &ast.BinaryExpr{
					Token:    tok,
					Left:     left,
					Operator: operator[:len(operator)-1], // убираем '='
					Right:    right,
				},
			}
		}

		return &ast.AssignmentExpr{
			Token:    tok,
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}

	return expr
}

// parseLogicalOr парсит логическое ИЛИ (||)
func (p *Parser) parseLogicalOr() ast.Expression {
	expr := p.parseLogicalAnd()

	for p.current.Type == token.OR {
		tok := p.current
		p.nextToken()
		right := p.parseLogicalAnd()
		expr = &ast.BinaryExpr{
			Token:    tok,
			Left:     expr,
			Operator: tok.Lexeme,
			Right:    right,
		}
	}

	return expr
}

// parseLogicalAnd парсит логическое И (&&)
func (p *Parser) parseLogicalAnd() ast.Expression {
	expr := p.parseEquality()

	for p.current.Type == token.AND {
		tok := p.current
		p.nextToken()
		right := p.parseEquality()
		expr = &ast.BinaryExpr{
			Token:    tok,
			Left:     expr,
			Operator: tok.Lexeme,
			Right:    right,
		}
	}

	return expr
}

// parseEquality парсит сравнение на равенство (==, !=)
func (p *Parser) parseEquality() ast.Expression {
	expr := p.parseRelational()

	for p.current.Type == token.EQ || p.current.Type == token.NOT_EQ {
		tok := p.current
		p.nextToken()
		right := p.parseRelational()
		expr = &ast.BinaryExpr{
			Token:    tok,
			Left:     expr,
			Operator: tok.Lexeme,
			Right:    right,
		}
	}

	return expr
}

// parseRelational парсит операторы сравнения (<, <=, >, >=)
func (p *Parser) parseRelational() ast.Expression {
	expr := p.parseAdditive()

	for p.current.Type == token.LT || p.current.Type == token.LT_EQ ||
		p.current.Type == token.GT || p.current.Type == token.GT_EQ {
		tok := p.current
		p.nextToken()
		right := p.parseAdditive()
		expr = &ast.BinaryExpr{
			Token:    tok,
			Left:     expr,
			Operator: tok.Lexeme,
			Right:    right,
		}
	}

	return expr
}

// parseAdditive парсит сложение и вычитание (+, -)
func (p *Parser) parseAdditive() ast.Expression {
	expr := p.parseMultiplicative()

	for p.current.Type == token.PLUS || p.current.Type == token.MINUS {
		tok := p.current
		p.nextToken()
		right := p.parseMultiplicative()
		expr = &ast.BinaryExpr{
			Token:    tok,
			Left:     expr,
			Operator: tok.Lexeme,
			Right:    right,
		}
	}

	return expr
}

// parseMultiplicative парсит умножение, деление и остаток (*, /, %)
func (p *Parser) parseMultiplicative() ast.Expression {
	expr := p.parseUnary()

	for p.current.Type == token.MULTIPLY || p.current.Type == token.DIVIDE ||
		p.current.Type == token.MODULO {
		tok := p.current
		p.nextToken()
		right := p.parseUnary()
		expr = &ast.BinaryExpr{
			Token:    tok,
			Left:     expr,
			Operator: tok.Lexeme,
			Right:    right,
		}
	}

	return expr
}

// parseUnary парсит унарные операторы (-, !)
func (p *Parser) parseUnary() ast.Expression {
	// Унарный минус
	if p.current.Type == token.MINUS {
		tok := p.current
		operator := tok.Lexeme
		p.nextToken()
		right := p.parseUnary()
		return &ast.UnaryExpr{
			Token:    tok,
			Operator: operator,
			Right:    right,
		}
	}

	// Логическое отрицание
	if p.current.Type == token.NOT_EQ {
		tok := p.current
		operator := tok.Lexeme
		p.nextToken()
		right := p.parseUnary()
		return &ast.UnaryExpr{
			Token:    tok,
			Operator: operator,
			Right:    right,
		}
	}

	return p.parseCall()
}

// parseCall парсит вызов функции и доступ к полям структур
func (p *Parser) parseCall() ast.Expression {
	expr := p.parsePrimary()

	for p.current.Type == token.LPAREN || p.current.Type == token.DOT {
		if p.current.Type == token.LPAREN {
			// Вызов функции
			tok := p.current
			p.nextToken() // потребляем '('
			args := []ast.Expression{}

			// Парсим аргументы, если они есть
			if p.current.Type != token.RPAREN {
				args = append(args, p.parseExpression())
				for p.current.Type == token.COMMA {
					p.nextToken() // потребляем ','
					args = append(args, p.parseExpression())
				}
			}

			if !p.expect(token.RPAREN) {
				return expr
			}

			expr = &ast.CallExpr{
				Token:     tok,
				Function:  expr,
				Arguments: args,
			}
		} else if p.current.Type == token.DOT {
			// Доступ к полю структуры
			p.nextToken() // потребляем '.'
			if p.current.Type == token.IDENTIFIER {
				// Создаем выражение доступа к полю
				// Пока просто возвращаем идентификатор, в будущем будет MemberAccess
				field := &ast.Identifier{
					Token: p.current,
					Value: p.current.Lexeme,
				}
				p.nextToken()
				expr = field
			}
		}
	}

	return expr
}

// parsePrimary парсит первичные выражения
func (p *Parser) parsePrimary() ast.Expression {
	switch p.current.Type {
	case token.IDENTIFIER:
		ident := &ast.Identifier{
			Token: p.current,
			Value: p.current.Lexeme,
		}
		p.nextToken()
		return ident
	case token.INT_LITERAL:
		lit := &ast.LiteralExpr{
			Token:    p.current,
			Type:     "int",
			IntValue: p.current.Literal.IntValue,
		}
		p.nextToken()
		return lit
	case token.FLOAT_LITERAL:
		lit := &ast.LiteralExpr{
			Token:      p.current,
			Type:       "float",
			FloatValue: p.current.Literal.FloatValue,
		}
		p.nextToken()
		return lit
	case token.STRING_LITERAL:
		lit := &ast.LiteralExpr{
			Token:       p.current,
			Type:        "string",
			StringValue: p.current.Literal.StringValue,
		}
		p.nextToken()
		return lit
	case token.KW_TRUE:
		lit := &ast.LiteralExpr{
			Token:     p.current,
			Type:      "bool",
			BoolValue: true,
		}
		p.nextToken()
		return lit
	case token.KW_FALSE:
		lit := &ast.LiteralExpr{
			Token:     p.current,
			Type:      "bool",
			BoolValue: false,
		}
		p.nextToken()
		return lit
	case token.LPAREN:
		p.nextToken() // потребляем '('
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

// parseType парсит тип
func (p *Parser) parseType() ast.Type {
	switch p.current.Type {
	case token.KW_INT:
		t := ast.Type{Token: p.current, Kind: "int"}
		p.nextToken()
		return t
	case token.KW_FLOAT:
		t := ast.Type{Token: p.current, Kind: "float"}
		p.nextToken()
		return t
	case token.KW_BOOL:
		t := ast.Type{Token: p.current, Kind: "bool"}
		p.nextToken()
		return t
	case token.KW_VOID:
		t := ast.Type{Token: p.current, Kind: "void"}
		p.nextToken()
		return t
	case token.IDENTIFIER:
		// Пользовательский тип (struct)
		t := ast.Type{Token: p.current, Kind: "identifier", Name: p.current.Lexeme}
		p.nextToken()
		return t
	default:
		p.addError(fmt.Sprintf("ожидался тип, получен %s (строка %d, колонка %d)",
			p.current.Type, p.current.Line, p.current.Column))
		return ast.Type{Kind: "unknown"}
	}
}
