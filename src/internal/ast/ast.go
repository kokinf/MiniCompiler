package ast

import (
	"mikrocompiler/src/internal/token"
)

type Node interface {
	TokenLiteral() string
	String() string
	Line() int
	Column() int
}

type Program struct {
	Declarations []Declaration
}

func (p *Program) TokenLiteral() string {
	if len(p.Declarations) > 0 {
		return p.Declarations[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out string
	for _, decl := range p.Declarations {
		out += decl.String() + "\n"
	}
	return out
}

func (p *Program) Line() int {
	if len(p.Declarations) > 0 {
		return p.Declarations[0].Line()
	}
	return 0
}

func (p *Program) Column() int {
	if len(p.Declarations) > 0 {
		return p.Declarations[0].Column()
	}
	return 0
}

type Declaration interface {
	Node
	declarationNode()
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type FunctionDecl struct {
	Token      token.Token
	Name       *Identifier
	Parameters []*Parameter
	ReturnType Type
	Body       *BlockStmt
}

func (fd *FunctionDecl) declarationNode()     {}
func (fd *FunctionDecl) TokenLiteral() string { return fd.Token.Lexeme }
func (fd *FunctionDecl) String() string       { return "FunctionDecl" }
func (fd *FunctionDecl) Line() int            { return fd.Token.Line }
func (fd *FunctionDecl) Column() int          { return fd.Token.Column }

type StructDecl struct {
	Token  token.Token
	Name   *Identifier
	Fields []*VarDecl
}

func (sd *StructDecl) declarationNode()     {}
func (sd *StructDecl) TokenLiteral() string { return sd.Token.Lexeme }
func (sd *StructDecl) String() string       { return "StructDecl" }
func (sd *StructDecl) Line() int            { return sd.Token.Line }
func (sd *StructDecl) Column() int          { return sd.Token.Column }

type VarDecl struct {
	Token       token.Token
	Type        Type
	Name        *Identifier
	Initializer Expression
}

func (vd *VarDecl) declarationNode()     {}
func (vd *VarDecl) statementNode()       {}
func (vd *VarDecl) TokenLiteral() string { return vd.Token.Lexeme }
func (vd *VarDecl) String() string       { return "VarDecl" }
func (vd *VarDecl) Line() int            { return vd.Token.Line }
func (vd *VarDecl) Column() int          { return vd.Token.Column }

type BlockStmt struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BlockStmt) statementNode()       {}
func (bs *BlockStmt) TokenLiteral() string { return bs.Token.Lexeme }
func (bs *BlockStmt) String() string       { return "BlockStmt" }
func (bs *BlockStmt) Line() int            { return bs.Token.Line }
func (bs *BlockStmt) Column() int          { return bs.Token.Column }

type IfStmt struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStmt
	Alternative Statement
}

func (is *IfStmt) statementNode()       {}
func (is *IfStmt) TokenLiteral() string { return is.Token.Lexeme }
func (is *IfStmt) String() string       { return "IfStmt" }
func (is *IfStmt) Line() int            { return is.Token.Line }
func (is *IfStmt) Column() int          { return is.Token.Column }

type WhileStmt struct {
	Token     token.Token
	Condition Expression
	Body      *BlockStmt
}

func (ws *WhileStmt) statementNode()       {}
func (ws *WhileStmt) TokenLiteral() string { return ws.Token.Lexeme }
func (ws *WhileStmt) String() string       { return "WhileStmt" }
func (ws *WhileStmt) Line() int            { return ws.Token.Line }
func (ws *WhileStmt) Column() int          { return ws.Token.Column }

type ForStmt struct {
	Token     token.Token
	Init      Statement
	Condition Expression
	Update    Expression
	Body      *BlockStmt
}

func (fs *ForStmt) statementNode()       {}
func (fs *ForStmt) TokenLiteral() string { return fs.Token.Lexeme }
func (fs *ForStmt) String() string       { return "ForStmt" }
func (fs *ForStmt) Line() int            { return fs.Token.Line }
func (fs *ForStmt) Column() int          { return fs.Token.Column }

type ReturnStmt struct {
	Token    token.Token
	RetValue Expression
}

func (rs *ReturnStmt) statementNode()       {}
func (rs *ReturnStmt) TokenLiteral() string { return rs.Token.Lexeme }
func (rs *ReturnStmt) String() string       { return "ReturnStmt" }
func (rs *ReturnStmt) Line() int            { return rs.Token.Line }
func (rs *ReturnStmt) Column() int          { return rs.Token.Column }

type ExprStmt struct {
	Token      token.Token
	Expression Expression
}

func (es *ExprStmt) statementNode()       {}
func (es *ExprStmt) TokenLiteral() string { return es.Token.Lexeme }
func (es *ExprStmt) String() string       { return "ExprStmt" }
func (es *ExprStmt) Line() int            { return es.Token.Line }
func (es *ExprStmt) Column() int          { return es.Token.Column }

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Lexeme }
func (i *Identifier) String() string       { return i.Value }
func (i *Identifier) Line() int            { return i.Token.Line }
func (i *Identifier) Column() int          { return i.Token.Column }

type LiteralExpr struct {
	Token       token.Token
	Type        string
	IntValue    int32
	FloatValue  float64
	StringValue string
	BoolValue   bool
}

func (le *LiteralExpr) expressionNode()      {}
func (le *LiteralExpr) TokenLiteral() string { return le.Token.Lexeme }
func (le *LiteralExpr) String() string       { return le.Token.Lexeme }
func (le *LiteralExpr) Line() int            { return le.Token.Line }
func (le *LiteralExpr) Column() int          { return le.Token.Column }

type BinaryExpr struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (be *BinaryExpr) expressionNode()      {}
func (be *BinaryExpr) TokenLiteral() string { return be.Token.Lexeme }
func (be *BinaryExpr) String() string       { return "BinaryExpr" }
func (be *BinaryExpr) Line() int            { return be.Token.Line }
func (be *BinaryExpr) Column() int          { return be.Token.Column }

type UnaryExpr struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (ue *UnaryExpr) expressionNode()      {}
func (ue *UnaryExpr) TokenLiteral() string { return ue.Token.Lexeme }
func (ue *UnaryExpr) String() string       { return "UnaryExpr" }
func (ue *UnaryExpr) Line() int            { return ue.Token.Line }
func (ue *UnaryExpr) Column() int          { return ue.Token.Column }

type CallExpr struct {
	Token     token.Token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpr) expressionNode()      {}
func (ce *CallExpr) TokenLiteral() string { return ce.Token.Lexeme }
func (ce *CallExpr) String() string       { return "CallExpr" }
func (ce *CallExpr) Line() int            { return ce.Token.Line }
func (ce *CallExpr) Column() int          { return ce.Token.Column }

type AssignmentExpr struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ae *AssignmentExpr) expressionNode()      {}
func (ae *AssignmentExpr) TokenLiteral() string { return ae.Token.Lexeme }
func (ae *AssignmentExpr) String() string       { return "AssignmentExpr" }
func (ae *AssignmentExpr) Line() int            { return ae.Token.Line }
func (ae *AssignmentExpr) Column() int          { return ae.Token.Column }

type Type struct {
	Token token.Token
	Kind  string
	Name  string
}

func (t *Type) String() string {
	if t.Kind == "identifier" {
		return t.Name
	}
	return t.Kind
}

type Parameter struct {
	Token token.Token
	Type  Type
	Name  *Identifier
}

func (p *Parameter) String() string {
	return p.Type.String() + " " + p.Name.String()
}
