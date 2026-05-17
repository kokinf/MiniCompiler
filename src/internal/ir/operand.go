package ir

import (
	"fmt"
	"strconv"
)

// OperandType определяет тип операнда
type OperandType int

const (
	OperandTemp OperandType = iota
	OperandVar
	OperandLiteral
	OperandLabel
	OperandGlobal
)

// Operand представляет операнд в IR инструкции
type Operand struct {
	Type  OperandType
	Name  string
	Value interface{}
}

// NewTempOperand создает новую временную переменную
func NewTempOperand(id int) *Operand {
	return &Operand{
		Type: OperandTemp,
		Name: fmt.Sprintf("%%t%d", id),
	}
}

// NewVarOperand создает операнд для переменной
func NewVarOperand(name string) *Operand {
	return &Operand{
		Type: OperandVar,
		Name: name,
	}
}

// NewLiteralOperand создает операнд-литерал
func NewLiteralOperand(value interface{}) *Operand {
	return &Operand{
		Type:  OperandLiteral,
		Value: value,
	}
}

// NewLabelOperand создает операнд-метку
func NewLabelOperand(name string) *Operand {
	return &Operand{
		Type: OperandLabel,
		Name: name,
	}
}

// NewGlobalOperand создает операнд для глобальной переменной
func NewGlobalOperand(name string) *Operand {
	return &Operand{
		Type: OperandGlobal,
		Name: name,
	}
}

func (o *Operand) String() string {
	switch o.Type {
	case OperandTemp:
		return o.Name
	case OperandVar:
		return o.Name
	case OperandLiteral:
		switch v := o.Value.(type) {
		case int, int32, int64:
			return fmt.Sprintf("%d", v)
		case float32, float64:
			return fmt.Sprintf("%g", v)
		case bool:
			if v {
				return "true"
			}
			return "false"
		case string:
			return strconv.Quote(v)
		default:
			return fmt.Sprintf("%v", v)
		}
	case OperandLabel:
		return o.Name
	case OperandGlobal:
		return "@" + o.Name
	default:
		return "?"
	}
}

// IsConstant возвращает true, если операнд - константа
func (o *Operand) IsConstant() bool {
	return o.Type == OperandLiteral
}

// GetIntValue возвращает целочисленное значение литерала
func (o *Operand) GetIntValue() (int, bool) {
	if o.Type != OperandLiteral {
		return 0, false
	}
	switch v := o.Value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	default:
		return 0, false
	}
}

// GetFloatValue возвращает значение с плавающей точкой
func (o *Operand) GetFloatValue() (float64, bool) {
	if o.Type != OperandLiteral {
		return 0, false
	}
	switch v := o.Value.(type) {
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

// GetBoolValue возвращает булево значение
func (o *Operand) GetBoolValue() (bool, bool) {
	if o.Type != OperandLiteral {
		return false, false
	}
	if v, ok := o.Value.(bool); ok {
		return v, true
	}
	return false, false
}
