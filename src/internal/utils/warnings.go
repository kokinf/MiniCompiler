package utils

import (
	"fmt"
	"os"
)

type WarningCode string

const (
	WarnUnusedVariable     WarningCode = "W001"
	WarnImplicitConversion WarningCode = "W002"
	WarnUnreachableCode    WarningCode = "W003"
)

var WarningMessages = map[WarningCode]string{
	WarnUnusedVariable:     "неиспользуемая переменная",
	WarnImplicitConversion: "неявное преобразование типов",
	WarnUnreachableCode:    "недостижимый код",
}

type Warning struct {
	Code     WarningCode
	Message  string
	Line     int
	Column   int
	Context  string
	Filename string
}

type WarningCollector struct {
	warnings []*Warning
	enabled  map[WarningCode]bool
	asErrors bool
	maxCount int
}

func NewWarningCollector() *WarningCollector {
	wc := &WarningCollector{
		warnings: make([]*Warning, 0),
		enabled:  make(map[WarningCode]bool),
		maxCount: 50,
	}
	// По умолчанию включены базовые предупреждения
	wc.enabled[WarnUnusedVariable] = true
	wc.enabled[WarnImplicitConversion] = true
	return wc
}

func (wc *WarningCollector) EnableAll() {
	for code := range WarningMessages {
		wc.enabled[code] = true
	}
}

func (wc *WarningCollector) Disable(code WarningCode) {
	wc.enabled[code] = false
}

func (wc *WarningCollector) SetAsErrors(asErrors bool) {
	wc.asErrors = asErrors
}

func (wc *WarningCollector) Add(code WarningCode, message string, line, column int, context string) {
	if !wc.enabled[code] {
		return
	}
	if len(wc.warnings) >= wc.maxCount {
		return
	}
	wc.warnings = append(wc.warnings, &Warning{
		Code:    code,
		Message: message,
		Line:    line,
		Column:  column,
		Context: context,
	})
}

func (wc *WarningCollector) PrintWarnings() {
	if len(wc.warnings) == 0 {
		return
	}

	for _, w := range wc.warnings {
		prefix := ""
		if w.Filename != "" {
			prefix = fmt.Sprintf("%s:%d:%d: ", BoldText(w.Filename), w.Line, w.Column)
		}

		if wc.asErrors {
			fmt.Fprintf(os.Stderr, "%s%s %s: %s\n",
				prefix,
				RedText("ошибка"),
				YellowText(string(w.Code)),
				w.Message,
			)
		} else {
			fmt.Fprintf(os.Stderr, "%s%s %s: %s\n",
				prefix,
				YellowText("предупреждение"),
				YellowText(string(w.Code)),
				w.Message,
			)
		}

		if w.Context != "" {
			fmt.Fprintf(os.Stderr, "  %s %s\n", CyanText("-->"), w.Context)
		}
	}
}

func (wc *WarningCollector) HasWarnings() bool {
	return len(wc.warnings) > 0
}

func (wc *WarningCollector) Count() int {
	return len(wc.warnings)
}
