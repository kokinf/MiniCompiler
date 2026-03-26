package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"mikrocompiler/src/internal/lexer"
	"mikrocompiler/src/internal/parser"
	"mikrocompiler/src/internal/semantic"
	mytoken "mikrocompiler/src/internal/token"
)

func main() {
	lexCmd := flag.NewFlagSet("lex", flag.ExitOnError)
	inputFile := lexCmd.String("input", "", "Входной исходный файл")
	outputFile := lexCmd.String("output", "", "Выходной файл с токенами (необязательно)")

	parseCmd := flag.NewFlagSet("parse", flag.ExitOnError)
	parseInput := parseCmd.String("input", "", "Входной исходный файл")
	parseOutput := parseCmd.String("output", "", "Выходной файл с AST")
	parseFormat := parseCmd.String("format", "text", "Формат вывода AST: text, dot, json")

	checkCmd := flag.NewFlagSet("check", flag.ExitOnError)
	checkInput := checkCmd.String("input", "", "Входной исходный файл")
	checkOutput := checkCmd.String("output", "", "Выходной файл для результатов семантического анализа")
	checkVerbose := checkCmd.Bool("verbose", false, "Подробный вывод")
	checkShowTypes := checkCmd.Bool("show-types", false, "Показывать типы выражений")

	symbolsCmd := flag.NewFlagSet("symbols", flag.ExitOnError)
	symbolsInput := symbolsCmd.String("input", "", "Входной исходный файл")
	symbolsFormat := symbolsCmd.String("format", "text", "Формат вывода: text, json")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "lex":
		lexCmd.Parse(os.Args[2:])
		runLexer(*inputFile, *outputFile)
	case "parse":
		parseCmd.Parse(os.Args[2:])
		runParser(*parseInput, *parseOutput, *parseFormat)
	case "check":
		checkCmd.Parse(os.Args[2:])
		runSemanticCheck(*checkInput, *checkOutput, *checkVerbose, *checkShowTypes)
	case "symbols":
		symbolsCmd.Parse(os.Args[2:])
		runSymbolTable(*symbolsInput, *symbolsFormat)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Использование: compiler <команда> [опции]")
	fmt.Println("Команды:")
	fmt.Println("  lex --input <файл> [--output <файл>]                     Запуск лексера")
	fmt.Println("  parse --input <файл> [--output <файл>] [--format text|dot|json]  Запуск парсера")
	fmt.Println("  check --input <файл> [--output <файл>] [--verbose] [--show-types]  Семантический анализ")
	fmt.Println("  symbols --input <файл> [--format text|json]             Вывод таблицы символов")
}

func runLexer(inputFile, outputFile string) {
	// ... existing code ...
}

func runParser(inputFile, outputFile, format string) {
	// ... existing code ...
}

func runSemanticCheck(inputFile, outputFile string, verbose, showTypes bool) {
	if inputFile == "" {
		fmt.Fprintln(os.Stderr, "Ошибка: необходимо указать входной файл")
		fmt.Fprintln(os.Stderr, "Использование: compiler check --input <файл> [--output <файл>]")
		os.Exit(1)
	}

	content, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения входного файла: %v\n", err)
		os.Exit(1)
	}

	// Lexical analysis
	scanner := lexer.NewScanner(string(content))
	var tokens []mytoken.Token

	for {
		tok := scanner.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == mytoken.EOF {
			break
		}
	}

	hasLexErrors := false
	for _, tok := range tokens {
		if tok.Type == mytoken.ILLEGAL {
			hasLexErrors = true
			fmt.Fprintf(os.Stderr, "Лексическая ошибка: %s\n", tok.Lexeme)
		}
	}
	if hasLexErrors {
		os.Exit(1)
	}

	// Parsing
	p := parser.NewParser(tokens)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		fmt.Fprintln(os.Stderr, "Ошибки парсинга:")
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", err)
		}
		os.Exit(1)
	}

	// Semantic analysis
	analyzer := semantic.NewSemanticAnalyzer()
	symbolTable, errors := analyzer.Analyze(program)

	var output strings.Builder

	if verbose {
		output.WriteString("=== Семантический анализ ===\n\n")
	}

	if showTypes {
		output.WriteString("=== Типы выражений ===\n")
		// TODO: Print type annotations
		output.WriteString("\n")
	}

	// Print symbol table
	output.WriteString(symbolTable.String())

	if len(errors.Errors()) > 0 {
		output.WriteString("\n=== Ошибки семантического анализа ===\n")
		output.WriteString(errors.String())
	}

	if outputFile != "" {
		err = os.WriteFile(outputFile, []byte(output.String()), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка записи выходного файла: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Результаты семантического анализа записаны в %s\n", outputFile)
	} else {
		fmt.Print(output.String())
	}

	if len(errors.Errors()) > 0 {
		os.Exit(1)
	}
}

func runSymbolTable(inputFile, format string) {
	if inputFile == "" {
		fmt.Fprintln(os.Stderr, "Ошибка: необходимо указать входной файл")
		fmt.Fprintln(os.Stderr, "Использование: compiler symbols --input <файл> [--format text|json]")
		os.Exit(1)
	}

	content, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения входного файла: %v\n", err)
		os.Exit(1)
	}

	// Lexical analysis
	scanner := lexer.NewScanner(string(content))
	var tokens []mytoken.Token

	for {
		tok := scanner.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == mytoken.EOF {
			break
		}
	}

	// Parsing
	p := parser.NewParser(tokens)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		fmt.Fprintln(os.Stderr, "Ошибки парсинга:")
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", err)
		}
		os.Exit(1)
	}

	// Semantic analysis
	analyzer := semantic.NewSemanticAnalyzer()
	symbolTable, _ := analyzer.Analyze(program)

	switch format {
	case "text":
		fmt.Print(symbolTable.String())
	case "json":
		// TODO: Implement JSON output
		fmt.Println("{\"error\": \"JSON format not yet implemented\"}")
	default:
		fmt.Fprintf(os.Stderr, "Неизвестный формат: %s\n", format)
		os.Exit(1)
	}
}
