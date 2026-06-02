package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type ErrorFormatter struct {
	Format string
}

func NewErrorFormatter(format string) *ErrorFormatter {
	return &ErrorFormatter{Format: format}
}

type JSONError struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Context  string `json:"context,omitempty"`
	Severity string `json:"severity"`
}

type JSONReport struct {
	Errors   []JSONError `json:"errors"`
	Warnings []JSONError `json:"warnings"`
	Summary  string      `json:"summary"`
}

var jsonErrors []JSONError
var jsonWarnings []JSONError

func AddJSONError(code, message, filename, context string, line, column int) {
	jsonErrors = append(jsonErrors, JSONError{
		File:     filename,
		Line:     line,
		Column:   column,
		Code:     code,
		Message:  message,
		Context:  context,
		Severity: "error",
	})
}

func AddJSONWarning(code, message, filename, context string, line, column int) {
	jsonWarnings = append(jsonWarnings, JSONError{
		File:     filename,
		Line:     line,
		Column:   column,
		Code:     code,
		Message:  message,
		Context:  context,
		Severity: "warning",
	})
}

func PrintJSONReport() {
	report := JSONReport{
		Errors:   jsonErrors,
		Warnings: jsonWarnings,
		Summary:  fmt.Sprintf("%d errors, %d warnings", len(jsonErrors), len(jsonWarnings)),
	}

	data, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(data))
}

// PrintError форматирует ошибку в зависимости от формата
func PrintErrorFormatted(filename string, line, column int, code, message, sourceLine, context string, format string) {
	switch format {
	case "json":
		AddJSONError(code, message, filename, context, line, column)
	case "gcc":
		// GCC-совместимый формат
		prefix := ""
		if filename != "" {
			prefix = fmt.Sprintf("%s:%d:%d: ", filename, line, column)
		}
		fmt.Fprintf(os.Stderr, "%serror: %s: %s\n", prefix, code, message)
		if context != "" {
			fmt.Fprintf(os.Stderr, "   in %s\n", context)
		}
	default:
		// Человеко-читаемый формат с цветами
		prefix := ""
		if filename != "" {
			prefix = fmt.Sprintf("%s:%d:%d: ", BoldText(filename), line, column)
		} else {
			prefix = fmt.Sprintf("%d:%d: ", line, column)
		}

		fmt.Fprintf(os.Stderr, "%s%s %s: %s\n",
			prefix,
			RedText("ошибка"),
			YellowText(string(code)),
			message,
		)

		if context != "" {
			fmt.Fprintf(os.Stderr, "  %s %s\n", CyanText("-->"), context)
		}

		if sourceLine != "" {
			fmt.Fprintf(os.Stderr, "   |\n")
			fmt.Fprintf(os.Stderr, " %s | %s\n", GrayText(fmt.Sprintf("%d", line)), sourceLine)
			fmt.Fprintf(os.Stderr, "   | %s\n", GreenText(strings.Repeat(" ", column-1)+"^"))
		}

		fmt.Fprintf(os.Stderr, "   | %s\n", GrayText(message))
	}
}
