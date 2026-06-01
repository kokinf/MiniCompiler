package ast

import (
	"fmt"
	"mikrocompiler/src/internal/token"
)

// Node - базовый интерфейс для всех узлов AST
type Node interface {
	TokenLiteral() string
	String() string
	Line() int
	Column() int
	Accept(Visitor) interface{}
}

// Visitor pattern для обхода AST
type Visitor interface {
	VisitProgram(node *ProgramNode) interface{}
	VisitFunctionDecl(node *FunctionDeclNode) interface{}
	VisitStructDecl(node *StructDeclNode) interface{}
	VisitVarDecl(node *VarDeclNode) interface{}
	VisitBlockStmt(node *BlockStmtNode) interface{}
	VisitIfStmt(node *IfStmtNode) interface{}
	VisitWhileStmt(node *WhileStmtNode) interface{}
	VisitForStmt(node *ForStmtNode) interface{}
	VisitReturnStmt(node *ReturnStmtNode) interface{}
	VisitExprStmt(node *ExprStmtNode) interface{}
	VisitIdentifier(node *IdentifierNode) interface{}
	VisitLiteralExpr(node *LiteralExprNode) interface{}
	VisitBinaryExpr(node *BinaryExprNode) interface{}
	VisitUnaryExpr(node *UnaryExprNode) interface{}
	VisitCallExpr(node *CallExprNode) interface{}
	VisitAssignmentExpr(node *AssignmentExprNode) interface{}
	VisitParameter(node *ParameterNode) interface{}
	// Sprint 7: новые узлы
	VisitExternFuncDecl(node *ExternFuncDeclNode) interface{}
	VisitIndexExpr(node *IndexExprNode) interface{}
	VisitArrayLiteralExpr(node *ArrayLiteralExprNode) interface{}
}

// ProgramNode - корневой узел программы
type ProgramNode struct {
	Declarations []DeclarationNode
	LinePos      int
	ColumnPos    int
}

func (p *ProgramNode) TokenLiteral() string {
	if len(p.Declarations) > 0 {
		return p.Declarations[0].TokenLiteral()
	}
	return ""
}

func (p *ProgramNode) String() string {
	var out string
	for _, decl := range p.Declarations {
		out += decl.String() + "\n"
	}
	return out
}

func (p *ProgramNode) Line() int {
	if len(p.Declarations) > 0 && p.Declarations[0] != nil {
		return p.Declarations[0].Line()
	}
	return p.LinePos
}

func (p *ProgramNode) Column() int {
	if len(p.Declarations) > 0 && p.Declarations[0] != nil {
		return p.Declarations[0].Column()
	}
	return p.ColumnPos
}

func (p *ProgramNode) Accept(v Visitor) interface{} {
	return v.VisitProgram(p)
}

// DeclarationNode интерфейс для объявлений
type DeclarationNode interface {
	Node
	declarationNode()
}

// StatementNode интерфейс для операторов
type StatementNode interface {
	Node
	statementNode()
}

// ExpressionNode интерфейс для выражений
type ExpressionNode interface {
	Node
	expressionNode()
	Type() *TypeAnnotation
	SetType(*TypeAnnotation)
}

// TypeAnnotation аннотация типа для выражений
type TypeAnnotation struct {
	Kind string
	Name string
}

// FunctionDeclNode объявление функции
type FunctionDeclNode struct {
	Token      token.Token
	Name       *IdentifierNode
	Parameters []*ParameterNode
	ReturnType *TypeNode
	Body       *BlockStmtNode
}

func (fd *FunctionDeclNode) declarationNode()     {}
func (fd *FunctionDeclNode) TokenLiteral() string { return fd.Token.Lexeme }
func (fd *FunctionDeclNode) String() string       { return "FunctionDecl" }
func (fd *FunctionDeclNode) Line() int            { return fd.Token.Line }
func (fd *FunctionDeclNode) Column() int          { return fd.Token.Column }
func (fd *FunctionDeclNode) Accept(v Visitor) interface{} {
	return v.VisitFunctionDecl(fd)
}

// ParameterNode параметр функции
type ParameterNode struct {
	Token token.Token
	Type  *TypeNode
	Name  *IdentifierNode
}

func (p *ParameterNode) TokenLiteral() string { return p.Token.Lexeme }
func (p *ParameterNode) String() string       { return p.Type.String() + " " + p.Name.String() }
func (p *ParameterNode) Line() int            { return p.Token.Line }
func (p *ParameterNode) Column() int          { return p.Token.Column }
func (p *ParameterNode) Accept(v Visitor) interface{} {
	return v.VisitParameter(p)
}

// StructDeclNode объявление структуры
type StructDeclNode struct {
	Token  token.Token
	Name   *IdentifierNode
	Fields []*VarDeclNode
}

