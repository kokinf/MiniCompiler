package ir

import (
	"fmt"
	"strings"
)

// OptimizationPass представляет один проход оптимизации
type OptimizationPass struct {
	Name        string
	Description string
	Enabled     bool
}

// PeepholeOptimizer выполняет оптимизации IR
type PeepholeOptimizer struct {
	program    *Program
	changes    []string
	iterations int
	passes     []OptimizationPass
	stats      map[string]int
}

// NewPeepholeOptimizer создает новый оптимизатор
func NewPeepholeOptimizer(program *Program) *PeepholeOptimizer {
	return &PeepholeOptimizer{
		program: program,
		changes: make([]string, 0),
		stats:   make(map[string]int),
		passes: []OptimizationPass{
			{Name: "constant_folding", Description: "Свёртка константных выражений", Enabled: true},
			{Name: "algebraic", Description: "Алгебраические упрощения (x+0→x, x*1→x)", Enabled: true},
			{Name: "strength_reduction", Description: "Понижение силы операций (x*2→x+x)", Enabled: true},
			{Name: "constant_propagation", Description: "Распространение констант", Enabled: true},
			{Name: "branch_simplification", Description: "Упрощение условных переходов с константами", Enabled: true},
			{Name: "dead_allocation", Description: "Удаление неиспользуемых malloc/free", Enabled: true},
			{Name: "dead_code", Description: "Удаление мёртвого кода", Enabled: true},
			{Name: "jump_optimization", Description: "Оптимизация цепочек переходов", Enabled: true},
			{Name: "empty_block", Description: "Удаление пустых базовых блоков", Enabled: true},
		},
	}
}

// EnablePass включает конкретный проход
func (o *PeepholeOptimizer) EnablePass(name string) {
	for i := range o.passes {
		if o.passes[i].Name == name {
			o.passes[i].Enabled = true
			return
		}
	}
}

// DisablePass выключает конкретный проход
func (o *PeepholeOptimizer) DisablePass(name string) {
	for i := range o.passes {
		if o.passes[i].Name == name {
			o.passes[i].Enabled = false
			return
		}
	}
}

// IsPassEnabled проверяет, включен ли проход
func (o *PeepholeOptimizer) IsPassEnabled(name string) bool {
	for _, p := range o.passes {
		if p.Name == name {
			return p.Enabled
		}
	}
	return false
}

// Optimize выполняет все включенные оптимизации итеративно
func (o *PeepholeOptimizer) Optimize() {
	maxIterations := 20

	for i := 0; i < maxIterations; i++ {
		changed := false
		o.iterations++

		for _, fn := range o.program.Functions {
			// Pass 1: Свёртка констант + алгебраические + strength reduction + dead code
			if o.IsPassEnabled("constant_folding") || o.IsPassEnabled("algebraic") ||
				o.IsPassEnabled("strength_reduction") || o.IsPassEnabled("dead_code") {
				for _, block := range fn.Blocks {
					if o.optimizeBlock(block, fn) {
						changed = true
					}
				}
			}

			// Pass 2: Распространение констант
			if o.IsPassEnabled("constant_propagation") {
				if o.constantPropagation(fn) {
					changed = true
				}
			}

			// Pass 3: Упрощение условных переходов
			if o.IsPassEnabled("branch_simplification") {
				if o.simplifyConstantBranches(fn) {
					changed = true
				}
			}

			// Pass 4: Удаление неиспользуемых malloc/free
			if o.IsPassEnabled("dead_allocation") {
				if o.eliminateDeadAllocations(fn) {
					changed = true
				}
			}

			// Pass 5: Оптимизация переходов
			if o.IsPassEnabled("jump_optimization") {
				if o.optimizeJumps(fn) {
					changed = true
				}
			}

			// Pass 6: Удаление пустых блоков
			if o.IsPassEnabled("empty_block") {
				if o.eliminateEmptyBlocks(fn) {
					changed = true
				}
			}
		}

		if !changed {
			break
		}
	}
}

// ============================================================================
// PASS 1: Свёртка констант, алгебраические упрощения, strength reduction, dead code
// ============================================================================

