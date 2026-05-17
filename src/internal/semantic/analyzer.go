package semantic

import (
	"fmt"

	"mikrocompiler/src/internal/ast"
)

type SemanticAnalyzer struct {
	symbolTable  *SymbolTable
	errors       *ErrorCollector
	currentFunc  *Symbol
	currentScope string
	inLoop       bool

	// Для отслеживания объявленных переменных в текущем блоке
	declaredInBlock map[string]bool
}

func NewSemanticAnalyzer() *SemanticAnalyzer {
	return &SemanticAnalyzer{
		symbolTable:     NewSymbolTable(),
		errors:          NewErrorCollector(),
		currentFunc:     nil,
		currentScope:    "global",
		inLoop:          false,
		declaredInBlock: make(map[string]bool),
	}
}

func (sa *SemanticAnalyzer) Analyze(program *ast.ProgramNode) (*SymbolTable, *ErrorCollector, *ast.ProgramNode) {
	sa.collectDeclarations(program)

	decoratedProgram := sa.analyzeProgram(program)

	return sa.symbolTable, sa.errors, decoratedProgram
}

func (sa *SemanticAnalyzer) collectDeclarations(program *ast.ProgramNode) {
	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.FunctionDeclNode:
			sa.collectFunctionDecl(d)
		case *ast.StructDeclNode:
			sa.collectStructDecl(d)
		case *ast.VarDeclNode:
			sa.collectGlobalVarDecl(d)
		}
	}
}

func (sa *SemanticAnalyzer) collectFunctionDecl(fd *ast.FunctionDeclNode) {
	funcName := fd.Name.Value

	if existing := sa.symbolTable.Lookup(funcName); existing != nil {
		sa.errors.Add(ErrDuplicateDeclaration,
			fmt.Sprintf("function '%s' already declared at line %d", funcName, existing.Line),
			fd.Line(), fd.Column(), sa.currentScope)
		return
	}

	returnType := sa.typeFromAST(fd.ReturnType)
	if returnType == nil {
		returnType = NewType(TypeVoid)
	}

	params := make([]*Symbol, 0)
	paramTypes := make([]*Type, 0)

	for _, param := range fd.Parameters {
		paramType := sa.typeFromAST(param.Type)
		if paramType == nil {
			paramType = NewType(TypeVoid)
		}
		paramTypes = append(paramTypes, paramType)

		paramSym := &Symbol{
			Name:   param.Name.Value,
			Kind:   SymbolParameter,
			Type:   paramType,
			Line:   param.Name.Line(),
			Column: param.Name.Column(),
		}
		params = append(params, paramSym)
	}

	funcType := NewFunctionType(returnType, paramTypes)

	sym := &Symbol{
		Name:       funcName,
		Kind:       SymbolFunction,
		Type:       funcType,
		Line:       fd.Line(),
		Column:     fd.Column(),
		Parameters: params,
	}

	sa.symbolTable.Insert(sym)
}

func (sa *SemanticAnalyzer) collectStructDecl(sd *ast.StructDeclNode) {
	structName := sd.Name.Value

	if existing := sa.symbolTable.Lookup(structName); existing != nil {
		sa.errors.Add(ErrDuplicateDeclaration,
			fmt.Sprintf("struct '%s' already declared at line %d", structName, existing.Line),
			sd.Line(), sd.Column(), sa.currentScope)
		return
	}

	structType := NewStructType(structName)

	fields := make(map[string]*Symbol)
	for _, field := range sd.Fields {
		fieldType := sa.typeFromAST(field.Type)
		if fieldType == nil {
			fieldType = NewType(TypeVoid)
		}
		fieldSym := &Symbol{
			Name:   field.Name.Value,
			Kind:   SymbolField,
			Type:   fieldType,
			Line:   field.Line(),
			Column: field.Column(),
		}
		if _, exists := fields[field.Name.Value]; exists {
			sa.errors.Add(ErrDuplicateDeclaration,
				fmt.Sprintf("field '%s' already declared in struct", field.Name.Value),
				field.Line(), field.Column(), "struct "+structName)
		}
		fields[field.Name.Value] = fieldSym
		structType.Fields[field.Name.Value] = fieldType
	}

	sym := &Symbol{
		Name:   structName,
		Kind:   SymbolStruct,
		Type:   structType,
		Line:   sd.Line(),
		Column: sd.Column(),
		Fields: fields,
	}

	sa.symbolTable.Insert(sym)
}

