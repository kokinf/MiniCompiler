package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"mikrocompiler/src/internal/ast"
	"mikrocompiler/src/internal/codegen"
	"mikrocompiler/src/internal/ir"
	"mikrocompiler/src/internal/lexer"
	"mikrocompiler/src/internal/parser"
	"mikrocompiler/src/internal/semantic"
	mytoken "mikrocompiler/src/internal/token"
)

func main() {
	// Команда lex
	lexCmd := flag.NewFlagSet("lex", flag.ExitOnError)
	inputFile := lexCmd.String("input", "", "Входной исходный файл")
	outputFile := lexCmd.String("output", "", "Выходной файл с токенами (необязательно)")

	// Команда parse
	parseCmd := flag.NewFlagSet("parse", flag.ExitOnError)
	parseInput := parseCmd.String("input", "", "Входной исходный файл")
	parseOutput := parseCmd.String("output", "", "Выходной файл с AST")
	parseFormat := parseCmd.String("format", "text", "Формат вывода AST: text, dot, json")
	parseVerbose := parseCmd.Bool("verbose", false, "Подробный вывод")

	// Команда check
	checkCmd := flag.NewFlagSet("check", flag.ExitOnError)
	checkInput := checkCmd.String("input", "", "Входной исходный файл")
	checkOutput := checkCmd.String("output", "", "Выходной файл для результатов")
	checkVerbose := checkCmd.Bool("verbose", false, "Подробный вывод")
	checkShowTypes := checkCmd.Bool("show-types", false, "Показывать типы выражений")

	// Команда symbols
	symbolsCmd := flag.NewFlagSet("symbols", flag.ExitOnError)
	symbolsInput := symbolsCmd.String("input", "", "Входной исходный файл")
	symbolsFormat := symbolsCmd.String("format", "text", "Формат вывода: text, json")

	// Команда ir
	irCmd := flag.NewFlagSet("ir", flag.ExitOnError)
	irInput := irCmd.String("input", "", "Входной исходный файл")
	irOutput := irCmd.String("output", "", "Выходной файл для IR")
	irFormat := irCmd.String("format", "text", "Формат вывода IR: text, dot, json")
	irOptimize := irCmd.Bool("optimize", false, "Применить базовые оптимизации")
	irStats := irCmd.Bool("stats", false, "Показать статистику IR")

	// Команда compile
	compileCmd := flag.NewFlagSet("compile", flag.ExitOnError)
	compileInput := compileCmd.String("input", "", "Входной исходный файл")
	compileOutput := compileCmd.String("output", "", "Выходной файл с ассемблером")
	compileTarget := compileCmd.String("target", "x86_64", "Целевая архитектура: x86_64")
	compileOptimize := compileCmd.Bool("optimize", false, "Включить оптимизации")

	// Команда test
	testCmd := flag.NewFlagSet("test", flag.ExitOnError)
	testType := testCmd.String("type", "all", "Тип тестов: all, lexer, parser, semantic, ir, codegen")

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
		runParser(*parseInput, *parseOutput, *parseFormat, *parseVerbose)
	case "check":
		checkCmd.Parse(os.Args[2:])
		runSemanticCheck(*checkInput, *checkOutput, *checkVerbose, *checkShowTypes)
	case "symbols":
		symbolsCmd.Parse(os.Args[2:])
		runSymbolTable(*symbolsInput, *symbolsFormat)
	case "ir":
		irCmd.Parse(os.Args[2:])
		runIRGenerator(*irInput, *irOutput, *irFormat, *irOptimize, *irStats)
	case "compile":
		compileCmd.Parse(os.Args[2:])
		runCompiler(*compileInput, *compileOutput, *compileTarget, *compileOptimize)
	case "test":
		testCmd.Parse(os.Args[2:])
		runTests(*testType)
	default:
		fmt.Printf("Неизвестная команда: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Использование: compiler <команда> [опции]")
	fmt.Println()
	fmt.Println("Команды:")
	fmt.Println("  lex     --input <файл> [--output <файл>]")
	fmt.Println("  parse   --input <файл> [--output <файл>] [--format text|dot|json] [--verbose]")
	fmt.Println("  check   --input <файл> [--output <файл>] [--verbose] [--show-types]")
	fmt.Println("  symbols --input <файл> [--format text|json]")
	fmt.Println("  ir      --input <файл> [--output <файл>] [--format text|dot|json] [--optimize] [--stats]")
	fmt.Println("  compile --input <файл> [--output <файл>] [--target x86_64] [--optimize]")
	fmt.Println("  test    --type <all|lexer|parser|semantic|ir|codegen>")
}