func (o *PeepholeOptimizer) optimizeBlock(block *BasicBlock, fn *Function) bool {
	changed := false
	newInsts := make([]*Instruction, 0, len(block.Instructions))

	for i := 0; i < len(block.Instructions); i++ {
		inst := block.Instructions[i]

		// Свёртка констант
		if o.IsPassEnabled("constant_folding") {
			if folded := o.tryConstantFolding(inst); folded != nil {
				newInsts = append(newInsts, folded)
				o.stats["constant_folding"]++
				o.changes = append(o.changes, fmt.Sprintf("[%s] folded: %s → %s", fn.Name, inst.String(), folded.String()))
				changed = true
				continue
			}
		}

		// Алгебраические упрощения
		if o.IsPassEnabled("algebraic") {
			if simplified := o.tryAlgebraicSimplification(inst); simplified != nil {
				newInsts = append(newInsts, simplified)
				o.stats["algebraic"]++
				o.changes = append(o.changes, fmt.Sprintf("[%s] algebraic: %s → %s", fn.Name, inst.String(), simplified.String()))
				changed = true
				continue
			}
		}

		// Strength reduction
		if o.IsPassEnabled("strength_reduction") {
			if reduced := o.tryStrengthReduction(inst); reduced != nil {
				newInsts = append(newInsts, reduced)
				o.stats["strength_reduction"]++
				o.changes = append(o.changes, fmt.Sprintf("[%s] strength: %s → %s", fn.Name, inst.String(), reduced.String()))
				changed = true
				continue
			}
		}

		// Удаление мёртвого кода
		if o.IsPassEnabled("dead_code") {
			if o.isDeadCode(inst, block, i, fn) {
				o.stats["dead_code"]++
				o.changes = append(o.changes, fmt.Sprintf("[%s] dead: %s", fn.Name, inst.String()))
				changed = true
				continue
			}
		}

		// Слияние CMP + JUMP_IF
		if isComparison(inst.Opcode) && i < len(block.Instructions)-1 {
			nextInst := block.Instructions[i+1]
			if (nextInst.Opcode == OpJmpIf || nextInst.Opcode == OpJmpIfNot) &&
				inst.Dest != nil && nextInst.Src1 != nil &&
				inst.Dest.Name == nextInst.Src1.Name {
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
					mergedOp := nextInst.Opcode
					merged := &Instruction{
						Opcode:  mergedOp,
						Src1:    inst.Src1,
						Src2:    inst.Src2,
						Comment: "merged cmp+jump",
						Line:    inst.Line,
						Column:  inst.Column,
					}
					newInsts = append(newInsts, merged)
					newInsts = append(newInsts, &Instruction{
						Opcode:  OpJmp,
						Src1:    NewLabelOperand(nextInst.Src2.Name),
						Comment: "fallthrough",
					})
					o.changes = append(o.changes, fmt.Sprintf("[%s] merged cmp+jump", fn.Name))
					changed = true
					i++
					continue
				}
			}
		}

		// Удаление jmp на следующую метку
		if inst.Opcode == OpJmp && i < len(block.Instructions)-1 {
			nextInst := block.Instructions[i+1]
			if nextInst.Opcode == OpLabel && inst.Src1.Name == nextInst.Src1.Name {
				o.changes = append(o.changes, fmt.Sprintf("[%s] removed jmp to next label", fn.Name))
				changed = true
				continue
			}
		}

		newInsts = append(newInsts, inst)
	}

	if changed {
		block.Instructions = newInsts
	}
	return changed
}

// ============================================================================
// Свёртка констант
// ============================================================================

