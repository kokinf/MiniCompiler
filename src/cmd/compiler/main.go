package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"mikrocompiler/src/internal/ast"
	"mikrocompiler/src/internal/lexer"
	"mikrocompiler/src/internal/parser"
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
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Использование: compiler <команда> [опции]")
	fmt.Println("Команды:")
	fmt.Println("  lex --input <файл> [--output <файл>]           Запуск лексера на входном файле")
	fmt.Println("  parse --input <файл> [--output <файл>] [--format text|dot|json]  Запуск парсера")
}

func runLexer(inputFile, outputFile string) {
	if inputFile == "" {
		fmt.Fprintln(os.Stderr, "Ошибка: необходимо указать входной файл")
		fmt.Fprintln(os.Stderr, "Использование: compiler lex --input <файл> [--output <файл>]")
		os.Exit(1)
	}

	content, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения входного файла: %v\n", err)
		os.Exit(1)
	}

	scanner := lexer.NewScanner(string(content))

	var output strings.Builder
	hasErrors := false
	firstToken := true

	for {
		tok := scanner.NextToken()

		if tok.Type == mytoken.ILLEGAL {
			if !firstToken {
				output.WriteString("\n")
			}
			output.WriteString(fmt.Sprintf("%d:%d ILLEGAL \"%s\"", tok.Line, tok.Column, tok.Lexeme))
			firstToken = false
			hasErrors = true
		} else {
			if !firstToken {
				output.WriteString("\n")
			}
			output.WriteString(tok.String())
			firstToken = false
		}

		if tok.Type == mytoken.EOF {
			break
		}
	}

	if outputFile != "" {
		err = os.WriteFile(outputFile, []byte(output.String()), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка записи выходного файла: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Токены записаны в %s\n", outputFile)
	} else {
		fmt.Print(output.String())
	}

	if hasErrors {
		os.Exit(1)
	}
}

func runParser(inputFile, outputFile, format string) {
	if inputFile == "" {
		fmt.Fprintln(os.Stderr, "Ошибка: необходимо указать входной файл")
		fmt.Fprintln(os.Stderr, "Использование: compiler parse --input <файл> [--output <файл>] [--format text|dot|json]")
		os.Exit(1)
	}

	content, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения входного файла: %v\n", err)
		os.Exit(1)
	}

	scanner := lexer.NewScanner(string(content))
	var tokens []mytoken.Token // используем наш тип Token

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

	p := parser.NewParser(tokens)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		fmt.Fprintln(os.Stderr, "Ошибки парсинга:")
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", err)
		}
		os.Exit(1)
	}

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
		err = os.WriteFile(outputFile, []byte(output), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка записи выходного файла: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("AST записан в %s\n", outputFile)
	} else {
		fmt.Print(output)
	}
}
