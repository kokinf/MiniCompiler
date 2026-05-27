package ir

import (
	"fmt"
	"strings"
)

// PeepholeOptimizer выполняет простые локальные оптимизации
type PeepholeOptimizer struct {
	program    *Program
	changes    []string
	iterations int
}

// NewPeepholeOptimizer создает новый оптимизатор
func NewPeepholeOptimizer(program *Program) *PeepholeOptimizer {
	return &PeepholeOptimizer{
		program: program,
		changes: make([]string, 0),
	}
}

// Optimize выполняет все оптимизации итеративно
func (o *PeepholeOptimizer) Optimize() {
	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		changed := false
		o.iterations++

		for _, fn := range o.program.Functions {
			for _, block := range fn.Blocks {
				if o.optimizeBlock(block, fn) {
					changed = true
				}
			}
		}

		if !changed {
			break
		}
	}
}

// isComparison проверяет, является ли опкод сравнением
func isComparison(op Opcode) bool {
	switch op {
	case OpCmpEq, OpCmpNe, OpCmpLt, OpCmpLe, OpCmpGt, OpCmpGe:
		return true
	}
	return false
}

// optimizeBlock оптимизирует один базовый блок
func (o *PeepholeOptimizer) optimizeBlock(block *BasicBlock, fn *Function) bool {
	changed := false
	newInsts := make([]*Instruction, 0, len(block.Instructions))

	for i := 0; i < len(block.Instructions); i++ {
		inst := block.Instructions[i]

		// Оптимизация: CMP + JUMP_IF -> эффективный условный переход
		if isComparison(inst.Opcode) && i < len(block.Instructions)-1 {
			nextInst := block.Instructions[i+1]
			if (nextInst.Opcode == OpJmpIf || nextInst.Opcode == OpJmpIfNot) && inst.Dest != nil && nextInst.Src1 != nil {
				// Проверяем, что результат сравнения используется только для перехода
				if inst.Dest.Name == nextInst.Src1.Name {
					// Проверяем, что результат сравнения не используется где-то ещё
					usedElsewhere := false
					for _, b := range fn.Blocks {
						for _, use := range b.Uses[inst.Dest.Name] {
							if use != nextInst {
								usedElsewhere = true
								break
							}
						}
						if usedElsewhere {
							break
						}
					}

					if !usedElsewhere {
						// Объединяем сравнение и условный переход
						condMap := map[Opcode]string{
							OpCmpEq: "e", OpCmpNe: "ne",
							OpCmpLt: "l", OpCmpLe: "le",
							OpCmpGt: "g", OpCmpGe: "ge",
						}

						condition, ok := condMap[inst.Opcode]
						if ok {
							// Создаем комбинированную инструкцию
							mergedOp := OpJmpIf
							if nextInst.Opcode == OpJmpIfNot {
								mergedOp = OpJmpIfNot
							}

							merged := &Instruction{
								Opcode: mergedOp,
								Src1:   inst.Src1,
								Src2:   inst.Src2,
								Comment: fmt.Sprintf("optimized: cmp %s + jump -> conditional jump %s %s",
									condition, condition, nextInst.Src2.Name),
								Line:   inst.Line,
								Column: inst.Column,
							}

							newInsts = append(newInsts, merged)
							// Добавляем безусловный переход к цели, если нужно
							if nextInst.Opcode == OpJmpIf {
								// После условного перехода нужен безусловный
								jmpInst := &Instruction{
									Opcode:  OpJmp,
									Src1:    NewLabelOperand(nextInst.Src2.Name),
									Comment: "fallthrough jump",
								}
								newInsts = append(newInsts, jmpInst)
							}

							o.changes = append(o.changes, fmt.Sprintf(
								"Optimized: merged CMP + JUMP in %s: %s",
								fn.Name, merged.Comment))
							changed = true
							i++ // Пропускаем следующую инструкцию
							continue
						}
					}
				}
			}
		}

		// Алгебраические упрощения
		if simplified := o.tryAlgebraicSimplification(inst, fn); simplified != nil {
			newInsts = append(newInsts, simplified)
			o.changes = append(o.changes, fmt.Sprintf("Algebraic simplification in %s: %s -> %s",
				fn.Name, inst.String(), simplified.String()))
			changed = true
			continue
		}

		// Свёртка констант
		if folded := o.tryConstantFolding(inst, fn); folded != nil {
			newInsts = append(newInsts, folded)
			o.changes = append(o.changes, fmt.Sprintf("Constant folding in %s: %s -> %s",
				fn.Name, inst.String(), folded.String()))
			changed = true
			continue
		}

		// Удаление мёртвого кода (MOVE x, x)
		if inst.Opcode == OpMove {
			if inst.Dest.String() == inst.Src1.String() {
				o.changes = append(o.changes, fmt.Sprintf("Removed dead move in %s: %s",
					fn.Name, inst.String()))
				changed = true
				continue
			}
		}

		// Strength reduction: MUL x, 2 -> ADD x, x
		if inst.Opcode == OpMul {
			if val, ok := inst.Src2.GetIntValue(); ok && val == 2 {
				newInst := NewBinaryInst(OpAdd, inst.Dest, inst.Src1, inst.Src1)
				newInst.Comment = "strength reduction: *2 -> +"
				newInsts = append(newInsts, newInst)
				o.changes = append(o.changes, fmt.Sprintf("Strength reduction in %s: *2 -> +", fn.Name))
				changed = true
				continue
			}
		}

		// Dead code elimination: удаление инструкций, результат которых не используется
		if o.isDeadCode(inst, block, i, fn) {
			o.changes = append(o.changes, fmt.Sprintf("Removed dead code in %s: %s",
				fn.Name, inst.String()))
			changed = true
			continue
		}

		// Jump chaining
		if i < len(block.Instructions)-1 {
			if inst.Opcode == OpJmp {
				nextInst := block.Instructions[i+1]
				if nextInst.Opcode == OpLabel && inst.Src1.Name == nextInst.Src1.Name {
					o.changes = append(o.changes, fmt.Sprintf("Removed jump to next label in %s", fn.Name))
					changed = true
					continue
				}
			}
		}

		newInsts = append(newInsts, inst)
	}

	block.Instructions = newInsts
	return changed
}