func (sa *SemanticAnalyzer) collectGlobalVarDecl(vd *ast.VarDeclNode) {
	varName := vd.Name.Value
	varType := sa.typeFromAST(vd.Type)

	if varType == nil {
		sa.errors.Add(ErrTypeMismatch,
			fmt.Sprintf("invalid type for variable '%s'", varName),
			vd.Line(), vd.Column(), sa.currentScope)
		return
	}

	if existing := sa.symbolTable.Lookup(varName); existing != nil {
		sa.errors.Add(ErrDuplicateDeclaration,
			fmt.Sprintf("variable '%s' already declared at line %d", varName, existing.Line),
			vd.Line(), vd.Column(), sa.currentScope)
		return
	}

	sym := &Symbol{
		Name:   varName,
		Kind:   SymbolVariable,
		Type:   varType,
		Line:   vd.Line(),
		Column: vd.Column(),
	}

	sa.symbolTable.Insert(sym)
}

func (sa *SemanticAnalyzer) analyzeProgram(program *ast.ProgramNode) *ast.ProgramNode {
	decoratedProgram := &ast.ProgramNode{
		Declarations: make([]ast.DeclarationNode, 0),
		LinePos:      program.LinePos,
		ColumnPos:    program.ColumnPos,
	}

	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.FunctionDeclNode:
			decoratedFunc := sa.analyzeFunction(d)
			decoratedProgram.Declarations = append(decoratedProgram.Declarations, decoratedFunc)
		case *ast.StructDeclNode:
			sa.analyzeStruct(d)
			decoratedProgram.Declarations = append(decoratedProgram.Declarations, d)
		case *ast.VarDeclNode:
			sa.analyzeGlobalVarDecl(d)
			decoratedProgram.Declarations = append(decoratedProgram.Declarations, d)
		}
	}

	return decoratedProgram
}

func (sa *SemanticAnalyzer) analyzeFunction(fd *ast.FunctionDeclNode) *ast.FunctionDeclNode {
	funcName := fd.Name.Value
	sym := sa.symbolTable.Lookup(funcName)
	if sym == nil {
		return fd
	}

	sa.currentFunc = sym
	sa.currentScope = "function " + funcName
	sa.declaredInBlock = make(map[string]bool)

	sa.symbolTable.EnterScope(funcName)

	for _, param := range fd.Parameters {
		paramType := sa.typeFromAST(param.Type)
		if paramType == nil {
			paramType = NewType(TypeVoid)
		}
		paramSym := &Symbol{
			Name:   param.Name.Value,
			Kind:   SymbolParameter,
			Type:   paramType,
			Line:   param.Name.Line(),
			Column: param.Name.Column(),
		}
		if existing := sa.symbolTable.LookupLocal(param.Name.Value); existing != nil {
			sa.errors.Add(ErrDuplicateDeclaration,
				fmt.Sprintf("parameter '%s' already declared", param.Name.Value),
				param.Name.Line(), param.Name.Column(), sa.currentScope)
		}
		sa.symbolTable.Insert(paramSym)
		sa.declaredInBlock[param.Name.Value] = true
	}

	decoratedFunc := &ast.FunctionDeclNode{
		Token:      fd.Token,
		Name:       fd.Name,
		Parameters: fd.Parameters,
		ReturnType: fd.ReturnType,
	}

	if fd.Body != nil {
		decoratedFunc.Body = sa.analyzeBlockStmt(fd.Body, sym.Type.Return)

		// Проверка возврата для non-void функций
		if !sym.Type.Return.IsVoid() {
			if !sa.hasReturnStatement(fd.Body) {
				sa.errors.Add(ErrInvalidReturn,
					fmt.Sprintf("function '%s' must return a value of type %s", funcName, sym.Type.Return.String()),
					fd.Line(), fd.Column(), sa.currentScope)
			}
		}
	}

	sa.symbolTable.ExitScope()
	sa.currentFunc = nil
	sa.currentScope = "global"

	return decoratedFunc
}

