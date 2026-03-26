package semantic

type TypeKind string

const (
	TypeInt    TypeKind = "int"
	TypeFloat  TypeKind = "float"
	TypeBool   TypeKind = "bool"
	TypeVoid   TypeKind = "void"
	TypeString TypeKind = "string"
	TypeStruct TypeKind = "struct"
	TypeFunc   TypeKind = "function"
)

type Type struct {
	Kind   TypeKind
	Name   string
	Fields map[string]*Type
	Return *Type
	Params []*Type
}

func NewType(kind TypeKind) *Type {
	return &Type{
		Kind:   kind,
		Fields: nil,
		Return: nil,
		Params: nil,
	}
}

func NewStructType(name string) *Type {
	return &Type{
		Kind:   TypeStruct,
		Name:   name,
		Fields: make(map[string]*Type),
	}
}

func NewFunctionType(returnType *Type, params []*Type) *Type {
	return &Type{
		Kind:   TypeFunc,
		Return: returnType,
		Params: params,
	}
}

func (t *Type) String() string {
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
	default:
		return t.Kind == other.Kind
	}
}

func (t *Type) IsNumeric() bool {
	return t.Kind == TypeInt || t.Kind == TypeFloat
}

func (t *Type) IsInteger() bool {
	return t.Kind == TypeInt
}

func (t *Type) IsFloat() bool {
	return t.Kind == TypeFloat
}

func (t *Type) IsBool() bool {
	return t.Kind == TypeBool
}

func (t *Type) IsVoid() bool {
	return t.Kind == TypeVoid
}

func (t *Type) IsString() bool {
	return t.Kind == TypeString
}

func (t *Type) IsStruct() bool {
	return t.Kind == TypeStruct
}

func (t *Type) IsFunction() bool {
	return t.Kind == TypeFunc
}

func (t *Type) IsAssignableTo(target *Type) bool {
	if t.Equals(target) {
		return true
	}

	if t.Kind == TypeInt && target.Kind == TypeFloat {
		return true
	}

	return false
}