func (o *PeepholeOptimizer) tryConstantFolding(inst *Instruction) *Instruction {
	// Унарные операции
	if inst.Src2 == nil && inst.Src1 != nil && inst.Src1.IsConstant() {
		switch inst.Opcode {
		case OpNeg:
			if val, ok := inst.Src1.GetIntValue(); ok {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(-val), Comment: "folded -const"}
			}
			if val, ok := inst.Src1.GetFloatValue(); ok {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(-val), Comment: "folded -const"}
			}
		case OpNot:
			if val, ok := inst.Src1.GetBoolValue(); ok {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(!val), Comment: "folded !const"}
			}
			if v, ok := inst.Src1.GetIntValue(); ok {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(v == 0), Comment: "folded !const"}
			}
		}
		return nil
	}

	// Бинарные операции
	if inst.Src1 == nil || inst.Src2 == nil || !inst.Src1.IsConstant() || !inst.Src2.IsConstant() {
		return nil
	}

	switch inst.Opcode {
	case OpAdd:
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v2, ok2 := inst.Src2.GetIntValue(); ok2 {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(v1 + v2), Comment: "folded add"}
			}
		}
	case OpSub:
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v2, ok2 := inst.Src2.GetIntValue(); ok2 {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(v1 - v2), Comment: "folded sub"}
			}
		}
	case OpMul:
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v2, ok2 := inst.Src2.GetIntValue(); ok2 {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(v1 * v2), Comment: "folded mul"}
			}
		}
	case OpDiv:
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v2, ok2 := inst.Src2.GetIntValue(); ok2 && v2 != 0 {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(v1 / v2), Comment: "folded div"}
			}
		}
	case OpMod:
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v2, ok2 := inst.Src2.GetIntValue(); ok2 && v2 != 0 {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(v1 % v2), Comment: "folded mod"}
			}
		}
	case OpCmpEq:
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v2, ok2 := inst.Src2.GetIntValue(); ok2 {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(v1 == v2), Comment: "folded cmp_eq"}
			}
		}
	case OpCmpNe:
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v2, ok2 := inst.Src2.GetIntValue(); ok2 {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(v1 != v2), Comment: "folded cmp_ne"}
			}
		}
	case OpCmpLt:
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v2, ok2 := inst.Src2.GetIntValue(); ok2 {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(v1 < v2), Comment: "folded cmp_lt"}
			}
		}
	case OpCmpLe:
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v2, ok2 := inst.Src2.GetIntValue(); ok2 {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(v1 <= v2), Comment: "folded cmp_le"}
			}
		}
	case OpCmpGt:
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v2, ok2 := inst.Src2.GetIntValue(); ok2 {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(v1 > v2), Comment: "folded cmp_gt"}
			}
		}
	case OpCmpGe:
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v2, ok2 := inst.Src2.GetIntValue(); ok2 {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(v1 >= v2), Comment: "folded cmp_ge"}
			}
		}
	case OpAnd:
		if v1, ok1 := inst.Src1.GetBoolValue(); ok1 {
			if v2, ok2 := inst.Src2.GetBoolValue(); ok2 {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(v1 && v2), Comment: "folded and"}
			}
		}
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v1 != 0 {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: inst.Src2, Comment: "folded: true && x → x"}
			}
			return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(false), Comment: "folded: false && x → false"}
		}
	case OpOr:
		if v1, ok1 := inst.Src1.GetBoolValue(); ok1 {
			if v2, ok2 := inst.Src2.GetBoolValue(); ok2 {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(v1 || v2), Comment: "folded or"}
			}
		}
		if v1, ok1 := inst.Src1.GetIntValue(); ok1 {
			if v1 != 0 {
				return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(true), Comment: "folded: true || x → true"}
			}
			return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: inst.Src2, Comment: "folded: false || x → x"}
		}
	}

	return nil
}

// ============================================================================
// Алгебраические упрощения
// ============================================================================

func (o *PeepholeOptimizer) tryAlgebraicSimplification(inst *Instruction) *Instruction {
	if inst.Opcode != OpAdd && inst.Opcode != OpSub &&
		inst.Opcode != OpMul && inst.Opcode != OpDiv && inst.Opcode != OpMod {
		return nil
	}

	if inst.Opcode == OpAdd && isZero(inst.Src2) {
		return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: inst.Src1, Comment: "x+0→x"}
	}
	if inst.Opcode == OpAdd && isZero(inst.Src1) {
		return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: inst.Src2, Comment: "0+x→x"}
	}
	if inst.Opcode == OpSub && isZero(inst.Src2) {
		return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: inst.Src1, Comment: "x-0→x"}
	}
	if inst.Opcode == OpMul && (isZero(inst.Src1) || isZero(inst.Src2)) {
		return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(0), Comment: "x*0→0"}
	}
	if inst.Opcode == OpMul && isOne(inst.Src2) {
		return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: inst.Src1, Comment: "x*1→x"}
	}
	if inst.Opcode == OpMul && isOne(inst.Src1) {
		return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: inst.Src2, Comment: "1*x→x"}
	}
	if inst.Opcode == OpDiv && isOne(inst.Src2) {
		return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: inst.Src1, Comment: "x/1→x"}
	}
	if inst.Opcode == OpMod && isOne(inst.Src2) {
		return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(0), Comment: "x%1→0"}
	}
	if inst.Opcode == OpSub && inst.Src1.String() == inst.Src2.String() {
		return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(0), Comment: "x-x→0"}
	}
	if inst.Opcode == OpDiv && inst.Src1.String() == inst.Src2.String() {
		return &Instruction{Opcode: OpMove, Dest: inst.Dest, Src1: NewLiteralOperand(1), Comment: "x/x→1"}
	}

	return nil
}

