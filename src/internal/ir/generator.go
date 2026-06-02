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

	breakTargets    []*BasicBlock
	continueTargets []*BasicBlock

	varAddresses map[string]*Operand

	labelCounter int

	mergeBlocks map[string]*BasicBlock

	// Sprint 7: отслеживание массивов для автоматического освобождения
	arrayAllocs map[string]*Operand
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
		arrayAllocs:     make(map[string]*Operand),
	}
}

// Generate генерирует IR для всей программы
func (g *IRGenerator) Generate(program *ast.ProgramNode) *Program {
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
		case *ast.ExternFuncDeclNode:
			g.generateExternFunction(d)
		case *ast.StructDeclNode:
		}
	}

	for _, fn := range g.program.Functions {
		if !fn.IsExtern {
			fn.BuildDominatorTree()
			fn.BuildGlobalDefUseChains()
		}
	}

	return g.program
}

func (g *IRGenerator) generateGlobalVar(vd *ast.VarDeclNode) {
	var init *Operand
	if vd.Initializer != nil {
		init = g.generateExpression(vd.Initializer)
	}
	g.program.AddGlobal(vd.Name.Value, vd.Type.String(), init)
}

func (g *IRGenerator) generateExternFunction(ed *ast.ExternFuncDeclNode) {
	funcName := ed.Name.Value
	returnType := "void"
	if ed.ReturnType != nil && ed.ReturnType.Kind != "" {
		returnType = ed.ReturnType.String()
	}

	extFunc := NewFunction(funcName, returnType)
	extFunc.IsExtern = true

	for _, param := range ed.Parameters {
		extFunc.AddParam(param.Name.Value, param.Type.String())
	}

	g.program.AddFunction(extFunc)
}

func (g *IRGenerator) generateFunction(fd *ast.FunctionDeclNode) {
	funcName := fd.Name.Value
	returnType := "void"
	if fd.ReturnType != nil && fd.ReturnType.Kind != "" {
		returnType = fd.ReturnType.String()
	}

	g.currentFunc = NewFunction(funcName, returnType)
	g.currentBlock = g.currentFunc.EntryBlock
	g.varAddresses = make(map[string]*Operand)
	g.arrayAllocs = make(map[string]*Operand)
	g.mergeBlocks = make(map[string]*BasicBlock)
	g.labelCounter = 0

	// Добавляем параметры
	for i, param := range fd.Parameters {
		paramType := param.Type.String()
		g.currentFunc.AddParam(param.Name.Value, paramType)
		// Параметры доступны напрямую через VarOperand
		g.varAddresses[param.Name.Value] = NewVarOperand(param.Name.Value)

		g.currentFunc.Locals[param.Name.Value] = &VarInfo{
			Name:   param.Name.Value,
			Type:   paramType,
			Offset: 8 + i*8,
			Size:   g.getTypeSize(paramType),
		}
	}

	// Генерируем тело функции
	if fd.Body != nil {
		g.generateBlockStmt(fd.Body)
	}

	// Вставляем освобождение массивов и return если нужно
	if !g.currentBlock.IsTerminated() {
		g.emitArrayFrees()
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

func (g *IRGenerator) emitArrayFrees() {
	if len(g.arrayAllocs) == 0 {
		return
	}

	// Собираем имена массивов
	var names []string
	for name := range g.arrayAllocs {
		names = append(names, name)
	}

	// Освобождаем в обратном порядке
	for i := len(names) - 1; i >= 0; i-- {
		name := names[i]
		ptrOp := g.arrayAllocs[name]

		// PARAM для free
		g.currentBlock.AddInstruction(&Instruction{
			Opcode: OpParam,
			Src1:   NewLiteralOperand(0),
			Src2:   ptrOp,
		})

		// Вызов free
		freeCall := g.currentFunc.NewTemp()
		g.currentBlock.AddInstruction(&Instruction{
			Opcode:  OpCall,
			Dest:    freeCall,
			Src1:    NewVarOperand("free"),
			Args:    []*Operand{ptrOp},
			Comment: fmt.Sprintf("free(%s)", name),
		})

		delete(g.arrayAllocs, name)
	}
}

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
	case "pointer", "void*", "int[]", "float[]", "bool[]", "string[]":
		return 8
	default:
		return 8
	}
}

