package ir

import (
	"fmt"
	"mikrocompiler/src/internal/ast"
	"mikrocompiler/src/internal/semantic"
)

// IRGenerator генерирует IR из декорированного AST
type IRGenerator struct {
	program     *Program
	symbolTable *semantic.SymbolTable
	typeSystem  *semantic.TypeSystem

	currentFunc  *Function
	currentBlock *BasicBlock

	// Для циклов
	breakTargets    []*BasicBlock
	continueTargets []*BasicBlock

	// Переменные -> их адреса в памяти
	varAddresses map[string]*Operand

	// Счетчики для меток
	labelCounter int

	// Для PHI узлов
	mergeBlocks map[string]*BasicBlock
}

// NewIRGenerator создает новый генератор IR
func NewIRGenerator(symbolTable *semantic.SymbolTable, typeSystem *semantic.TypeSystem) *IRGenerator {
	return &IRGenerator{
		program:         NewProgram(),
		symbolTable:     symbolTable,
		typeSystem:      typeSystem,
		breakTargets:    make([]*BasicBlock, 0),
		continueTargets: make([]*BasicBlock, 0),
		varAddresses:    make(map[string]*Operand),
		labelCounter:    0,
		mergeBlocks:     make(map[string]*BasicBlock),
	}
}

// Generate генерирует IR для всей программы
func (g *IRGenerator) Generate(program *ast.ProgramNode) *Program {
	// Сначала собираем глобальные переменные
	for _, decl := range program.Declarations {
		if vd, ok := decl.(*ast.VarDeclNode); ok {
			g.generateGlobalVar(vd)
		}
	}

	// Затем генерируем функции
	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.FunctionDeclNode:
			g.generateFunction(d)
		case *ast.StructDeclNode:
		}
	}

	// Строим CFG и DOM дерево для всех функций
	for _, fn := range g.program.Functions {
		fn.BuildDominatorTree()
		fn.BuildGlobalDefUseChains()
	}

	return g.program
}

// generateGlobalVar генерирует глобальную переменную
func (g *IRGenerator) generateGlobalVar(vd *ast.VarDeclNode) {
	var init *Operand
	if vd.Initializer != nil {
		init = g.generateExpression(vd.Initializer)
	}
	g.program.AddGlobal(vd.Name.Value, vd.Type.String(), init)
}

// generateFunction генерирует IR для функции
func (g *IRGenerator) generateFunction(fd *ast.FunctionDeclNode) {
	funcName := fd.Name.Value
	returnType := "void"
	if fd.ReturnType != nil && fd.ReturnType.Kind != "" {
		returnType = fd.ReturnType.String()
	}

	g.currentFunc = NewFunction(funcName, returnType)
	g.currentBlock = g.currentFunc.EntryBlock
	g.varAddresses = make(map[string]*Operand)
	g.mergeBlocks = make(map[string]*BasicBlock)
	g.labelCounter = 0

	// Добавляем параметры
	for i, param := range fd.Parameters {
		g.currentFunc.AddParam(param.Name.Value, param.Type.String())
		paramAddr := g.currentFunc.NewTemp()
		typeSize := g.getTypeSize(param.Type.String())
		g.currentBlock.AddInstruction(&Instruction{
			Opcode:  OpAlloca,
			Dest:    paramAddr,
			Src1:    NewLiteralOperand(typeSize),
			Comment: fmt.Sprintf("param %s: %s", param.Name.Value, param.Type.String()),
		})
		g.currentBlock.AddInstruction(NewStoreInst(paramAddr, NewVarOperand(param.Name.Value)))
		g.varAddresses[param.Name.Value] = paramAddr

		// Устанавливаем смещение параметра
		offset := 8 + i*8
		g.currentFunc.Locals[param.Name.Value] = &VarInfo{
			Name:   param.Name.Value,
			Type:   param.Type.String(),
			Offset: offset,
			Size:   typeSize,
		}
	}

	// Генерируем тело функции
	if fd.Body != nil {
		g.generateBlockStmt(fd.Body)
	}

	// Если функция void и нет return, добавляем RETURN
	if !g.currentBlock.IsTerminated() {
		if returnType == "void" {
			g.currentBlock.AddInstruction(NewRetInst(nil))
		} else {
			g.currentBlock.AddInstruction(NewRetInst(NewLiteralOperand(0)))
		}
	}

	g.program.AddFunction(g.currentFunc)
	g.currentFunc = nil
	g.currentBlock = nil
}

