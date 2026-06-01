package lexer

import (
	"fmt"
	"strconv"
	"strings"

	"mikrocompiler/src/internal/token"
)

type Scanner struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int

	peekedToken *token.Token
	hasPeeked   bool
}

func NewScanner(input string) *Scanner {
	s := &Scanner{
		input:    input,
		line:     1,
		column:   0,
		position: 0,
	}
	s.readChar()
	return s
}

func (s *Scanner) readChar() {
	if s.readPosition >= len(s.input) {
		s.ch = 0
	} else {
		s.ch = s.input[s.readPosition]
	}

	s.position = s.readPosition
	s.readPosition++

	if s.ch == '\n' {
		s.line++
		s.column = 0
	} else if s.ch == '\r' {
		if s.readPosition < len(s.input) && s.input[s.readPosition] == '\n' {
			s.readPosition++
		}
		s.line++
		s.column = 0
	} else if s.ch != 0 {
		s.column++
	}
}

func (s *Scanner) peekChar() byte {
	if s.readPosition >= len(s.input) {
		return 0
	}
	return s.input[s.readPosition]
}

func (s *Scanner) IsAtEnd() bool {
	return s.ch == 0 && s.position >= len(s.input)
}

func (s *Scanner) GetLine() int {
	return s.line
}

func (s *Scanner) GetColumn() int {
	return s.column
}

func (s *Scanner) PeekToken() token.Token {
	if s.hasPeeked {
		return *s.peekedToken
	}

	savedPos := s.position
	savedReadPos := s.readPosition
	savedCh := s.ch
	savedLine := s.line
	savedColumn := s.column

	tok := s.NextToken()

	s.position = savedPos
	s.readPosition = savedReadPos
	s.ch = savedCh
	s.line = savedLine
	s.column = savedColumn

	s.peekedToken = &tok
	s.hasPeeked = true

	return tok
}