// ============================================================================
// Strength reduction
// ============================================================================

func (o *PeepholeOptimizer) tryStrengthReduction(inst *Instruction) *Instruction {
	if inst.Opcode == OpMul {
		if val, ok := inst.Src2.GetIntValue(); ok && val == 2 {
			return &Instruction{Opcode: OpAdd, Dest: inst.Dest, Src1: inst.Src1, Src2: inst.Src1, Comment: "*2→x+x"}
		}
		if val, ok := inst.Src1.GetIntValue(); ok && val == 2 {
			return &Instruction{Opcode: OpAdd, Dest: inst.Dest, Src1: inst.Src2, Src2: inst.Src2, Comment: "2*→x+x"}
		}
	}
	return nil
}

// ============================================================================
// PASS 2: Распространение констант
// ============================================================================

func (o *PeepholeOptimizer) constantPropagation(fn *Function) bool {
	changed := false
	constants := make(map[string]*Operand)

	for _, block := range fn.Blocks {
		blockConstants := make(map[string]*Operand)
		for _, pred := range block.Predecessors {
			if predConstants, ok := fn.blockConstants[pred.Label]; ok {
				for k, v := range predConstants {
					blockConstants[k] = v
				}
			}
		}
		for k, v := range constants {
			blockConstants[k] = v
		}

		newInsts := make([]*Instruction, 0, len(block.Instructions))

		for _, inst := range block.Instructions {
			if inst.Opcode == OpMove && inst.Dest != nil && inst.Src1.IsConstant() {
				blockConstants[inst.Dest.Name] = inst.Src1
				constants[inst.Dest.Name] = inst.Src1
				newInsts = append(newInsts, inst)
				continue
			}

			if inst.Opcode == OpStore && inst.Src2.IsConstant() && inst.Src1.Name != "" {
				constants[inst.Src1.Name] = inst.Src2
			}

			if inst.Opcode == OpLoad && inst.Dest != nil {
				if c, ok := constants[inst.Src1.Name]; ok {
					newInsts = append(newInsts, &Instruction{
						Opcode: OpMove, Dest: inst.Dest, Src1: c,
						Comment: "propagated constant",
					})
					o.stats["constant_propagation"]++
					changed = true
					continue
				}
			}

			if inst.Src1 != nil && !inst.Src1.IsConstant() {
				if c, ok := blockConstants[inst.Src1.Name]; ok {
					inst.Src1 = c
					o.stats["constant_propagation"]++
					changed = true
				}
			}
			if inst.Src2 != nil && !inst.Src2.IsConstant() && inst.Src2.Type == OperandTemp {
				if c, ok := blockConstants[inst.Src2.Name]; ok {
					inst.Src2 = c
					o.stats["constant_propagation"]++
					changed = true
				}
			}

			if inst.Dest != nil && !inst.IsTerminator() {
				delete(blockConstants, inst.Dest.Name)
				delete(constants, inst.Dest.Name)
			}

			newInsts = append(newInsts, inst)
		}

		if fn.blockConstants == nil {
			fn.blockConstants = make(map[string]map[string]*Operand)
		}
		fn.blockConstants[block.Label] = blockConstants
		block.Instructions = newInsts
	}

	return changed
}

// ============================================================================
// PASS 3: Упрощение условных переходов с константами
// ============================================================================