// getTypeSize возвращает размер типа в байтах
func (g *IRGenerator) getTypeSize(typeName string) int {
	switch typeName {
	case "int":
		return 4
	case "float":
		return 8
	case "bool":
		return 1
	case "string":
		return 16
	default:
		return 8
	}
}

// newLabel генерирует уникальную метку
func (g *IRGenerator) newLabel(prefix string) string {
	g.labelCounter++
	return g.currentFunc.NewLabel(prefix)
}

// generateBlockStmt генерирует блок операторов
func (g *IRGenerator) generateBlockStmt(block *ast.BlockStmtNode) {
	for _, stmt := range block.Statements {
		if g.currentBlock.IsTerminated() {
			break
		}
		g.generateStatement(stmt)
	}
}

// generateStatement генерирует оператор
func (g *IRGenerator) generateStatement(stmt ast.StatementNode) {
	switch s := stmt.(type) {
	case *ast.VarDeclNode:
		g.generateVarDecl(s)
	case *ast.ExprStmtNode:
		if assign, ok := s.Expression.(*ast.AssignmentExprNode); ok {
			g.generateAssignment(assign)
		} else {
			g.generateExpression(s.Expression)
		}
	case *ast.IfStmtNode:
		g.generateIfStmt(s)
	case *ast.WhileStmtNode:
		g.generateWhileStmt(s)
	case *ast.ForStmtNode:
		g.generateForStmt(s)
	case *ast.ReturnStmtNode:
		g.generateReturnStmt(s)
	case *ast.BlockStmtNode:
		g.generateBlockStmt(s)
	}
}

// generateVarDecl генерирует объявление переменной
func (g *IRGenerator) generateVarDecl(vd *ast.VarDeclNode) {
	varName := vd.Name.Value

	// Создаем временную для адреса переменной
	addrTemp := g.currentFunc.NewTemp()

	// Выделяем память
	typeSize := g.getTypeSize(vd.Type.String())

	allocInst := &Instruction{
		Opcode:  OpAlloca,
		Dest:    addrTemp,
		Src1:    NewLiteralOperand(typeSize),
		Comment: fmt.Sprintf("var %s: %s", varName, vd.Type.String()),
	}
	g.currentBlock.AddInstruction(allocInst)
	g.varAddresses[varName] = addrTemp

	// Добавляем в локальные переменные функции
	offset := len(g.currentFunc.Locals) * 8
	g.currentFunc.Locals[varName] = &VarInfo{
		Name:   varName,
		Type:   vd.Type.String(),
		Offset: offset,
		Size:   typeSize,
	}

	// Инициализатор
	if vd.Initializer != nil {
		initVal := g.generateExpression(vd.Initializer)
		if initVal != nil {
			g.currentBlock.AddInstruction(NewStoreInst(addrTemp, initVal))
		}
	}
}