func (s *Scanner) NextToken() token.Token {
	if s.hasPeeked {
		s.hasPeeked = false
		return *s.peekedToken
	}

	s.skipWhitespace()

	line, column := s.line, s.column

	for s.ch == '/' && (s.peekChar() == '/' || s.peekChar() == '*') {
		if s.peekChar() == '/' {
			s.skipSingleLineComment()
		} else {
			if s.skipMultiLineComment() {
				return s.newIllegalToken("незакрытый многострочный комментарий", line, column)
			}
		}
		s.skipWhitespace()
		line, column = s.line, s.column
	}

	if s.ch == 0 {
		return token.NewToken(token.EOF, "", s.line, 1)
	}

	line, column = s.line, s.column

	switch s.ch {
	case '-':
		if s.peekChar() == '>' {
			s.readChar()
			tok := token.NewToken(token.ARROW, "->", line, column)
			s.readChar()
			return tok
		}
		if s.peekChar() == '=' {
			s.readChar()
			tok := token.NewToken(token.MINUS_ASSIGN, "-=", line, column)
			s.readChar()
			return tok
		}
		if isDigit(s.peekChar()) {
			return s.readNumber(line, column)
		}
		tok := token.NewToken(token.MINUS, "-", line, column)
		s.readChar()
		return tok

	case '=':
		if s.peekChar() == '=' {
			s.readChar()
			tok := token.NewToken(token.EQ, "==", line, column)
			s.readChar()
			return tok
		}
		tok := token.NewToken(token.ASSIGN, "=", line, column)
		s.readChar()
		return tok

	case '+':
		if s.peekChar() == '=' {
			s.readChar()
			tok := token.NewToken(token.PLUS_ASSIGN, "+=", line, column)
			s.readChar()
			return tok
		}
		tok := token.NewToken(token.PLUS, "+", line, column)
		s.readChar()
		return tok

	case '*':
		if s.peekChar() == '=' {
			s.readChar()
			tok := token.NewToken(token.MULTIPLY_ASSIGN, "*=", line, column)
			s.readChar()
			return tok
		}
		tok := token.NewToken(token.MULTIPLY, "*", line, column)
		s.readChar()
		return tok

	case '/':
		if s.peekChar() == '=' {
			s.readChar()
			tok := token.NewToken(token.DIVIDE_ASSIGN, "/=", line, column)
			s.readChar()
			return tok
		}
		tok := token.NewToken(token.DIVIDE, "/", line, column)
		s.readChar()
		return tok

	case '%':
		tok := token.NewToken(token.MODULO, "%", line, column)
		s.readChar()
		return tok

	case '!':
		if s.peekChar() == '=' {
			s.readChar()
			tok := token.NewToken(token.NOT_EQ, "!=", line, column)
			s.readChar()
			return tok
		}
		tok := token.NewToken(token.NOT, "!", line, column)
		s.readChar()
		return tok

	case '<':
		if s.peekChar() == '=' {
			s.readChar()
			tok := token.NewToken(token.LT_EQ, "<=", line, column)
			s.readChar()
			return tok
		}
		tok := token.NewToken(token.LT, "<", line, column)
		s.readChar()
		return tok

	case '>':
		if s.peekChar() == '=' {
			s.readChar()
			tok := token.NewToken(token.GT_EQ, ">=", line, column)
			s.readChar()
			return tok
		}
		tok := token.NewToken(token.GT, ">", line, column)
		s.readChar()
		return tok

	case '&':
		if s.peekChar() == '&' {
			s.readChar()
			tok := token.NewToken(token.AND, "&&", line, column)
			s.readChar()
			return tok
		}
		tok := s.newIllegalToken(fmt.Sprintf("недопустимый символ '%c'", s.ch), line, column)
		s.readChar()
		return tok

	case '|':
		if s.peekChar() == '|' {
			s.readChar()
			tok := token.NewToken(token.OR, "||", line, column)
			s.readChar()
			return tok
		}
		tok := s.newIllegalToken(fmt.Sprintf("недопустимый символ '%c'", s.ch), line, column)
		s.readChar()
		return tok

	case '(':
		tok := token.NewToken(token.LPAREN, "(", line, column)
		s.readChar()
		return tok
	case ')':
		tok := token.NewToken(token.RPAREN, ")", line, column)
		s.readChar()
		return tok
	case '{':
		tok := token.NewToken(token.LBRACE, "{", line, column)
		s.readChar()
		return tok
	case '}':
		tok := token.NewToken(token.RBRACE, "}", line, column)
		s.readChar()
		return tok
	case '[':
		tok := token.NewToken(token.LBRACKET, "[", line, column)
		s.readChar()
		return tok
	case ']':
		tok := token.NewToken(token.RBRACKET, "]", line, column)
		s.readChar()
		return tok
	case ';':
		tok := token.NewToken(token.SEMICOLON, ";", line, column)
		s.readChar()
		return tok
	case ',':
		tok := token.NewToken(token.COMMA, ",", line, column)
		s.readChar()
		return tok
	case '.':
		// Проверяем на variadic (...)
		if s.peekChar() == '.' && s.readPosition+1 < len(s.input) && s.input[s.readPosition+1] == '.' {
			s.readChar()
			s.readChar()
			tok := token.NewToken(token.ELLIPSIS, "...", line, column)
			s.readChar()
			return tok
		}
		tok := token.NewToken(token.DOT, ".", line, column)
		s.readChar()
		return tok

	case '"':
		return s.readString(line, column)

	default:
		if isLetter(s.ch) || s.ch == '_' {
			return s.readIdentifier(line, column)
		}
		if isDigit(s.ch) {
			return s.readNumber(line, column)
		}
		tok := s.newIllegalToken(fmt.Sprintf("недопустимый символ '%c'", s.ch), line, column)
		s.readChar()
		for s.ch != '\n' && s.ch != '\r' && s.ch != 0 {
			s.readChar()
		}
		return tok
	}
}

func (s *Scanner) readIdentifier(line, column int) token.Token {
	start := s.position
	for isLetter(s.ch) || isDigit(s.ch) || s.ch == '_' {
		s.readChar()
	}
	lexeme := s.input[start:s.position]

	if len(lexeme) > 255 {
		return s.newIllegalToken("идентификатор слишком длинный (максимум 255 символов)", line, column)
	}

	tokenType := token.LookupIdentifier(lexeme)

	if tokenType == token.KW_TRUE {
		return token.NewLiteralToken(token.KW_TRUE, lexeme, line, column, &token.LiteralValue{BoolValue: true})
	}
	if tokenType == token.KW_FALSE {
		return token.NewLiteralToken(token.KW_FALSE, lexeme, line, column, &token.LiteralValue{BoolValue: false})
	}

	return token.NewToken(tokenType, lexeme, line, column)
}

