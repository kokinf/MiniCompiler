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
}

func NewSemanticAnalyzer() *SemanticAnalyzer {
	return &SemanticAnalyzer{
		symbolTable:  NewSymbolTable(),
		errors:       NewErrorCollector(),
		currentFunc:  nil,
		currentScope: "global",
		inLoop:       false,
	}
}

func (sa *SemanticAnalyzer) Analyze(program *ast.Program) (*SymbolTable, *ErrorCollector) {
	sa.collectDeclarations(program)

	sa.analyzeProgram(program)

	return sa.symbolTable, sa.errors
}

func (sa *SemanticAnalyzer) collectDeclarations(program *ast.Program) {
	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.FunctionDecl:
			sa.collectFunctionDecl(d)
		case *ast.StructDecl:
			sa.collectStructDecl(d)
		case *ast.VarDecl:
			sa.collectGlobalVarDecl(d)
		}
	}
}

func (sa *SemanticAnalyzer) collectFunctionDecl(fd *ast.FunctionDecl) {
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

func (sa *SemanticAnalyzer) collectStructDecl(sd *ast.StructDecl) {
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
		// Проверка на дубликат поля
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

func (sa *SemanticAnalyzer) collectGlobalVarDecl(vd *ast.VarDecl) {
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

func (sa *SemanticAnalyzer) analyzeProgram(program *ast.Program) {
	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.FunctionDecl:
			sa.analyzeFunction(d)
		case *ast.StructDecl:
			sa.analyzeStruct(d)
		case *ast.VarDecl:
			sa.analyzeGlobalVarDecl(d)
		}
	}
}

func (sa *SemanticAnalyzer) analyzeFunction(fd *ast.FunctionDecl) {
	funcName := fd.Name.Value
	sym := sa.symbolTable.Lookup(funcName)
	if sym == nil {
		return
	}

	sa.currentFunc = sym
	sa.currentScope = "function " + funcName

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
		// Проверка на дубликат параметра
		if existing := sa.symbolTable.LookupLocal(param.Name.Value); existing != nil {
			sa.errors.Add(ErrDuplicateDeclaration,
				fmt.Sprintf("parameter '%s' already declared", param.Name.Value),
				param.Name.Line(), param.Name.Column(), sa.currentScope)
		}
		sa.symbolTable.Insert(paramSym)
	}

	if fd.Body != nil {
		sa.analyzeBlockStmt(fd.Body, sym.Type.Return)
	}

	if fd.Body != nil && !sym.Type.Return.IsVoid() {
		if !sa.hasReturnStatement(fd.Body) {
			sa.errors.Add(ErrInvalidReturn,
				fmt.Sprintf("function '%s' must return a value of type %s", funcName, sym.Type.Return.String()),
				fd.Line(), fd.Column(), sa.currentScope)
		}
	}

	sa.symbolTable.ExitScope()
	sa.currentFunc = nil
	sa.currentScope = "global"
}

