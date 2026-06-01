package ast

import (
	"fmt"
	"strings"
)

type DOTPrinter struct {
	nodeCounter int
	output      strings.Builder
}

func NewDOTPrinter() *DOTPrinter {
	return &DOTPrinter{
		nodeCounter: 0,
	}
}

func (p *DOTPrinter) Print(program *ProgramNode) string {
	p.output.Reset()
	p.nodeCounter = 0

	p.output.WriteString("digraph AST {\n")
	p.output.WriteString("  node [shape=box, style=filled, fillcolor=lightblue];\n")
	p.output.WriteString("  edge [arrowhead=vee];\n\n")

	rootID := p.nextID()
	p.output.WriteString(fmt.Sprintf("  node%d [label=\"Program\"];\n", rootID))

	for _, decl := range program.Declarations {
		declID := p.printNode(decl)
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d;\n", rootID, declID))
	}

	p.output.WriteString("}\n")
	return p.output.String()
}

func (p *DOTPrinter) nextID() int {
	p.nodeCounter++
	return p.nodeCounter
}

func (p *DOTPrinter) printNode(node Node) int {
	id := p.nextID()

	switch n := node.(type) {
	case *ExternFuncDeclNode:
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"Extern\\n%s -> %s\", fillcolor=lightblue];\n",
			id, n.Name.Value, n.ReturnType.String()))

		paramsID := p.nextID()
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"Parameters\", shape=box, fillcolor=lightsalmon];\n", paramsID))
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d;\n", id, paramsID))

		for _, param := range n.Parameters {
			paramID := p.nextID()
			p.output.WriteString(fmt.Sprintf("  node%d [label=\"%s\", shape=box, fillcolor=lightsalmon];\n",
				paramID, param.String()))
			p.output.WriteString(fmt.Sprintf("  node%d -> node%d;\n", paramsID, paramID))
		}

	case *FunctionDeclNode:
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"Function\\n%s -> %s\", fillcolor=lightgreen];\n",
			id, n.Name.Value, n.ReturnType.String()))

		paramsID := p.nextID()
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"Parameters\", shape=box, fillcolor=lightsalmon];\n", paramsID))
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d;\n", id, paramsID))

		for _, param := range n.Parameters {
			paramID := p.nextID()
			p.output.WriteString(fmt.Sprintf("  node%d [label=\"%s\", shape=box, fillcolor=lightsalmon];\n",
				paramID, param.String()))
			p.output.WriteString(fmt.Sprintf("  node%d -> node%d;\n", paramsID, paramID))
		}

		bodyID := p.printBlockStmt(n.Body)
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"body\"];\n", id, bodyID))

	case *StructDeclNode:
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"Struct\\n%s\", fillcolor=lightgreen];\n",
			id, n.Name.Value))

		for _, field := range n.Fields {
			fieldID := p.nextID()
			p.output.WriteString(fmt.Sprintf("  node%d [label=\"%s %s\", shape=box, fillcolor=lightyellow];\n",
				fieldID, field.Type.String(), field.Name.Value))
			p.output.WriteString(fmt.Sprintf("  node%d -> node%d;\n", id, fieldID))
		}

	case *VarDeclNode:
		label := fmt.Sprintf("VarDecl\\n%s %s", n.Type.String(), n.Name.Value)
		if n.Initializer != nil {
			label += " = ..."
		}
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"%s\", fillcolor=lightyellow];\n", id, label))

		if n.Initializer != nil {
			initID := p.printExpression(n.Initializer)
			p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"init\"];\n", id, initID))
		}

	case *ForStmtNode:
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"For\", fillcolor=lightcoral];\n", id))

		if n.Init != nil {
			initID := p.printStatement(n.Init)
			p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"init\"];\n", id, initID))
		} else {
			emptyID := p.nextID()
			p.output.WriteString(fmt.Sprintf("  node%d [label=\"<empty>\", shape=box, fillcolor=lightgray];\n", emptyID))
			p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"init\"];\n", id, emptyID))
		}

		if n.Condition != nil {
			condID := p.printExpression(n.Condition)
			p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"cond\"];\n", id, condID))
		} else {
			emptyID := p.nextID()
			p.output.WriteString(fmt.Sprintf("  node%d [label=\"<empty>\", shape=box, fillcolor=lightgray];\n", emptyID))
			p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"cond\"];\n", id, emptyID))
		}

		if n.Update != nil {
			updateID := p.printExpression(n.Update)
			p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"update\"];\n", id, updateID))
		} else {
			emptyID := p.nextID()
			p.output.WriteString(fmt.Sprintf("  node%d [label=\"<empty>\", shape=box, fillcolor=lightgray];\n", emptyID))
			p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"update\"];\n", id, emptyID))
		}

		bodyID := p.printBlockStmt(n.Body)
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"body\"];\n", id, bodyID))
	}

	return id
}

func (p *DOTPrinter) printBlockStmt(block *BlockStmtNode) int {
	id := p.nextID()
	p.output.WriteString(fmt.Sprintf("  node%d [label=\"Block\", fillcolor=lightcoral];\n", id))

	for _, stmt := range block.Statements {
		stmtID := p.printStatement(stmt)
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d;\n", id, stmtID))
	}

	return id
}