func (sa *SemanticAnalyzer) analyzeStruct(sd *ast.StructDeclNode) {
	structSym := sa.symbolTable.Lookup(sd.Name.Value)
	if structSym == nil {
		return
	}

	for _, field := range sd.Fields {
		fieldType := sa.typeFromAST(field.Type)
		if fieldType == nil {
			sa.errors.Add(ErrTypeMismatch,
				fmt.Sprintf("invalid type for field '%s'", field.Name.Value),
				field.Line(), field.Column(), "struct "+sd.Name.Value)
		}
	}
}

func (sa *SemanticAnalyzer) analyzeGlobalVarDecl(vd *ast.VarDeclNode) {
	sym := sa.symbolTable.Lookup(vd.Name.Value)
	if sym == nil {
		return
	}

	if vd.Initializer != nil {
		initType := sa.analyzeExpression(vd.Initializer)
		if initType != nil && !initType.IsAssignableTo(sym.Type) {
			sa.errors.Add(ErrTypeMismatch,
				fmt.Sprintf("cannot assign %s to %s", initType.String(), sym.Type.String()),
				vd.Line(), vd.Column(), sa.currentScope)
		}
	}
}

func (sa *SemanticAnalyzer) analyzeBlockStmt(block *ast.BlockStmtNode, expectedReturn *Type) *ast.BlockStmtNode {
	sa.symbolTable.EnterScope("block")
	defer sa.symbolTable.ExitScope()

	prevDeclared := sa.declaredInBlock
	sa.declaredInBlock = make(map[string]bool)

	decoratedBlock := &ast.BlockStmtNode{
		Token:      block.Token,
		Statements: make([]ast.StatementNode, 0),
	}

	for _, stmt := range block.Statements {
		decoratedStmt := sa.analyzeStatement(stmt, expectedReturn)
		if decoratedStmt != nil {
			decoratedBlock.Statements = append(decoratedBlock.Statements, decoratedStmt)
		}
	}

	sa.declaredInBlock = prevDeclared
	return decoratedBlock
}

func (sa *SemanticAnalyzer) analyzeStatement(stmt ast.StatementNode, expectedReturn *Type) ast.StatementNode {
	switch s := stmt.(type) {
	case *ast.BlockStmtNode:
		return sa.analyzeBlockStmt(s, expectedReturn)

	case *ast.VarDeclNode:
		sa.collectLocalVarDecl(s)
		sa.analyzeLocalVarDecl(s)
		return s

	case *ast.ExprStmtNode:
		sa.analyzeExpression(s.Expression)
		return s

	case *ast.IfStmtNode:
		condType := sa.analyzeExpression(s.Condition)
		if condType != nil && !condType.IsBool() {
			sa.errors.Add(ErrInvalidCondition,
				fmt.Sprintf("if condition must be bool, got %s", condType.String()),
				s.Line(), s.Column(), sa.currentScope)
		}

		sa.analyzeBlockStmt(s.Consequence, expectedReturn)

		if s.Alternative != nil {
			if altBlock, ok := s.Alternative.(*ast.BlockStmtNode); ok {
				sa.analyzeBlockStmt(altBlock, expectedReturn)
			} else {
				sa.analyzeStatement(s.Alternative, expectedReturn)
			}
		}
		return s

	case *ast.WhileStmtNode:
		oldInLoop := sa.inLoop
		sa.inLoop = true

		condType := sa.analyzeExpression(s.Condition)
		if condType != nil && !condType.IsBool() {
			sa.errors.Add(ErrInvalidCondition,
				fmt.Sprintf("while condition must be bool, got %s", condType.String()),
				s.Line(), s.Column(), sa.currentScope)
		}

		sa.analyzeBlockStmt(s.Body, expectedReturn)
		sa.inLoop = oldInLoop
		return s

	case *ast.ForStmtNode:
		oldInLoop := sa.inLoop
		sa.inLoop = true

		if s.Init != nil {
			sa.analyzeStatement(s.Init, expectedReturn)
		}

		if s.Condition != nil {
			condType := sa.analyzeExpression(s.Condition)
			if condType != nil && !condType.IsBool() {
				sa.errors.Add(ErrInvalidCondition,
					fmt.Sprintf("for condition must be bool, got %s", condType.String()),
					s.Line(), s.Column(), sa.currentScope)
			}
		}

		if s.Update != nil {
			sa.analyzeExpression(s.Update)
		}

		sa.analyzeBlockStmt(s.Body, expectedReturn)
		sa.inLoop = oldInLoop
		return s

	case *ast.ReturnStmtNode:
		if sa.currentFunc == nil {
			sa.errors.Add(ErrInvalidReturn,
				"return statement outside function",
				s.Line(), s.Column(), sa.currentScope)
			return s
		}

		expectedType := sa.currentFunc.Type.Return

		if s.RetValue == nil {
			if !expectedType.IsVoid() {
				sa.errors.Add(ErrInvalidReturn,
					fmt.Sprintf("function returns %s, but no value provided", expectedType.String()),
					s.Line(), s.Column(), sa.currentScope)
			}
			return s
		}

		retType := sa.analyzeExpression(s.RetValue)
		if retType != nil && !retType.IsAssignableTo(expectedType) {
			sa.errors.Add(ErrInvalidReturn,
				fmt.Sprintf("cannot return %s, expected %s", retType.String(), expectedType.String()),
				s.Line(), s.Column(), sa.currentScope)
		}
		return s

	default:
		return stmt
	}
}