func (sa *SemanticAnalyzer) analyzeStruct(sd *ast.StructDecl) {
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

func (sa *SemanticAnalyzer) analyzeGlobalVarDecl(vd *ast.VarDecl) {
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

func (sa *SemanticAnalyzer) analyzeBlockStmt(block *ast.BlockStmt, expectedReturn *Type) *Type {
	sa.symbolTable.EnterScope("block")
	defer sa.symbolTable.ExitScope()

	var lastType *Type

	for _, stmt := range block.Statements {
		stmtType := sa.analyzeStatement(stmt, expectedReturn)
		if stmtType != nil {
			lastType = stmtType
		}
	}

	return lastType
}

func (sa *SemanticAnalyzer) analyzeStatement(stmt ast.Statement, expectedReturn *Type) *Type {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		return sa.analyzeBlockStmt(s, expectedReturn)

	case *ast.VarDecl:
		sa.collectLocalVarDecl(s)
		sa.analyzeLocalVarDecl(s)
		return nil

	case *ast.ExprStmt:
		return sa.analyzeExpression(s.Expression)

	case *ast.IfStmt:
		condType := sa.analyzeExpression(s.Condition)
		if condType != nil && !condType.IsBool() {
			sa.errors.Add(ErrInvalidCondition,
				fmt.Sprintf("if condition must be bool, got %s", condType.String()),
				s.Line(), s.Column(), sa.currentScope)
		}

		sa.analyzeBlockStmt(s.Consequence, expectedReturn)

		if s.Alternative != nil {
			if altBlock, ok := s.Alternative.(*ast.BlockStmt); ok {
				sa.analyzeBlockStmt(altBlock, expectedReturn)
			} else {
				sa.analyzeStatement(s.Alternative, expectedReturn)
			}
		}
		return nil

	case *ast.WhileStmt:
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
		return nil

	case *ast.ForStmt:
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
		return nil

	case *ast.ReturnStmt:
		if sa.currentFunc == nil {
			sa.errors.Add(ErrInvalidReturn,
				"return statement outside function",
				s.Line(), s.Column(), sa.currentScope)
			return nil
		}

		expectedType := sa.currentFunc.Type.Return

		if s.RetValue == nil {
			if !expectedType.IsVoid() {
				sa.errors.Add(ErrInvalidReturn,
					fmt.Sprintf("function returns %s, but no value provided", expectedType.String()),
					s.Line(), s.Column(), sa.currentScope)
			}
			return expectedType
		}

		retType := sa.analyzeExpression(s.RetValue)
		if retType != nil && !retType.IsAssignableTo(expectedType) {
			sa.errors.Add(ErrInvalidReturn,
				fmt.Sprintf("cannot return %s, expected %s", retType.String(), expectedType.String()),
				s.Line(), s.Column(), sa.currentScope)
		}
		return expectedType

	default:
		return nil
	}
}

func (sa *SemanticAnalyzer) collectLocalVarDecl(vd *ast.VarDecl) {
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
}

func (sa *SemanticAnalyzer) analyzeLocalVarDecl(vd *ast.VarDecl) {
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

func (sa *SemanticAnalyzer) analyzeExpression(expr ast.Expression) *Type {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.Identifier:
		sym := sa.symbolTable.Lookup(e.Value)
		if sym == nil {
			sa.errors.Add(ErrUndeclaredIdentifier,
				fmt.Sprintf("identifier '%s' not declared", e.Value),
				e.Line(), e.Column(), sa.currentScope)
			return nil
		}
		return sym.Type

	case *ast.LiteralExpr:
		switch e.Type {
		case "int":
			return NewType(TypeInt)
		case "float":
			return NewType(TypeFloat)
		case "bool":
			return NewType(TypeBool)
		case "string":
			return NewType(TypeString)
		default:
			return nil
		}

	case *ast.BinaryExpr:
		leftType := sa.analyzeExpression(e.Left)
		rightType := sa.analyzeExpression(e.Right)

		if leftType == nil || rightType == nil {
			return nil
		}

		switch e.Operator {
		case "+", "-", "*", "/", "%":
			if leftType.IsNumeric() && rightType.IsNumeric() {
				if leftType.IsFloat() || rightType.IsFloat() {
					return NewType(TypeFloat)
				}
				return NewType(TypeInt)
			}
			sa.errors.Add(ErrInvalidBinaryOp,
				fmt.Sprintf("operator %s requires numeric operands, got %s and %s", e.Operator, leftType.String(), rightType.String()),
				e.Line(), e.Column(), sa.currentScope)
			return nil

		case "==", "!=", "<", "<=", ">", ">=":
			if leftType.IsAssignableTo(rightType) || rightType.IsAssignableTo(leftType) {
				return NewType(TypeBool)
			}
			sa.errors.Add(ErrInvalidBinaryOp,
				fmt.Sprintf("cannot compare %s and %s", leftType.String(), rightType.String()),
				e.Line(), e.Column(), sa.currentScope)
			return nil

		case "&&", "||":
			if leftType.IsBool() && rightType.IsBool() {
				return NewType(TypeBool)
			}
			sa.errors.Add(ErrInvalidBinaryOp,
				fmt.Sprintf("operator %s requires bool operands, got %s and %s", e.Operator, leftType.String(), rightType.String()),
				e.Line(), e.Column(), sa.currentScope)
			return nil

		default:
			return nil
		}

	case *ast.UnaryExpr:
		rightType := sa.analyzeExpression(e.Right)
		if rightType == nil {
			return nil
		}

		switch e.Operator {
		case "-":
			if rightType.IsNumeric() {
				return rightType
			}
			sa.errors.Add(ErrInvalidUnaryOp,
				fmt.Sprintf("operator - requires numeric operand, got %s", rightType.String()),
				e.Line(), e.Column(), sa.currentScope)
			return nil

		case "!":
			if rightType.IsBool() {
				return NewType(TypeBool)
			}
			sa.errors.Add(ErrInvalidUnaryOp,
				fmt.Sprintf("operator ! requires bool operand, got %s", rightType.String()),
				e.Line(), e.Column(), sa.currentScope)
			return nil

		default:
			return nil
		}

	case *ast.CallExpr:
		funcType := sa.analyzeExpression(e.Function)
		if funcType == nil {
			return nil
		}

		if !funcType.IsFunction() {
			sa.errors.Add(ErrFunctionNotFound,
				fmt.Sprintf("'%s' is not a function", e.Function.String()),
				e.Line(), e.Column(), sa.currentScope)
			return nil
		}

		expectedParams := funcType.Params
		if len(e.Arguments) != len(expectedParams) {
			sa.errors.Add(ErrArgumentCount,
				fmt.Sprintf("expected %d arguments, got %d", len(expectedParams), len(e.Arguments)),
				e.Line(), e.Column(), sa.currentScope)
			return funcType.Return
		}

		for i, arg := range e.Arguments {
			argType := sa.analyzeExpression(arg)
			if argType != nil && !argType.IsAssignableTo(expectedParams[i]) {
				sa.errors.Add(ErrArgumentType,
					fmt.Sprintf("argument %d: expected %s, got %s", i+1, expectedParams[i].String(), argType.String()),
					arg.Line(), arg.Column(), sa.currentScope)
			}
		}

		return funcType.Return

	case *ast.AssignmentExpr:
		leftType := sa.analyzeExpression(e.Left)
		rightType := sa.analyzeExpression(e.Right)

		if leftType == nil || rightType == nil {
			return nil
		}

		if ident, ok := e.Left.(*ast.Identifier); ok {
			sym := sa.symbolTable.Lookup(ident.Value)
			if sym != nil && sym.Kind == SymbolFunction {
				sa.errors.Add(ErrInvalidAssignment,
					fmt.Sprintf("cannot assign to function '%s'", ident.Value),
					e.Line(), e.Column(), sa.currentScope)
				return nil
			}
		}

		if !rightType.IsAssignableTo(leftType) {
			sa.errors.Add(ErrTypeMismatch,
				fmt.Sprintf("cannot assign %s to %s", rightType.String(), leftType.String()),
				e.Line(), e.Column(), sa.currentScope)
		}

		return leftType

	default:
		return nil
	}
}

func (sa *SemanticAnalyzer) typeFromAST(t ast.Type) *Type {
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

func (sa *SemanticAnalyzer) hasReturnStatement(block *ast.BlockStmt) bool {
	for _, stmt := range block.Statements {
		switch s := stmt.(type) {
		case *ast.ReturnStmt:
			return true
		case *ast.BlockStmt:
			if sa.hasReturnStatement(s) {
				return true
			}
		}
	}
	return false
}

func (sa *SemanticAnalyzer) GetSymbolTable() *SymbolTable {
	return sa.symbolTable
}

func (sa *SemanticAnalyzer) GetErrors() []*SemanticError {
	return sa.errors.Errors()
}