func (s *Scanner) readNumber(line, column int) token.Token {
	start := s.position
	isFloat := false
	isNegative := false

	if s.ch == '-' {
		isNegative = true
		s.readChar()
	}

	for isDigit(s.ch) {
		s.readChar()
	}

	if s.ch == '.' && isDigit(s.peekChar()) {
		isFloat = true
		s.readChar()
		for isDigit(s.ch) {
			s.readChar()
		}
	}

	end := s.position
	lexeme := s.input[start:end]

	if isFloat {
		floatVal, err := strconv.ParseFloat(lexeme, 64)
		if err != nil {
			return s.newIllegalToken("недопустимый литерал с плавающей точкой", line, column)
		}
		return token.NewLiteralToken(token.FLOAT_LITERAL, lexeme, line, column,
			&token.LiteralValue{FloatValue: floatVal})
	}

	intVal, err := strconv.ParseInt(lexeme, 10, 64)
	if err != nil {
		return s.newIllegalToken("целочисленный литерал вне диапазона [-2^31, 2^31-1]", line, column)
	}

	if intVal < -2147483648 || intVal > 2147483647 {
		return s.newIllegalToken("целочисленный литерал вне диапазона [-2^31, 2^31-1]", line, column)
	}

	if isNegative {
		return token.NewLiteralToken(token.INT_LITERAL, lexeme, line, column,
			&token.LiteralValue{IntValue: int32(intVal)})
	}

	return token.NewLiteralToken(token.INT_LITERAL, lexeme, line, column,
		&token.LiteralValue{IntValue: int32(intVal)})
}

func (s *Scanner) readString(line, column int) token.Token {
	s.readChar()
	var content strings.Builder

	for {
		if s.ch == '"' {
			break
		}
		if s.ch == 0 || s.ch == '\n' || s.ch == '\r' {
			return s.newIllegalToken("незакрытый строковый литерал", line, column)
		}

		if s.ch == '\\' {
			s.readChar()
			switch s.ch {
			case 'n':
				content.WriteByte('\n')
			case 't':
				content.WriteByte('\t')
			case 'r':
				content.WriteByte('\r')
			case '"':
				content.WriteByte('"')
			case '\\':
				content.WriteByte('\\')
			case '0':
				content.WriteByte(0)
			default:
				content.WriteByte('\\')
				content.WriteByte(s.ch)
			}
			s.readChar()
			continue
		}

		content.WriteByte(s.ch)
		s.readChar()
	}

	if s.ch == 0 {
		return s.newIllegalToken("незакрытый строковый литерал", line, column)
	}

	strContent := content.String()
	s.readChar()

	displayLexeme := "\"" + strContent + "\""

	return token.NewLiteralToken(token.STRING_LITERAL, displayLexeme, line, column,
		&token.LiteralValue{StringValue: strContent})
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.readChar()
	}
}

func (s *Scanner) skipSingleLineComment() {
	s.readChar()
	s.readChar()

	for s.ch != '\n' && s.ch != '\r' && s.ch != 0 {
		s.readChar()
	}
}

func (s *Scanner) skipMultiLineComment() bool {
	s.readChar()
	s.readChar()

	nesting := 1
	for nesting > 0 && s.ch != 0 {
		if s.ch == '/' && s.peekChar() == '*' {
			nesting++
			s.readChar()
			s.readChar()
		} else if s.ch == '*' && s.peekChar() == '/' {
			nesting--
			s.readChar()
			s.readChar()
		} else {
			s.readChar()
		}
	}

	return nesting > 0
}

func (s *Scanner) newIllegalToken(message string, line, column int) token.Token {
	return token.NewToken(token.ILLEGAL, fmt.Sprintf("Ошибка в %d:%d: %s", line, column, message), line, column)
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
