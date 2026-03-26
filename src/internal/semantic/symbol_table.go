package semantic

import (
	"fmt"
)

type SymbolKind string

const (
	SymbolVariable  SymbolKind = "variable"
	SymbolFunction  SymbolKind = "function"
	SymbolParameter SymbolKind = "parameter"
	SymbolStruct    SymbolKind = "struct"
	SymbolField     SymbolKind = "field"
)

type Symbol struct {
	Name   string
	Kind   SymbolKind
	Type   *Type
	Line   int
	Column int
	Scope  *Scope

	Parameters []*Symbol

	Fields map[string]*Symbol

	Offset int
	Size   int
}

type Scope struct {
	symbols map[string]*Symbol
	parent  *Scope
	level   int
	name    string
}

func NewScope(parent *Scope, level int, name string) *Scope {
	return &Scope{
		symbols: make(map[string]*Symbol),
		parent:  parent,
		level:   level,
		name:    name,
	}
}

func (s *Scope) Insert(symbol *Symbol) bool {
	if _, exists := s.symbols[symbol.Name]; exists {
		return false
	}
	symbol.Scope = s
	s.symbols[symbol.Name] = symbol
	return true
}

func (s *Scope) Lookup(name string) *Symbol {
	if sym, exists := s.symbols[name]; exists {
		return sym
	}
	if s.parent != nil {
		return s.parent.Lookup(name)
	}
	return nil
}

func (s *Scope) LookupLocal(name string) *Symbol {
	if sym, exists := s.symbols[name]; exists {
		return sym
	}
	return nil
}

func (s *Scope) GetAllSymbols() []*Symbol {
	result := make([]*Symbol, 0, len(s.symbols))
	for _, sym := range s.symbols {
		result = append(result, sym)
	}
	return result
}

type SymbolTable struct {
	global     *Scope
	current    *Scope
	scopeLevel int
}

func NewSymbolTable() *SymbolTable {
	st := &SymbolTable{
		scopeLevel: 0,
	}
	st.global = NewScope(nil, 0, "global")
	st.current = st.global
	return st
}

func (st *SymbolTable) EnterScope(name string) {
	st.scopeLevel++
	st.current = NewScope(st.current, st.scopeLevel, name)
}

func (st *SymbolTable) ExitScope() {
	if st.current.parent != nil {
		st.current = st.current.parent
		st.scopeLevel--
	}
}

func (st *SymbolTable) Insert(symbol *Symbol) bool {
	return st.current.Insert(symbol)
}

func (st *SymbolTable) Lookup(name string) *Symbol {
	return st.current.Lookup(name)
}

func (st *SymbolTable) LookupLocal(name string) *Symbol {
	return st.current.LookupLocal(name)
}

func (st *SymbolTable) GetCurrentScope() *Scope {
	return st.current
}

func (st *SymbolTable) GetGlobalScope() *Scope {
	return st.global
}

func (st *SymbolTable) String() string {
	result := "Symbol Table:\n"
	st.printScope(st.global, 0, &result)
	return result
}

func (st *SymbolTable) printScope(scope *Scope, indent int, result *string) {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}
	*result += fmt.Sprintf("%s%s scope (level %d):\n", prefix, scope.name, scope.level)
	for _, sym := range scope.GetAllSymbols() {
		*result += fmt.Sprintf("%s  %s: %s %s (line %d)\n",
			prefix, sym.Name, sym.Kind, sym.Type.String(), sym.Line)
	}
	*result += "\n"
}