func (sa *SemanticAnalyzer) collectLocalVarDecl(vd *ast.VarDeclNode) {
	varName := vd.Name.Value
	varType := sa.typeFromAST(vd.Type)

	if varType == nil {
		sa.errors.Add(ErrTypeMismatch,
			fmt.Sprintf("invalid type for variable '%s'", varName),
			vd.Line(), vd.Column(), sa.currentScope)
		return
	}

	if existing := sa.symbolTable.LookupLocal(varName); existing != nil {
		sa.errors.Add(ErrDuplicateDeclaration,
			fmt.Sprintf("variable '%s' already declared in this scope at line %d", varName, existing.Line),
			vd.Line(), vd.Column(), sa.currentScope)
		return
	}

	sym := &Symbol{
		Name:   varName,
		Kind:   SymbolVariable,
		Type:   varType,
		Line:   vd.Line(),
		Column: vd.Column(),
	}

	sa.symbolTable.Insert(sym)
	sa.declaredInBlock[varName] = true
}

func (sa *SemanticAnalyzer) analyzeLocalVarDecl(vd *ast.VarDeclNode) {
	sym := sa.symbolTable.LookupLocal(vd.Name.Value)
	if sym == nil {
		return
	}

	if vd.Initializer != nil {
		initType := sa.analyzeExpression(vd.Initializer)
		if initType != nil && !initType.IsAssignableTo(sym.Type) {
			sa.errors.Add(ErrTypeMismatch,
				fmt.Sprintf("cannot assign %s to %s", initType.String(), sym.Type.String()),
				vd.Line(), vd.Column(), sa.currentScope)
		}
	}
}