func (p *DOTPrinter) printStatement(stmt StatementNode) int {
	if stmt == nil {
		id := p.nextID()
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"<empty>\", shape=box, fillcolor=lightgray];\n", id))
		return id
	}

	id := p.nextID()

	switch s := stmt.(type) {
	case *BlockStmtNode:
		return p.printBlockStmt(s)

	case *IfStmtNode:
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"If\", fillcolor=lightcoral];\n", id))

		condID := p.printExpression(s.Condition)
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"cond\"];\n", id, condID))

		thenID := p.printBlockStmt(s.Consequence)
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"then\"];\n", id, thenID))

		if s.Alternative != nil {
			elseID := p.printStatement(s.Alternative)
			p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"else\"];\n", id, elseID))
		}

	case *WhileStmtNode:
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"While\", fillcolor=lightcoral];\n", id))

		condID := p.printExpression(s.Condition)
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"cond\"];\n", id, condID))

		bodyID := p.printBlockStmt(s.Body)
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"body\"];\n", id, bodyID))

	case *ForStmtNode:
		return p.printNode(s)

	case *ReturnStmtNode:
		label := "Return"
		if s.RetValue != nil {
			label += " (with value)"
		}
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"%s\", fillcolor=lightcoral];\n", id, label))

		if s.RetValue != nil {
			valID := p.printExpression(s.RetValue)
			p.output.WriteString(fmt.Sprintf("  node%d -> node%d;\n", id, valID))
		}

	case *ExprStmtNode:
		exprID := p.printExpression(s.Expression)
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"ExprStmt\", fillcolor=lightcoral];\n", id))
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d;\n", id, exprID))

	case *VarDeclNode:
		return p.printNode(s)
	}

	return id
}

func (p *DOTPrinter) printExpression(expr ExpressionNode) int {
	if expr == nil {
		id := p.nextID()
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"<nil>\", shape=box, fillcolor=lightgray];\n", id))
		return id
	}

	id := p.nextID()

	switch e := expr.(type) {
	case *IdentifierNode:
		label := fmt.Sprintf("Ident: %s", e.Value)
		if t := e.Type(); t != nil {
			label += fmt.Sprintf("\\n[%s]", t.Kind)
		}
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"%s\", shape=box, fillcolor=lightgray];\n", id, label))

	case *LiteralExprNode:
		label := fmt.Sprintf("Literal: %s", e.Token.Lexeme)
		if t := e.Type(); t != nil {
			label += fmt.Sprintf("\\n[%s]", t.Kind)
		}
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"%s\", shape=box, fillcolor=lightgray];\n", id, label))

	case *BinaryExprNode:
		label := e.Operator
		if t := e.Type(); t != nil {
			label += fmt.Sprintf("\\n[%s]", t.Kind)
		}
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"%s\", fillcolor=lightgray];\n", id, label))

		leftID := p.printExpression(e.Left)
		rightID := p.printExpression(e.Right)
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"left\"];\n", id, leftID))
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"right\"];\n", id, rightID))

	case *UnaryExprNode:
		label := e.Operator
		if t := e.Type(); t != nil {
			label += fmt.Sprintf("\\n[%s]", t.Kind)
		}
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"%s\", fillcolor=lightgray];\n", id, label))

		rightID := p.printExpression(e.Right)
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d;\n", id, rightID))

	case *CallExprNode:
		label := "Call"
		if t := e.Type(); t != nil {
			label += fmt.Sprintf("\\n[%s]", t.Kind)
		}
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"%s\", fillcolor=lightgray];\n", id, label))

		funcID := p.printExpression(e.Function)
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"func\"];\n", id, funcID))

		for i, arg := range e.Arguments {
			argID := p.printExpression(arg)
			p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"arg%d\"];\n", id, argID, i))
		}

	case *AssignmentExprNode:
		label := e.Operator
		if t := e.Type(); t != nil {
			label += fmt.Sprintf("\\n[%s]", t.Kind)
		}
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"%s\", fillcolor=lightgray];\n", id, label))

		leftID := p.printExpression(e.Left)
		rightID := p.printExpression(e.Right)
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"left\"];\n", id, leftID))
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"right\"];\n", id, rightID))

	case *IndexExprNode:
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"Index[]\", fillcolor=lightgray];\n", id))
		arrayID := p.printExpression(e.Array)
		indexID := p.printExpression(e.Index)
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"array\"];\n", id, arrayID))
		p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"index\"];\n", id, indexID))

	case *ArrayLiteralExprNode:
		p.output.WriteString(fmt.Sprintf("  node%d [label=\"ArrayLiteral{%d}\", fillcolor=lightgray];\n", id, len(e.Elements)))
		for i, elem := range e.Elements {
			elemID := p.printExpression(elem)
			p.output.WriteString(fmt.Sprintf("  node%d -> node%d [label=\"[%d]\"];\n", id, elemID, i))
		}
	}

	return id
}
