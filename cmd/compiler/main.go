package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"mikrocompiler/internal/lexer"
)

func main() {
	lexCmd := flag.NewFlagSet("lex", flag.ExitOnError)
	inputFile := lexCmd.String("input", "", "Входной исходный файл")
	outputFile := lexCmd.String("output", "", "Выходной файл с токенами (необязательно)")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "lex":
		lexCmd.Parse(os.Args[2:])
		runLexer(*inputFile, *outputFile)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Использование: compiler <команда> [опции]")
	fmt.Println("Команды:")
	fmt.Println("  lex --input <файл> [--output <файл>]  Запуск лексера на входном файле")
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

		if tok.Type == "ILLEGAL" {
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

		if tok.Type == "END_OF_FILE" {
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
