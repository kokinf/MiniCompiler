package ast

import (
	"fmt"
	"strings"
)

type PrettyPrinter struct {
	indentLevel int
	output      strings.Builder
}

func NewPrettyPrinter() *PrettyPrinter {
	return &PrettyPrinter{
		indentLevel: 0,
	}
}

func (p *PrettyPrinter) indent() string {
	return strings.Repeat("  ", p.indentLevel)
}

func (p *PrettyPrinter) Print(program *ProgramNode) string {
	p.output.Reset()
	p.indentLevel = 0

	p.output.WriteString("Program:\n")
	p.indentLevel++

	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *FunctionDeclNode:
			p.printFunctionDecl(d)
		case *StructDeclNode:
			p.printStructDecl(d)
		case *VarDeclNode:
			p.printVarDecl(d, true)
		}
	}

	return p.output.String()
}

func (p *PrettyPrinter) printFunctionDecl(fd *FunctionDeclNode) {
	p.output.WriteString(fmt.Sprintf("%sFunctionDecl: %s -> %s [line %d]:\n",
		p.indent(), fd.Name.Value, fd.ReturnType.String(), fd.Line()))
	p.indentLevel++

	p.output.WriteString(fmt.Sprintf("%sParameters:\n", p.indent()))
	p.indentLevel++
	for _, param := range fd.Parameters {
		p.output.WriteString(fmt.Sprintf("%s%s\n", p.indent(), param.String()))
	}
	p.indentLevel--

	p.output.WriteString(fmt.Sprintf("%sBody:\n", p.indent()))
	p.indentLevel++
	if fd.Body != nil {
		p.printBlockStmt(fd.Body, false)
	} else {
		p.output.WriteString(fmt.Sprintf("%s<empty>\n", p.indent()))
	}
	p.indentLevel -= 2
}

func (p *PrettyPrinter) printStructDecl(sd *StructDeclNode) {
	p.output.WriteString(fmt.Sprintf("%sStructDecl: %s [line %d]:\n",
		p.indent(), sd.Name.Value, sd.Line()))
	p.indentLevel++
	p.output.WriteString(fmt.Sprintf("%sFields:\n", p.indent()))
	p.indentLevel++
	for _, field := range sd.Fields {
		p.output.WriteString(fmt.Sprintf("%s%s %s\n",
			p.indent(), field.Type.String(), field.Name.Value))
	}
	p.indentLevel -= 2
}

func (p *PrettyPrinter) printVarDecl(vd *VarDeclNode, topLevel bool) {
	if topLevel {
		p.output.WriteString(p.indent())
	}
	typeStr := vd.Type.String()
	p.output.WriteString(fmt.Sprintf("VarDecl: %s %s", typeStr, vd.Name.Value))

	// Вывод аннотации типа для инициализатора
	if vd.Initializer != nil {
		p.output.WriteString(" = ")
		p.printExpression(vd.Initializer)
	}
	p.output.WriteString("\n")
}

func (p *PrettyPrinter) printBlockStmt(bs *BlockStmtNode, addHeader bool) {
	if addHeader {
		p.output.WriteString(fmt.Sprintf("%sBlock [line %d-%d]:\n",
			p.indent(), bs.Line(), bs.Line()))
		p.indentLevel++
	}

	for _, stmt := range bs.Statements {
		switch s := stmt.(type) {
		case *BlockStmtNode:
			p.printBlockStmt(s, true)
		case *IfStmtNode:
			p.printIfStmt(s)
		case *WhileStmtNode:
			p.printWhileStmt(s)
		case *ForStmtNode:
			p.printForStmt(s)
		case *ReturnStmtNode:
			p.printReturnStmt(s)
		case *VarDeclNode:
			p.output.WriteString(p.indent())
			p.printVarDecl(s, false)
		case *ExprStmtNode:
			p.output.WriteString(p.indent())
			p.output.WriteString("Expr: ")
			p.printExpression(s.Expression)
			p.output.WriteString("\n")
		}
	}

	if addHeader {
		p.indentLevel--
	}
}

