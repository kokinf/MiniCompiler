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

func (p *PrettyPrinter) Print(program *Program) string {
	p.output.Reset()
	p.indentLevel = 0

	p.output.WriteString("Program:\n")
	p.indentLevel++

	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *FunctionDecl:
			p.printFunctionDecl(d)
		case *StructDecl:
			p.printStructDecl(d)
		case *VarDecl:
			p.printVarDecl(d, true)
		}
	}

	return p.output.String()
}

func (p *PrettyPrinter) printFunctionDecl(fd *FunctionDecl) {
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

func (p *PrettyPrinter) printStructDecl(sd *StructDecl) {
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

func (p *PrettyPrinter) printVarDecl(vd *VarDecl, topLevel bool) {
	if topLevel {
		p.output.WriteString(p.indent())
	}
	p.output.WriteString(fmt.Sprintf("VarDecl: %s %s", vd.Type.String(), vd.Name.Value))
	if vd.Initializer != nil {
		p.output.WriteString(" = ")
		p.printExpression(vd.Initializer)
	}
	p.output.WriteString("\n")
}

func (p *PrettyPrinter) printBlockStmt(bs *BlockStmt, addHeader bool) {
	if addHeader {
		lastLine := bs.Line()
		if len(bs.Statements) > 0 {
			lastLine = bs.Statements[len(bs.Statements)-1].Line()
		}
		p.output.WriteString(fmt.Sprintf("%sBlock [line %d-%d]:\n",
			p.indent(), bs.Line(), lastLine))
		p.indentLevel++
	}

	for _, stmt := range bs.Statements {
		switch s := stmt.(type) {
		case *BlockStmt:
			p.printBlockStmt(s, true)
		case *IfStmt:
			p.printIfStmt(s)
		case *WhileStmt:
			p.printWhileStmt(s)
		case *ForStmt:
			p.printForStmt(s)
		case *ReturnStmt:
			p.printReturnStmt(s)
		case *VarDecl:
			p.output.WriteString(p.indent())
			p.printVarDecl(s, false)
		case *ExprStmt:
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

func (p *PrettyPrinter) printIfStmt(is *IfStmt) {
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
		case *BlockStmt:
			p.printBlockStmt(alt, true)
		case *IfStmt:
			p.printIfStmt(alt)
		}
		p.indentLevel--
	}
	p.indentLevel--
}

func (p *PrettyPrinter) printWhileStmt(ws *WhileStmt) {
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

func (p *PrettyPrinter) printForStmt(fs *ForStmt) {
	p.output.WriteString(fmt.Sprintf("%sForStmt [line %d]:\n", p.indent(), fs.Line()))
	p.indentLevel++
	if fs.Init != nil {
		p.output.WriteString(fmt.Sprintf("%sInit: ", p.indent()))
		switch init := fs.Init.(type) {
		case *VarDecl:
			p.printVarDecl(init, false)
		case *ExprStmt:
			p.printExpression(init.Expression)
			p.output.WriteString("\n")
		}
	} else {
		p.output.WriteString(fmt.Sprintf("%sInit: <empty>\n", p.indent()))
	}

	if fs.Condition != nil {
		p.output.WriteString(fmt.Sprintf("%sCondition: ", p.indent()))
		p.printExpression(fs.Condition)
		p.output.WriteString("\n")
	} else {
		p.output.WriteString(fmt.Sprintf("%sCondition: <empty>\n", p.indent()))
	}

	if fs.Update != nil {
		p.output.WriteString(fmt.Sprintf("%sUpdate: ", p.indent()))
		p.printExpression(fs.Update)
		p.output.WriteString("\n")
	} else {
		p.output.WriteString(fmt.Sprintf("%sUpdate: <empty>\n", p.indent()))
	}

	p.output.WriteString(fmt.Sprintf("%sBody:\n", p.indent()))
	p.indentLevel++
	p.printBlockStmt(fs.Body, true)
	p.indentLevel -= 2
}

func (p *PrettyPrinter) printReturnStmt(rs *ReturnStmt) {
	p.output.WriteString(fmt.Sprintf("%sReturn", p.indent()))
	if rs.RetValue != nil {
		p.output.WriteString(": ")
		p.printExpression(rs.RetValue)
	}
	p.output.WriteString("\n")
}

func (p *PrettyPrinter) printExpression(expr Expression) {
	switch e := expr.(type) {
	case *Identifier:
		p.output.WriteString(e.Value)
	case *LiteralExpr:
		switch e.Type {
		case "int":
			p.output.WriteString(fmt.Sprintf("%d", e.IntValue))
		case "float":
			p.output.WriteString(fmt.Sprintf("%g", e.FloatValue))
		case "string":
			p.output.WriteString(fmt.Sprintf("\"%s\"", e.StringValue))
		case "bool":
			p.output.WriteString(fmt.Sprintf("%t", e.BoolValue))
		}
	case *BinaryExpr:
		p.output.WriteString("(")
		p.printExpression(e.Left)
		p.output.WriteString(" " + e.Operator + " ")
		p.printExpression(e.Right)
		p.output.WriteString(")")
	case *UnaryExpr:
		p.output.WriteString("(" + e.Operator)
		p.printExpression(e.Right)
		p.output.WriteString(")")
	case *CallExpr:
		p.printExpression(e.Function)
		p.output.WriteString("(")
		for i, arg := range e.Arguments {
			if i > 0 {
				p.output.WriteString(", ")
			}
			p.printExpression(arg)
		}
		p.output.WriteString(")")
	case *AssignmentExpr:
		p.printExpression(e.Left)
		p.output.WriteString(" " + e.Operator + " ")
		p.printExpression(e.Right)
	}
}
