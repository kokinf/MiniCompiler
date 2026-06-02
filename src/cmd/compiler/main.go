package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"mikrocompiler/src/internal/ast"
	"mikrocompiler/src/internal/codegen"
	"mikrocompiler/src/internal/ir"
	"mikrocompiler/src/internal/lexer"
	"mikrocompiler/src/internal/parser"
	"mikrocompiler/src/internal/semantic"
	mytoken "mikrocompiler/src/internal/token"
	"mikrocompiler/src/internal/utils"
)

const (
	VERSION = "1.0.0"
	AUTHOR  = "Fokin Nikita"
	BUILD   = "2026-06-01"
	TARGET  = "x86_64-linux-gnu"
)

var (
	optOutput       string
	optVerbose      bool
	optHelp         bool
	optVersion      bool
	optOptimize     bool
	optAssemblyOnly bool
	optCompileOnly  bool
	optPreprocess   bool
	optShowAST      bool
	optShowIR       bool
	optTarget       string
	optFormat       string
	optStats        bool
	optShowTypes    bool
	optColor        string
	optMaxErrors    int
	optErrorFormat  string
	optWall         bool
	optWerror       bool
	optWno          string
)

func main() {
	flag.StringVar(&optOutput, "o", "", "Выходной файл")
	flag.StringVar(&optOutput, "output", "", "Выходной файл")
	flag.BoolVar(&optAssemblyOnly, "S", false, "Только ассемблер (без линковки)")
	flag.BoolVar(&optCompileOnly, "c", false, "Компиляция в объектный файл")
	flag.BoolVar(&optPreprocess, "E", false, "Препроцессор (вывод токенов)")
	flag.BoolVar(&optVerbose, "v", false, "Подробный вывод")
	flag.BoolVar(&optVerbose, "verbose", false, "Подробный вывод")
	flag.BoolVar(&optHelp, "h", false, "Показать справку")
	flag.BoolVar(&optHelp, "help", false, "Показать справку")
	flag.BoolVar(&optOptimize, "O", false, "Включить оптимизации")
	flag.BoolVar(&optOptimize, "optimize", false, "Включить оптимизации")
	flag.BoolVar(&optShowAST, "ast", false, "Вывести AST")
	flag.BoolVar(&optShowIR, "ir", false, "Вывести IR")
	flag.StringVar(&optTarget, "target", "x86_64", "Целевая архитектура")
	flag.StringVar(&optFormat, "format", "text", "Формат вывода: text, dot, json")
	flag.BoolVar(&optStats, "stats", false, "Показать статистику")
	flag.BoolVar(&optShowTypes, "show-types", false, "Показывать типы выражений")
	flag.StringVar(&optColor, "color", "auto", "Цветной вывод: auto, always, never")
	flag.IntVar(&optMaxErrors, "max-errors", 20, "Максимальное количество ошибок")
	flag.StringVar(&optErrorFormat, "error-format", "human", "Формат ошибок: human, json, gcc")
	flag.BoolVar(&optWall, "Wall", false, "Включить все предупреждения")
	flag.BoolVar(&optWerror, "Werror", false, "Считать предупреждения ошибками")
	flag.StringVar(&optWno, "Wno", "", "Отключить предупреждение")

	flag.Parse()

	// Ручной парсинг -o если он после файла
	for i, arg := range os.Args {
		if (arg == "-o" || arg == "--output") && i+1 < len(os.Args) {
			if optOutput == "" {
				optOutput = os.Args[i+1]
			}
		}
	}

	utils.SetColorMode(optColor)

	if optVersion {
		printVersion()
		os.Exit(0)
	}

	if optHelp || len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) == 0 {
		printHelp()
		os.Exit(1)
	}

	var srcFiles []string
	for _, arg := range args {
		if strings.HasSuffix(arg, ".src") {
			srcFiles = append(srcFiles, arg)
		}
	}

	if len(srcFiles) > 1 {
		runMultiFileCompiler(srcFiles, optOutput)
		return
	}

	command := args[0]

	if strings.HasSuffix(command, ".src") {
		if optOutput == "" {
			if optCompileOnly {
				optOutput = strings.TrimSuffix(command, ".src") + ".o"
			} else {
				optOutput = strings.TrimSuffix(command, ".src") + ".asm"
			}
		}
		runCompiler(command, optOutput, optTarget, optOptimize)
		return
	}

	if command == "compile" {
		inputFile := ""
		if len(args) > 1 && strings.HasSuffix(args[1], ".src") {
			inputFile = args[1]
		}
		if inputFile == "" {
			utils.PrintError("", 0, 0, "E000", "укажите исходный файл (.src)", "", "main")
			fmt.Fprintln(os.Stderr, "Пример: compiler compile program.src -o output.asm")
			os.Exit(1)
		}
		if optOutput == "" {
			if optCompileOnly {
				optOutput = strings.TrimSuffix(inputFile, ".src") + ".o"
			} else {
				optOutput = strings.TrimSuffix(inputFile, ".src") + ".asm"
			}
		}
		runCompiler(inputFile, optOutput, optTarget, optOptimize)
		return
	}

	switch command {
	case "lex":
		runLexerCommand(args[1:])
	case "parse":
		runParserCommand(args[1:])
	case "check":
		runCheckCommand(args[1:])
	case "symbols":
		runSymbolsCommand(args[1:])
	case "ir":
		runIRCommand(args[1:])
	case "test":
		runTestCommand(args[1:])
	default:
		utils.PrintError("", 0, 0, "E000",
			fmt.Sprintf("неизвестная команда: %s", command), "", "main")
		fmt.Fprintln(os.Stderr, "Используйте --help для справки")
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Printf("%s %s\n", utils.BoldText("mikrocompiler"), utils.CyanText(VERSION))
	fmt.Printf("Автор: %s\n", AUTHOR)
	fmt.Printf("Сборка: %s\n", BUILD)
	fmt.Printf("Цель: %s\n", TARGET)
	fmt.Println()
	fmt.Println("Мини-компилятор языка C-like в x86-64 ассемблер")
	fmt.Println("Поддерживает: переменные, функции, массивы, структуры, указатели")
	fmt.Println("System V AMD64 ABI, NASM синтаксис")
}