func (o *PeepholeOptimizer) simplifyConstantBranches(fn *Function) bool {
	changed := false

	for _, block := range fn.Blocks {
		for i, inst := range block.Instructions {
			if inst.Opcode != OpJmpIf && inst.Opcode != OpJmpIfNot {
				continue
			}
			if inst.Src1 == nil || !inst.Src1.IsConstant() {
				continue
			}

			condVal, ok := inst.Src1.GetBoolValue()
			if !ok {
				if intVal, ok2 := inst.Src1.GetIntValue(); ok2 {
					condVal = intVal != 0
				} else {
					continue
				}
			}

			targetLabel := inst.Src2.Name

			if (inst.Opcode == OpJmpIf && condVal) || (inst.Opcode == OpJmpIfNot && !condVal) {
				block.Instructions[i] = &Instruction{
					Opcode: OpJmp, Src1: NewLabelOperand(targetLabel),
					Comment: "constant condition → unconditional jump",
				}
				o.stats["branch_simplification"]++
				changed = true
			} else {
				block.Instructions = append(block.Instructions[:i], block.Instructions[i+1:]...)
				o.stats["branch_simplification"]++
				changed = true
			}
		}
	}

	return changed
}

// ============================================================================
// PASS 4: Удаление неиспользуемых malloc/free
// ============================================================================

func (o *PeepholeOptimizer) eliminateDeadAllocations(fn *Function) bool {
	changed := false

	for _, block := range fn.Blocks {
		usedVars := make(map[string]bool)
		for _, inst := range block.Instructions {
			for _, op := range inst.GetUsedOperands() {
				usedVars[op.Name] = true
			}
		}

		type AllocInfo struct {
			mallocInst  *Instruction
			mallocIndex int
			freeIndex   int
			destName    string
		}

		allocs := make(map[string]*AllocInfo)

		for i, inst := range block.Instructions {
			if inst.Opcode == OpCall && inst.Src1.Name == "malloc" && inst.Dest != nil {
				allocs[inst.Dest.Name] = &AllocInfo{
					mallocInst: inst, mallocIndex: i, destName: inst.Dest.Name,
				}
			}
			if inst.Opcode == OpCall && inst.Src1.Name == "free" && len(inst.Args) > 0 {
				argName := inst.Args[0].Name
				if info, ok := allocs[argName]; ok {
					info.freeIndex = i
				}
			}
		}

		toRemove := make(map[int]bool)
		for destName, info := range allocs {
			if info.freeIndex > 0 && !usedVars[destName] {
				used := false
				for j := info.mallocIndex + 1; j < info.freeIndex; j++ {
					for _, op := range block.Instructions[j].GetUsedOperands() {
						if op.Name == destName {
							used = true
							break
						}
					}
					if used {
						break
					}
				}
				if !used {
					toRemove[info.mallocIndex] = true
					toRemove[info.freeIndex] = true
					for j := info.mallocIndex + 1; j < info.freeIndex; j++ {
						inst := block.Instructions[j]
						if inst.Opcode == OpJmpIf || inst.Opcode == OpJmpIfNot ||
							inst.Opcode == OpJmp || inst.Opcode == OpCmpEq {
							toRemove[j] = true
						}
					}
					o.stats["dead_allocation"]++
					o.changes = append(o.changes,
						fmt.Sprintf("[%s] removed unused malloc/free for %s", fn.Name, destName))
				}
			}
		}

		if len(toRemove) > 0 {
			newInsts := make([]*Instruction, 0)
			for i, inst := range block.Instructions {
				if !toRemove[i] {
					newInsts = append(newInsts, inst)
				}
			}
			block.Instructions = newInsts
			changed = true
		}
	}

	return changed
}

// ============================================================================
// PASS 5: Оптимизация переходов
// ============================================================================

func (o *PeepholeOptimizer) optimizeJumps(fn *Function) bool {
	changed := false

	labelToBlock := make(map[string]*BasicBlock)
	for _, block := range fn.Blocks {
		labelToBlock[block.Label] = block
	}

	for _, block := range fn.Blocks {
		for i, inst := range block.Instructions {
			if inst.Opcode == OpJmp && inst.Src1 != nil {
				targetBlock := labelToBlock[inst.Src1.Name]
				if targetBlock != nil && len(targetBlock.Instructions) == 1 &&
					targetBlock.Instructions[0].Opcode == OpJmp {
					block.Instructions[i] = &Instruction{
						Opcode: OpJmp, Src1: targetBlock.Instructions[0].Src1,
						Comment: "jump chaining",
					}
					o.stats["jump_optimization"]++
					changed = true
				}
			}
		}
	}

	return changed
}

