package ir

import (
	"strings"
)

// Program представляет всю программу в IR
type Program struct {
	Functions   []*Function
	Globals     map[string]*GlobalVar
	CurrentFunc *Function
}

// GlobalVar представляет глобальную переменную
type GlobalVar struct {
	Name string
	Type string
	Init *Operand
}

// NewProgram создает новую программу
func NewProgram() *Program {
	return &Program{
		Functions: make([]*Function, 0),
		Globals:   make(map[string]*GlobalVar),
	}
}

// AddFunction добавляет функцию
func (p *Program) AddFunction(f *Function) {
	p.Functions = append(p.Functions, f)
}

// AddGlobal добавляет глобальную переменную
func (p *Program) AddGlobal(name, typ string, init *Operand) {
	p.Globals[name] = &GlobalVar{
		Name: name,
		Type: typ,
		Init: init,
	}
}

// String возвращает строковое представление программы
func (p *Program) String() string {
	var sb strings.Builder

	if len(p.Globals) > 0 {
		sb.WriteString("; Global variables:\n")
		for name, g := range p.Globals {
			if g.Init != nil {
				sb.WriteString(";   @")
				sb.WriteString(name)
				sb.WriteString(": ")
				sb.WriteString(g.Type)
				sb.WriteString(" = ")
				sb.WriteString(g.Init.String())
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n")
	}

	for i, f := range p.Functions {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(f.String())
	}

	return sb.String()
}