func readSourceFile(inputFile string) (string, []mytoken.Token) {
	if inputFile == "" {
		fmt.Fprintln(os.Stderr, "Ошибка: необходимо указать входной файл (--input)")
		os.Exit(1)
	}

	content, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения файла '%s': %v\n", inputFile, err)
		os.Exit(1)
	}

	scanner := lexer.NewScanner(string(content))
	var tokens []mytoken.Token

	for !scanner.IsAtEnd() {
		tok := scanner.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == mytoken.EOF {
			break
		}
	}

	return string(content), tokens
}

func checkLexErrors(tokens []mytoken.Token) bool {
	hasErrors := false
	for _, tok := range tokens {
		if tok.Type == mytoken.ILLEGAL {
			fmt.Fprintf(os.Stderr, "Лексическая ошибка: %s\n", tok.Lexeme)
			hasErrors = true
		}
	}
	if hasErrors {
		os.Exit(1)
	}
	return false
}

func parseSource(tokens []mytoken.Token) *ast.ProgramNode {
	p := parser.NewParser(tokens)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		fmt.Fprintln(os.Stderr, "Ошибки парсинга:")
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", err)
		}
		os.Exit(1)
	}

	if program == nil {
		fmt.Fprintln(os.Stderr, "Ошибка: не удалось распарсить программу")
		os.Exit(1)
	}

	return program
}

func runLexer(inputFile, outputFile string) {
	_, tokens := readSourceFile(inputFile)
	checkLexErrors(tokens)

	var output strings.Builder
	for _, tok := range tokens {
		output.WriteString(tok.String() + "\n")
	}

	if outputFile != "" {
		os.WriteFile(outputFile, []byte(output.String()), 0644)
		fmt.Printf("Токены записаны в %s\n", outputFile)
	} else {
		fmt.Print(output.String())
	}
}

func runParser(inputFile, outputFile, format string, verbose bool) {
	_, tokens := readSourceFile(inputFile)
	checkLexErrors(tokens)
	program := parseSource(tokens)

	var output string
	switch format {
	case "text":
		printer := ast.NewPrettyPrinter()
		output = printer.Print(program)
	case "dot":
		printer := ast.NewDOTPrinter()
		output = printer.Print(program)
	case "json":
		printer := ast.NewJSONPrinter()
		output = printer.Print(program)
	default:
		fmt.Fprintf(os.Stderr, "Неизвестный формат: %s\n", format)
		os.Exit(1)
	}

	if outputFile != "" {
		os.WriteFile(outputFile, []byte(output), 0644)
		fmt.Printf("AST записан в %s\n", outputFile)
	} else {
		fmt.Print(output)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "\nСтатистика:\n")
		fmt.Fprintf(os.Stderr, "  Объявлений: %d\n", len(program.Declarations))
	}
}

