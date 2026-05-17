package ir

import (
	"fmt"
	"strings"
)

// BasicBlock представляет базовый блок в CFG
type BasicBlock struct {
	Label        string
	Instructions []*Instruction
	Function     *Function
	Predecessors []*BasicBlock
	Successors   []*BasicBlock

	// Для SSA формы
	PhiNodes map[string]*Instruction

	// Для dominator tree
	Dominator *BasicBlock
	Children  []*BasicBlock

	// Def-Use цепочки
	Defs map[string]*Instruction   // инструкция определения
	Uses map[string][]*Instruction // список использований
}

// NewBasicBlock создает новый базовый блок
func NewBasicBlock(label string) *BasicBlock {
	return &BasicBlock{
		Label:        label,
		Instructions: make([]*Instruction, 0),
		Predecessors: make([]*BasicBlock, 0),
		Successors:   make([]*BasicBlock, 0),
		PhiNodes:     make(map[string]*Instruction),
		Children:     make([]*BasicBlock, 0),
		Defs:         make(map[string]*Instruction),
		Uses:         make(map[string][]*Instruction),
	}
}

// AddInstruction добавляет инструкцию в блок
func (b *BasicBlock) AddInstruction(inst *Instruction) {
	b.Instructions = append(b.Instructions, inst)

	// Обновляем Def-Use цепочки
	if inst.Dest != nil {
		b.Defs[inst.Dest.Name] = inst
	}

	// Регистрируем использования
	for _, operand := range inst.GetUsedOperands() {
		b.Uses[operand.Name] = append(b.Uses[operand.Name], inst)
	}
}

// AddPredecessor добавляет предшественника
func (b *BasicBlock) AddPredecessor(pred *BasicBlock) {
	for _, p := range b.Predecessors {
		if p == pred {
			return
		}
	}
	b.Predecessors = append(b.Predecessors, pred)
}

// AddSuccessor добавляет последователя
func (b *BasicBlock) AddSuccessor(succ *BasicBlock) {
	for _, s := range b.Successors {
		if s == succ {
			return
		}
	}
	b.Successors = append(b.Successors, succ)
}

// GetTerminator возвращает завершающую инструкцию блока
func (b *BasicBlock) GetTerminator() *Instruction {
	if len(b.Instructions) == 0 {
		return nil
	}
	last := b.Instructions[len(b.Instructions)-1]
	if last.IsTerminator() {
		return last
	}
	return nil
}

// IsTerminated возвращает true, если блок завершен
func (b *BasicBlock) IsTerminated() bool {
	return b.GetTerminator() != nil
}

// BuildDefUseChains строит Def-Use цепочки для блока
func (b *BasicBlock) BuildDefUseChains() {
	b.Defs = make(map[string]*Instruction)
	b.Uses = make(map[string][]*Instruction)

	for _, inst := range b.Instructions {
		if inst.Dest != nil {
			b.Defs[inst.Dest.Name] = inst
		}

		for _, operand := range inst.GetUsedOperands() {
			b.Uses[operand.Name] = append(b.Uses[operand.Name], inst)
		}
	}
}

// GetDefs возвращает все определения в блоке
func (b *BasicBlock) GetDefs() map[string]*Instruction {
	return b.Defs
}

// GetUses возвращает все использования переменной в блоке
func (b *BasicBlock) GetUses(varName string) []*Instruction {
	return b.Uses[varName]
}

// String возвращает строковое представление блока
func (b *BasicBlock) String() string {
	var sb strings.Builder

	sb.WriteString(b.Label)
	sb.WriteString(":\n")

	// Предшественники в комментарии
	if len(b.Predecessors) > 0 {
		preds := make([]string, len(b.Predecessors))
		for i, p := range b.Predecessors {
			preds[i] = p.Label
		}
		fmt.Fprintf(&sb, "  ; preds: %s\n", strings.Join(preds, ", "))
	}

	// Dominator информация
	if b.Dominator != nil {
		fmt.Fprintf(&sb, "  ; idom: %s\n", b.Dominator.Label)
	}

	for _, inst := range b.Instructions {
		fmt.Fprintf(&sb, "  %s\n", inst.String())
	}

	return sb.String()
}
