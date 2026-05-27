package codegen

import (
	"mikrocompiler/src/internal/ir"
)

// ExpressionGenerator генерирует x86-64 код для выражений с short-circuit поддержкой
type ExpressionGenerator struct {
	gen          *X86Generator
	labelManager *LabelManager
}

// NewExpressionGenerator создает новый генератор выражений
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

	// Загружаем первый операнд в eax
	eg.gen.loadToRegister("eax", inst.Src1)
	// Загружаем второй операнд в ecx
	eg.gen.loadToRegister("ecx", inst.Src2)
	// Выполняем операцию
	eg.gen.emit("    %s eax, ecx", op)
	// Сохраняем результат
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
	eg.gen.emit("    cmp eax, ecx")
	eg.gen.emit("    set%s al", condition)
	eg.gen.emit("    movzx eax, al")
	eg.gen.storeFromRegister("eax", inst.Dest)
}

// GenerateLogicalAnd генерирует код для short-circuit AND (&&)
func (eg *ExpressionGenerator) GenerateLogicalAnd(inst *ir.Instruction, falseLabel, endLabel string) {
	// Вычисляем левый операнд
	eg.gen.loadToRegister("eax", inst.Src1)

	// Если false, переходим к falseLabel (short-circuit)
	eg.gen.emit("    test eax, eax")
	eg.gen.emit("    jz %s", falseLabel)

	// Вычисляем правый операнд
	eg.gen.loadToRegister("eax", inst.Src2)

	// Сохраняем результат
	eg.gen.storeFromRegister("eax", inst.Dest)
	eg.gen.emit("    jmp %s", endLabel)
}

// GenerateLogicalOr генерирует код для short-circuit OR (||)
func (eg *ExpressionGenerator) GenerateLogicalOr(inst *ir.Instruction, trueLabel, endLabel string) {
	// Вычисляем левый операнд
	eg.gen.loadToRegister("eax", inst.Src1)

	// Если true, переходим к trueLabel (short-circuit)
	eg.gen.emit("    test eax, eax")
	eg.gen.emit("    jnz %s", trueLabel)

	// Вычисляем правый операнд
	eg.gen.loadToRegister("eax", inst.Src2)

	// Сохраняем результат
	eg.gen.storeFromRegister("eax", inst.Dest)
	eg.gen.emit("    jmp %s", endLabel)
}

// GenerateUnaryExpr генерирует код для унарного выражения
func (eg *ExpressionGenerator) GenerateUnaryExpr(inst *ir.Instruction, op string) {
	eg.gen.loadToRegister("eax", inst.Src1)
	eg.gen.emit("    %s eax", op)
	eg.gen.storeFromRegister("eax", inst.Dest)
}

// GenerateNotExpr генерирует код для логического NOT (!)
func (eg *ExpressionGenerator) GenerateNotExpr(inst *ir.Instruction) {
	eg.gen.loadToRegister("eax", inst.Src1)
	eg.gen.emit("    test eax, eax")
	eg.gen.emit("    sete al")
	eg.gen.emit("    movzx eax, al")
	eg.gen.storeFromRegister("eax", inst.Dest)
}

// GenerateDivModExpr генерирует код для деления и остатка
func (eg *ExpressionGenerator) GenerateDivModExpr(inst *ir.Instruction, isMod bool) {
	eg.gen.loadToRegister("eax", inst.Src1)
	eg.gen.emit("    cdq")
	eg.gen.loadToRegister("ecx", inst.Src2)
	eg.gen.emit("    idiv ecx")

	if isMod {
		eg.gen.storeFromRegister("edx", inst.Dest)
	} else {
		eg.gen.storeFromRegister("eax", inst.Dest)
	}
}

// GenerateLoadExpr генерирует код для загрузки из памяти
func (eg *ExpressionGenerator) GenerateLoadExpr(inst *ir.Instruction) {
	addr := eg.gen.getOperandAddr(inst.Src1)
	eg.gen.emit("    mov eax, [%s]", addr)
	eg.gen.storeFromRegister("eax", inst.Dest)
}

// GenerateStoreExpr генерирует код для сохранения в память
func (eg *ExpressionGenerator) GenerateStoreExpr(inst *ir.Instruction) {
	addr := eg.gen.getOperandAddr(inst.Src1)

	if inst.Src2.Type == ir.OperandLiteral {
		if val, ok := inst.Src2.GetIntValue(); ok {
			eg.gen.emit("    mov dword [%s], %d", addr, val)
		} else if val, ok := inst.Src2.GetBoolValue(); ok {
			if val {
				eg.gen.emit("    mov dword [%s], 1", addr)
			} else {
				eg.gen.emit("    mov dword [%s], 0", addr)
			}
		} else {
			eg.gen.emit("    mov dword [%s], %v", addr, inst.Src2.Value)
		}
	} else {
		eg.gen.loadToRegister("eax", inst.Src2)
		eg.gen.emit("    mov [%s], eax", addr)
	}
}

// GenerateMoveExpr генерирует код для перемещения данных
func (eg *ExpressionGenerator) GenerateMoveExpr(inst *ir.Instruction) {
	eg.gen.loadToRegister("eax", inst.Src1)
	eg.gen.storeFromRegister("eax", inst.Dest)
}

// GenerateCallExpr генерирует код для вызова функции
func (eg *ExpressionGenerator) GenerateCallExpr(inst *ir.Instruction) {
	// Обработка параметров
	for _, param := range eg.gen.pendingParams {
		paramIdx := 0
		if param.Src1 != nil {
			if val, ok := param.Src1.GetIntValue(); ok {
				paramIdx = val
			}
		}

		if paramIdx < 6 {
			reg := eg.gen.abi.IntParamRegs[paramIdx]
			eg.gen.loadToRegister(string(GetReg32(reg)), param.Src2)
		}
	}

	eg.gen.emit("    call %s", inst.Src1.Name)

	if inst.Dest != nil {
		eg.gen.storeFromRegister("eax", inst.Dest)
	}
}

// GeneratePhiExpr генерирует код для PHI узла
func (eg *ExpressionGenerator) GeneratePhiExpr(inst *ir.Instruction) {
	// В реальном SSA разрушении PHI заменяется на MOV в каждом предшественнике
	// Здесь используем первое значение как fallback
	if inst.Dest != nil && len(inst.PhiPairs) > 0 {
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

	// Загружаем операнды
	eg.gen.loadToRegister("eax", inst.Src1)
	eg.gen.loadToRegister("ecx", inst.Src2)
	eg.gen.emit("    cmp eax, ecx")

	if jumpTrue {
		eg.gen.emit("    j%s %s", condition, inst.Dest.Name)
	} else {
		// Инвертируем условие
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
