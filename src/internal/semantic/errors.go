package semantic

import (
	"fmt"
)

type ErrorCode string

const (
	ErrUndeclaredIdentifier ErrorCode = "undeclared_identifier"
	ErrDuplicateDeclaration ErrorCode = "duplicate_declaration"
	ErrTypeMismatch         ErrorCode = "type_mismatch"
	ErrArgumentCount        ErrorCode = "argument_count_mismatch"
	ErrArgumentType         ErrorCode = "argument_type_mismatch"
	ErrInvalidReturn        ErrorCode = "invalid_return"
	ErrInvalidCondition     ErrorCode = "invalid_condition"
	ErrUseBeforeDeclaration ErrorCode = "use_before_declaration"
	ErrInvalidAssignment    ErrorCode = "invalid_assignment"
	ErrInvalidUnaryOp       ErrorCode = "invalid_unary_operator"
	ErrInvalidBinaryOp      ErrorCode = "invalid_binary_operator"
	ErrFunctionNotFound     ErrorCode = "function_not_found"
	ErrStructNotFound       ErrorCode = "struct_not_found"
	ErrFieldNotFound        ErrorCode = "field_not_found"
)

type SemanticError struct {
	Code    ErrorCode
	Message string
	Line    int
	Column  int
	Context string
}

func (e *SemanticError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("%s: %s\n  --> line %d, column %d\n  |\n  = in %s\n  = %s",
			e.Code, e.Message, e.Line, e.Column, e.Context, e.Message)
	}
	return fmt.Sprintf("%s: %s\n  --> line %d, column %d", e.Code, e.Message, e.Line, e.Column)
}

type ErrorCollector struct {
	errors []*SemanticError
}

func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors: make([]*SemanticError, 0),
	}
}

func (ec *ErrorCollector) Add(code ErrorCode, message string, line, column int, context string) {
	ec.errors = append(ec.errors, &SemanticError{
		Code:    code,
		Message: message,
		Line:    line,
		Column:  column,
		Context: context,
	})
}

func (ec *ErrorCollector) AddError(err *SemanticError) {
	ec.errors = append(ec.errors, err)
}

func (ec *ErrorCollector) HasErrors() bool {
	return len(ec.errors) > 0
}

func (ec *ErrorCollector) Errors() []*SemanticError {
	return ec.errors
}

func (ec *ErrorCollector) String() string {
	result := ""
	for _, err := range ec.errors {
		result += err.Error() + "\n\n"
	}
	return result
}