// generateIfStmt генерирует if оператор с PHI узлами
func (g *IRGenerator) generateIfStmt(is *ast.IfStmtNode) {
	condVal := g.generateExpression(is.Condition)
	if condVal == nil {
		return
	}

	thenBlock := g.currentFunc.NewBlock(g.newLabel("if_then"))
	var elseBlock *BasicBlock
	if is.Alternative != nil {
		elseBlock = g.currentFunc.NewBlock(g.newLabel("if_else"))
	}
	mergeBlock := g.currentFunc.NewBlock(g.newLabel("if_merge"))

	// Сохраняем состояние для PHI
	savedVarAddresses := make(map[string]*Operand)
	for k, v := range g.varAddresses {
		savedVarAddresses[k] = v
	}

	// Условный переход
	g.currentBlock.AddInstruction(NewCondJumpInst(OpJmpIf, condVal, thenBlock))
	if elseBlock != nil {
		g.currentBlock.AddInstruction(NewJumpInst(elseBlock))
	} else {
		g.currentBlock.AddInstruction(NewJumpInst(mergeBlock))
	}

	// Связываем блоки
	g.currentBlock.AddSuccessor(thenBlock)
	thenBlock.AddPredecessor(g.currentBlock)
	if elseBlock != nil {
		g.currentBlock.AddSuccessor(elseBlock)
		elseBlock.AddPredecessor(g.currentBlock)
	} else {
		g.currentBlock.AddSuccessor(mergeBlock)
		mergeBlock.AddPredecessor(g.currentBlock)
	}

	// Then блок
	g.currentBlock = thenBlock
	g.generateStatement(is.Consequence)

	// Собираем изменённые переменные для PHI
	thenVarAddrs := make(map[string]*Operand)
	for k, v := range g.varAddresses {
		thenVarAddrs[k] = v
	}

	if !g.currentBlock.IsTerminated() {
		g.currentBlock.AddInstruction(NewJumpInst(mergeBlock))
		g.currentBlock.AddSuccessor(mergeBlock)
		mergeBlock.AddPredecessor(g.currentBlock)
	}

	// Else блок
	if elseBlock != nil {
		g.varAddresses = savedVarAddresses
		g.currentBlock = elseBlock
		g.generateStatement(is.Alternative)

		if !g.currentBlock.IsTerminated() {
			g.currentBlock.AddInstruction(NewJumpInst(mergeBlock))
			g.currentBlock.AddSuccessor(mergeBlock)
			mergeBlock.AddPredecessor(g.currentBlock)
		}
	} else {
		g.varAddresses = savedVarAddresses
	}

	// Генерируем PHI узлы в merge блоке
	g.currentBlock = mergeBlock
	for varName, thenAddr := range thenVarAddrs {
		elseAddr, exists := g.varAddresses[varName]
		if !exists || thenAddr.Name != elseAddr.Name {
			// Переменная была изменена в одной из веток - создаём PHI
			phiTemp := g.currentFunc.NewTemp()
			pairs := []PhiPair{
				{Value: thenAddr, Block: thenBlock},
			}
			if elseBlock != nil {
				pairs = append(pairs, PhiPair{Value: elseAddr, Block: elseBlock})
			} else {
				pairs = append(pairs, PhiPair{Value: savedVarAddresses[varName], Block: mergeBlock.Predecessors[0]})
			}

			g.currentBlock.AddInstruction(NewPhiInst(phiTemp, pairs))
			g.varAddresses[varName] = phiTemp
		}
	}
}

// generateWhileStmt генерирует while оператор
func (g *IRGenerator) generateWhileStmt(ws *ast.WhileStmtNode) {
	headerBlock := g.currentFunc.NewBlock(g.newLabel("while_header"))
	bodyBlock := g.currentFunc.NewBlock(g.newLabel("while_body"))
	exitBlock := g.currentFunc.NewBlock(g.newLabel("while_exit"))

	// Сохраняем цели для break/continue
	g.breakTargets = append(g.breakTargets, exitBlock)
	g.continueTargets = append(g.continueTargets, headerBlock)

	// Переход к заголовку
	g.currentBlock.AddInstruction(NewJumpInst(headerBlock))
	g.currentBlock.AddSuccessor(headerBlock)
	headerBlock.AddPredecessor(g.currentBlock)

	// Заголовок - проверка условия
	g.currentBlock = headerBlock
	condVal := g.generateExpression(ws.Condition)
	if condVal != nil {
		g.currentBlock.AddInstruction(NewCondJumpInst(OpJmpIf, condVal, bodyBlock))
		g.currentBlock.AddInstruction(NewJumpInst(exitBlock))
		g.currentBlock.AddSuccessor(bodyBlock)
		g.currentBlock.AddSuccessor(exitBlock)
		bodyBlock.AddPredecessor(g.currentBlock)
		exitBlock.AddPredecessor(g.currentBlock)
	}

	// Тело цикла
	g.currentBlock = bodyBlock
	g.generateStatement(ws.Body)
	if !g.currentBlock.IsTerminated() {
		g.currentBlock.AddInstruction(NewJumpInst(headerBlock))
		g.currentBlock.AddSuccessor(headerBlock)
		headerBlock.AddPredecessor(g.currentBlock)
	}

	// Восстанавливаем цели
	g.breakTargets = g.breakTargets[:len(g.breakTargets)-1]
	g.continueTargets = g.continueTargets[:len(g.continueTargets)-1]

	g.currentBlock = exitBlock
}