// tryAlgebraicSimplification пытается упростить алгебраические выражения
func (o *PeepholeOptimizer) tryAlgebraicSimplification(inst *Instruction, fn *Function) *Instruction {
	if inst.Opcode != OpAdd && inst.Opcode != OpSub &&
		inst.Opcode != OpMul && inst.Opcode != OpDiv {
		return nil
	}

	// x + 0 => x
	if inst.Opcode == OpAdd && isZero(inst.Src2) {
		return NewMoveInst(inst.Dest, inst.Src1)
	}

	// 0 + x => x
	if inst.Opcode == OpAdd && isZero(inst.Src1) {
		return NewMoveInst(inst.Dest, inst.Src2)
	}

	// x - 0 => x
	if inst.Opcode == OpSub && isZero(inst.Src2) {
		return NewMoveInst(inst.Dest, inst.Src1)
	}

	// x * 1 => x
	if inst.Opcode == OpMul && isOne(inst.Src2) {
		return NewMoveInst(inst.Dest, inst.Src1)
	}

	// 1 * x => x
	if inst.Opcode == OpMul && isOne(inst.Src1) {
		return NewMoveInst(inst.Dest, inst.Src2)
	}

	// x * 0 => 0
	if inst.Opcode == OpMul && isZero(inst.Src2) {
		return NewMoveInst(inst.Dest, NewLiteralOperand(0))
	}

	// 0 * x => 0
	if inst.Opcode == OpMul && isZero(inst.Src1) {
		return NewMoveInst(inst.Dest, NewLiteralOperand(0))
	}

	// x / 1 => x
	if inst.Opcode == OpDiv && isOne(inst.Src2) {
		return NewMoveInst(inst.Dest, inst.Src1)
	}

	// x - x => 0
	if inst.Opcode == OpSub && inst.Src1.String() == inst.Src2.String() {
		return NewMoveInst(inst.Dest, NewLiteralOperand(0))
	}

	return nil
}

