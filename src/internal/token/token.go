package token

import (
	"fmt"
	"strings"
)

type TokenType string

const (
	ILLEGAL    TokenType = "ILLEGAL"
	EOF        TokenType = "END_OF_FILE"
	COMMENT    TokenType = "COMMENT"
	WHITESPACE TokenType = "WHITESPACE"

	KW_IF     TokenType = "KW_IF"
	KW_ELSE   TokenType = "KW_ELSE"
	KW_WHILE  TokenType = "KW_WHILE"
	KW_FOR    TokenType = "KW_FOR"
	KW_INT    TokenType = "KW_INT"
	KW_FLOAT  TokenType = "KW_FLOAT"
	KW_BOOL   TokenType = "KW_BOOL"
	KW_RETURN TokenType = "KW_RETURN"
	KW_TRUE   TokenType = "KW_TRUE"
	KW_FALSE  TokenType = "KW_FALSE"
	KW_VOID   TokenType = "KW_VOID"
	KW_STRUCT TokenType = "KW_STRUCT"
	KW_FN     TokenType = "KW_FN"

	IDENTIFIER     TokenType = "IDENTIFIER"
	INT_LITERAL    TokenType = "INT_LITERAL"
	FLOAT_LITERAL  TokenType = "FLOAT_LITERAL"
	STRING_LITERAL TokenType = "STRING_LITERAL"

	ASSIGN   TokenType = "ASSIGN"   // =
	PLUS     TokenType = "PLUS"     // +
	MINUS    TokenType = "MINUS"    // -
	MULTIPLY TokenType = "MULTIPLY" // *
	DIVIDE   TokenType = "DIVIDE"   // /
	MODULO   TokenType = "MODULO"   // %

	PLUS_ASSIGN     TokenType = "PLUS_ASSIGN"     // +=
	MINUS_ASSIGN    TokenType = "MINUS_ASSIGN"    // -=
	MULTIPLY_ASSIGN TokenType = "MULTIPLY_ASSIGN" // *=
	DIVIDE_ASSIGN   TokenType = "DIVIDE_ASSIGN"   // /=

	EQ     TokenType = "EQ"     // ==
	NOT_EQ TokenType = "NOT_EQ" // !=
	LT     TokenType = "LT"     // <
	LT_EQ  TokenType = "LT_EQ"  // <=
	GT     TokenType = "GT"     // >
	GT_EQ  TokenType = "GT_EQ"  // >=

	AND TokenType = "AND" // &&
	OR  TokenType = "OR"  // ||

	LPAREN    TokenType = "LPAREN"    // (
	RPAREN    TokenType = "RPAREN"    // )
	LBRACE    TokenType = "LBRACE"    // {
	RBRACE    TokenType = "RBRACE"    // }
	LBRACKET  TokenType = "LBRACKET"  // [
	RBRACKET  TokenType = "RBRACKET"  // ]
	SEMICOLON TokenType = "SEMICOLON" // ;
	COMMA     TokenType = "COMMA"     // ,
	DOT       TokenType = "DOT"       // .
	ARROW     TokenType = "ARROW"     // ->
)

type LiteralValue struct {
	IntValue    int32
	FloatValue  float64
	StringValue string
	BoolValue   bool
}

type Token struct {
	Type    TokenType
	Lexeme  string
	Line    int
	Column  int
	Literal *LiteralValue
}

func NewToken(tokenType TokenType, lexeme string, line, column int) Token {
	return Token{
		Type:    tokenType,
		Lexeme:  lexeme,
		Line:    line,
		Column:  column,
		Literal: nil,
	}
}

func NewLiteralToken(tokenType TokenType, lexeme string, line, column int, literal *LiteralValue) Token {
	return Token{
		Type:    tokenType,
		Lexeme:  lexeme,
		Line:    line,
		Column:  column,
		Literal: literal,
	}
}

func (t Token) String() string {
	if t.Literal != nil {
		switch t.Type {
		case INT_LITERAL:
			return fmt.Sprintf("%d:%d %s \"%s\" %d", t.Line, t.Column, t.Type, t.Lexeme, t.Literal.IntValue)
		case FLOAT_LITERAL:
			return fmt.Sprintf("%d:%d %s \"%s\" %g", t.Line, t.Column, t.Type, t.Lexeme, t.Literal.FloatValue)
		case STRING_LITERAL:
			escapedLexeme := strings.ReplaceAll(t.Lexeme, `"`, `\"`)
			if t.Literal.StringValue == "" {
				return fmt.Sprintf("%d:%d %s \"%s\"", t.Line, t.Column, t.Type, escapedLexeme)
			}
			return fmt.Sprintf("%d:%d %s \"%s\" %s", t.Line, t.Column, t.Type, escapedLexeme, t.Literal.StringValue)
		case KW_TRUE, KW_FALSE:
			return fmt.Sprintf("%d:%d %s \"%s\" %t", t.Line, t.Column, t.Type, t.Lexeme, t.Literal.BoolValue)
		}
	}
	return fmt.Sprintf("%d:%d %s \"%s\"", t.Line, t.Column, t.Type, t.Lexeme)
}

var Keywords = map[string]TokenType{
	"if":     KW_IF,
	"else":   KW_ELSE,
	"while":  KW_WHILE,
	"for":    KW_FOR,
	"int":    KW_INT,
	"float":  KW_FLOAT,
	"bool":   KW_BOOL,
	"return": KW_RETURN,
	"true":   KW_TRUE,
	"false":  KW_FALSE,
	"void":   KW_VOID,
	"struct": KW_STRUCT,
	"fn":     KW_FN,
}

func LookupIdentifier(ident string) TokenType {
	if tokType, ok := Keywords[ident]; ok {
		return tokType
	}
	return IDENTIFIER
}
