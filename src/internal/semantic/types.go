package semantic

import "fmt"

type TypeKind string

const (
	TypeInt     TypeKind = "int"
	TypeFloat   TypeKind = "float"
	TypeBool    TypeKind = "bool"
	TypeVoid    TypeKind = "void"
	TypeString  TypeKind = "string"
	TypeStruct  TypeKind = "struct"
	TypeFunc    TypeKind = "function"
	TypeArray   TypeKind = "array"   // Sprint 7
	TypePointer TypeKind = "pointer" // Sprint 7
)

type Type struct {
	Kind      TypeKind
	Name      string
	Fields    map[string]*Type
	Return    *Type
	Params    []*Type
	BaseType  *Type // Sprint 7: базовый тип для массивов/указателей
	ArraySize int   // Sprint 7: размер массива (-1 если неизвестен)
}

func NewType(kind TypeKind) *Type {
	return &Type{
		Kind:      kind,
		Fields:    nil,
		Return:    nil,
		Params:    nil,
		BaseType:  nil,
		ArraySize: -1,
	}
}

func NewStructType(name string) *Type {
	return &Type{
		Kind:      TypeStruct,
		Name:      name,
		Fields:    make(map[string]*Type),
		BaseType:  nil,
		ArraySize: -1,
	}
}

func NewFunctionType(returnType *Type, params []*Type) *Type {
	return &Type{
		Kind:      TypeFunc,
		Return:    returnType,
		Params:    params,
		BaseType:  nil,
		ArraySize: -1,
	}
}

// NewArrayType создает тип массива (Sprint 7)
func NewArrayType(baseType *Type, size int) *Type {
	return &Type{
		Kind:      TypeArray,
		BaseType:  baseType,
		ArraySize: size,
	}
}

// NewPointerType создает тип указателя (Sprint 7)
func NewPointerType(baseType *Type) *Type {
	return &Type{
		Kind:     TypePointer,
		BaseType: baseType,
	}
}

func (t *Type) String() string {
	if t == nil {
		return "<nil>"
	}

	switch t.Kind {
	case TypeInt, TypeFloat, TypeBool, TypeVoid, TypeString:
		return string(t.Kind)
	case TypeStruct:
		return "struct " + t.Name
	case TypeFunc:
		if t.Return == nil {
			return "function"
		}
		return "function -> " + t.Return.String()
	case TypeArray:
		if t.BaseType != nil {
			if t.ArraySize >= 0 {
				return fmt.Sprintf("%s[%d]", t.BaseType.String(), t.ArraySize)
			}
			return t.BaseType.String() + "[]"
		}
		return "array"
	case TypePointer:
		if t.BaseType != nil {
			return "*" + t.BaseType.String()
		}
		return "pointer"
	default:
		return "unknown"
	}
}

func (t *Type) Equals(other *Type) bool {
	if t == nil || other == nil {
		return t == other
	}

	if t.Kind != other.Kind {
		if t.Kind == TypeInt && other.Kind == TypeFloat {
			return true
		}
		if t.Kind == TypeArray && other.Kind == TypePointer {
			return t.BaseType.Equals(other.BaseType)
		}
		return false
	}

	switch t.Kind {
	case TypeStruct:
		return t.Name == other.Name
	case TypeFunc:
		if t.Return == nil || other.Return == nil {
			return t.Return == other.Return
		}
		if !t.Return.Equals(other.Return) {
			return false
		}
		if len(t.Params) != len(other.Params) {
			return false
		}
		for i := range t.Params {
			if !t.Params[i].Equals(other.Params[i]) {
				return false
			}
		}
		return true
	case TypeArray:
		if t.BaseType == nil || other.BaseType == nil {
			return t.BaseType == other.BaseType
		}
		return t.BaseType.Equals(other.BaseType)
	case TypePointer:
		if t.BaseType == nil || other.BaseType == nil {
			return t.BaseType == other.BaseType
		}
		return t.BaseType.Equals(other.BaseType)
	default:
		return t.Kind == other.Kind
	}
}

func (t *Type) IsNumeric() bool {
	if t == nil {
		return false
	}
	return t.Kind == TypeInt || t.Kind == TypeFloat
}

func (t *Type) IsInteger() bool {
	if t == nil {
		return false
	}
	return t.Kind == TypeInt
}

func (t *Type) IsFloat() bool {
	if t == nil {
		return false
	}
	return t.Kind == TypeFloat
}

func (t *Type) IsBool() bool {
	if t == nil {
		return false
	}
	return t.Kind == TypeBool
}

func (t *Type) IsVoid() bool {
	if t == nil {
		return false
	}
	return t.Kind == TypeVoid
}

func (t *Type) IsString() bool {
	if t == nil {
		return false
	}
	return t.Kind == TypeString
}

func (t *Type) IsStruct() bool {
	if t == nil {
		return false
	}
	return t.Kind == TypeStruct
}

func (t *Type) IsFunction() bool {
	if t == nil {
		return false
	}
	return t.Kind == TypeFunc
}

// IsArray проверяет, является ли тип массивом (Sprint 7)
func (t *Type) IsArray() bool {
	if t == nil {
		return false
	}
	return t.Kind == TypeArray
}

// IsPointer проверяет, является ли тип указателем (Sprint 7)
func (t *Type) IsPointer() bool {
	if t == nil {
		return false
	}
	return t.Kind == TypePointer
}

func (t *Type) IsAssignableTo(target *Type) bool {
	if t == nil || target == nil {
		return false
	}

	if t.Equals(target) {
		return true
	}

	// int -> float widening
	if t.Kind == TypeInt && target.Kind == TypeFloat {
		return true
	}

	// array -> pointer decay
	if t.Kind == TypeArray && target.Kind == TypePointer {
		return t.BaseType.Equals(target.BaseType)
	}

	return false
}

func (t *Type) IsAssignable(other *Type) bool {
	if t == nil || other == nil {
		return false
	}

	if t.Equals(other) {
		return true
	}

	if t.Kind == TypeInt && other.Kind == TypeFloat {
		return true
	}

	return false
}