// ============================================================================
// PASS 6: Удаление пустых блоков
// ============================================================================

func (o *PeepholeOptimizer) eliminateEmptyBlocks(fn *Function) bool {
	changed := false
	newBlocks := make([]*BasicBlock, 0)

	for _, block := range fn.Blocks {
		if block == fn.EntryBlock {
			newBlocks = append(newBlocks, block)
			continue
		}

		if len(block.Instructions) == 1 && block.Instructions[0].Opcode == OpJmp {
			targetLabel := block.Instructions[0].Src1.Name
			for _, pred := range block.Predecessors {
				for i, inst := range pred.Instructions {
					if inst.IsTerminator() {
						targets := inst.GetBranchTargets()
						for j, t := range targets {
							if t == block.Label {
								targets[j] = targetLabel
							}
						}
						if inst.Opcode == OpJmp {
							pred.Instructions[i] = &Instruction{
								Opcode: OpJmp, Src1: NewLabelOperand(targetLabel),
								Comment: "redirected through empty block",
							}
						}
					}
				}
				for _, succ := range block.Successors {
					pred.AddSuccessor(succ)
					succ.AddPredecessor(pred)
				}
			}
			o.stats["empty_block"]++
			changed = true
			continue
		}

		newBlocks = append(newBlocks, block)
	}

	if changed {
		fn.Blocks = newBlocks
	}
	return changed
}

// ============================================================================
// Вспомогательные функции
// ============================================================================

func (o *PeepholeOptimizer) isDeadCode(inst *Instruction, block *BasicBlock, index int, fn *Function) bool {
	if inst.Dest == nil {
		return false
	}
	if inst.Opcode == OpStore || inst.Opcode == OpCall || inst.Opcode == OpRet ||
		inst.Opcode == OpJmp || inst.Opcode == OpJmpIf || inst.Opcode == OpJmpIfNot ||
		inst.Opcode == OpPhi || inst.Opcode == OpParam {
		return false
	}
	for _, b := range fn.Blocks {
		for _, use := range b.Uses[inst.Dest.Name] {
			if use != inst {
				return false
			}
		}
	}
	return true
}

func isComparison(op Opcode) bool {
	switch op {
	case OpCmpEq, OpCmpNe, OpCmpLt, OpCmpLe, OpCmpGt, OpCmpGe:
		return true
	}
	return false
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

// ============================================================================
// Отчёт об оптимизациях
// ============================================================================

func (o *PeepholeOptimizer) GetOptimizationReport() string {
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════\n")
	sb.WriteString("   OPTIMIZATION REPORT\n")
	sb.WriteString("═══════════════════════════════════\n\n")
	sb.WriteString(fmt.Sprintf("Total iterations: %d\n\n", o.iterations))
	sb.WriteString("Passes executed:\n")
	sb.WriteString("─────────────────────────────────\n")

	totalChanges := 0
	for _, pass := range o.passes {
		count := o.stats[pass.Name]
		totalChanges += count
		status := "✓"
		if !pass.Enabled {
			status = "✗"
		}
		sb.WriteString(fmt.Sprintf("  %s %-22s: %3d changes  (%s)\n",
			status, pass.Name, count, pass.Description))
	}

	sb.WriteString("\n─────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("  Total changes: %d\n", totalChanges))
	sb.WriteString("═══════════════════════════════════\n")

	if len(o.changes) > 0 && len(o.changes) <= 20 {
		sb.WriteString("\nDetailed changes:\n")
		for _, change := range o.changes {
			sb.WriteString(fmt.Sprintf("  • %s\n", change))
		}
	} else if len(o.changes) > 20 {
		sb.WriteString(fmt.Sprintf("\n... and %d more changes\n", len(o.changes)-20))
	}

	return sb.String()
}