func printHelp() {
	fmt.Printf("%s %s - %s\n\n",
		utils.BoldText("mikrocompiler"),
		utils.CyanText(VERSION),
		"Компилятор MiniLang",
	)

	fmt.Println(utils.BoldText("ИСПОЛЬЗОВАНИЕ:"))
	fmt.Println("  compiler [опции] <файл.src>              # Компиляция программы")
	fmt.Println("  compiler [опции] <команда> [аргументы]     # Специальные операции")
	fmt.Println()

	fmt.Println(utils.BoldText("ОПЦИИ:"))
	fmt.Printf("  %-26s %s\n", utils.YellowText("-o, --output <файл>"), "Выходной файл")
	fmt.Printf("  %-26s %s\n", utils.YellowText("-S"), "Только ассемблер (без линковки)")
	fmt.Printf("  %-26s %s\n", utils.YellowText("-c"), "Компиляция в объектный файл")
	fmt.Printf("  %-26s %s\n", utils.YellowText("-E"), "Препроцессор (вывод токенов)")
	fmt.Printf("  %-26s %s\n", utils.YellowText("-v, --verbose"), "Подробный вывод")
	fmt.Printf("  %-26s %s\n", utils.YellowText("-O, --optimize"), "Включить оптимизации")
	fmt.Printf("  %-26s %s\n", utils.YellowText("--ast"), "Вывести AST")
	fmt.Printf("  %-26s %s\n", utils.YellowText("--ir"), "Вывести IR")
	fmt.Printf("  %-26s %s\n", utils.YellowText("--target <arch>"), "Целевая архитектура (x86_64)")
	fmt.Printf("  %-26s %s\n", utils.YellowText("--format <fmt>"), "Формат: text, dot, json")
	fmt.Printf("  %-26s %s\n", utils.YellowText("--stats"), "Показать статистику")
	fmt.Printf("  %-26s %s\n", utils.YellowText("--show-types"), "Показывать типы выражений")
	fmt.Printf("  %-26s %s\n", utils.YellowText("--color <mode>"), "Цвета: auto, always, never")
	fmt.Printf("  %-26s %s\n", utils.YellowText("--max-errors <n>"), "Максимум ошибок (по умолчанию: 20)")
	fmt.Printf("  %-26s %s\n", utils.YellowText("--error-format <fmt>"), "Формат ошибок: human, json, gcc")
	fmt.Printf("  %-26s %s\n", utils.YellowText("-Wall"), "Включить все предупреждения")
	fmt.Printf("  %-26s %s\n", utils.YellowText("-Werror"), "Считать предупреждения ошибками")
	fmt.Printf("  %-26s %s\n", utils.YellowText("-Wno=<warn>"), "Отключить предупреждение")
	fmt.Printf("  %-26s %s\n", utils.YellowText("--version"), "Показать версию")
	fmt.Printf("  %-26s %s\n", utils.YellowText("-h, --help"), "Показать справку")
	fmt.Println()

	fmt.Println(utils.BoldText("КОМАНДЫ:"))
	fmt.Printf("  %-22s %s\n", utils.GreenText("lex     <файл>"), "Лексический анализ (токенизация)")
	fmt.Printf("  %-22s %s\n", utils.GreenText("parse   <файл>"), "Синтаксический анализ (AST)")
	fmt.Printf("  %-22s %s\n", utils.GreenText("check   <файл>"), "Семантический анализ (типы)")
	fmt.Printf("  %-22s %s\n", utils.GreenText("symbols <файл>"), "Таблица символов")
	fmt.Printf("  %-22s %s\n", utils.GreenText("ir      <файл>"), "Промежуточное представление")
	fmt.Printf("  %-22s %s\n", utils.GreenText("compile <файл>"), "Компиляция в ассемблер")
	fmt.Printf("  %-22s %s\n", utils.GreenText("test    <тип>"), "Запуск тестов")
	fmt.Println()

	fmt.Println(utils.BoldText("ПРИМЕРЫ:"))
	fmt.Printf("  %s\n", utils.GrayText("compiler test.src -o program"))
	fmt.Printf("  %s\n", utils.GrayText("compiler -S test.src -o test.asm"))
	fmt.Printf("  %s\n", utils.GrayText("compiler -c test.src -o test.o"))
	fmt.Printf("  %s\n", utils.GrayText("compiler -E test.src"))
	fmt.Printf("  %s\n", utils.GrayText("compiler test.src -O -v"))
	fmt.Printf("  %s\n", utils.GrayText("compiler --ast test.src"))
	fmt.Printf("  %s\n", utils.GrayText("compiler --ir test.src"))
	fmt.Printf("  %s\n", utils.GrayText("compiler -Wall -Werror test.src"))
	fmt.Printf("  %s\n", utils.GrayText("compiler --error-format=json test.src"))
	fmt.Println()

	fmt.Println(utils.BoldText("СБОРКА ИСПОЛНЯЕМОГО ФАЙЛА:"))
	fmt.Printf("  %s\n", utils.GrayText("nasm -f elf64 -o runtime.o src/runtime/runtime.asm"))
	fmt.Printf("  %s\n", utils.GrayText("nasm -f elf64 -o program.o program.asm"))
	fmt.Printf("  %s\n", utils.GrayText("gcc -no-pie -o program runtime.o program.o"))
	fmt.Println()

	fmt.Println(utils.BoldText("ЯЗЫК MINILANG:"))
	fmt.Println("  • Статическая типизация (int, float, bool, string)")
	fmt.Println("  • Функции с параметрами и возвращаемыми значениями")
	fmt.Println("  • Массивы с динамическим выделением (malloc/free)")
	fmt.Println("  • Структуры (struct)")
	fmt.Println("  • Внешние функции C (extern)")
	fmt.Println("  • Управляющие конструкции: if/else, while, for")
}