func (g *IRGenerator) newLabel(prefix string) string {
	g.labelCounter++
	return fmt.Sprintf("%s.%s_%d", g.currentFunc.Name, prefix, g.labelCounter)
}

func (g *IRGenerator) generateBlockStmt(block *ast.BlockStmtNode) {
	for _, stmt := range block.Statements {
		if g.currentBlock.IsTerminated() {
			break
		}
		g.generateStatement(stmt)
	}
}

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

func (g *IRGenerator) generateVarDecl(vd *ast.VarDeclNode) {
	varName := vd.Name.Value
	typeStr := vd.Type.String()

	if vd.Type.Kind == "array" {
		g.generateArrayVarDecl(vd)
		return
	}

	addrTemp := g.currentFunc.NewTemp()
	typeSize := g.getTypeSize(typeStr)

	g.currentBlock.AddInstruction(&Instruction{
		Opcode:  OpAlloca,
		Dest:    addrTemp,
		Src1:    NewLiteralOperand(typeSize),
		Comment: fmt.Sprintf("var %s: %s", varName, typeStr),
	})
	g.varAddresses[varName] = addrTemp

	offset := len(g.currentFunc.Locals) * 8
	g.currentFunc.Locals[varName] = &VarInfo{
		Name:   varName,
		Type:   typeStr,
		Offset: offset,
		Size:   typeSize,
	}

	if vd.Initializer != nil {
		initVal := g.generateExpression(vd.Initializer)
		if initVal != nil {
			g.currentBlock.AddInstruction(NewStoreInst(addrTemp, initVal))
		}
	}
}

func (g *IRGenerator) generateArrayVarDecl(vd *ast.VarDeclNode) {
	varName := vd.Name.Value
	baseType := vd.Type.BaseType

	if baseType == nil {
		return
	}

	elemSize := g.getTypeSize(baseType.Kind)
	if elemSize == 0 {
		elemSize = g.getTypeSize(baseType.String())
	}

	var totalSize int
	if vd.Type.ArraySize >= 0 {
		totalSize = vd.Type.ArraySize * elemSize
	} else if vd.Initializer != nil {
		if arrLit, ok := vd.Initializer.(*ast.ArrayLiteralExprNode); ok {
			totalSize = len(arrLit.Elements) * elemSize
		} else {
			totalSize = 8
		}
	} else {
		totalSize = 8
	}

	// Вызываем malloc(totalSize)
	sizeOp := NewLiteralOperand(totalSize)

	g.currentBlock.AddInstruction(&Instruction{
		Opcode: OpParam,
		Src1:   NewLiteralOperand(0),
		Src2:   sizeOp,
	})

	mallocResult := g.currentFunc.NewTemp()
	g.currentBlock.AddInstruction(&Instruction{
		Opcode:  OpCall,
		Dest:    mallocResult,
		Src1:    NewVarOperand("malloc"),
		Args:    []*Operand{sizeOp},
		Comment: fmt.Sprintf("arr %s = malloc(%d)", varName, totalSize),
	})

	// Сохраняем результат malloc как переменную
	g.varAddresses[varName] = mallocResult
	g.arrayAllocs[varName] = mallocResult

	g.currentFunc.Locals[varName] = &VarInfo{
		Name: varName,
		Type: fmt.Sprintf("%s[]", baseType.String()),
		Size: 8,
	}

	// Инициализация элементов массива
	if vd.Initializer != nil {
		if arrLit, ok := vd.Initializer.(*ast.ArrayLiteralExprNode); ok {
			for i, elem := range arrLit.Elements {
				elemVal := g.generateExpression(elem)
				if elemVal == nil {
					continue
				}

				idxOp := NewLiteralOperand(i)
				elemAddr := g.currentFunc.NewTemp()
				g.currentBlock.AddInstruction(NewGepInst(elemAddr, mallocResult, idxOp))
				g.currentBlock.AddInstruction(NewStoreInst(elemAddr, elemVal))
			}
		}
	}
}