func runSemanticCheck(inputFile, outputFile string, verbose, showTypes bool) {
	_, tokens := readSourceFile(inputFile)
	checkLexErrors(tokens)
	program := parseSource(tokens)

	analyzer := semantic.NewSemanticAnalyzer()
	symbolTable, errors, decoratedAST := analyzer.Analyze(program)

	var output strings.Builder

	if verbose {
		output.WriteString("=== Семантический анализ ===\n\n")
	}

	output.WriteString(symbolTable.String())

	if showTypes && decoratedAST != nil {
		output.WriteString("\n=== Декорированный AST ===\n")
		printer := ast.NewPrettyPrinter()
		output.WriteString(printer.Print(decoratedAST))
	}

	if len(errors.Errors()) > 0 {
		output.WriteString("\n=== Ошибки ===\n")
		output.WriteString(errors.String())
	}

	if verbose {
		output.WriteString("\n=== Отчёт валидации ===\n")
		allErrors := errors.Errors()
		output.WriteString(fmt.Sprintf("Ошибок: %d\n", len(allErrors)))

		globalScope := symbolTable.GetGlobalScope()
		output.WriteString(fmt.Sprintf("\nГлобальных символов: %d\n", len(globalScope.GetAllSymbols())))
		for _, sym := range globalScope.GetAllSymbols() {
			externTag := ""
			if sym.IsExtern {
				externTag = " [extern]"
			}
			output.WriteString(fmt.Sprintf("  %s: %s %s (line %d)%s\n", sym.Name, sym.Kind, sym.Type.String(), sym.Line, externTag))
		}

		typeSystem := semantic.NewTypeSystem()
		output.WriteString(fmt.Sprintf("\nРазмеры типов:\n"))
		output.WriteString(fmt.Sprintf("  int: %d байт\n", typeSystem.GetSize(typeSystem.IntType)))
		output.WriteString(fmt.Sprintf("  float: %d байт\n", typeSystem.GetSize(typeSystem.FloatType)))
		output.WriteString(fmt.Sprintf("  bool: %d байт\n", typeSystem.GetSize(typeSystem.BoolType)))
		output.WriteString(fmt.Sprintf("  string: %d байт\n", typeSystem.GetSize(typeSystem.StringType)))
	}

	if outputFile != "" {
		os.WriteFile(outputFile, []byte(output.String()), 0644)
		fmt.Printf("Результаты записаны в %s\n", outputFile)
	} else {
		fmt.Print(output.String())
	}

	if len(errors.Errors()) > 0 {
		os.Exit(1)
	}
}

func runSymbolTable(inputFile, format string) {
	_, tokens := readSourceFile(inputFile)
	checkLexErrors(tokens)
	program := parseSource(tokens)

	analyzer := semantic.NewSemanticAnalyzer()
	symbolTable, _, _ := analyzer.Analyze(program)

	if format == "text" {
		fmt.Print(symbolTable.String())
	} else if format == "json" {
		fmt.Println("{")
		fmt.Println("  \"scopes\": [")
		globalScope := symbolTable.GetGlobalScope()
		symbols := globalScope.GetAllSymbols()
		for i, sym := range symbols {
			comma := ","
			if i == len(symbols)-1 {
				comma = ""
			}
			externFlag := "false"
			if sym.IsExtern {
				externFlag = "true"
			}
			fmt.Printf("    {\"name\": \"%s\", \"kind\": \"%s\", \"type\": \"%s\", \"line\": %d, \"extern\": %s}%s\n",
				sym.Name, sym.Kind, sym.Type.String(), sym.Line, externFlag, comma)
		}
		fmt.Println("  ]")
		fmt.Println("}")
	}
}

