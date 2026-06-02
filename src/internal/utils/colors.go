package utils

import (
	"fmt"
	"os"
	"strings"
)

type Color string

const (
	Reset  Color = "\033[0m"
	Red    Color = "\033[0;31m"
	Green  Color = "\033[0;32m"
	Yellow Color = "\033[1;33m"
	Blue   Color = "\033[0;34m"
	Cyan   Color = "\033[0;36m"
	Bold   Color = "\033[1m"
	Gray   Color = "\033[0;90m"
)

var (
	UseColors = true
	ColorMode = "auto" // auto, always, never
)

func init() {
	// Проверяем, является ли вывод терминалом
	info, _ := os.Stdout.Stat()
	if (info.Mode() & os.ModeCharDevice) == 0 {
		UseColors = false
	}

	// Проверяем переменную окружения
	if val := os.Getenv("NO_COLOR"); val != "" {
		UseColors = false
	}
	if val := os.Getenv("CLICOLOR_FORCE"); val != "" && val != "0" {
		UseColors = true
	}
}

func SetColorMode(mode string) {
	switch mode {
	case "always":
		UseColors = true
	case "never":
		UseColors = false
	case "auto":
		info, _ := os.Stdout.Stat()
		UseColors = (info.Mode() & os.ModeCharDevice) != 0
	}
}

func colorize(color Color, text string) string {
	if !UseColors {
		return text
	}
	return string(color) + text + string(Reset)
}

func RedText(text string) string    { return colorize(Red, text) }
func GreenText(text string) string  { return colorize(Green, text) }
func YellowText(text string) string { return colorize(Yellow, text) }
func BlueText(text string) string   { return colorize(Blue, text) }
func CyanText(text string) string   { return colorize(Cyan, text) }
func BoldText(text string) string   { return colorize(Bold, text) }
func GrayText(text string) string   { return colorize(Gray, text) }

// PrintError печатает ошибку в стиле gcc/clang
func PrintError(filename string, line, column int, code, message, sourceLine, context string) {
	// Формат: filename:line:column: error: CODE: message
	prefix := ""
	if filename != "" {
		prefix = fmt.Sprintf("%s:%d:%d: ", filename, line, column)
	} else {
		prefix = fmt.Sprintf("%d:%d: ", line, column)
	}

	fmt.Fprintf(os.Stderr, "%s%s %s: %s\n",
		BoldText(prefix),
		RedText("ошибка"),
		YellowText(string(code)),
		message,
	)

	if context != "" {
		fmt.Fprintf(os.Stderr, "  %s\n", CyanText("--> "+context))
	}

	if sourceLine != "" {
		fmt.Fprintf(os.Stderr, "   |\n")
		fmt.Fprintf(os.Stderr, " %s | %s\n", GrayText(fmt.Sprintf("%d", line)), sourceLine)
		fmt.Fprintf(os.Stderr, "   | %s\n", GreenText(strings.Repeat(" ", column-1)+"^"))
	}

	fmt.Fprintf(os.Stderr, "   | %s\n", GrayText(message))
}

// PrintWarning печатает предупреждение
func PrintWarning(filename string, line, column int, code, message, sourceLine string) {
	prefix := ""
	if filename != "" {
		prefix = fmt.Sprintf("%s:%d:%d: ", filename, line, column)
	} else {
		prefix = fmt.Sprintf("%d:%d: ", line, column)
	}

	fmt.Fprintf(os.Stderr, "%s%s %s: %s\n",
		BoldText(prefix),
		YellowText("предупреждение"),
		YellowText(string(code)),
		message,
	)

	if sourceLine != "" {
		fmt.Fprintf(os.Stderr, "   |\n")
		fmt.Fprintf(os.Stderr, " %s | %s\n", GrayText(fmt.Sprintf("%d", line)), sourceLine)
	}
}

// ProgressBar показывает прогресс
type ProgressBar struct {
	Total     int
	Current   int
	Prefix    string
	BarLength int
}

func NewProgressBar(total int, prefix string) *ProgressBar {
	return &ProgressBar{
		Total:     total,
		Current:   0,
		Prefix:    prefix,
		BarLength: 40,
	}
}

func (pb *ProgressBar) Increment() {
	pb.Current++
}

func (pb *ProgressBar) Print() {
	percent := float64(pb.Current) / float64(pb.Total)
	filled := int(percent * float64(pb.BarLength))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", pb.BarLength-filled)
	fmt.Fprintf(os.Stderr, "\r%s [%s] %3.0f%%", pb.Prefix, bar, percent*100)
}

func (pb *ProgressBar) Finish() {
	fmt.Fprintf(os.Stderr, "\n")
}