// generateForStmt генерирует for оператор
func (g *IRGenerator) generateForStmt(fs *ast.ForStmtNode) {
	// Инициализация
	if fs.Init != nil {
		g.generateStatement(fs.Init)
	}

	headerBlock := g.currentFunc.NewBlock(g.newLabel("for_header"))
	bodyBlock := g.currentFunc.NewBlock(g.newLabel("for_body"))
	updateBlock := g.currentFunc.NewBlock(g.newLabel("for_update"))
	exitBlock := g.currentFunc.NewBlock(g.newLabel("for_exit"))

	// Сохраняем цели
	g.breakTargets = append(g.breakTargets, exitBlock)
	g.continueTargets = append(g.continueTargets, updateBlock)

	// Переход к заголовку
	g.currentBlock.AddInstruction(NewJumpInst(headerBlock))
	g.currentBlock.AddSuccessor(headerBlock)
	headerBlock.AddPredecessor(g.currentBlock)

	// Заголовок - проверка условия
	g.currentBlock = headerBlock
	if fs.Condition != nil {
		condVal := g.generateExpression(fs.Condition)
		if condVal != nil {
			g.currentBlock.AddInstruction(NewCondJumpInst(OpJmpIf, condVal, bodyBlock))
			g.currentBlock.AddInstruction(NewJumpInst(exitBlock))
			g.currentBlock.AddSuccessor(bodyBlock)
			g.currentBlock.AddSuccessor(exitBlock)
			bodyBlock.AddPredecessor(g.currentBlock)
			exitBlock.AddPredecessor(g.currentBlock)
		}
	} else {
		g.currentBlock.AddInstruction(NewJumpInst(bodyBlock))
		g.currentBlock.AddSuccessor(bodyBlock)
		bodyBlock.AddPredecessor(g.currentBlock)
	}

	// Тело цикла
	g.currentBlock = bodyBlock
	g.generateStatement(fs.Body)
	if !g.currentBlock.IsTerminated() {
		g.currentBlock.AddInstruction(NewJumpInst(updateBlock))
		g.currentBlock.AddSuccessor(updateBlock)
		updateBlock.AddPredecessor(g.currentBlock)
	}

	// Обновление
	g.currentBlock = updateBlock
	if fs.Update != nil {
		g.generateExpression(fs.Update)
	}
	g.currentBlock.AddInstruction(NewJumpInst(headerBlock))
	g.currentBlock.AddSuccessor(headerBlock)
	headerBlock.AddPredecessor(g.currentBlock)

	// Восстанавливаем цели
	g.breakTargets = g.breakTargets[:len(g.breakTargets)-1]
	g.continueTargets = g.continueTargets[:len(g.continueTargets)-1]

	g.currentBlock = exitBlock
}

// generateReturnStmt генерирует return оператор
func (g *IRGenerator) generateReturnStmt(rs *ast.ReturnStmtNode) {
	var retVal *Operand
	if rs.RetValue != nil {
		retVal = g.generateExpression(rs.RetValue)
	}
	g.currentBlock.AddInstruction(NewRetInst(retVal))
}

// generateAssignment генерирует присваивание
func (g *IRGenerator) generateAssignment(ae *ast.AssignmentExprNode) {
	rightVal := g.generateExpression(ae.Right)
	if rightVal == nil {
		return
	}

	if ident, ok := ae.Left.(*ast.IdentifierNode); ok {
		if addr, exists := g.varAddresses[ident.Value]; exists {
			g.currentBlock.AddInstruction(NewStoreInst(addr, rightVal))
		} else {
			// Глобальная переменная
			g.currentBlock.AddInstruction(NewStoreInst(NewGlobalOperand(ident.Value), rightVal))
		}
	}
}

