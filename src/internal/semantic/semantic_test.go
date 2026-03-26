package semantic

import (
	"testing"
)

func TestSymbolTableInsert(t *testing.T) {
	st := NewSymbolTable()

	sym := &Symbol{
		Name: "x",
		Kind: SymbolVariable,
		Type: NewType(TypeInt),
		Line: 1,
	}

	if !st.Insert(sym) {
		t.Error("Expected insert to succeed")
	}

	if st.Insert(sym) {
		t.Error("Expected duplicate insert to fail")
	}
}

func TestSymbolTableLookup(t *testing.T) {
	st := NewSymbolTable()

	sym := &Symbol{
		Name: "x",
		Kind: SymbolVariable,
		Type: NewType(TypeInt),
		Line: 1,
	}

	st.Insert(sym)

	found := st.Lookup("x")
	if found == nil {
		t.Error("Expected to find symbol 'x'")
	}

	if found.Name != "x" {
		t.Errorf("Expected name 'x', got '%s'", found.Name)
	}

	notFound := st.Lookup("y")
	if notFound != nil {
		t.Error("Expected not to find symbol 'y'")
	}
}

func TestSymbolTableScopeNesting(t *testing.T) {
	st := NewSymbolTable()

	globalSym := &Symbol{
		Name: "global",
		Kind: SymbolVariable,
		Type: NewType(TypeInt),
		Line: 1,
	}
	st.Insert(globalSym)

	st.EnterScope("func")

	localSym := &Symbol{
		Name: "local",
		Kind: SymbolVariable,
		Type: NewType(TypeInt),
		Line: 2,
	}
	st.Insert(localSym)

	foundGlobal := st.Lookup("global")
	if foundGlobal == nil {
		t.Error("Should find global symbol from inner scope")
	}

	foundLocal := st.Lookup("local")
	if foundLocal == nil {
		t.Error("Should find local symbol")
	}

	st.ExitScope()

	afterExit := st.Lookup("local")
	if afterExit != nil {
		t.Error("Local symbol should not be accessible after scope exit")
	}

	stillGlobal := st.Lookup("global")
	if stillGlobal == nil {
		t.Error("Global symbol should still be accessible")
	}
}

func TestTypeCompatibility(t *testing.T) {
	ts := NewTypeSystem()

	intType := ts.IntType
	floatType := ts.FloatType
	boolType := ts.BoolType

	if !ts.IsAssignable(floatType, intType) {
		t.Error("int should be assignable to float")
	}

	if ts.IsAssignable(intType, floatType) {
		t.Error("float should not be assignable to int")
	}

	if ts.IsAssignable(intType, boolType) {
		t.Error("bool should not be assignable to int")
	}
}

func TestBinaryOperationResult(t *testing.T) {
	ts := NewTypeSystem()

	intType := ts.IntType
	floatType := ts.FloatType
	boolType := ts.BoolType

	result, err := ts.BinaryOperationResult("+", intType, intType)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !result.Equals(intType) {
		t.Errorf("Expected int, got %s", result.String())
	}

	result, err = ts.BinaryOperationResult("+", intType, floatType)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !result.Equals(floatType) {
		t.Errorf("Expected float, got %s", result.String())
	}

	result, err = ts.BinaryOperationResult("==", intType, intType)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !result.Equals(boolType) {
		t.Errorf("Expected bool, got %s", result.String())
	}

	_, err = ts.BinaryOperationResult("+", intType, boolType)
	if err == nil {
		t.Error("Expected error for int + bool")
	}
}

func TestUnaryOperationResult(t *testing.T) {
	ts := NewTypeSystem()

	intType := ts.IntType
	boolType := ts.BoolType

	result, err := ts.UnaryOperationResult("-", intType)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !result.Equals(intType) {
		t.Errorf("Expected int, got %s", result.String())
	}

	result, err = ts.UnaryOperationResult("!", boolType)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !result.Equals(boolType) {
		t.Errorf("Expected bool, got %s", result.String())
	}

	_, err = ts.UnaryOperationResult("-", boolType)
	if err == nil {
		t.Error("Expected error for -bool")
	}

	_, err = ts.UnaryOperationResult("!", intType)
	if err == nil {
		t.Error("Expected error for !int")
	}
}

func TestTypeSize(t *testing.T) {
	ts := NewTypeSystem()

	if ts.GetSize(ts.IntType) != 4 {
		t.Errorf("Expected int size 4, got %d", ts.GetSize(ts.IntType))
	}

	if ts.GetSize(ts.FloatType) != 8 {
		t.Errorf("Expected float size 8, got %d", ts.GetSize(ts.FloatType))
	}

	if ts.GetSize(ts.BoolType) != 1 {
		t.Errorf("Expected bool size 1, got %d", ts.GetSize(ts.BoolType))
	}
}