func runLexerCommand(args []string) {
	inputFile := findInputFile(args)
	if inputFile == "" {
		return
	}

	source, tokens := readSourceFile(inputFile)
	checkLexErrors(tokens)

	if optVerbose {
		fmt.Printf("Лексический анализ: %s\n", utils.CyanText(inputFile))
		fmt.Printf("Найдено токенов: %d\n\n", len(tokens))
	}

	var output strings.Builder
	for _, tok := range tokens {
		if optVerbose {
			output.WriteString(tok.String() + "\n")
		} else {
			output.WriteString(fmt.Sprintf("%d:%d %s \"%s\"\n",
				tok.Line, tok.Column, tok.Type, tok.Lexeme))
		}
	}

	_ = source
	writeOutput(output.String(), optOutput)
}

func runParserCommand(args []string) {
	inputFile := findInputFile(args)
	if inputFile == "" {
		return
	}

	source, tokens := readSourceFile(inputFile)
	checkLexErrors(tokens)
	program := parseSource(tokens, source)

	if optVerbose {
		fmt.Printf("Синтаксический анализ: %s\n", utils.CyanText(inputFile))
		fmt.Printf("Объявлений: %d\n", len(program.Declarations))
	}

	var output string
	switch optFormat {
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
		utils.PrintError("", 0, 0, "E000",
			fmt.Sprintf("неизвестный формат: %s", optFormat), "", "parser")
		fmt.Fprintln(os.Stderr, "Доступные форматы: text, dot, json")
		os.Exit(1)
	}

	writeOutput(output, optOutput)
}