func (sd *StructDeclNode) declarationNode()     {}
func (sd *StructDeclNode) TokenLiteral() string { return sd.Token.Lexeme }
func (sd *StructDeclNode) String() string       { return "StructDecl" }
func (sd *StructDeclNode) Line() int            { return sd.Token.Line }
func (sd *StructDeclNode) Column() int          { return sd.Token.Column }
func (sd *StructDeclNode) Accept(v Visitor) interface{} {
	return v.VisitStructDecl(sd)
}

// ExternFuncDeclNode объявление внешней функции (Sprint 7)
type ExternFuncDeclNode struct {
	Token      token.Token
	Name       *IdentifierNode
	Parameters []*ParameterNode
	ReturnType *TypeNode
}

func (ed *ExternFuncDeclNode) declarationNode()     {}
func (ed *ExternFuncDeclNode) TokenLiteral() string { return ed.Token.Lexeme }
func (ed *ExternFuncDeclNode) String() string       { return "ExternFuncDecl" }
func (ed *ExternFuncDeclNode) Line() int            { return ed.Token.Line }
func (ed *ExternFuncDeclNode) Column() int          { return ed.Token.Column }
func (ed *ExternFuncDeclNode) Accept(v Visitor) interface{} {
	return v.VisitExternFuncDecl(ed)
}

// VarDeclNode объявление переменной
type VarDeclNode struct {
	Token       token.Token
	Type        *TypeNode
	Name        *IdentifierNode
	Initializer ExpressionNode
}

func (vd *VarDeclNode) declarationNode()     {}
func (vd *VarDeclNode) statementNode()       {}
func (vd *VarDeclNode) TokenLiteral() string { return vd.Token.Lexeme }
func (vd *VarDeclNode) String() string       { return "VarDecl" }
func (vd *VarDeclNode) Line() int            { return vd.Token.Line }
func (vd *VarDeclNode) Column() int          { return vd.Token.Column }
func (vd *VarDeclNode) Accept(v Visitor) interface{} {
	return v.VisitVarDecl(vd)
}

// BlockStmtNode блок операторов
type BlockStmtNode struct {
	Token      token.Token
	Statements []StatementNode
}

func (bs *BlockStmtNode) statementNode()       {}
func (bs *BlockStmtNode) TokenLiteral() string { return bs.Token.Lexeme }
func (bs *BlockStmtNode) String() string       { return "BlockStmt" }
func (bs *BlockStmtNode) Line() int            { return bs.Token.Line }
func (bs *BlockStmtNode) Column() int          { return bs.Token.Column }
func (bs *BlockStmtNode) Accept(v Visitor) interface{} {
	return v.VisitBlockStmt(bs)
}

// IfStmtNode условный оператор
type IfStmtNode struct {
	Token       token.Token
	Condition   ExpressionNode
	Consequence *BlockStmtNode
	Alternative StatementNode
}

func (is *IfStmtNode) statementNode()       {}
func (is *IfStmtNode) TokenLiteral() string { return is.Token.Lexeme }
func (is *IfStmtNode) String() string       { return "IfStmt" }
func (is *IfStmtNode) Line() int            { return is.Token.Line }
func (is *IfStmtNode) Column() int          { return is.Token.Column }
func (is *IfStmtNode) Accept(v Visitor) interface{} {
	return v.VisitIfStmt(is)
}

// WhileStmtNode цикл while
type WhileStmtNode struct {
	Token     token.Token
	Condition ExpressionNode
	Body      *BlockStmtNode
}

func (ws *WhileStmtNode) statementNode()       {}
func (ws *WhileStmtNode) TokenLiteral() string { return ws.Token.Lexeme }
func (ws *WhileStmtNode) String() string       { return "WhileStmt" }
func (ws *WhileStmtNode) Line() int            { return ws.Token.Line }
func (ws *WhileStmtNode) Column() int          { return ws.Token.Column }
func (ws *WhileStmtNode) Accept(v Visitor) interface{} {
	return v.VisitWhileStmt(ws)
}

// ForStmtNode цикл for
type ForStmtNode struct {
	Token     token.Token
	Init      StatementNode
	Condition ExpressionNode
	Update    ExpressionNode
	Body      *BlockStmtNode
}

func (fs *ForStmtNode) statementNode()       {}
func (fs *ForStmtNode) TokenLiteral() string { return fs.Token.Lexeme }
func (fs *ForStmtNode) String() string       { return "ForStmt" }
func (fs *ForStmtNode) Line() int            { return fs.Token.Line }
func (fs *ForStmtNode) Column() int          { return fs.Token.Column }
func (fs *ForStmtNode) Accept(v Visitor) interface{} {
	return v.VisitForStmt(fs)
}

