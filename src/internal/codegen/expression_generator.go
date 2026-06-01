// src/internal/codegen/expression_generator.go
package codegen

import (
	"mikrocompiler/src/internal/ir"
)

type ExpressionGenerator struct {
	gen          *X86Generator
	labelManager *LabelManager
}

func NewExpressionGenerator(gen *X86Generator, labelManager *LabelManager) *ExpressionGenerator {
	return &ExpressionGenerator{
		gen:          gen,
		labelManager: labelManager,
	}
}

// GenerateBinaryExpr генерирует код для бинарного выражения
func (eg *ExpressionGenerator) GenerateBinaryExpr(inst *ir.Instruction) {
	opMap := map[ir.Opcode]string{
		ir.OpAdd: "add",
		ir.OpSub: "sub",
		ir.OpMul: "imul",
		ir.OpAnd: "and",
		ir.OpOr:  "or",
	}

	op, ok := opMap[inst.Opcode]
	if !ok {
		return
	}

	eg.gen.loadToRegister("eax", inst.Src1)
	eg.gen.loadToRegister("ecx", inst.Src2)
	eg.gen.emit("    %s eax, ecx                          ; %s = %s %s, %s", op, inst.Dest, inst.Opcode, inst.Src1, inst.Src2)
	eg.gen.storeFromRegister("eax", inst.Dest)
}

// GenerateComparison генерирует код для сравнения
func (eg *ExpressionGenerator) GenerateComparison(inst *ir.Instruction) {
	condMap := map[ir.Opcode]string{
		ir.OpCmpEq: "e",
		ir.OpCmpNe: "ne",
		ir.OpCmpLt: "l",
		ir.OpCmpLe: "le",
		ir.OpCmpGt: "g",
		ir.OpCmpGe: "ge",
	}

	condition, ok := condMap[inst.Opcode]
	if !ok {
		return
	}

	eg.gen.loadToRegister("eax", inst.Src1)
	eg.gen.loadToRegister("ecx", inst.Src2)
	eg.gen.emit("    cmp eax, ecx                          ; %s = %s %s, %s", inst.Dest, inst.Opcode, inst.Src1, inst.Src2)
	eg.gen.emit("    set%s al", condition)
	eg.gen.emit("    movzx eax, al")
	eg.gen.storeFromRegister("eax", inst.Dest)
}

// GenerateLogicalAnd генерирует код для short-circuit AND (&&)
func (eg *ExpressionGenerator) GenerateLogicalAnd(inst *ir.Instruction, falseLabel, endLabel string) {
	eg.gen.loadToRegister("eax", inst.Src1)
	eg.gen.emit("    test eax, eax                         ; short-circuit AND: %s = %s && %s", inst.Dest, inst.Src1, inst.Src2)
	eg.gen.emit("    jz %s", falseLabel)
	eg.gen.loadToRegister("eax", inst.Src2)
	eg.gen.storeFromRegister("eax", inst.Dest)
	eg.gen.emit("    jmp %s", endLabel)
}

// GenerateLogicalOr генерирует код для short-circuit OR (||)
func (eg *ExpressionGenerator) GenerateLogicalOr(inst *ir.Instruction, trueLabel, endLabel string) {
	eg.gen.loadToRegister("eax", inst.Src1)
	eg.gen.emit("    test eax, eax                         ; short-circuit OR: %s = %s || %s", inst.Dest, inst.Src1, inst.Src2)
	eg.gen.emit("    jnz %s", trueLabel)
	eg.gen.loadToRegister("eax", inst.Src2)
	eg.gen.storeFromRegister("eax", inst.Dest)
	eg.gen.emit("    jmp %s", endLabel)
}

// GenerateUnaryExpr генерирует код для унарного выражения
func (eg *ExpressionGenerator) GenerateUnaryExpr(inst *ir.Instruction, op string) {
	eg.gen.loadToRegister("eax", inst.Src1)
	eg.gen.emit("    %s eax                                 ; %s = %s %s", op, inst.Dest, inst.Opcode, inst.Src1)
	eg.gen.storeFromRegister("eax", inst.Dest)
}