func runCheckCommand(args []string) {
	inputFile := findInputFile(args)
	if inputFile == "" {
		return
	}

	source, tokens := readSourceFile(inputFile)
	checkLexErrors(tokens)
	program := parseSource(tokens, source)

	analyzer := semantic.NewSemanticAnalyzer()
	symbolTable, errors, decoratedAST := analyzer.Analyze(program)

	var output strings.Builder

	if optVerbose {
		output.WriteString(fmt.Sprintf("Семантический анализ: %s\n", utils.CyanText(inputFile)))
		output.WriteString(fmt.Sprintf(strings.Repeat("─", 60) + "\n\n"))
	}

	output.WriteString(symbolTable.String())

	if optShowTypes && decoratedAST != nil {
		output.WriteString("\n=== Декорированный AST (с типами) ===\n")
		printer := ast.NewPrettyPrinter()
		output.WriteString(printer.Print(decoratedAST))
	}

	if len(errors.Errors()) > 0 {
		output.WriteString("\n" + utils.RedText("=== Ошибки ===") + "\n")
		output.WriteString(errors.String())
	}

	writeOutput(output.String(), optOutput)

	if len(errors.Errors()) > 0 {
		os.Exit(1)
	}
}

func runSymbolsCommand(args []string) {
	inputFile := findInputFile(args)
	if inputFile == "" {
		return
	}

	_, tokens := readSourceFile(inputFile)
	checkLexErrors(tokens)
	program := parseSource(tokens, "")

	analyzer := semantic.NewSemanticAnalyzer()
	symbolTable, _, _ := analyzer.Analyze(program)

	if optFormat == "json" {
		output := convertSymbolTableToJSON(symbolTable)
		fmt.Print(output)
	} else {
		fmt.Print(symbolTable.String())
	}
}

func convertSymbolTableToJSON(st *semantic.SymbolTable) string {
	var sb strings.Builder
	sb.WriteString("{\n")
	sb.WriteString("  \"scopes\": [\n")

	globalScope := st.GetGlobalScope()
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
		sb.WriteString(fmt.Sprintf("    {\"name\": \"%s\", \"kind\": \"%s\", \"type\": \"%s\", \"line\": %d, \"extern\": %s}%s\n",
			sym.Name, sym.Kind, sym.Type.String(), sym.Line, externFlag, comma))
	}

	sb.WriteString("  ]\n")
	sb.WriteString("}\n")
	return sb.String()
}

