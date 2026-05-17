package ir

import (
	"fmt"
	"sort"
	"strings"
)

// Function представляет функцию в IR
type Function struct {
	Name       string
	ReturnType string
	Params     []*Param
	Blocks     []*BasicBlock
	EntryBlock *BasicBlock
	ExitBlock  *BasicBlock

	Locals map[string]*VarInfo

	tempCounter  int
	labelCounter int

	// Dominator tree
	DominatorTree map[string]*BasicBlock

	// Def-Use chain для всей функции
	GlobalDefs map[string]*Instruction
	GlobalUses map[string][]*Instruction
}

// Param представляет параметр функции
type Param struct {
	Name string
	Type string
}

// VarInfo содержит информацию о локальной переменной
type VarInfo struct {
	Name   string
	Type   string
	Offset int // смещение в стеке
	Size   int // в байтах
}

// NewFunction создает новую функцию
func NewFunction(name, returnType string) *Function {
	f := &Function{
		Name:          name,
		ReturnType:    returnType,
		Params:        make([]*Param, 0),
		Blocks:        make([]*BasicBlock, 0),
		Locals:        make(map[string]*VarInfo),
		DominatorTree: make(map[string]*BasicBlock),
		GlobalDefs:    make(map[string]*Instruction),
		GlobalUses:    make(map[string][]*Instruction),
	}
	f.EntryBlock = f.NewBlock("entry")
	return f
}

// AddParam добавляет параметр
func (f *Function) AddParam(name, typ string) {
	f.Params = append(f.Params, &Param{Name: name, Type: typ})
}

// NewBlock создает новый базовый блок
func (f *Function) NewBlock(label string) *BasicBlock {
	block := NewBasicBlock(label)
	block.Function = f
	f.Blocks = append(f.Blocks, block)
	return block
}

// NewLabel генерирует уникальную метку
func (f *Function) NewLabel(prefix string) string {
	f.labelCounter++
	return fmt.Sprintf("%s_%d", prefix, f.labelCounter)
}

// NewTemp генерирует новую временную переменную
func (f *Function) NewTemp() *Operand {
	f.tempCounter++
	return NewTempOperand(f.tempCounter)
}

// AddLocal добавляет локальную переменную
func (f *Function) AddLocal(name, typ string) {
	f.Locals[name] = &VarInfo{
		Name: name,
		Type: typ,
	}
}

// BuildDominatorTree строит дерево доминаторов
func (f *Function) BuildDominatorTree() {
	if len(f.Blocks) == 0 {
		return
	}

	// Для каждого блока находим ближайшего доминатора
	for _, block := range f.Blocks {
		if block == f.EntryBlock {
			block.Dominator = nil
			continue
		}

		// Ищем общего доминатора среди предшественников
		var dominator *BasicBlock
		for _, pred := range block.Predecessors {
			if dominator == nil {
				dominator = pred
			} else {
				dominator = findCommonDominator(dominator, pred)
			}
		}
		block.Dominator = dominator
		if dominator != nil {
			dominator.Children = append(dominator.Children, block)
		}
		f.DominatorTree[block.Label] = dominator
	}
}

func findCommonDominator(a, b *BasicBlock) *BasicBlock {
	if a == nil || b == nil {
		return nil
	}

	visited := make(map[string]bool)
	current := a
	for current != nil {
		visited[current.Label] = true
		current = current.Dominator
	}

	current = b
	for current != nil {
		if visited[current.Label] {
			return current
		}
		current = current.Dominator
	}

	return nil
}

// BuildGlobalDefUseChains строит глобальные Def-Use цепочки
func (f *Function) BuildGlobalDefUseChains() {
	f.GlobalDefs = make(map[string]*Instruction)
	f.GlobalUses = make(map[string][]*Instruction)

	for _, block := range f.Blocks {
		block.BuildDefUseChains()

		for name, def := range block.Defs {
			f.GlobalDefs[name] = def
		}

		for name, uses := range block.Uses {
			f.GlobalUses[name] = append(f.GlobalUses[name], uses...)
		}
	}
}

// GetDefUseChain возвращает цепочку def-use для переменной
func (f *Function) GetDefUseChain(varName string) (*Instruction, []*Instruction) {
	def := f.GlobalDefs[varName]
	uses := f.GlobalUses[varName]
	return def, uses
}

// String возвращает строковое представление функции
func (f *Function) String() string {
	var sb strings.Builder

	// Сигнатура функции
	fmt.Fprintf(&sb, "function %s: %s (", f.Name, f.ReturnType)
	params := make([]string, len(f.Params))
	for i, p := range f.Params {
		params[i] = fmt.Sprintf("%s %s", p.Type, p.Name)
	}
	sb.WriteString(strings.Join(params, ", "))
	sb.WriteString(")\n")

	// Локальные переменные
	if len(f.Locals) > 0 {
		names := make([]string, 0, len(f.Locals))
		for name := range f.Locals {
			names = append(names, name)
		}
		sort.Strings(names)

		sb.WriteString("  ; Locals:\n")
		for _, name := range names {
			info := f.Locals[name]
			fmt.Fprintf(&sb, "  ;   %s: %s (offset %d, size %d)\n", name, info.Type, info.Offset, info.Size)
		}
	}

	// Базовые блоки
	for _, block := range f.Blocks {
		sb.WriteString("\n")
		sb.WriteString(block.String())
	}

	return sb.String()
}
