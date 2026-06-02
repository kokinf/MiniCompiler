package semantic

import (
	"fmt"
	"mikrocompiler/src/internal/utils"
	"strings"
)

type ErrorCode string

const (
	ErrUndeclaredIdentifier ErrorCode = "E001"
	ErrDuplicateDeclaration ErrorCode = "E002"
	ErrTypeMismatch         ErrorCode = "E003"
	ErrArgumentCount        ErrorCode = "E004"
	ErrArgumentType         ErrorCode = "E005"
	ErrInvalidReturn        ErrorCode = "E006"
	ErrInvalidCondition     ErrorCode = "E007"
	ErrUseBeforeDeclaration ErrorCode = "E008"
	ErrInvalidAssignment    ErrorCode = "E009"
	ErrInvalidUnaryOp       ErrorCode = "E010"
	ErrInvalidBinaryOp      ErrorCode = "E011"
	ErrFunctionNotFound     ErrorCode = "E012"
	ErrStructNotFound       ErrorCode = "E013"
	ErrFieldNotFound        ErrorCode = "E014"
	ErrInvalidArraySize     ErrorCode = "E015"
	ErrInvalidArrayIndex    ErrorCode = "E016"
	ErrExternMismatch       ErrorCode = "E017"
)

var ErrorMessages = map[ErrorCode]string{
	ErrUndeclaredIdentifier: "необъявленная переменная",
	ErrDuplicateDeclaration: "повторное объявление",
	ErrTypeMismatch:         "несовместимые типы",
	ErrArgumentCount:        "неверное количество аргументов",
	ErrArgumentType:         "неверный тип аргумента",
	ErrInvalidReturn:        "неверное возвращаемое значение",
	ErrInvalidCondition:     "условие должно быть типа bool",
	ErrUseBeforeDeclaration: "использование до объявления",
	ErrInvalidAssignment:    "недопустимое присваивание",
	ErrInvalidUnaryOp:       "недопустимая унарная операция",
	ErrInvalidBinaryOp:      "недопустимая бинарная операция",
	ErrFunctionNotFound:     "функция не найдена",
	ErrStructNotFound:       "структура не найдена",
	ErrFieldNotFound:        "поле не найдено",
	ErrInvalidArraySize:     "неверный размер массива",
	ErrInvalidArrayIndex:    "неверный индекс массива",
	ErrExternMismatch:       "несоответствие extern объявления",
}

type SemanticError struct {
	Code     ErrorCode
	Message  string
	Line     int
	Column   int
	Context  string
	Filename string
	Source   string // исходная строка для отображения
}

func (e *SemanticError) Error() string {
	var sb strings.Builder

	// Формат: filename:line:column: error: CODE: message
	if e.Filename != "" {
		fmt.Fprintf(&sb, "%s:%d:%d: ", utils.BoldText(e.Filename), e.Line, e.Column)
	} else {
		fmt.Fprintf(&sb, "%s:%d: ", utils.BoldText(fmt.Sprintf("строка %d", e.Line)), e.Column)
	}

	msg, ok := ErrorMessages[e.Code]
	if !ok {
		msg = "неизвестная ошибка"
	}

	fmt.Fprintf(&sb, "%s: %s: %s\n", utils.RedText("ошибка"), utils.YellowText(string(e.Code)), msg)

	if e.Context != "" {
		fmt.Fprintf(&sb, "  %s %s\n", utils.CyanText("-->"), e.Context)
	}

	if e.Source != "" {
		fmt.Fprintf(&sb, "   |\n")
		fmt.Fprintf(&sb, " %s | %s\n", utils.GrayText(fmt.Sprintf("%d", e.Line)), e.Source)
		fmt.Fprintf(&sb, "   | %s\n", utils.GreenText(strings.Repeat(" ", e.Column-1)+"^"))
	}

	fmt.Fprintf(&sb, "   | %s\n", utils.GrayText(e.Message))

	return sb.String()
}

func (ec *ErrorCollector) String() string {
	if len(ec.errors) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n%s: %d\n", utils.RedText("Ошибок"), len(ec.errors)))
	sb.WriteString(strings.Repeat("─", 60) + "\n\n")

	for i, err := range ec.errors {
		sb.WriteString(fmt.Sprintf("%s %d:\n", utils.BoldText("Ошибка"), i+1))
		sb.WriteString(err.Error())
		sb.WriteString("\n")
	}

	return sb.String()
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