func (g *IRGenerator) generateIfStmt(is *ast.IfStmtNode) {
	thenBlock := g.currentFunc.NewBlock(g.newLabel("if_then"))
	var elseBlock *BasicBlock
	if is.Alternative != nil {
		elseBlock = g.currentFunc.NewBlock(g.newLabel("if_else"))
	}
	mergeBlock := g.currentFunc.NewBlock(g.newLabel("if_merge"))

	savedVarAddresses := make(map[string]*Operand)
	for k, v := range g.varAddresses {
		savedVarAddresses[k] = v
	}

	if is.Alternative != nil {
		g.generateCondition(is.Condition, thenBlock, elseBlock)
	} else {
		g.generateCondition(is.Condition, thenBlock, mergeBlock)
	}

	if !g.currentBlock.IsTerminated() {
		g.currentBlock.AddSuccessor(thenBlock)
		thenBlock.AddPredecessor(g.currentBlock)
		if elseBlock != nil {
			g.currentBlock.AddSuccessor(elseBlock)
			elseBlock.AddPredecessor(g.currentBlock)
		} else {
			g.currentBlock.AddSuccessor(mergeBlock)
			mergeBlock.AddPredecessor(g.currentBlock)
		}
	}

	g.currentBlock = thenBlock
	g.generateStatement(is.Consequence)

	thenVarAddrs := make(map[string]*Operand)
	for k, v := range g.varAddresses {
		thenVarAddrs[k] = v
	}

	if !g.currentBlock.IsTerminated() {
		g.currentBlock.AddInstruction(NewJumpInst(mergeBlock))
		g.currentBlock.AddSuccessor(mergeBlock)
		mergeBlock.AddPredecessor(g.currentBlock)
	}

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

	g.currentBlock = mergeBlock
	for varName, thenAddr := range thenVarAddrs {
		elseAddr, exists := g.varAddresses[varName]
		if !exists || thenAddr.Name != elseAddr.Name {
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

func (g *IRGenerator) generateWhileStmt(ws *ast.WhileStmtNode) {
	headerBlock := g.currentFunc.NewBlock(g.newLabel("while_header"))
	bodyBlock := g.currentFunc.NewBlock(g.newLabel("while_body"))
	exitBlock := g.currentFunc.NewBlock(g.newLabel("while_exit"))

	g.breakTargets = append(g.breakTargets, exitBlock)
	g.continueTargets = append(g.continueTargets, headerBlock)

	g.currentBlock.AddInstruction(NewJumpInst(headerBlock))
	g.currentBlock.AddSuccessor(headerBlock)
	headerBlock.AddPredecessor(g.currentBlock)

	g.currentBlock = headerBlock
	g.generateCondition(ws.Condition, bodyBlock, exitBlock)

	if !g.currentBlock.IsTerminated() {
		g.currentBlock.AddSuccessor(bodyBlock)
		g.currentBlock.AddSuccessor(exitBlock)
		bodyBlock.AddPredecessor(g.currentBlock)
		exitBlock.AddPredecessor(g.currentBlock)
	}

	g.currentBlock = bodyBlock
	g.generateStatement(ws.Body)
	if !g.currentBlock.IsTerminated() {
		g.currentBlock.AddInstruction(NewJumpInst(headerBlock))
		g.currentBlock.AddSuccessor(headerBlock)
		headerBlock.AddPredecessor(g.currentBlock)
	}

	g.breakTargets = g.breakTargets[:len(g.breakTargets)-1]
	g.continueTargets = g.continueTargets[:len(g.continueTargets)-1]

	g.currentBlock = exitBlock
}

func (g *IRGenerator) generateForStmt(fs *ast.ForStmtNode) {
	if fs.Init != nil {
		g.generateStatement(fs.Init)
	}

	headerBlock := g.currentFunc.NewBlock(g.newLabel("for_header"))
	bodyBlock := g.currentFunc.NewBlock(g.newLabel("for_body"))
	updateBlock := g.currentFunc.NewBlock(g.newLabel("for_update"))
	exitBlock := g.currentFunc.NewBlock(g.newLabel("for_exit"))

	g.breakTargets = append(g.breakTargets, exitBlock)
	g.continueTargets = append(g.continueTargets, updateBlock)

	g.currentBlock.AddInstruction(NewJumpInst(headerBlock))
	g.currentBlock.AddSuccessor(headerBlock)
	headerBlock.AddPredecessor(g.currentBlock)

	g.currentBlock = headerBlock
	if fs.Condition != nil {
		g.generateCondition(fs.Condition, bodyBlock, exitBlock)
		if !g.currentBlock.IsTerminated() {
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

	g.currentBlock = bodyBlock
	g.generateStatement(fs.Body)
	if !g.currentBlock.IsTerminated() {
		g.currentBlock.AddInstruction(NewJumpInst(updateBlock))
		g.currentBlock.AddSuccessor(updateBlock)
		updateBlock.AddPredecessor(g.currentBlock)
	}

	g.currentBlock = updateBlock
	if fs.Update != nil {
		updateStmt := &ast.ExprStmtNode{
			Token:      fs.Token,
			Expression: fs.Update,
		}
		g.generateStatement(updateStmt)
	}
	g.currentBlock.AddInstruction(NewJumpInst(headerBlock))
	g.currentBlock.AddSuccessor(headerBlock)
	headerBlock.AddPredecessor(g.currentBlock)

	g.breakTargets = g.breakTargets[:len(g.breakTargets)-1]
	g.continueTargets = g.continueTargets[:len(g.continueTargets)-1]

	g.currentBlock = exitBlock
}

func (g *IRGenerator) generateCondition(expr ast.ExpressionNode, trueBlock, falseBlock *BasicBlock) {
	if binExpr, ok := expr.(*ast.BinaryExprNode); ok {
		switch binExpr.Operator {
		case "&&":
			midBlock := g.currentFunc.NewBlock(g.newLabel("sc_and"))
			g.generateCondition(binExpr.Left, midBlock, falseBlock)

			if !g.currentBlock.IsTerminated() {
				g.currentBlock.AddSuccessor(midBlock)
				midBlock.AddPredecessor(g.currentBlock)
			}

			g.currentBlock = midBlock
			g.generateCondition(binExpr.Right, trueBlock, falseBlock)
			return

		case "||":
			midBlock := g.currentFunc.NewBlock(g.newLabel("sc_or"))
			g.generateCondition(binExpr.Left, trueBlock, midBlock)

			if !g.currentBlock.IsTerminated() {
				g.currentBlock.AddSuccessor(midBlock)
				midBlock.AddPredecessor(g.currentBlock)
			}

			g.currentBlock = midBlock
			g.generateCondition(binExpr.Right, trueBlock, falseBlock)
			return
		}
	}

	val := g.generateExpression(expr)
	if val != nil {
		g.currentBlock.AddInstruction(NewCondJumpInst(OpJmpIf, val, trueBlock))
		g.currentBlock.AddInstruction(NewJumpInst(falseBlock))
	}
}

func (g *IRGenerator) generateReturnStmt(rs *ast.ReturnStmtNode) {
	var retVal *Operand
	if rs.RetValue != nil {
		retVal = g.generateExpression(rs.RetValue)
	}
	g.currentBlock.AddInstruction(NewRetInst(retVal))
}

func (g *IRGenerator) generateAssignment(ae *ast.AssignmentExprNode) {
	rightVal := g.generateExpression(ae.Right)
	if rightVal == nil {
		return
	}

	switch left := ae.Left.(type) {
	case *ast.IdentifierNode:
		if addr, exists := g.varAddresses[left.Value]; exists {
			g.currentBlock.AddInstruction(NewStoreInst(addr, rightVal))
		} else {
			g.currentBlock.AddInstruction(NewStoreInst(NewGlobalOperand(left.Value), rightVal))
		}

	case *ast.IndexExprNode:
		arrayPtr := g.generateExpression(left.Array)
		index := g.generateExpression(left.Index)
		if arrayPtr != nil && index != nil {
			elemAddr := g.currentFunc.NewTemp()
			g.currentBlock.AddInstruction(NewGepInst(elemAddr, arrayPtr, index))
			g.currentBlock.AddInstruction(NewStoreInst(elemAddr, rightVal))
		}
	}
}

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
		switch e.Operator {
		case "&&":
			return g.generateLogicalAnd(e)
		case "||":
			return g.generateLogicalOr(e)
		default:
			return g.generateBinaryExpr(e)
		}
	case *ast.UnaryExprNode:
		return g.generateUnaryExpr(e)
	case *ast.CallExprNode:
		return g.generateCallExpr(e)
	case *ast.AssignmentExprNode:
		return g.generateAssignmentExpr(e)
	case *ast.IndexExprNode:
		return g.generateIndexExpr(e)
	case *ast.ArrayLiteralExprNode:
		return g.generateArrayLiteral(e)
	default:
		return nil
	}
}

func (g *IRGenerator) generateIndexExpr(ie *ast.IndexExprNode) *Operand {
	arrayPtr := g.generateExpression(ie.Array)
	index := g.generateExpression(ie.Index)

	if arrayPtr == nil || index == nil {
		return nil
	}

	elemAddr := g.currentFunc.NewTemp()
	g.currentBlock.AddInstruction(NewGepInst(elemAddr, arrayPtr, index))

	result := g.currentFunc.NewTemp()
	g.currentBlock.AddInstruction(NewLoadInst(result, elemAddr))

	return result
}

func (g *IRGenerator) generateArrayLiteral(al *ast.ArrayLiteralExprNode) *Operand {
	size := len(al.Elements)
	elemSize := 4

	if size > 0 {
		if lit, ok := al.Elements[0].(*ast.LiteralExprNode); ok {
			switch lit.TypeName {
			case "float":
				elemSize = 8
			case "bool":
				elemSize = 1
			}
		}
	}

	totalSize := size * elemSize

	sizeOp := NewLiteralOperand(totalSize)
	g.currentBlock.AddInstruction(&Instruction{
		Opcode: OpParam,
		Src1:   NewLiteralOperand(0),
		Src2:   sizeOp,
	})

	arrayPtr := g.currentFunc.NewTemp()
	g.currentBlock.AddInstruction(&Instruction{
		Opcode:  OpCall,
		Dest:    arrayPtr,
		Src1:    NewVarOperand("malloc"),
		Args:    []*Operand{sizeOp},
		Comment: fmt.Sprintf("malloc(%d) for array literal", totalSize),
	})

	for i, elem := range al.Elements {
		elemVal := g.generateExpression(elem)
		if elemVal == nil {
			continue
		}

		idxOp := NewLiteralOperand(i)
		elemAddr := g.currentFunc.NewTemp()
		g.currentBlock.AddInstruction(NewGepInst(elemAddr, arrayPtr, idxOp))
		g.currentBlock.AddInstruction(NewStoreInst(elemAddr, elemVal))
	}

	return arrayPtr
}

func (g *IRGenerator) generateLogicalAnd(be *ast.BinaryExprNode) *Operand {
	leftVal := g.generateExpression(be.Left)
	if leftVal == nil {
		return nil
	}

	result := g.currentFunc.NewTemp()
	rightBlock := g.currentFunc.NewBlock(g.newLabel("and_right"))
	mergeBlock := g.currentFunc.NewBlock(g.newLabel("and_merge"))

	g.currentBlock.AddInstruction(NewCondJumpInst(OpJmpIfNot, leftVal, mergeBlock))
	g.currentBlock.AddInstruction(NewMoveInst(result, NewLiteralOperand(false)))
	g.currentBlock.AddInstruction(NewJumpInst(mergeBlock))

	g.currentBlock.AddSuccessor(mergeBlock)
	mergeBlock.AddPredecessor(g.currentBlock)

	g.currentBlock = rightBlock
	rightVal := g.generateExpression(be.Right)
	if rightVal == nil {
		return nil
	}

	if !g.currentBlock.IsTerminated() {
		g.currentBlock.AddInstruction(NewMoveInst(result, rightVal))
		g.currentBlock.AddInstruction(NewJumpInst(mergeBlock))
		g.currentBlock.AddSuccessor(mergeBlock)
		mergeBlock.AddPredecessor(g.currentBlock)
	}

	g.currentBlock = mergeBlock
	phiPairs := []PhiPair{
		{Value: NewLiteralOperand(false), Block: mergeBlock.Predecessors[0]},
		{Value: rightVal, Block: rightBlock},
	}
	g.currentBlock.AddInstruction(NewPhiInst(result, phiPairs))

	return result
}

func (g *IRGenerator) generateLogicalOr(be *ast.BinaryExprNode) *Operand {
	leftVal := g.generateExpression(be.Left)
	if leftVal == nil {
		return nil
	}

	result := g.currentFunc.NewTemp()
	rightBlock := g.currentFunc.NewBlock(g.newLabel("or_right"))
	mergeBlock := g.currentFunc.NewBlock(g.newLabel("or_merge"))

	g.currentBlock.AddInstruction(NewCondJumpInst(OpJmpIf, leftVal, mergeBlock))
	g.currentBlock.AddInstruction(NewMoveInst(result, NewLiteralOperand(true)))
	g.currentBlock.AddInstruction(NewJumpInst(mergeBlock))

	g.currentBlock.AddSuccessor(mergeBlock)
	mergeBlock.AddPredecessor(g.currentBlock)

	g.currentBlock = rightBlock
	rightVal := g.generateExpression(be.Right)
	if rightVal == nil {
		return nil
	}

	if !g.currentBlock.IsTerminated() {
		g.currentBlock.AddInstruction(NewMoveInst(result, rightVal))
		g.currentBlock.AddInstruction(NewJumpInst(mergeBlock))
		g.currentBlock.AddSuccessor(mergeBlock)
		mergeBlock.AddPredecessor(g.currentBlock)
	}

	g.currentBlock = mergeBlock
	phiPairs := []PhiPair{
		{Value: NewLiteralOperand(true), Block: mergeBlock.Predecessors[0]},
		{Value: rightVal, Block: rightBlock},
	}
	g.currentBlock.AddInstruction(NewPhiInst(result, phiPairs))

	return result
}

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

func (g *IRGenerator) generateIdentifier(ident *ast.IdentifierNode) *Operand {
	// Проверяем, массив ли это
	if _, isArray := g.arrayAllocs[ident.Value]; isArray {
		// Возвращаем указатель напрямую
		if addr, exists := g.varAddresses[ident.Value]; exists {
			return addr
		}
	}

	// Проверяем, параметр ли это
	for _, param := range g.currentFunc.Params {
		if param.Name == ident.Value {
			return NewVarOperand(ident.Value)
		}
	}

	// Локальная переменная
	if addr, exists := g.varAddresses[ident.Value]; exists {
		temp := g.currentFunc.NewTemp()
		g.currentBlock.AddInstruction(NewLoadInst(temp, addr))
		return temp
	}

	// Глобальная переменная
	temp := g.currentFunc.NewTemp()
	g.currentBlock.AddInstruction(NewLoadInst(temp, NewGlobalOperand(ident.Value)))
	return temp
}

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
	default:
		return nil
	}

	g.currentBlock.AddInstruction(NewBinaryInst(op, temp, left, right))
	return temp
}

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

func (g *IRGenerator) generateCallExpr(ce *ast.CallExprNode) *Operand {
	var funcName string
	if ident, ok := ce.Function.(*ast.IdentifierNode); ok {
		funcName = ident.Value
	} else {
		return nil
	}

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

func (g *IRGenerator) generateAssignmentExpr(ae *ast.AssignmentExprNode) *Operand {
	rightVal := g.generateExpression(ae.Right)
	if rightVal == nil {
		return nil
	}

	switch left := ae.Left.(type) {
	case *ast.IdentifierNode:
		if addr, exists := g.varAddresses[left.Value]; exists {
			g.currentBlock.AddInstruction(NewStoreInst(addr, rightVal))
			return rightVal
		} else {
			g.currentBlock.AddInstruction(NewStoreInst(NewGlobalOperand(left.Value), rightVal))
			return rightVal
		}

	case *ast.IndexExprNode:
		arrayPtr := g.generateExpression(left.Array)
		index := g.generateExpression(left.Index)
		if arrayPtr != nil && index != nil {
			elemAddr := g.currentFunc.NewTemp()
			g.currentBlock.AddInstruction(NewGepInst(elemAddr, arrayPtr, index))
			g.currentBlock.AddInstruction(NewStoreInst(elemAddr, rightVal))
		}
		return rightVal
	}

	return rightVal
}

func (g *IRGenerator) GetProgram() *Program {
	return g.program
}
