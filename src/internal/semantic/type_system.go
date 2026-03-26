package semantic

import (
	"fmt"
)

type TypeSystem struct {
	IntType    *Type
	FloatType  *Type
	BoolType   *Type
	VoidType   *Type
	StringType *Type
}

func NewTypeSystem() *TypeSystem {
	return &TypeSystem{
		IntType:    NewType(TypeInt),
		FloatType:  NewType(TypeFloat),
		BoolType:   NewType(TypeBool),
		VoidType:   NewType(TypeVoid),
		StringType: NewType(TypeString),
	}
}

func (ts *TypeSystem) IsNumeric(t *Type) bool {
	return t.IsNumeric()
}

func (ts *TypeSystem) IsInteger(t *Type) bool {
	return t.IsInteger()
}

func (ts *TypeSystem) IsFloat(t *Type) bool {
	return t.IsFloat()
}

func (ts *TypeSystem) IsBool(t *Type) bool {
	return t.IsBool()
}

func (ts *TypeSystem) IsVoid(t *Type) bool {
	return t.IsVoid()
}

func (ts *TypeSystem) IsString(t *Type) bool {
	return t.IsString()
}

func (ts *TypeSystem) IsStruct(t *Type) bool {
	return t.IsStruct()
}

func (ts *TypeSystem) IsFunction(t *Type) bool {
	return t.IsFunction()
}

func (ts *TypeSystem) BinaryOperationResult(op string, left, right *Type) (*Type, error) {
	if left == nil || right == nil {
		return nil, fmt.Errorf("operand types cannot be nil")
	}

	switch op {
	case "+", "-", "*", "/", "%":
		if ts.IsNumeric(left) && ts.IsNumeric(right) {
			if ts.IsFloat(left) || ts.IsFloat(right) {
				return ts.FloatType, nil
			}
			return ts.IntType, nil
		}
		return nil, fmt.Errorf("arithmetic operator %s requires numeric operands, got %s and %s",
			op, left.String(), right.String())

	case "==", "!=", "<", "<=", ">", ">=":
		if left.Equals(right) {
			return ts.BoolType, nil
		}
		if (ts.IsInteger(left) && ts.IsFloat(right)) || (ts.IsFloat(left) && ts.IsInteger(right)) {
			return ts.BoolType, nil
		}
		return nil, fmt.Errorf("comparison operator %s requires compatible operands, got %s and %s",
			op, left.String(), right.String())

	case "&&", "||":
		if ts.IsBool(left) && ts.IsBool(right) {
			return ts.BoolType, nil
		}
		return nil, fmt.Errorf("logical operator %s requires bool operands, got %s and %s",
			op, left.String(), right.String())

	default:
		return nil, fmt.Errorf("unknown binary operator: %s", op)
	}
}

func (ts *TypeSystem) UnaryOperationResult(op string, operand *Type) (*Type, error) {
	if operand == nil {
		return nil, fmt.Errorf("operand type cannot be nil")
	}

	switch op {
	case "-":
		if ts.IsNumeric(operand) {
			return operand, nil
		}
		return nil, fmt.Errorf("unary operator - requires numeric operand, got %s", operand.String())

	case "!":
		if ts.IsBool(operand) {
			return ts.BoolType, nil
		}
		return nil, fmt.Errorf("unary operator ! requires bool operand, got %s", operand.String())

	default:
		return nil, fmt.Errorf("unknown unary operator: %s", op)
	}
}

func (ts *TypeSystem) IsAssignable(target, source *Type) bool {
	if source == nil || target == nil {
		return false
	}

	if source.Equals(target) {
		return true
	}

	if source.IsInteger() && target.IsFloat() {
		return true
	}

	return false
}

func (ts *TypeSystem) GetCommonType(a, b *Type) *Type {
	if a == nil || b == nil {
		return nil
	}

	if ts.IsFloat(a) || ts.IsFloat(b) {
		return ts.FloatType
	}

	if ts.IsInteger(a) || ts.IsInteger(b) {
		return ts.IntType
	}

	if ts.IsBool(a) && ts.IsBool(b) {
		return ts.BoolType
	}

	if ts.IsString(a) && ts.IsString(b) {
		return ts.StringType
	}

	return nil
}

func (ts *TypeSystem) IsComparable(t *Type) bool {
	if t == nil {
		return false
	}
	return ts.IsNumeric(t) || ts.IsBool(t) || ts.IsString(t)
}

func (ts *TypeSystem) GetSize(t *Type) int {
	if t == nil {
		return 0
	}

	switch t.Kind {
	case TypeInt:
		return 4
	case TypeFloat:
		return 8
	case TypeBool:
		return 1
	case TypeString:
		return 16
	case TypeStruct:
		size := 0
		for _, fieldType := range t.Fields {
			size += ts.GetSize(fieldType)
		}
		return size
	case TypeFunc:
		return 8
	default:
		return 0
	}
}

func (ts *TypeSystem) GetAlignment(t *Type) int {
	if t == nil {
		return 0
	}

	switch t.Kind {
	case TypeInt:
		return 4
	case TypeFloat:
		return 8
	case TypeBool:
		return 1
	case TypeString:
		return 8
	case TypeStruct:
		maxAlign := 0
		for _, fieldType := range t.Fields {
			align := ts.GetAlignment(fieldType)
			if align > maxAlign {
				maxAlign = align
			}
		}
		return maxAlign
	default:
		return 4
	}
}

func (ts *TypeSystem) GetTypeName(t *Type) string {
	if t == nil {
		return "<nil>"
	}
	return t.String()
}

func (ts *TypeSystem) IsValidType(t *Type) bool {
	if t == nil {
		return false
	}

	switch t.Kind {
	case TypeInt, TypeFloat, TypeBool, TypeVoid, TypeString:
		return true
	case TypeStruct:
		return t.Name != ""
	case TypeFunc:
		return t.Return != nil
	default:
		return false
	}
}

func (ts *TypeSystem) CanBeUsedInExpression(t *Type) bool {
	if t == nil {
		return false
	}

	switch t.Kind {
	case TypeInt, TypeFloat, TypeBool, TypeString:
		return true
	case TypeStruct:
		return true
	default:
		return false
	}
}