// tryConstantFolding сворачивает константные выражения
func (o *PeepholeOptimizer) tryConstantFolding(inst *Instruction, fn *Function) *Instruction {
	if !inst.Src1.IsConstant() {
		return nil
	}

	// Унарные операции
	if inst.Src2 == nil {
		switch inst.Opcode {
		case OpNeg:
			if val, ok := inst.Src1.GetIntValue(); ok {
				return NewMoveInst(inst.Dest, NewLiteralOperand(-val))
			}
			if val, ok := inst.Src1.GetFloatValue(); ok {
				return NewMoveInst(inst.Dest, NewLiteralOperand(-val))
			}
		case OpNot:
			if val, ok := inst.Src1.GetBoolValue(); ok {
				return NewMoveInst(inst.Dest, NewLiteralOperand(!val))
			}
		}
		return nil
	}

	if !inst.Src2.IsConstant() {
		return nil
	}

	// Бинарные операции
	switch inst.Opcode {
	case OpAdd:
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v2, ok2 := inst.Src2.GetIntValue(); ok2 {
				return NewMoveInst(inst.Dest, NewLiteralOperand(v1+v2))
			}
		}
		if v1, ok1 := inst.Src1.GetFloatValue(); ok1 {
			if v2, ok2 := inst.Src2.GetFloatValue(); ok2 {
				return NewMoveInst(inst.Dest, NewLiteralOperand(v1+v2))
			}
		}
	case OpSub:
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v2, ok2 := inst.Src2.GetIntValue(); ok2 {
				return NewMoveInst(inst.Dest, NewLiteralOperand(v1-v2))
			}
		}
		if v1, ok1 := inst.Src1.GetFloatValue(); ok1 {
			if v2, ok2 := inst.Src2.GetFloatValue(); ok2 {
				return NewMoveInst(inst.Dest, NewLiteralOperand(v1-v2))
			}
		}
	case OpMul:
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v2, ok2 := inst.Src2.GetIntValue(); ok2 {
				return NewMoveInst(inst.Dest, NewLiteralOperand(v1*v2))
			}
		}
		if v1, ok1 := inst.Src1.GetFloatValue(); ok1 {
			if v2, ok2 := inst.Src2.GetFloatValue(); ok2 {
				return NewMoveInst(inst.Dest, NewLiteralOperand(v1*v2))
			}
		}
	case OpDiv:
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v2, ok2 := inst.Src2.GetIntValue(); ok2 && v2 != 0 {
				return NewMoveInst(inst.Dest, NewLiteralOperand(v1/v2))
			}
		}
		if v1, ok1 := inst.Src1.GetFloatValue(); ok1 {
			if v2, ok2 := inst.Src2.GetFloatValue(); ok2 && v2 != 0 {
				return NewMoveInst(inst.Dest, NewLiteralOperand(v1/v2))
			}
		}
	case OpCmpEq:
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v2, ok2 := inst.Src2.GetIntValue(); ok2 {
				return NewMoveInst(inst.Dest, NewLiteralOperand(v1 == v2))
			}
		}
	case OpCmpLt:
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v2, ok2 := inst.Src2.GetIntValue(); ok2 {
				return NewMoveInst(inst.Dest, NewLiteralOperand(v1 < v2))
			}
		}
	}

	return nil
}

// isDeadCode проверяет, является ли инструкция мёртвым кодом
func (o *PeepholeOptimizer) isDeadCode(inst *Instruction, block *BasicBlock, index int, fn *Function) bool {
	if inst.Dest == nil {
		return false
	}

	// Проверяем, используется ли результат инструкции далее
	used := false
	for _, b := range fn.Blocks {
		for _, use := range b.Uses[inst.Dest.Name] {
			if use != inst {
				used = true
				break
			}
		}
		if used {
			break
		}
	}

	// PHI узлы не считаются мёртвым кодом
	if inst.Opcode == OpPhi {
		return false
	}

	// Инструкции с побочными эффектами не удаляем
	if inst.Opcode == OpStore || inst.Opcode == OpCall ||
		inst.Opcode == OpRet || inst.Opcode == OpJmp ||
		inst.Opcode == OpJmpIf || inst.Opcode == OpJmpIfNot {
		return false
	}

	return !used
}

func isZero(op *Operand) bool {
	if v, ok := op.GetIntValue(); ok {
		return v == 0
	}
	if v, ok := op.GetFloatValue(); ok {
		return v == 0.0
	}
	return false
}

func isOne(op *Operand) bool {
	if v, ok := op.GetIntValue(); ok {
		return v == 1
	}
	if v, ok := op.GetFloatValue(); ok {
		return v == 1.0
	}
	return false
}

// GetOptimizationReport возвращает отчёт об оптимизациях
func (o *PeepholeOptimizer) GetOptimizationReport() string {
	var sb strings.Builder

	sb.WriteString("Optimization Report:\n")
	sb.WriteString("===================\n\n")

	fmt.Fprintf(&sb, "Iterations: %d\n", o.iterations)
	fmt.Fprintf(&sb, "Changes made: %d\n\n", len(o.changes))

	if len(o.changes) > 0 {
		sb.WriteString("Changes:\n")
		for _, change := range o.changes {
			fmt.Fprintf(&sb, "  - %s\n", change)
		}
	}

	return sb.String()
}