func (sa *SemanticAnalyzer) analyzeExpression(expr ast.ExpressionNode) *Type {
	if expr == nil {
		return nil
	}

	var exprType *Type

	switch e := expr.(type) {
	case *ast.IdentifierNode:
		sym := sa.symbolTable.Lookup(e.Value)
		if sym == nil {
			sa.errors.Add(ErrUndeclaredIdentifier,
				fmt.Sprintf("identifier '%s' not declared", e.Value),
				e.Line(), e.Column(), sa.currentScope)
			exprType = nil
		} else {
			// Проверка use before declaration
			if sym.Kind == SymbolVariable && !sa.declaredInBlock[e.Value] {
				currentScope := sa.symbolTable.GetCurrentScope()
				if sym.Scope == currentScope {
					sa.errors.Add(ErrUseBeforeDeclaration,
						fmt.Sprintf("variable '%s' used before declaration", e.Value),
						e.Line(), e.Column(), sa.currentScope)
				}
			}
			exprType = sym.Type
		}
		e.SetType(typeToAnnotation(exprType))

	case *ast.LiteralExprNode:
		switch e.TypeName {
		case "int":
			exprType = NewType(TypeInt)
		case "float":
			exprType = NewType(TypeFloat)
		case "bool":
			exprType = NewType(TypeBool)
		case "string":
			exprType = NewType(TypeString)
		}
		e.SetType(typeToAnnotation(exprType))

	case *ast.BinaryExprNode:
		leftType := sa.analyzeExpression(e.Left)
		rightType := sa.analyzeExpression(e.Right)

		if leftType == nil || rightType == nil {
			exprType = nil
		} else {
			exprType = sa.checkBinaryOp(e.Operator, leftType, rightType, e.Line(), e.Column())
		}
		e.SetType(typeToAnnotation(exprType))

	case *ast.UnaryExprNode:
		rightType := sa.analyzeExpression(e.Right)
		if rightType == nil {
			exprType = nil
		} else {
			exprType = sa.checkUnaryOp(e.Operator, rightType, e.Line(), e.Column())
		}
		e.SetType(typeToAnnotation(exprType))

	case *ast.CallExprNode:
		funcType := sa.analyzeExpression(e.Function)
		if funcType == nil {
			exprType = nil
		} else if !funcType.IsFunction() {
			sa.errors.Add(ErrFunctionNotFound,
				fmt.Sprintf("'%s' is not a function", e.Function.String()),
				e.Line(), e.Column(), sa.currentScope)
			exprType = nil
		} else {
			// Проверка аргументов
			expectedParams := funcType.Params
			if len(e.Arguments) != len(expectedParams) {
				sa.errors.Add(ErrArgumentCount,
					fmt.Sprintf("expected %d arguments, got %d", len(expectedParams), len(e.Arguments)),
					e.Line(), e.Column(), sa.currentScope)
			}

			for i, arg := range e.Arguments {
				argType := sa.analyzeExpression(arg)
				if i < len(expectedParams) && argType != nil && !argType.IsAssignableTo(expectedParams[i]) {
					sa.errors.Add(ErrArgumentType,
						fmt.Sprintf("argument %d: expected %s, got %s", i+1, expectedParams[i].String(), argType.String()),
						arg.Line(), arg.Column(), sa.currentScope)
				}
			}

			exprType = funcType.Return
		}
		e.SetType(typeToAnnotation(exprType))

	case *ast.AssignmentExprNode:
		leftType := sa.analyzeExpression(e.Left)
		rightType := sa.analyzeExpression(e.Right)

		if leftType != nil && rightType != nil {
			if ident, ok := e.Left.(*ast.IdentifierNode); ok {
				sym := sa.symbolTable.Lookup(ident.Value)
				if sym != nil && sym.Kind == SymbolFunction {
					sa.errors.Add(ErrInvalidAssignment,
						fmt.Sprintf("cannot assign to function '%s'", ident.Value),
						e.Line(), e.Column(), sa.currentScope)
				}
			}

			if !rightType.IsAssignableTo(leftType) {
				sa.errors.Add(ErrTypeMismatch,
					fmt.Sprintf("cannot assign %s to %s", rightType.String(), leftType.String()),
					e.Line(), e.Column(), sa.currentScope)
			}
		}
		exprType = leftType
		e.SetType(typeToAnnotation(exprType))
	}

	return exprType
}

func (sa *SemanticAnalyzer) checkBinaryOp(op string, left, right *Type, line, column int) *Type {
	switch op {
	case "+", "-", "*", "/", "%":
		if left.IsNumeric() && right.IsNumeric() {
			if left.IsFloat() || right.IsFloat() {
				return NewType(TypeFloat)
			}
			return NewType(TypeInt)
		}
		sa.errors.Add(ErrInvalidBinaryOp,
			fmt.Sprintf("operator %s requires numeric operands, got %s and %s", op, left.String(), right.String()),
			line, column, sa.currentScope)
		return nil

	case "==", "!=", "<", "<=", ">", ">=":
		if left.IsAssignableTo(right) || right.IsAssignableTo(left) {
			return NewType(TypeBool)
		}
		sa.errors.Add(ErrInvalidBinaryOp,
			fmt.Sprintf("cannot compare %s and %s", left.String(), right.String()),
			line, column, sa.currentScope)
		return nil

	case "&&", "||":
		if left.IsBool() && right.IsBool() {
			return NewType(TypeBool)
		}
		sa.errors.Add(ErrInvalidBinaryOp,
			fmt.Sprintf("operator %s requires bool operands, got %s and %s", op, left.String(), right.String()),
			line, column, sa.currentScope)
		return nil

	default:
		return nil
	}
}

