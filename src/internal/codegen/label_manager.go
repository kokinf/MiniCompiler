package codegen

import (
	"fmt"
	"strings"
)

// LabelManager управляет генерацией уникальных меток для ассемблерного кода
type LabelManager struct {
	counters map[string]int
	labels   map[string]string // маппинг IR меток на ассемблерные
}

// NewLabelManager создает новый менеджер меток
func NewLabelManager() *LabelManager {
	return &LabelManager{
		counters: make(map[string]int),
		labels:   make(map[string]string),
	}
}

// NewLabel генерирует уникальную метку с заданным префиксом
func (lm *LabelManager) NewLabel(prefix string) string {
	lm.counters[prefix]++
	return fmt.Sprintf(".L%s%d", prefix, lm.counters[prefix])
}

// GetLabel возвращает существующую метку или создает новую
func (lm *LabelManager) GetLabel(name string) string {
	if label, exists := lm.labels[name]; exists {
		return label
	}
	label := lm.sanitizeLabel(name)
	lm.labels[name] = label
	return label
}

// sanitizeLabel преобразует имя метки в допустимый ассемблерный идентификатор
func (lm *LabelManager) sanitizeLabel(name string) string {
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")

	if !strings.HasPrefix(name, ".L") {
		name = ".L" + name
	}

	return name
}

// IfLabel генерирует метки для if-else конструкций
func (lm *LabelManager) IfLabels() (thenLabel, elseLabel, endLabel string) {
	thenLabel = lm.NewLabel("if_then")
	elseLabel = lm.NewLabel("if_else")
	endLabel = lm.NewLabel("if_end")
	return
}

// WhileLabel генерирует метки для while циклов
func (lm *LabelManager) WhileLabels() (startLabel, bodyLabel, endLabel string) {
	startLabel = lm.NewLabel("while_start")
	bodyLabel = lm.NewLabel("while_body")
	endLabel = lm.NewLabel("while_end")
	return
}

// ForLabel генерирует метки для for циклов
func (lm *LabelManager) ForLabels() (startLabel, bodyLabel, updateLabel, endLabel string) {
	startLabel = lm.NewLabel("for_start")
	bodyLabel = lm.NewLabel("for_body")
	updateLabel = lm.NewLabel("for_update")
	endLabel = lm.NewLabel("for_end")
	return
}

// LogicalLabel генерирует метки для логических операторов (short-circuit)
func (lm *LabelManager) LogicalLabels(op string) (trueLabel, falseLabel, endLabel string) {
	prefix := "log_" + op
	trueLabel = lm.NewLabel(prefix + "_true")
	falseLabel = lm.NewLabel(prefix + "_false")
	endLabel = lm.NewLabel(prefix + "_end")
	return
}

// StringLabel генерирует метку для строкового литерала
func (lm *LabelManager) StringLabel(index int) string {
	return fmt.Sprintf(".L.str%d", index)
}

// Reset сбрасывает все счетчики
func (lm *LabelManager) Reset() {
	lm.counters = make(map[string]int)
	lm.labels = make(map[string]string)
}