func runIRGenerator(inputFile, outputFile, format string, optimize, showStats bool) {
	_, tokens := readSourceFile(inputFile)
	checkLexErrors(tokens)
	program := parseSource(tokens)

	analyzer := semantic.NewSemanticAnalyzer()
	symbolTable, errors, _ := analyzer.Analyze(program)

	if len(errors.Errors()) > 0 {
		fmt.Fprintln(os.Stderr, "Семантические ошибки:")
		fmt.Fprint(os.Stderr, errors.String())
		os.Exit(1)
	}

	typeSystem := semantic.NewTypeSystem()
	irGenerator := ir.NewIRGenerator(symbolTable, typeSystem)
	irProgram := irGenerator.Generate(program)

	if optimize {
		optimizer := ir.NewPeepholeOptimizer(irProgram)
		optimizer.Optimize()
	}

	var output strings.Builder

	if showStats {
		stats := ir.CollectStats(irProgram)
		output.WriteString(stats.String())
		output.WriteString("\n")
	}

	switch format {
	case "text":
		printer := ir.NewTextPrinter()
		output.WriteString(printer.Print(irProgram))
	case "dot":
		printer := ir.NewDOTPrinter()
		output.WriteString(printer.Print(irProgram))
	case "json":
		printer := ir.NewJSONPrinter()
		output.WriteString(printer.Print(irProgram))
	default:
		fmt.Fprintf(os.Stderr, "Неизвестный формат: %s\n", format)
		os.Exit(1)
	}

	if outputFile != "" {
		os.WriteFile(outputFile, []byte(output.String()), 0644)
		fmt.Printf("IR записан в %s\n", outputFile)
	} else {
		fmt.Print(output.String())
	}
}

func runCompiler(inputFile, outputFile, target string, optimize bool) {
	if inputFile == "" {
		fmt.Fprintln(os.Stderr, "Ошибка: необходимо указать входной файл")
		os.Exit(1)
	}

	if target != "x86_64" {
		fmt.Fprintf(os.Stderr, "Ошибка: неподдерживаемая архитектура '%s'. Поддерживается только x86_64\n", target)
		os.Exit(1)
	}

	_, tokens := readSourceFile(inputFile)
	checkLexErrors(tokens)
	program := parseSource(tokens)

	// Семантический анализ
	analyzer := semantic.NewSemanticAnalyzer()
	symbolTable, errors, decoratedAST := analyzer.Analyze(program)

	if len(errors.Errors()) > 0 {
		fmt.Fprintln(os.Stderr, "Семантические ошибки:")
		fmt.Fprint(os.Stderr, errors.String())
		os.Exit(1)
	}

	// Генерация IR
	typeSystem := semantic.NewTypeSystem()
	irGenerator := ir.NewIRGenerator(symbolTable, typeSystem)
	irProgram := irGenerator.Generate(decoratedAST)

	// Оптимизация IR
	if optimize {
		optimizer := ir.NewPeepholeOptimizer(irProgram)
		optimizer.Optimize()
		if outputFile == "" {
			fmt.Fprintf(os.Stderr, "\n%s\n", optimizer.GetOptimizationReport())
		}
	}

	// Генерация x86-64 ассемблера
	codeGenerator := codegen.NewX86Generator(irProgram, symbolTable, typeSystem)
	asmCode := codeGenerator.Generate()

	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(asmCode), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка записи: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Ассемблерный код записан в %s\n", outputFile)
		fmt.Println()
		fmt.Println("Для сборки исполняемого файла выполните:")
		fmt.Println("  nasm -f elf64 -o runtime.o src/runtime/runtime.asm")
		fmt.Printf("  nasm -f elf64 -o program.o %s\n", outputFile)
		fmt.Println("  gcc -no-pie -o program runtime.o program.o")
		fmt.Println()
		fmt.Println("Примечание: используется gcc для линковки с libc (malloc, free, printf, abort)")
	} else {
		fmt.Print(asmCode)
	}
}

func runTests(testType string) {
	fmt.Printf("=== Запуск тестов: %s ===\n\n", testType)

	switch testType {
	case "all", "lexer", "parser", "semantic", "ir", "codegen", "control-flow":
		fmt.Println("Для запуска тестов используйте: make test")
		fmt.Println("Или запустите скрипт: tests/test_runner/run_tests.sh")
	default:
		fmt.Printf("Неизвестный тип тестов: %s\n", testType)
		fmt.Println("Доступные типы: all, lexer, parser, semantic, ir, codegen, control-flow")
	}
}