func (sa *SemanticAnalyzer) checkUnaryOp(op string, operand *Type, line, column int) *Type {
	switch op {
	case "-":
		if operand.IsNumeric() {
			return operand
		}
		sa.errors.Add(ErrInvalidUnaryOp,
			fmt.Sprintf("operator - requires numeric operand, got %s", operand.String()),
			line, column, sa.currentScope)
		return nil

	case "!":
		if operand.IsBool() {
			return NewType(TypeBool)
		}
		sa.errors.Add(ErrInvalidUnaryOp,
			fmt.Sprintf("operator ! requires bool operand, got %s", operand.String()),
			line, column, sa.currentScope)
		return nil

	default:
		return nil
	}
}

func (sa *SemanticAnalyzer) typeFromAST(t *ast.TypeNode) *Type {
	if t == nil {
		return nil
	}

	switch t.Kind {
	case "int":
		return NewType(TypeInt)
	case "float":
		return NewType(TypeFloat)
	case "bool":
		return NewType(TypeBool)
	case "void":
		return NewType(TypeVoid)
	case "string":
		return NewType(TypeString)
	case "identifier":
		sym := sa.symbolTable.Lookup(t.Name)
		if sym != nil && sym.Kind == SymbolStruct {
			return sym.Type
		}
		sa.errors.Add(ErrStructNotFound,
			fmt.Sprintf("undefined struct type '%s'", t.Name),
			t.Token.Line, t.Token.Column, sa.currentScope)
		return nil
	default:
		return nil
	}
}

// hasReturnStatement проверяет, гарантирует ли блок возврат значения
func (sa *SemanticAnalyzer) hasReturnStatement(block *ast.BlockStmtNode) bool {
	if block == nil {
		return false
	}

	// Проверяем все statements с конца
	for i := len(block.Statements) - 1; i >= 0; i-- {
		stmt := block.Statements[i]

		switch s := stmt.(type) {
		case *ast.ReturnStmtNode:
			return true

		case *ast.IfStmtNode:
			// If с else  проверяем обе ветки
			if s.Alternative != nil {
				consHasRet := sa.blockHasReturnStmt(s.Consequence)
				altHasRet := sa.blockHasReturnStmt(s.Alternative)
				if consHasRet && altHasRet {
					return true
				}
			}
			// If без else  продолжаем проверку дальше
			continue

		case *ast.ForStmtNode:
			// Бесконечный цикл for (;;) с return внутри
			if s.Condition == nil && s.Update == nil {
				if sa.blockHasReturnStmt(s.Body) {
					return true
				}
			}
			// Обычный for  не гарантирует возврат
			continue

		case *ast.WhileStmtNode:
			// Бесконечный цикл while (true) с return внутри
			if sa.isAlwaysTrue(s.Condition) {
				if sa.blockHasReturnStmt(s.Body) {
					return true
				}
			}
			continue

		case *ast.BlockStmtNode:
			if sa.hasReturnStatement(s) {
				return true
			}

		default:
			return false
		}
	}

	return false
}

// blockHasReturnStmt проверяет наличие return в блоке
func (sa *SemanticAnalyzer) blockHasReturnStmt(stmt ast.StatementNode) bool {
	if stmt == nil {
		return false
	}

	switch s := stmt.(type) {
	case *ast.BlockStmtNode:
		return sa.hasReturnStatement(s)
	case *ast.ReturnStmtNode:
		return true
	case *ast.IfStmtNode:
		if s.Alternative == nil {
			return false
		}
		return sa.blockHasReturnStmt(s.Consequence) && sa.blockHasReturnStmt(s.Alternative)
	default:
		return false
	}
}

// isAlwaysTrue проверяет, является ли условие всегда истинным
func (sa *SemanticAnalyzer) isAlwaysTrue(expr ast.ExpressionNode) bool {
	if lit, ok := expr.(*ast.LiteralExprNode); ok {
		return lit.TypeName == "bool" && lit.BoolValue
	}
	return false
}

func typeToAnnotation(t *Type) *ast.TypeAnnotation {
	if t == nil {
		return nil
	}
	return &ast.TypeAnnotation{
		Kind: t.String(),
		Name: t.Name,
	}
}

func (sa *SemanticAnalyzer) GetSymbolTable() *SymbolTable {
	return sa.symbolTable
}

func (sa *SemanticAnalyzer) GetErrors() []*SemanticError {
	return sa.errors.Errors()
}