func runIRCommand(args []string) {
	inputFile := findInputFile(args)
	if inputFile == "" {
		return
	}

	_, tokens := readSourceFile(inputFile)
	checkLexErrors(tokens)
	program := parseSource(tokens, "")

	analyzer := semantic.NewSemanticAnalyzer()
	symbolTable, errors, _ := analyzer.Analyze(program)

	if len(errors.Errors()) > 0 {
		fmt.Fprintln(os.Stderr, utils.RedText("Семантические ошибки:"))
		fmt.Fprint(os.Stderr, errors.String())
		os.Exit(1)
	}

	typeSystem := semantic.NewTypeSystem()
	irGenerator := ir.NewIRGenerator(symbolTable, typeSystem)
	irProgram := irGenerator.Generate(program)

	if optOptimize {
		optimizer := ir.NewPeepholeOptimizer(irProgram)
		optimizer.Optimize()
		if optVerbose {
			fmt.Fprintf(os.Stderr, "%s\n", utils.GreenText("Оптимизации применены"))
			fmt.Fprintf(os.Stderr, "%s\n", optimizer.GetOptimizationReport())
		}
	}

	var output strings.Builder

	if optStats {
		stats := ir.CollectStats(irProgram)
		output.WriteString(stats.String())
		output.WriteString("\n")
	}

	switch optFormat {
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
		utils.PrintError("", 0, 0, "E000",
			fmt.Sprintf("неизвестный формат: %s", optFormat), "", "ir")
		os.Exit(1)
	}

	writeOutput(output.String(), optOutput)
}

func runTestCommand(args []string) {
	testType := "all"
	if len(args) > 0 {
		testType = args[0]
	}

	fmt.Printf("=== %s: %s ===\n\n", utils.BoldText("Запуск тестов"), utils.CyanText(testType))

	switch testType {
	case "all", "lexer", "parser", "semantic", "ir", "codegen", "control-flow", "final":
		fmt.Println("Для запуска тестов используйте: make test")
		fmt.Println("Или запустите скрипт: tests/test_runner/run_tests.sh")
		fmt.Println()
		fmt.Println("Примеры:")
		fmt.Printf("  %s\n", utils.GrayText("make test              # все тесты"))
		fmt.Printf("  %s\n", utils.GrayText("make test-lexer        # тесты лексера"))
		fmt.Printf("  %s\n", utils.GrayText("make test-final        # финальные тесты"))
	default:
		fmt.Printf("%s: %s\n", utils.RedText("Неизвестный тип тестов"), testType)
		fmt.Println("Доступные типы: all, lexer, parser, semantic, ir, codegen, control-flow, final")
	}
}

func runMultiFileCompiler(files []string, outputFile string) {
	if optVerbose {
		fmt.Fprintf(os.Stderr, "%s %d файлов...\n", utils.BoldText("Мульти-файловая компиляция"), len(files))
		fmt.Fprintf(os.Stderr, "%s\n", strings.Repeat("─", 50))
	}

	var allObjFiles []string

	for i, file := range files {
		asmFile := strings.TrimSuffix(file, ".src") + ".asm"
		objFile := strings.TrimSuffix(file, ".src") + ".o"

		if optVerbose {
			fmt.Fprintf(os.Stderr, "  [%d/%d] %s\n", i+1, len(files), file)
		}

		runCompilerQuiet(file, asmFile, objFile)
		allObjFiles = append(allObjFiles, objFile)
	}

	if optVerbose {
		fmt.Fprintf(os.Stderr, "%s\n", strings.Repeat("─", 50))
	}

	outputName := "program"
	if outputFile != "" {
		outputName = strings.TrimSuffix(outputFile, ".o")
	}

	fmt.Printf("\n%s\n", utils.GreenText("Компиляция завершена. Для сборки выполните:"))
	fmt.Printf("  %s\n", utils.GrayText("nasm -f elf64 -o runtime.o src/runtime/runtime.asm"))

	objList := strings.Join(allObjFiles, " ")
	fmt.Printf("  %s\n", utils.GrayText(fmt.Sprintf("gcc -no-pie -o %s runtime.o %s", outputName, objList)))
}