func (p *PrettyPrinter) printIfStmt(is *IfStmtNode) {
	p.output.WriteString(fmt.Sprintf("%sIfStmt [line %d]:\n", p.indent(), is.Line()))
	p.indentLevel++
	p.output.WriteString(fmt.Sprintf("%sCondition:\n", p.indent()))
	p.indentLevel++
	p.output.WriteString(p.indent())
	p.printExpression(is.Condition)
	p.output.WriteString("\n")
	p.indentLevel--

	p.output.WriteString(fmt.Sprintf("%sThen:\n", p.indent()))
	p.indentLevel++
	p.printBlockStmt(is.Consequence, true)
	p.indentLevel--

	if is.Alternative != nil {
		p.output.WriteString(fmt.Sprintf("%sElse:\n", p.indent()))
		p.indentLevel++
		switch alt := is.Alternative.(type) {
		case *BlockStmtNode:
			p.printBlockStmt(alt, true)
		case *IfStmtNode:
			p.printIfStmt(alt)
		}
		p.indentLevel--
	}
	p.indentLevel--
}

func (p *PrettyPrinter) printWhileStmt(ws *WhileStmtNode) {
	p.output.WriteString(fmt.Sprintf("%sWhileStmt [line %d]:\n", p.indent(), ws.Line()))
	p.indentLevel++
	p.output.WriteString(fmt.Sprintf("%sCondition:\n", p.indent()))
	p.indentLevel++
	p.output.WriteString(p.indent())
	p.printExpression(ws.Condition)
	p.output.WriteString("\n")
	p.indentLevel--

	p.output.WriteString(fmt.Sprintf("%sBody:\n", p.indent()))
	p.indentLevel++
	p.printBlockStmt(ws.Body, true)
	p.indentLevel -= 2
}

func (p *PrettyPrinter) printForStmt(fs *ForStmtNode) {
	p.output.WriteString(fmt.Sprintf("%sForStmt [line %d]:\n", p.indent(), fs.Line()))
	p.indentLevel++
	if fs.Init != nil {
		p.output.WriteString(fmt.Sprintf("%sInit: ", p.indent()))
		switch init := fs.Init.(type) {
		case *VarDeclNode:
			p.printVarDecl(init, false)
		case *ExprStmtNode:
			p.printExpression(init.Expression)
			p.output.WriteString("\n")
		}
	}

	if fs.Condition != nil {
		p.output.WriteString(fmt.Sprintf("%sCondition: ", p.indent()))
		p.printExpression(fs.Condition)
		p.output.WriteString("\n")
	}

	if fs.Update != nil {
		p.output.WriteString(fmt.Sprintf("%sUpdate: ", p.indent()))
		p.printExpression(fs.Update)
		p.output.WriteString("\n")
	}

	p.output.WriteString(fmt.Sprintf("%sBody:\n", p.indent()))
	p.indentLevel++
	p.printBlockStmt(fs.Body, true)
	p.indentLevel -= 2
}

func (p *PrettyPrinter) printReturnStmt(rs *ReturnStmtNode) {
	p.output.WriteString(fmt.Sprintf("%sReturn", p.indent()))
	if rs.RetValue != nil {
		p.output.WriteString(": ")
		p.printExpression(rs.RetValue)
	}
	p.output.WriteString("\n")
}

func (p *PrettyPrinter) printExpression(expr ExpressionNode) {
	if expr == nil {
		p.output.WriteString("<nil>")
		return
	}

	switch e := expr.(type) {
	case *IdentifierNode:
		p.output.WriteString(e.Value)
	case *LiteralExprNode:
		switch e.TypeName {
		case "int":
			p.output.WriteString(fmt.Sprintf("%d", e.IntValue))
		case "float":
			p.output.WriteString(fmt.Sprintf("%g", e.FloatValue))
		case "string":
			p.output.WriteString(fmt.Sprintf("\"%s\"", e.StringValue))
		case "bool":
			p.output.WriteString(fmt.Sprintf("%t", e.BoolValue))
		}
	case *BinaryExprNode:
		p.output.WriteString("(")
		p.printExpression(e.Left)
		p.output.WriteString(" " + e.Operator + " ")
		p.printExpression(e.Right)
		p.output.WriteString(")")
	case *UnaryExprNode:
		p.output.WriteString("(" + e.Operator)
		p.printExpression(e.Right)
		p.output.WriteString(")")
	case *CallExprNode:
		p.printExpression(e.Function)
		p.output.WriteString("(")
		for i, arg := range e.Arguments {
			if i > 0 {
				p.output.WriteString(", ")
			}
			p.printExpression(arg)
		}
		p.output.WriteString(")")
	case *AssignmentExprNode:
		p.printExpression(e.Left)
		p.output.WriteString(" " + e.Operator + " ")
		p.printExpression(e.Right)
	}

	// Вывод аннотации типа если есть
	if t := expr.Type(); t != nil {
		p.output.WriteString(fmt.Sprintf(" [type: %s]", t.Kind))
	}
}
