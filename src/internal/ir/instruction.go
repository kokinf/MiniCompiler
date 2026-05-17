package ir

import (
	"fmt"
	"strings"
)

// Opcode определяет код операции
type Opcode int

const (
	// Арифметические операции
	OpAdd Opcode = iota
	OpSub
	OpMul
	OpDiv
	OpMod
	OpNeg

	// Логические операции
	OpAnd
	OpOr
	OpNot
	OpXor

	// Сравнения
	OpCmpEq
	OpCmpNe
	OpCmpLt
	OpCmpLe
	OpCmpGt
	OpCmpGe

	// Работа с памятью
	OpLoad
	OpStore
	OpAlloca
	OpGep

	// Управление потоком
	OpJmp
	OpJmpIf
	OpJmpIfNot
	OpLabel
	OpPhi

	// Функции
	OpCall
	OpRet
	OpParam

	// Перемещение данных
	OpMove
)

var opcodeNames = map[Opcode]string{
	OpAdd: "ADD", OpSub: "SUB", OpMul: "MUL", OpDiv: "DIV", OpMod: "MOD", OpNeg: "NEG",
	OpAnd: "AND", OpOr: "OR", OpNot: "NOT", OpXor: "XOR",
	OpCmpEq: "CMP_EQ", OpCmpNe: "CMP_NE", OpCmpLt: "CMP_LT", OpCmpLe: "CMP_LE",
	OpCmpGt: "CMP_GT", OpCmpGe: "CMP_GE",
	OpLoad: "LOAD", OpStore: "STORE", OpAlloca: "ALLOCA", OpGep: "GEP",
	OpJmp: "JUMP", OpJmpIf: "JUMP_IF", OpJmpIfNot: "JUMP_IF_NOT", OpLabel: "LABEL", OpPhi: "PHI",
	OpCall: "CALL", OpRet: "RETURN", OpParam: "PARAM",
	OpMove: "MOVE",
}

func (o Opcode) String() string {
	if name, ok := opcodeNames[o]; ok {
		return name
	}
	return "UNKNOWN"
}

// Instruction представляет одну IR инструкцию
type Instruction struct {
	Opcode   Opcode
	Dest     *Operand
	Src1     *Operand
	Src2     *Operand
	PhiPairs []PhiPair
	Args     []*Operand
	Comment  string
	Line     int
	Column   int
}

// PhiPair представляет пару (значение, блок) для PHI узла
type PhiPair struct {
	Value *Operand
	Block *BasicBlock
}

// NewBinaryInst создает бинарную инструкцию
func NewBinaryInst(op Opcode, dest, src1, src2 *Operand) *Instruction {
	return &Instruction{
		Opcode: op,
		Dest:   dest,
		Src1:   src1,
		Src2:   src2,
	}
}

// NewUnaryInst создает унарную инструкцию
func NewUnaryInst(op Opcode, dest, src *Operand) *Instruction {
	return &Instruction{
		Opcode: op,
		Dest:   dest,
		Src1:   src,
	}
}

// NewJumpInst создает инструкцию безусловного перехода
func NewJumpInst(target *BasicBlock) *Instruction {
	return &Instruction{
		Opcode: OpJmp,
		Src1:   NewLabelOperand(target.Label),
	}
}

// NewCondJumpInst создает инструкцию условного перехода
func NewCondJumpInst(op Opcode, cond *Operand, target *BasicBlock) *Instruction {
	return &Instruction{
		Opcode: op,
		Src1:   cond,
		Src2:   NewLabelOperand(target.Label),
	}
}

// NewLabelInst создает инструкцию-метку
func NewLabelInst(label string) *Instruction {
	return &Instruction{
		Opcode: OpLabel,
		Src1:   NewLabelOperand(label),
	}
}

// NewPhiInst создает PHI инструкцию
func NewPhiInst(dest *Operand, pairs []PhiPair) *Instruction {
	return &Instruction{
		Opcode:   OpPhi,
		Dest:     dest,
		PhiPairs: pairs,
	}
}

// NewCallInst создает инструкцию вызова функции
func NewCallInst(dest *Operand, funcName string, args []*Operand) *Instruction {
	return &Instruction{
		Opcode: OpCall,
		Dest:   dest,
		Src1:   NewVarOperand(funcName),
		Args:   args,
	}
}

// NewRetInst создает инструкцию возврата
func NewRetInst(value *Operand) *Instruction {
	return &Instruction{
		Opcode: OpRet,
		Src1:   value,
	}
}

// NewStoreInst создает инструкцию сохранения в память
func NewStoreInst(addr, value *Operand) *Instruction {
	return &Instruction{
		Opcode: OpStore,
		Src1:   addr,
		Src2:   value,
	}
}

// NewLoadInst создает инструкцию загрузки из памяти
func NewLoadInst(dest, addr *Operand) *Instruction {
	return &Instruction{
		Opcode: OpLoad,
		Dest:   dest,
		Src1:   addr,
	}
}