func runCompilerQuiet(inputFile, asmFile, objFile string) {
	source, tokens := readSourceFile(inputFile)
	checkLexErrors(tokens)
	program := parseSource(tokens, source)

	analyzer := semantic.NewSemanticAnalyzer()
	symbolTable, errors, decoratedAST := analyzer.Analyze(program)
	if len(errors.Errors()) > 0 {
		fmt.Fprintln(os.Stderr, utils.RedText("Семантические ошибки:"))
		fmt.Fprint(os.Stderr, errors.String())
		os.Exit(1)
	}

	typeSystem := semantic.NewTypeSystem()
	irGenerator := ir.NewIRGenerator(symbolTable, typeSystem)
	irProgram := irGenerator.Generate(decoratedAST)

	if optOptimize {
		optimizer := ir.NewPeepholeOptimizer(irProgram)
		optimizer.Optimize()
	}

	codeGenerator := codegen.NewX86Generator(irProgram, symbolTable, typeSystem)
	asmCode := codeGenerator.Generate()

	if codeGenerator.HasErrors() {
		fmt.Fprintln(os.Stderr, utils.RedText("Ошибки кодогенерации:"))
		for _, err := range codeGenerator.GetErrors() {
			fmt.Fprintf(os.Stderr, "  %s\n", err)
		}
		os.Exit(1)
	}

	os.WriteFile(asmFile, []byte(asmCode), 0644)

	cmd := exec.Command("nasm", "-f", "elf64", "-o", objFile, asmFile)
	cmd.Run()
	os.Remove(asmFile)
}

