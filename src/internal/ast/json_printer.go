package ast

import (
	"encoding/json"
	"fmt"
)

type JSONPrinter struct{}

func NewJSONPrinter() *JSONPrinter {
	return &JSONPrinter{}
}

type jsonNode struct {
	Type     string      `json:"type"`
	Line     int         `json:"line"`
	Column   int         `json:"column"`
	Children interface{} `json:"children,omitempty"`
	Value    interface{} `json:"value,omitempty"`
	Name     string      `json:"name,omitempty"`
	Operator string      `json:"operator,omitempty"`
	Kind     string      `json:"kind,omitempty"`
	TypeAnn  string      `json:"typeAnnotation,omitempty"`
}

func (p *JSONPrinter) Print(program *ProgramNode) string {
	node := p.convertProgram(program)
	jsonBytes, err := json.MarshalIndent(node, "", "  ")
	if err != nil {
		return fmt.Sprintf("{\"error\": \"%v\"}", err)
	}
	return string(jsonBytes)
}

func (p *JSONPrinter) convertProgram(program *ProgramNode) jsonNode {
	children := make([]interface{}, len(program.Declarations))
	for i, decl := range program.Declarations {
		children[i] = p.convertDeclaration(decl)
	}

	return jsonNode{
		Type:     "Program",
		Line:     program.Line(),
		Column:   program.Column(),
		Children: children,
	}
}

func (p *JSONPrinter) convertDeclaration(decl DeclarationNode) interface{} {
	switch d := decl.(type) {
	case *FunctionDeclNode:
		params := make([]interface{}, len(d.Parameters))
		for i, param := range d.Parameters {
			params[i] = map[string]interface{}{
				"type": param.Type.String(),
				"name": param.Name.Value,
			}
		}

		return jsonNode{
			Type:   "FunctionDecl",
			Line:   d.Line(),
			Column: d.Column(),
			Name:   d.Name.Value,
			Value: map[string]interface{}{
				"returnType": d.ReturnType.String(),
				"parameters": params,
			},
			Children: []interface{}{p.convertStatement(d.Body)},
		}

	case *StructDeclNode:
		fields := make([]interface{}, len(d.Fields))
		for i, field := range d.Fields {
			fields[i] = map[string]interface{}{
				"type": field.Type.String(),
				"name": field.Name.Value,
			}
		}

		return jsonNode{
			Type:   "StructDecl",
			Line:   d.Line(),
			Column: d.Column(),
			Name:   d.Name.Value,
			Value: map[string]interface{}{
				"fields": fields,
			},
		}

	case *VarDeclNode:
		node := jsonNode{
			Type:   "VarDecl",
			Line:   d.Line(),
			Column: d.Column(),
			Name:   d.Name.Value,
			Kind:   d.Type.String(),
		}
		if d.Initializer != nil {
			node.Children = []interface{}{p.convertExpression(d.Initializer)}
		}
		return node
	}
	return nil
}