// NewMoveInst создает инструкцию перемещения
func NewMoveInst(dest, src *Operand) *Instruction {
	return &Instruction{
		Opcode: OpMove,
		Dest:   dest,
		Src1:   src,
	}
}

// NewGepInst создает инструкцию Get Element Pointer
func NewGepInst(dest, base, index *Operand) *Instruction {
	return &Instruction{
		Opcode: OpGep,
		Dest:   dest,
		Src1:   base,
		Src2:   index,
	}
}

// GetUsedOperands возвращает все используемые операнды
func (i *Instruction) GetUsedOperands() []*Operand {
	var operands []*Operand

	if i.Src1 != nil && i.Src1.Type != OperandLabel && i.Src1.Type != OperandLiteral {
		operands = append(operands, i.Src1)
	}
	if i.Src2 != nil && i.Src2.Type != OperandLabel && i.Src2.Type != OperandLiteral {
		operands = append(operands, i.Src2)
	}
	for _, arg := range i.Args {
		if arg != nil && arg.Type != OperandLiteral {
			operands = append(operands, arg)
		}
	}
	for _, pair := range i.PhiPairs {
		if pair.Value != nil {
			operands = append(operands, pair.Value)
		}
	}

	return operands
}

func (i *Instruction) String() string {
	var sb strings.Builder

	switch i.Opcode {
	case OpLabel:
		fmt.Fprintf(&sb, "%s:", i.Src1)
	case OpPhi:
		fmt.Fprintf(&sb, "%s = PHI ", i.Dest)
		pairs := make([]string, len(i.PhiPairs))
		for idx, p := range i.PhiPairs {
			pairs[idx] = fmt.Sprintf("(%s, %s)", p.Value, p.Block.Label)
		}
		sb.WriteString(strings.Join(pairs, ", "))
	case OpJmp:
		fmt.Fprintf(&sb, "JUMP %s", i.Src1)
	case OpJmpIf:
		fmt.Fprintf(&sb, "JUMP_IF %s, %s", i.Src1, i.Src2)
	case OpJmpIfNot:
		fmt.Fprintf(&sb, "JUMP_IF_NOT %s, %s", i.Src1, i.Src2)
	case OpCall:
		if i.Dest != nil {
			fmt.Fprintf(&sb, "%s = ", i.Dest)
		}
		fmt.Fprintf(&sb, "CALL %s(", i.Src1)
		args := make([]string, len(i.Args))
		for idx, arg := range i.Args {
			args[idx] = arg.String()
		}
		sb.WriteString(strings.Join(args, ", "))
		sb.WriteString(")")
	case OpRet:
		sb.WriteString("RETURN")
		if i.Src1 != nil {
			fmt.Fprintf(&sb, " %s", i.Src1)
		}
	case OpStore:
		fmt.Fprintf(&sb, "STORE [%s], %s", i.Src1, i.Src2)
	case OpLoad:
		fmt.Fprintf(&sb, "%s = LOAD [%s]", i.Dest, i.Src1)
	case OpAlloca:
		fmt.Fprintf(&sb, "%s = ALLOCA %s", i.Dest, i.Src1)
	case OpGep:
		fmt.Fprintf(&sb, "%s = GEP %s, %s", i.Dest, i.Src1, i.Src2)
	case OpParam:
		fmt.Fprintf(&sb, "PARAM %s, %s", i.Src1, i.Src2)
	case OpMove:
		fmt.Fprintf(&sb, "%s = MOVE %s", i.Dest, i.Src1)
	default:
		if i.Dest != nil {
			fmt.Fprintf(&sb, "%s = ", i.Dest)
		}
		fmt.Fprintf(&sb, "%s", i.Opcode)
		if i.Src1 != nil {
			fmt.Fprintf(&sb, " %s", i.Src1)
		}
		if i.Src2 != nil {
			fmt.Fprintf(&sb, ", %s", i.Src2)
		}
	}

	if i.Comment != "" {
		fmt.Fprintf(&sb, "  ; %s", i.Comment)
	}

	return sb.String()
}

// IsTerminator возвращает true, если инструкция завершает базовый блок
func (i *Instruction) IsTerminator() bool {
	switch i.Opcode {
	case OpJmp, OpJmpIf, OpJmpIfNot, OpRet:
		return true
	default:
		return false
	}
}

// IsBranch возвращает true, если инструкция - условный переход
func (i *Instruction) IsBranch() bool {
	return i.Opcode == OpJmpIf || i.Opcode == OpJmpIfNot
}

// GetBranchTargets возвращает цели переходов
func (i *Instruction) GetBranchTargets() []string {
	switch i.Opcode {
	case OpJmp:
		return []string{i.Src1.Name}
	case OpJmpIf, OpJmpIfNot:
		return []string{i.Src2.Name}
	default:
		return nil
	}
}