// GenerateNotExpr генерирует код для логического NOT (!)
func (eg *ExpressionGenerator) GenerateNotExpr(inst *ir.Instruction) {
	eg.gen.loadToRegister("eax", inst.Src1)
	eg.gen.emit("    test eax, eax                         ; %s = NOT %s", inst.Dest, inst.Src1)
	eg.gen.emit("    sete al")
	eg.gen.emit("    movzx eax, al")
	eg.gen.storeFromRegister("eax", inst.Dest)
}

// GenerateDivModExpr генерирует код для деления и остатка
func (eg *ExpressionGenerator) GenerateDivModExpr(inst *ir.Instruction, isMod bool) {
	eg.gen.loadToRegister("eax", inst.Src1)
	eg.gen.emit("    cdq")
	eg.gen.loadToRegister("ecx", inst.Src2)
	if isMod {
		eg.gen.emit("    idiv ecx                               ; %s = %s MOD %s", inst.Dest, inst.Src1, inst.Src2)
		eg.gen.storeFromRegister("edx", inst.Dest)
	} else {
		eg.gen.emit("    idiv ecx                               ; %s = %s DIV %s", inst.Dest, inst.Src1, inst.Src2)
		eg.gen.storeFromRegister("eax", inst.Dest)
	}
}

// GenerateLoadExpr генерирует код для загрузки из памяти
func (eg *ExpressionGenerator) GenerateLoadExpr(inst *ir.Instruction) {
	if inst.Src1 == nil || inst.Dest == nil {
		return
	}

	loadSize := eg.getOperandSize(inst.Src1)
	addr := eg.gen.getOperandAddr(inst.Src1)

	if loadSize == 8 {
		eg.gen.emit("    mov rax, [%s]                          ; %s = LOAD [%s] (загружаем указатель)", addr, inst.Dest, inst.Src1)
		eg.gen.emit("    mov eax, [rax]                         ; разыменовываем указатель -> значение")
		eg.gen.storeFromRegister("eax", inst.Dest)
	} else {
		eg.gen.emit("    mov eax, [%s]                          ; %s = LOAD [%s]", addr, inst.Dest, inst.Src1)
		eg.gen.storeFromRegister("eax", inst.Dest)
	}
}

// GenerateStoreExpr генерирует код для сохранения в память
func (eg *ExpressionGenerator) GenerateStoreExpr(inst *ir.Instruction) {
	if inst.Src1 == nil {
		return
	}

	storeSize := eg.getOperandSize(inst.Src1)

	if storeSize == 8 {
		ptrAddr := eg.gen.getOperandAddr(inst.Src1)
		eg.gen.emit("    mov rcx, [%s]                          ; STORE [%s], %s (загружаем адрес)", ptrAddr, inst.Src1, inst.Src2)

		if inst.Src2.Type == ir.OperandLiteral {
			if val, ok := inst.Src2.GetIntValue(); ok {
				eg.gen.emit("    mov dword [rcx], %d                      ; сохраняем значение по указателю", val)
			} else {
				eg.gen.loadToRegister("eax", inst.Src2)
				eg.gen.emit("    mov [rcx], eax                         ; сохраняем значение по указателю")
			}
		} else {
			eg.gen.loadToRegister("eax", inst.Src2)
			eg.gen.emit("    mov [rcx], eax                         ; сохраняем значение по указателю")
		}
		return
	}

	addr := eg.gen.getOperandAddr(inst.Src1)
	eg.gen.emit("    ; STORE [%s], %s", inst.Src1, inst.Src2)
	if inst.Src2.Type == ir.OperandLiteral {
		if val, ok := inst.Src2.GetIntValue(); ok {
			eg.gen.emit("    mov dword [%s], %d", addr, val)
		} else {
			eg.gen.loadToRegister("eax", inst.Src2)
			eg.gen.emit("    mov [%s], eax", addr)
		}
	} else {
		eg.gen.loadToRegister("eax", inst.Src2)
		eg.gen.emit("    mov [%s], eax", addr)
	}
}