// ReturnStmtNode оператор возврата
type ReturnStmtNode struct {
	Token    token.Token
	RetValue ExpressionNode
}

func (rs *ReturnStmtNode) statementNode()       {}
func (rs *ReturnStmtNode) TokenLiteral() string { return rs.Token.Lexeme }
func (rs *ReturnStmtNode) String() string       { return "ReturnStmt" }
func (rs *ReturnStmtNode) Line() int            { return rs.Token.Line }
func (rs *ReturnStmtNode) Column() int          { return rs.Token.Column }
func (rs *ReturnStmtNode) Accept(v Visitor) interface{} {
	return v.VisitReturnStmt(rs)
}

// ExprStmtNode выражение как оператор
type ExprStmtNode struct {
	Token      token.Token
	Expression ExpressionNode
}

func (es *ExprStmtNode) statementNode()       {}
func (es *ExprStmtNode) TokenLiteral() string { return es.Token.Lexeme }
func (es *ExprStmtNode) String() string       { return "ExprStmt" }
func (es *ExprStmtNode) Line() int            { return es.Token.Line }
func (es *ExprStmtNode) Column() int          { return es.Token.Column }
func (es *ExprStmtNode) Accept(v Visitor) interface{} {
	return v.VisitExprStmt(es)
}

// IdentifierNode идентификатор
type IdentifierNode struct {
	Token          token.Token
	Value          string
	TypeAnnotation *TypeAnnotation
}

func (i *IdentifierNode) expressionNode()      {}
func (i *IdentifierNode) TokenLiteral() string { return i.Token.Lexeme }
func (i *IdentifierNode) String() string       { return i.Value }
func (i *IdentifierNode) Line() int            { return i.Token.Line }
func (i *IdentifierNode) Column() int          { return i.Token.Column }
func (i *IdentifierNode) Accept(v Visitor) interface{} {
	return v.VisitIdentifier(i)
}
func (i *IdentifierNode) Type() *TypeAnnotation     { return i.TypeAnnotation }
func (i *IdentifierNode) SetType(t *TypeAnnotation) { i.TypeAnnotation = t }

// LiteralExprNode литерал
type LiteralExprNode struct {
	Token          token.Token
	TypeName       string
	IntValue       int32
	FloatValue     float64
	StringValue    string
	BoolValue      bool
	TypeAnnotation *TypeAnnotation
}

func (le *LiteralExprNode) expressionNode()      {}
func (le *LiteralExprNode) TokenLiteral() string { return le.Token.Lexeme }
func (le *LiteralExprNode) String() string       { return le.Token.Lexeme }
func (le *LiteralExprNode) Line() int            { return le.Token.Line }
func (le *LiteralExprNode) Column() int          { return le.Token.Column }
func (le *LiteralExprNode) Accept(v Visitor) interface{} {
	return v.VisitLiteralExpr(le)
}
func (le *LiteralExprNode) Type() *TypeAnnotation     { return le.TypeAnnotation }
func (le *LiteralExprNode) SetType(t *TypeAnnotation) { le.TypeAnnotation = t }

// BinaryExprNode бинарное выражение
type BinaryExprNode struct {
	Token          token.Token
	Left           ExpressionNode
	Operator       string
	Right          ExpressionNode
	TypeAnnotation *TypeAnnotation
}

func (be *BinaryExprNode) expressionNode()      {}
func (be *BinaryExprNode) TokenLiteral() string { return be.Token.Lexeme }
func (be *BinaryExprNode) String() string       { return "BinaryExpr" }
func (be *BinaryExprNode) Line() int            { return be.Token.Line }
func (be *BinaryExprNode) Column() int          { return be.Token.Column }
func (be *BinaryExprNode) Accept(v Visitor) interface{} {
	return v.VisitBinaryExpr(be)
}
func (be *BinaryExprNode) Type() *TypeAnnotation     { return be.TypeAnnotation }
func (be *BinaryExprNode) SetType(t *TypeAnnotation) { be.TypeAnnotation = t }

// UnaryExprNode унарное выражение
type UnaryExprNode struct {
	Token          token.Token
	Operator       string
	Right          ExpressionNode
	TypeAnnotation *TypeAnnotation
}

func (ue *UnaryExprNode) expressionNode()      {}
func (ue *UnaryExprNode) TokenLiteral() string { return ue.Token.Lexeme }
func (ue *UnaryExprNode) String() string       { return "UnaryExpr" }
func (ue *UnaryExprNode) Line() int            { return ue.Token.Line }
func (ue *UnaryExprNode) Column() int          { return ue.Token.Column }
func (ue *UnaryExprNode) Accept(v Visitor) interface{} {
	return v.VisitUnaryExpr(ue)
}
func (ue *UnaryExprNode) Type() *TypeAnnotation     { return ue.TypeAnnotation }
func (ue *UnaryExprNode) SetType(t *TypeAnnotation) { ue.TypeAnnotation = t }