func runCompiler(inputFile, outputFile, target string, optimize bool) {
	if inputFile == "" {
		utils.PrintError("", 0, 0, "E000", "необходимо указать входной файл", "", "main")
		os.Exit(1)
	}

	if target != "x86_64" {
		utils.PrintError("", 0, 0, "E000",
			fmt.Sprintf("неподдерживаемая архитектура '%s'. Поддерживается только x86_64", target), "", "compiler")
		os.Exit(1)
	}

	// Warning system
	warnings := utils.NewWarningCollector()
	if optWall {
		warnings.EnableAll()
	}
	if optWerror {
		warnings.SetAsErrors(true)
	}
	if optWno != "" {
		warnings.Disable(utils.WarningCode(optWno))
	}

	if optVerbose {
		fmt.Fprintf(os.Stderr, "%s %s...\n", utils.BoldText("Компиляция"), utils.CyanText(inputFile))
		fmt.Fprintf(os.Stderr, "  Выходной файл: %s\n", outputFile)
		fmt.Fprintf(os.Stderr, "  Целевая архитектура: %s\n", target)
		if optimize {
			fmt.Fprintf(os.Stderr, "  Оптимизации: %s\n", utils.GreenText("включены"))
		}
		if optCompileOnly {
			fmt.Fprintf(os.Stderr, "  Режим: %s\n", utils.YellowText("объектный файл (-c)"))
		}
		if optAssemblyOnly {
			fmt.Fprintf(os.Stderr, "  Режим: %s\n", utils.YellowText("только ассемблер (-S)"))
		}
		if optPreprocess {
			fmt.Fprintf(os.Stderr, "  Режим: %s\n", utils.YellowText("препроцессор (-E)"))
		}
		fmt.Fprintf(os.Stderr, "%s\n", strings.Repeat("─", 50))
	}

	if optPreprocess {
		_, tokens := readSourceFile(inputFile)
		for _, tok := range tokens {
			fmt.Println(tok.String())
		}
		return
	}

	source, tokens := readSourceFile(inputFile)
	checkLexErrors(tokens)

	if optVerbose {
		fmt.Fprintf(os.Stderr, "  [1/5] %s (%d токенов)\n", utils.GreenText("Лексический анализ: OK"), len(tokens))
	}

	program := parseSource(tokens, source)

	if optVerbose {
		fmt.Fprintf(os.Stderr, "  [2/5] %s (%d объявлений)\n", utils.GreenText("Синтаксический анализ: OK"), len(program.Declarations))
	}

	analyzer := semantic.NewSemanticAnalyzer()
	symbolTable, errors, decoratedAST := analyzer.Analyze(program)

	if len(errors.Errors()) > 0 {
		if optErrorFormat == "json" {
			for _, err := range errors.Errors() {
				utils.AddJSONError(string(err.Code), err.Message, inputFile, err.Context, err.Line, err.Column)
			}
			utils.PrintJSONReport()
		} else {
			fmt.Fprintln(os.Stderr, "\n"+utils.RedText("Семантические ошибки:"))
			fmt.Fprint(os.Stderr, errors.String())
		}
		os.Exit(1)
	}

	if optVerbose {
		fmt.Fprintf(os.Stderr, "  [3/5] %s\n", utils.GreenText("Семантический анализ: OK"))
	}

	if optShowAST {
		fmt.Println(utils.BoldText("\n=== AST ==="))
		printer := ast.NewPrettyPrinter()
		fmt.Println(printer.Print(decoratedAST))
	}

	typeSystem := semantic.NewTypeSystem()
	irGenerator := ir.NewIRGenerator(symbolTable, typeSystem)
	irProgram := irGenerator.Generate(decoratedAST)

	if optimize {
		optimizer := ir.NewPeepholeOptimizer(irProgram)
		optimizer.Optimize()
		if optVerbose {
			fmt.Fprintf(os.Stderr, "  [4/5] %s\n", utils.GreenText("Оптимизация IR: выполнена"))
			fmt.Fprintf(os.Stderr, "%s\n", optimizer.GetOptimizationReport())
		}
	}

	if optVerbose && !optimize {
		fmt.Fprintf(os.Stderr, "  [4/5] %s\n", utils.GreenText("Генерация IR: OK"))
	}

	if optShowIR {
		fmt.Println(utils.BoldText("\n=== IR ==="))
		irPrinter := ir.NewTextPrinter()
		fmt.Println(irPrinter.Print(irProgram))
	}

	if optShowAST || optShowIR {
		return
	}

	codeGenerator := codegen.NewX86Generator(irProgram, symbolTable, typeSystem)
	asmCode := codeGenerator.Generate()

	if codeGenerator.HasErrors() {
		if optErrorFormat == "json" {
			for _, err := range codeGenerator.GetErrors() {
				utils.AddJSONError("E018", err, inputFile, "codegen", 0, 0)
			}
			utils.PrintJSONReport()
		} else {
			fmt.Fprintln(os.Stderr, "\n"+utils.RedText("Ошибки кодогенерации:"))
			for _, err := range codeGenerator.GetErrors() {
				fmt.Fprintf(os.Stderr, "  %s\n", err)
			}
		}
		os.Exit(1)
	}

	if optVerbose {
		fmt.Fprintf(os.Stderr, "  [5/5] %s\n", utils.GreenText("Генерация кода: OK"))
		fmt.Fprintf(os.Stderr, "%s\n", strings.Repeat("─", 50))
	}

	if optCompileOnly {
		asmFile := outputFile
		if strings.HasSuffix(asmFile, ".o") {
			asmFile = strings.TrimSuffix(asmFile, ".o") + ".asm"
		} else if !strings.HasSuffix(asmFile, ".asm") {
			asmFile = asmFile + ".asm"
		}

		err := os.WriteFile(asmFile, []byte(asmCode), 0644)
		if err != nil {
			utils.PrintError(inputFile, 0, 0, "E000",
				fmt.Sprintf("ошибка записи файла '%s': %v", asmFile, err), "", "compiler")
			os.Exit(1)
		}

		objFile := outputFile
		if !strings.HasSuffix(objFile, ".o") {
			objFile = strings.TrimSuffix(objFile, ".asm") + ".o"
		}

		if optVerbose {
			fmt.Fprintf(os.Stderr, "  Ассемблирование %s -> %s...\n", asmFile, objFile)
		}

		cmd := exec.Command("nasm", "-f", "elf64", "-o", objFile, asmFile)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			utils.PrintError(inputFile, 0, 0, "E000",
				fmt.Sprintf("ошибка ассемблирования: %v", err), "", "compiler")
			os.Remove(asmFile)
			os.Exit(1)
		}

		os.Remove(asmFile)

		if optVerbose {
			fmt.Fprintf(os.Stderr, "%s: %s\n", utils.GreenText("Объектный файл создан"), objFile)
		} else {
			fmt.Printf("%s %s\n", utils.GreenText("Объектный файл:"), objFile)
		}
		return
	}

	err := os.WriteFile(outputFile, []byte(asmCode), 0644)
	if err != nil {
		utils.PrintError(inputFile, 0, 0, "E000",
			fmt.Sprintf("ошибка записи файла '%s': %v", outputFile, err), "", "compiler")
		os.Exit(1)
	}

	// Вывод предупреждений
	warnings.PrintWarnings()
	if optWerror && warnings.HasWarnings() {
		if optErrorFormat == "json" {
			utils.PrintJSONReport()
		}
		fmt.Fprintf(os.Stderr, "\n%s: компиляция остановлена из-за -Werror\n", utils.RedText("Ошибка"))
		os.Exit(1)
	}

	// JSON отчёт (успешная компиляция)
	if optErrorFormat == "json" {
		utils.PrintJSONReport()
		return
	}

	if optVerbose {
		fmt.Fprintf(os.Stderr, "\n%s: %s\n", utils.GreenText("Файл сохранён"), outputFile)
		fmt.Fprintf(os.Stderr, "Размер: %d байт\n", len(asmCode))
	} else {
		if optAssemblyOnly {
			fmt.Printf("%s %s\n", utils.GreenText("Ассемблерный код:"), outputFile)
		} else {
			fmt.Printf("%s %s\n", utils.GreenText("Ассемблерный код записан в"), outputFile)
			fmt.Println()
			fmt.Println("Для сборки исполняемого файла выполните:")
			fmt.Printf("  %s\n", utils.GrayText("nasm -f elf64 -o runtime.o src/runtime/runtime.asm"))
			fmt.Printf("  %s\n", utils.GrayText(fmt.Sprintf("nasm -f elf64 -o program.o %s", outputFile)))
			fmt.Printf("  %s\n", utils.GrayText("gcc -no-pie -o program runtime.o program.o"))
			fmt.Println()
			fmt.Printf("%s\n", utils.GrayText("Примечание: используется gcc для линковки с libc (malloc, free, printf, abort)"))
		}
	}
}