// generateExpression генерирует выражение и возвращает операнд с результатом
func (g *IRGenerator) generateExpression(expr ast.ExpressionNode) *Operand {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.LiteralExprNode:
		return g.generateLiteral(e)
	case *ast.IdentifierNode:
		return g.generateIdentifier(e)
	case *ast.BinaryExprNode:
		return g.generateBinaryExpr(e)
	case *ast.UnaryExprNode:
		return g.generateUnaryExpr(e)
	case *ast.CallExprNode:
		return g.generateCallExpr(e)
	default:
		return nil
	}
}

// generateLiteral генерирует литерал
func (g *IRGenerator) generateLiteral(lit *ast.LiteralExprNode) *Operand {
	switch lit.TypeName {
	case "int":
		return NewLiteralOperand(int(lit.IntValue))
	case "float":
		return NewLiteralOperand(lit.FloatValue)
	case "bool":
		return NewLiteralOperand(lit.BoolValue)
	case "string":
		return NewLiteralOperand(lit.StringValue)
	default:
		return nil
	}
}

// generateIdentifier генерирует идентификатор
func (g *IRGenerator) generateIdentifier(ident *ast.IdentifierNode) *Operand {
	// Проверяем локальные переменные
	if addr, exists := g.varAddresses[ident.Value]; exists {
		temp := g.currentFunc.NewTemp()
		g.currentBlock.AddInstruction(NewLoadInst(temp, addr))
		return temp
	}

	// Проверяем параметры
	for _, param := range g.currentFunc.Params {
		if param.Name == ident.Value {
			return NewVarOperand(ident.Value)
		}
	}

	// Глобальная переменная
	temp := g.currentFunc.NewTemp()
	g.currentBlock.AddInstruction(NewLoadInst(temp, NewGlobalOperand(ident.Value)))
	return temp
}

// generateBinaryExpr генерирует бинарное выражение
func (g *IRGenerator) generateBinaryExpr(be *ast.BinaryExprNode) *Operand {
	left := g.generateExpression(be.Left)
	right := g.generateExpression(be.Right)

	if left == nil || right == nil {
		return nil
	}

	temp := g.currentFunc.NewTemp()

	var op Opcode
	switch be.Operator {
	case "+":
		op = OpAdd
	case "-":
		op = OpSub
	case "*":
		op = OpMul
	case "/":
		op = OpDiv
	case "%":
		op = OpMod
	case "==":
		op = OpCmpEq
	case "!=":
		op = OpCmpNe
	case "<":
		op = OpCmpLt
	case "<=":
		op = OpCmpLe
	case ">":
		op = OpCmpGt
	case ">=":
		op = OpCmpGe
	case "&&":
		op = OpAnd
	case "||":
		op = OpOr
	default:
		return nil
	}

	g.currentBlock.AddInstruction(NewBinaryInst(op, temp, left, right))
	return temp
}

// generateUnaryExpr генерирует унарное выражение
func (g *IRGenerator) generateUnaryExpr(ue *ast.UnaryExprNode) *Operand {
	right := g.generateExpression(ue.Right)
	if right == nil {
		return nil
	}

	temp := g.currentFunc.NewTemp()

	var op Opcode
	switch ue.Operator {
	case "-":
		op = OpNeg
	case "!":
		op = OpNot
	default:
		return nil
	}

	g.currentBlock.AddInstruction(NewUnaryInst(op, temp, right))
	return temp
}

// generateCallExpr генерирует вызов функции
func (g *IRGenerator) generateCallExpr(ce *ast.CallExprNode) *Operand {
	var funcName string
	if ident, ok := ce.Function.(*ast.IdentifierNode); ok {
		funcName = ident.Value
	} else {
		return nil
	}

	// Генерируем аргументы
	args := make([]*Operand, len(ce.Arguments))
	for i, arg := range ce.Arguments {
		args[i] = g.generateExpression(arg)
		if args[i] == nil {
			return nil
		}
		g.currentBlock.AddInstruction(&Instruction{
			Opcode: OpParam,
			Src1:   NewLiteralOperand(i),
			Src2:   args[i],
		})
	}

	temp := g.currentFunc.NewTemp()
	g.currentBlock.AddInstruction(NewCallInst(temp, funcName, args))
	return temp
}

// GetProgram возвращает сгенерированную программу
func (g *IRGenerator) GetProgram() *Program {
	return g.program
}