// CallExprNode вызов функции
type CallExprNode struct {
	Token          token.Token
	Function       ExpressionNode
	Arguments      []ExpressionNode
	TypeAnnotation *TypeAnnotation
}

func (ce *CallExprNode) expressionNode()      {}
func (ce *CallExprNode) TokenLiteral() string { return ce.Token.Lexeme }
func (ce *CallExprNode) String() string       { return "CallExpr" }
func (ce *CallExprNode) Line() int            { return ce.Token.Line }
func (ce *CallExprNode) Column() int          { return ce.Token.Column }
func (ce *CallExprNode) Accept(v Visitor) interface{} {
	return v.VisitCallExpr(ce)
}
func (ce *CallExprNode) Type() *TypeAnnotation     { return ce.TypeAnnotation }
func (ce *CallExprNode) SetType(t *TypeAnnotation) { ce.TypeAnnotation = t }

// AssignmentExprNode присваивание
type AssignmentExprNode struct {
	Token          token.Token
	Left           ExpressionNode
	Operator       string
	Right          ExpressionNode
	TypeAnnotation *TypeAnnotation
}

func (ae *AssignmentExprNode) expressionNode()      {}
func (ae *AssignmentExprNode) TokenLiteral() string { return ae.Token.Lexeme }
func (ae *AssignmentExprNode) String() string       { return "AssignmentExpr" }
func (ae *AssignmentExprNode) Line() int            { return ae.Token.Line }
func (ae *AssignmentExprNode) Column() int          { return ae.Token.Column }
func (ae *AssignmentExprNode) Accept(v Visitor) interface{} {
	return v.VisitAssignmentExpr(ae)
}
func (ae *AssignmentExprNode) Type() *TypeAnnotation     { return ae.TypeAnnotation }
func (ae *AssignmentExprNode) SetType(t *TypeAnnotation) { ae.TypeAnnotation = t }

// IndexExprNode доступ к элементу массива (Sprint 7)
type IndexExprNode struct {
	Token          token.Token
	Array          ExpressionNode
	Index          ExpressionNode
	TypeAnnotation *TypeAnnotation
}

func (ie *IndexExprNode) expressionNode()      {}
func (ie *IndexExprNode) TokenLiteral() string { return ie.Token.Lexeme }
func (ie *IndexExprNode) String() string       { return "IndexExpr" }
func (ie *IndexExprNode) Line() int            { return ie.Token.Line }
func (ie *IndexExprNode) Column() int          { return ie.Token.Column }
func (ie *IndexExprNode) Accept(v Visitor) interface{} {
	return v.VisitIndexExpr(ie)
}
func (ie *IndexExprNode) Type() *TypeAnnotation     { return ie.TypeAnnotation }
func (ie *IndexExprNode) SetType(t *TypeAnnotation) { ie.TypeAnnotation = t }

// ArrayLiteralExprNode инициализатор массива (Sprint 7)
type ArrayLiteralExprNode struct {
	Token          token.Token
	Elements       []ExpressionNode
	TypeAnnotation *TypeAnnotation
}

func (al *ArrayLiteralExprNode) expressionNode()      {}
func (al *ArrayLiteralExprNode) TokenLiteral() string { return al.Token.Lexeme }
func (al *ArrayLiteralExprNode) String() string       { return "ArrayLiteral" }
func (al *ArrayLiteralExprNode) Line() int            { return al.Token.Line }
func (al *ArrayLiteralExprNode) Column() int          { return al.Token.Column }
func (al *ArrayLiteralExprNode) Accept(v Visitor) interface{} {
	return v.VisitArrayLiteralExpr(al)
}
func (al *ArrayLiteralExprNode) Type() *TypeAnnotation     { return al.TypeAnnotation }
func (al *ArrayLiteralExprNode) SetType(t *TypeAnnotation) { al.TypeAnnotation = t }

// TypeNode узел типа
type TypeNode struct {
	Token     token.Token
	Kind      string
	Name      string
	BaseType  *TypeNode // Sprint 7: базовый тип для массивов
	ArraySize int       // Sprint 7: размер массива (-1 если не указан)
}

func (t *TypeNode) String() string {
	if t.Kind == "identifier" {
		return t.Name
	}
	if t.Kind == "array" {
		if t.BaseType != nil {
			if t.ArraySize >= 0 {
				return fmt.Sprintf("%s[%d]", t.BaseType.String(), t.ArraySize)
			}
			return t.BaseType.String() + "[]"
		}
		return "unknown[]"
	}
	return t.Kind
}