func findInputFile(args []string) string {
	for _, arg := range args {
		if strings.HasSuffix(arg, ".src") {
			return arg
		}
	}
	utils.PrintError("", 0, 0, "E000", "укажите входной файл (.src)", "", "compiler")
	return ""
}

func writeOutput(content, outputFile string) {
	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(content), 0644)
		if err != nil {
			utils.PrintError("", 0, 0, "E000",
				fmt.Sprintf("ошибка записи файла '%s': %v", outputFile, err), "", "compiler")
			os.Exit(1)
		}
		if optVerbose {
			fmt.Fprintf(os.Stderr, "%s %s\n", utils.GreenText("Результат записан в"), outputFile)
		}
	} else {
		fmt.Print(content)
	}
}

func readSourceFile(inputFile string) (string, []mytoken.Token) {
	if inputFile == "" {
		utils.PrintError("", 0, 0, "E000", "необходимо указать входной файл", "", "compiler")
		os.Exit(1)
	}

	content, err := os.ReadFile(inputFile)
	if err != nil {
		utils.PrintError(inputFile, 0, 0, "E000",
			fmt.Sprintf("ошибка чтения файла: %v", err), "", "compiler")
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
			if optErrorFormat == "json" {
				utils.AddJSONError("E001", tok.Lexeme, "", "lexer", tok.Line, tok.Column)
			} else {
				utils.PrintError("", tok.Line, tok.Column, "E001", tok.Lexeme, "", "lexer")
			}
			hasErrors = true
		}
	}
	if hasErrors {
		if optErrorFormat == "json" {
			utils.PrintJSONReport()
		}
		os.Exit(1)
	}
	return false
}

func parseSource(tokens []mytoken.Token, source string) *ast.ProgramNode {
	p := parser.NewParser(tokens, source)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		if optErrorFormat == "json" {
			for _, err := range p.Errors() {
				utils.AddJSONError("E002", err, "", "parser", 0, 0)
			}
			utils.PrintJSONReport()
		} else {
			fmt.Fprintln(os.Stderr, "\n"+utils.RedText("Ошибки синтаксического анализа:"))
			fmt.Fprintln(os.Stderr, strings.Repeat("─", 60))
			for i, err := range p.Errors() {
				if i >= optMaxErrors {
					fmt.Fprintf(os.Stderr, "... и ещё %d ошибок\n", len(p.Errors())-optMaxErrors)
					break
				}
				fmt.Fprintf(os.Stderr, "%s\n", err)
			}
		}
		os.Exit(1)
	}

	if program == nil {
		utils.PrintError("", 0, 0, "E000", "не удалось распарсить программу", "", "parser")
		os.Exit(1)
	}

	return program
}