// GenerateMoveExpr генерирует код для перемещения данных
func (eg *ExpressionGenerator) GenerateMoveExpr(inst *ir.Instruction) {
	if inst.Src1 == nil || inst.Dest == nil {
		return
	}

	moveSize := eg.getOperandSize(inst.Src1)
	if destSize := eg.getOperandSize(inst.Dest); destSize > moveSize {
		moveSize = destSize
	}

	if moveSize == 8 {
		eg.gen.loadToRegister("rax", inst.Src1)
		eg.gen.storeFromRegister("rax", inst.Dest)
	} else {
		eg.gen.loadToRegister("eax", inst.Src1)
		eg.gen.storeFromRegister("eax", inst.Dest)
	}
}

func (eg *ExpressionGenerator) getOperandSize(op *ir.Operand) int {
	return eg.gen.getOperandSize(op)
}

// GenerateCallExpr генерирует код для вызова функции
func (eg *ExpressionGenerator) GenerateCallExpr(inst *ir.Instruction) {
	funcName := inst.Src1.Name

	var fn *ir.Function
	for _, f := range eg.gen.program.Functions {
		if f.Name == funcName {
			fn = f
			break
		}
	}

	for _, param := range eg.gen.pendingParams {
		paramIdx := 0
		if param.Src1 != nil {
			if val, ok := param.Src1.GetIntValue(); ok {
				paramIdx = val
			}
		}

		if paramIdx < 6 {
			reg := eg.gen.abi.IntParamRegs[paramIdx]
			use64bit := false
			if fn != nil && paramIdx < len(fn.Params) {
				pType := fn.Params[paramIdx].Type
				if pType == "int[]" || pType == "float[]" || pType == "bool[]" ||
					pType == "string[]" || pType == "pointer" || pType == "void*" ||
					pType == "string" {
					use64bit = true
				}
			}

			if use64bit {
				eg.gen.loadToRegister(string(reg), param.Src2)
			} else {
				eg.gen.loadToRegister(string(GetReg32(reg)), param.Src2)
			}
		}
	}

	if inst.Dest != nil {
		eg.gen.emit("    call %s                                ; %s = CALL %s", funcName, inst.Dest, funcName)
	} else {
		eg.gen.emit("    call %s", funcName)
	}

	if inst.Dest != nil {
		eg.gen.storeFromRegister("eax", inst.Dest)
	}
}

// GeneratePhiExpr генерирует код для PHI узла
func (eg *ExpressionGenerator) GeneratePhiExpr(inst *ir.Instruction) {
	if inst.Dest != nil && len(inst.PhiPairs) > 0 {
		eg.gen.emit("    ; PHI %s", inst.Dest)
		eg.gen.loadToRegister("eax", inst.PhiPairs[0].Value)
		eg.gen.storeFromRegister("eax", inst.Dest)
	}
}

// EmitCompareAndJump генерирует прямое сравнение с условным переходом
func (eg *ExpressionGenerator) EmitCompareAndJump(inst *ir.Instruction, jumpTrue bool) {
	condMap := map[ir.Opcode]string{
		ir.OpCmpEq: "e",
		ir.OpCmpNe: "ne",
		ir.OpCmpLt: "l",
		ir.OpCmpLe: "le",
		ir.OpCmpGt: "g",
		ir.OpCmpGe: "ge",
	}

	condition, ok := condMap[inst.Opcode]
	if !ok {
		return
	}

	eg.gen.loadToRegister("eax", inst.Src1)
	eg.gen.loadToRegister("ecx", inst.Src2)
	eg.gen.emit("    cmp eax, ecx                          ; сравнение для условного перехода")

	if jumpTrue {
		eg.gen.emit("    j%s %s", condition, inst.Dest.Name)
	} else {
		inverseCond := map[string]string{
			"e": "ne", "ne": "e",
			"l": "ge", "le": "g",
			"g": "le", "ge": "l",
		}
		if inv, ok := inverseCond[condition]; ok {
			eg.gen.emit("    j%s %s", inv, inst.Dest.Name)
		}
	}
}
