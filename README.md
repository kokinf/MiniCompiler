# MiniCompiler

Простой компилятор для языка программирования, C-подобного, реализованный на Go.


## Инструкции по сборке

### Требования

- Git - [Установить Git](https://git-scm.com/downloads)
- Go 1.25.1+ - [Установить Golang](https://go.dev/doc/install)

### Сборка проекта

```bash
# Клонирование репозитория
git clone https://github.com/kokinf/MiniCompiler
cd MiniCompiler

# Сборка для linux и macOS
make build

# Сборка для Windows
go build -o bin/compiler.exe ./cmd/compiler
```

## Использование

### 1. Запустите лексер

```bash
# Linux / macOS
./bin/compiler lex --input examples/hello.src

# Windows (PowerShell)
.\bin\compiler.exe lex --input examples\hello.src
```

### 2. Просмотрите вывод токенов

```
1:1 KW_FN "fn"
1:4 IDENTIFIER "main"
1:8 LPAREN "("
1:9 RPAREN ")"
1:11 LBRACE "{"
2:5 KW_INT "int"
2:9 IDENTIFIER "counter"
2:17 ASSIGN "="
2:19 INT_LITERAL "42" 42
2:21 SEMICOLON ";"
...
18:1 END_OF_FILE ""
```

## Тесты

### Запуск Linux / macOS

```bash
make test
```
## Валидные тесты

Расположены в: `tests/lexer/valid/`

| Тест | Файлы | Описание |
|------|-------|----------|
| **test_comments** | `test_comments.src`<br>`test_comments.expected` | Комментарии: однострочные (`//`), многострочные (`/* */`), вложенные, между токенами |
| **test_identifiers** | `test_identifiers.src`<br>`test_identifiers.expected` | Идентификаторы: простые, с цифрами, с подчёркиваниями, camelCase, PascalCase, максимальная длина (255 символов) |
| **test_keywords** | `test_keywords.src`<br>`test_keywords.expected` | Все ключевые слова: `fn`, `int`, `float`, `bool`, `if`, `else`, `while`, `for`, `return`, `true`, `false`, `void`, `struct` |
| **test_mixed** | `test_mixed.src`<br>`test_mixed.expected` | Комплексный тест: функция, переменные разных типов, условия, операторы |
| **test_numbers** | `test_numbers.src`<br>`test_numbers.expected` | Числовые литералы: int, float, отрицательные, граничные значения (2147483647, -2147483648) |
| **test_operators** | `test_operators.src`<br>`test_operators.expected` | Все операторы и разделители: `+`, `-`, `*`, `/`, `%`, `=`, `==`, `!=`, `<`, `<=`, `>`, `>=`, `&&`, `||`, `()`, `{}`, `[]`, `;`, `,`, `.` |
| **test_strings** | `test_strings.src`<br>`test_strings.expected` | Строковые литералы: простые, пустые, с пробелами, со специальными символами |

## Невалидные тесты

Расположены в: `tests/lexer/invalid/`

| Тест | Файлы | Описание |
|------|-------|----------|
| **test_invalid_char** | `test_invalid_char.src`<br>`test_invalid_char.expected` | Невалидные символы: `@`, `$`, `#`, `` ` `` |
| **test_long_identifier** | `test_long_identifier.src`<br>`test_long_identifier.expected` | Идентификатор длиннее 255 символов |
| **test_malformed_number** | `test_malformed_number.src`<br>`test_malformed_number.expected` | Некорректные числа: несколько точек, точка в начале, переполнение диапазона |
| **test_unterminated_comment** | `test_unterminated_comment.src`<br>`test_unterminated_comment.expected` | Незакрытый многострочный комментарий `/* ...` |
| **test_unterminated_string** | `test_unterminated_string.src`<br>`test_unterminated_string.expected` | Незакрытые строковые литералы `"...` |

## Спецификация языка

- Спецификация языка— грамматика EBNF, определения токенов и примеры

### Поддерживаемые возможности

| Категория | Токены |
|-----------|--------|
| **Ключевые слова** | `fn`, `int`, `float`, `bool`, `string`, `if`, `else`, `while`, `for`, `return`, `true`, `false`, `void`, `struct` |
| **Операторы** | `+`, `-`, `*`, `/`, `%`, `=`, `==`, `!=`, `<`, `<=`, `>`, `>=`, `&&`, `||` |
| **Разделители** | `()`, `{}`, `[]`, `;`, `,`, `.` |
| **Литералы** | Целые числа, числа с плавающей точкой, строки, булевы значения |
| **Комментарии** | Однострочные (`//`), многострочные (`/* */`) |

## Структура проекта

```
mikrocompiler/
├── cmd/
│   └── compiler/
│       └── main.go
├── internal/
│   ├── lexer/
│   │   └── scanner.go           # Лексический анализатор
│   └── token/
│       └── token.go             # Определения токенов
├── tests/
│   ├── lexer/
│   │   ├── valid/               # Валидные тесты
│   │   └── invalid/             # Тесты с ошибками
│   └── test_runner/
│       └── run_tests.sh
├── examples/
│   └── hello.src                # Пример исходного файла
├── docs/
│   └── language_spec.md         # Спецификация языка
├── Makefile                     # Автоматизация сборки
├── go.mod                       # Определение модуля Go
└── README.md                    # Этот файл
```