package codegen

import (
	"fmt"
	"mikrocompiler/src/internal/ir"
)

// ControlFlowGenerator генерирует x86-64 код для управляющих конструкций
type ControlFlowGenerator struct {
	gen           *X86Generator
	labelManager  *LabelManager
	exprGenerator *ExpressionGenerator
}

// NewControlFlowGenerator создает новый генератор control flow
func NewControlFlowGenerator(gen *X86Generator, labelManager *LabelManager, exprGen *ExpressionGenerator) *ControlFlowGenerator {
	return &ControlFlowGenerator{
		gen:           gen,
		labelManager:  labelManager,
		exprGenerator: exprGen,
	}
}

// GenerateConditionalJump генерирует условный переход
func (cfg *ControlFlowGenerator) GenerateConditionalJump(inst *ir.Instruction) {
	cfg.gen.emit("    test eax, eax")
	cfg.gen.emit("    jnz %s", inst.Src2.Name)
}

// GenerateConditionalJumpNot генерирует условный переход с отрицанием
func (cfg *ControlFlowGenerator) GenerateConditionalJumpNot(inst *ir.Instruction) {
	cfg.gen.emit("    test eax, eax")
	cfg.gen.emit("    jz %s", inst.Src2.Name)
}

// GenerateUnconditionalJump генерирует безусловный переход
func (cfg *ControlFlowGenerator) GenerateUnconditionalJump(inst *ir.Instruction) {
	cfg.gen.emit("    jmp %s", inst.Src1.Name)
}

// GenerateReturn генерирует код возврата из функции
func (cfg *ControlFlowGenerator) GenerateReturn(inst *ir.Instruction) {
	if inst.Src1 != nil {
		cfg.gen.loadToRegister("eax", inst.Src1)
	}

	cfg.gen.returnEmitted = true

	stackSize := cfg.gen.stackFrame.GetStackSize()
	if stackSize > 0 {
		cfg.gen.emit("    mov rsp, rbp")
	}
	cfg.gen.emit("    pop rbp")
	cfg.gen.emit("    ret")
}

// GenerateBasicBlock генерирует код для базового блока
func (cfg *ControlFlowGenerator) GenerateBasicBlock(block *ir.BasicBlock) {
	if block.Label != "entry" {
		cfg.gen.emit("%s:", block.Label)
	}

	for _, inst := range block.Instructions {
		cfg.generateInstruction(inst)
	}
}

// generateInstruction генерирует код для одной IR инструкции
func (cfg *ControlFlowGenerator) generateInstruction(inst *ir.Instruction) {
	switch inst.Opcode {
	case ir.OpAlloca:
		return

	case ir.OpParam:
		cfg.gen.pendingParams = append(cfg.gen.pendingParams, inst)
		return

	case ir.OpAdd, ir.OpSub, ir.OpMul, ir.OpAnd, ir.OpOr:
		cfg.exprGenerator.GenerateBinaryExpr(inst)

	case ir.OpDiv:
		cfg.exprGenerator.GenerateDivModExpr(inst, false)

	case ir.OpMod:
		cfg.exprGenerator.GenerateDivModExpr(inst, true)

	case ir.OpNeg:
		cfg.exprGenerator.GenerateUnaryExpr(inst, "neg")

	case ir.OpNot:
		cfg.exprGenerator.GenerateNotExpr(inst)

	case ir.OpCmpEq, ir.OpCmpNe, ir.OpCmpLt, ir.OpCmpLe, ir.OpCmpGt, ir.OpCmpGe:
		cfg.exprGenerator.GenerateComparison(inst)

	case ir.OpLoad:
		cfg.exprGenerator.GenerateLoadExpr(inst)

	case ir.OpStore:
		cfg.exprGenerator.GenerateStoreExpr(inst)

	case ir.OpJmp:
		cfg.GenerateUnconditionalJump(inst)

	case ir.OpJmpIf:
		cfg.GenerateConditionalJump(inst)

	case ir.OpJmpIfNot:
		cfg.GenerateConditionalJumpNot(inst)

	case ir.OpCall:
		if cfg.gen.isExternCall(inst.Src1.Name) {
			// Передаём pendingParams в inst.Args если их нет
			if len(inst.Args) == 0 && len(cfg.gen.pendingParams) > 0 {
				for _, p := range cfg.gen.pendingParams {
					if p.Src2 != nil {
						inst.Args = append(inst.Args, p.Src2)
					}
				}
			}
			cfg.gen.externGen.GenerateExternCall(inst)
		} else {
			cfg.exprGenerator.GenerateCallExpr(inst)
		}
		cfg.gen.pendingParams = nil

	case ir.OpRet:
		cfg.GenerateReturn(inst)

	case ir.OpMove:
		cfg.exprGenerator.GenerateMoveExpr(inst)

	case ir.OpPhi:
		cfg.exprGenerator.GeneratePhiExpr(inst)

	case ir.OpLabel:
		cfg.gen.emit("%s:", inst.Src1.Name)

	case ir.OpGep:
		cfg.gen.arrayGen.GenerateGEP(inst, 4)
		// Сохраняем размер результата GEP (всегда 8 байт — указатель)
		if inst.Dest != nil {
			cfg.gen.tempToAddrSize[inst.Dest.Name] = 8
		}

	default:
		cfg.gen.emit("    ; unknown opcode: %s", inst.Opcode.String())
	}
}

// GenerateFunctionPrologue генерирует пролог функции
func (cfg *ControlFlowGenerator) GenerateFunctionPrologue(fn *ir.Function) {
	cfg.gen.emit("    push rbp")
	cfg.gen.emit("    mov rbp, rsp")

	stackSize := cfg.gen.stackFrame.GetStackSize()
	if stackSize%16 != 0 {
		stackSize += 16 - (stackSize % 16)
	}
	if stackSize == 0 && cfg.gen.hasFunctionCalls(fn) {
		stackSize = 16
	}
	if stackSize > 0 {
		cfg.gen.emit("    sub rsp, %d", stackSize)
	}
}

// GenerateFunctionEpilogue генерирует эпилог функции
func (cfg *ControlFlowGenerator) GenerateFunctionEpilogue() {
	stackSize := cfg.gen.stackFrame.GetStackSize()
	if stackSize > 0 {
		cfg.gen.emit("    mov rsp, rbp")
	}
	cfg.gen.emit("    pop rbp")
	cfg.gen.emit("    ret")
}

func (cfg *ControlFlowGenerator) String() string {
	return fmt.Sprintf("ControlFlowGenerator")
}