func (p *JSONPrinter) convertStatement(stmt StatementNode) interface{} {
	if stmt == nil {
		return nil
	}

	switch s := stmt.(type) {
	case *BlockStmtNode:
		children := make([]interface{}, len(s.Statements))
		for i, stmt := range s.Statements {
			children[i] = p.convertStatement(stmt)
		}
		return jsonNode{
			Type:     "BlockStmt",
			Line:     s.Line(),
			Column:   s.Column(),
			Children: children,
		}

	case *IfStmtNode:
		node := jsonNode{
			Type:   "IfStmt",
			Line:   s.Line(),
			Column: s.Column(),
		}
		children := []interface{}{
			map[string]interface{}{"condition": p.convertExpression(s.Condition)},
			map[string]interface{}{"consequence": p.convertStatement(s.Consequence)},
		}
		if s.Alternative != nil {
			children = append(children, map[string]interface{}{"alternative": p.convertStatement(s.Alternative)})
		}
		node.Children = children
		return node

	case *WhileStmtNode:
		return jsonNode{
			Type:   "WhileStmt",
			Line:   s.Line(),
			Column: s.Column(),
			Children: []interface{}{
				map[string]interface{}{"condition": p.convertExpression(s.Condition)},
				map[string]interface{}{"body": p.convertStatement(s.Body)},
			},
		}

	case *ForStmtNode:
		children := []interface{}{}
		if s.Init != nil {
			children = append(children, map[string]interface{}{"init": p.convertStatement(s.Init)})
		}
		if s.Condition != nil {
			children = append(children, map[string]interface{}{"condition": p.convertExpression(s.Condition)})
		}
		if s.Update != nil {
			children = append(children, map[string]interface{}{"update": p.convertExpression(s.Update)})
		}
		children = append(children, map[string]interface{}{"body": p.convertStatement(s.Body)})
		return jsonNode{
			Type:     "ForStmt",
			Line:     s.Line(),
			Column:   s.Column(),
			Children: children,
		}

	case *ReturnStmtNode:
		node := jsonNode{
			Type:   "ReturnStmt",
			Line:   s.Line(),
			Column: s.Column(),
		}
		if s.RetValue != nil {
			node.Children = []interface{}{p.convertExpression(s.RetValue)}
		}
		return node

	case *ExprStmtNode:
		return jsonNode{
			Type:     "ExprStmt",
			Line:     s.Line(),
			Column:   s.Column(),
			Children: []interface{}{p.convertExpression(s.Expression)},
		}

	case *VarDeclNode:
		return p.convertDeclaration(s)
	}
	return nil
}

func (p *JSONPrinter) convertExpression(expr ExpressionNode) interface{} {
	if expr == nil {
		return nil
	}

	typeAnn := ""
	if t := expr.Type(); t != nil {
		typeAnn = t.Kind
	}

	switch e := expr.(type) {
	case *IdentifierNode:
		return jsonNode{
			Type:    "Identifier",
			Line:    e.Line(),
			Column:  e.Column(),
			Name:    e.Value,
			TypeAnn: typeAnn,
		}

	case *LiteralExprNode:
		var value interface{}
		switch e.TypeName {
		case "int":
			value = e.IntValue
		case "float":
			value = e.FloatValue
		case "string":
			value = e.StringValue
		case "bool":
			value = e.BoolValue
		}
		return jsonNode{
			Type:    "LiteralExpr",
			Line:    e.Line(),
			Column:  e.Column(),
			Kind:    e.TypeName,
			Value:   value,
			TypeAnn: typeAnn,
		}

	case *BinaryExprNode:
		return jsonNode{
			Type:     "BinaryExpr",
			Line:     e.Line(),
			Column:   e.Column(),
			Operator: e.Operator,
			TypeAnn:  typeAnn,
			Children: []interface{}{
				map[string]interface{}{"left": p.convertExpression(e.Left)},
				map[string]interface{}{"right": p.convertExpression(e.Right)},
			},
		}

	case *UnaryExprNode:
		return jsonNode{
			Type:     "UnaryExpr",
			Line:     e.Line(),
			Column:   e.Column(),
			Operator: e.Operator,
			TypeAnn:  typeAnn,
			Children: []interface{}{p.convertExpression(e.Right)},
		}

	case *CallExprNode:
		args := make([]interface{}, len(e.Arguments))
		for i, arg := range e.Arguments {
			args[i] = p.convertExpression(arg)
		}
		return jsonNode{
			Type:    "CallExpr",
			Line:    e.Line(),
			Column:  e.Column(),
			TypeAnn: typeAnn,
			Children: []interface{}{
				map[string]interface{}{"function": p.convertExpression(e.Function)},
				map[string]interface{}{"arguments": args},
			},
		}

	case *AssignmentExprNode:
		return jsonNode{
			Type:     "AssignmentExpr",
			Line:     e.Line(),
			Column:   e.Column(),
			Operator: e.Operator,
			TypeAnn:  typeAnn,
			Children: []interface{}{
				map[string]interface{}{"left": p.convertExpression(e.Left)},
				map[string]interface{}{"right": p.convertExpression(e.Right)},
			},
		}
	}
	return nil
}
