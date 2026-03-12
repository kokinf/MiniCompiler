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

# Сборка для Linux и macOS
make build

# Сборка для Windows
go build -o bin/compiler.exe ./cmd/compiler
```

## Использование

### 1. Лексический анализ (лексер)

```bash
# Linux / macOS
./bin/compiler lex --input examples/hello.src
./bin/compiler lex --input examples/hello.src --output tokens.txt

# Windows (PowerShell)
.\bin\compiler.exe lex --input examples\hello.src
```

**Пример вывода токенов:**
```
1:1 KW_FN "fn"
1:4 IDENTIFIER "main"
1:8 LPAREN "("
1:9 RPAREN ")"
1:10 LBRACE "{"
2:5 KW_INT "int"
2:9 IDENTIFIER "counter"
2:16 ASSIGN "="
2:18 INT_LITERAL "42" 42
2:20 SEMICOLON ";"
3:5 KW_RETURN "return"
3:11 SEMICOLON ";"
4:1 RBRACE "}"
5:1 END_OF_FILE ""
```

### 2. Синтаксический анализ (парсер)

```bash
# Базовый вывод AST в текстовом формате
./bin/compiler parse --input examples/factorial.src

# Сохранение AST в файл
./bin/compiler parse --input examples/factorial.src --output ast.txt

# Генерация DOT графа для визуализации
./bin/compiler parse --input examples/factorial.src --format dot --output ast.dot

# Генерация JSON для машинной обработки
./bin/compiler parse --input examples/struct.src --format json --output ast.json
```

**Пример вывода AST (текстовый формат):**
```
Program:
  FunctionDecl: factorial -> int [line 1]:
    Parameters:
      int n
    Body:
      IfStmt [line 2]:
        Condition:
          (n <= 1)
        Then:
          Block [line 2-3]:
            Return: 1
        Else:
          Block [line 4-5]:
            Return: (n * factorial((n - 1)))
  FunctionDecl: main -> void [line 9]:
    Parameters:
    Body:
      VarDecl: int result = factorial(5)
      Return
```

## Тесты

### Запуск всех тестов

```bash
make test
```

## Валидные тесты лексера

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

## Невалидные тесты лексера

Расположены в: `tests/lexer/invalid/`

| Тест | Файлы | Описание |
|------|-------|----------|
| **test_invalid_char** | `test_invalid_char.src`<br>`test_invalid_char.expected` | Невалидные символы: `@`, `$`, `#`, `` ` `` |
| **test_long_identifier** | `test_long_identifier.src`<br>`test_long_identifier.expected` | Идентификатор длиннее 255 символов |
| **test_malformed_number** | `test_malformed_number.src`<br>`test_malformed_number.expected` | Некорректные числа: несколько точек, точка в начале, переполнение диапазона |
| **test_unterminated_comment** | `test_unterminated_comment.src`<br>`test_unterminated_comment.expected` | Незакрытый многострочный комментарий `/* ...` |
| **test_unterminated_string** | `test_unterminated_string.src`<br>`test_unterminated_string.expected` | Незакрытые строковые литералы `"...` |

## Валидные тесты парсера

### Выражения (`tests/parser/valid/expressions/`)

| Тест | Описание |
|------|----------|
| **arithmetic** | Арифметические операции и приоритеты: `+`, `-`, `*`, `/`, `%`, `(`, `)` |
| **comparisons** | Операторы сравнения: `<`, `<=`, `>`, `>=`, `==`, `!=` |
| **logical** | Логические операторы: `&&`, `||`, `!` |
| **assignment** | Присваивание и составные операторы: `=`, `+=`, `-=`, `*=`, `/=` |
| **calls** | Вызовы функций с аргументами |

### Инструкции (`tests/parser/valid/statements/`)

| Тест | Описание |
|------|----------|
| **if_else** | Условные конструкции с вложенными `if-else` |
| **loops** | Циклы `while`, `for` (все варианты) |
| **returns** | Инструкции `return` с/без значения |
| **empty** | Пустые блоки, пустые инструкции |

### Объявления (`tests/parser/valid/declarations/`)

| Тест | Описание |
|------|----------|
| **variables** | Объявления переменных с/без инициализации |
| **functions** | Объявления функций с параметрами и без |
| **structs_complex** | Структуры с полями разных типов |

### Полные программы (`tests/parser/valid/full_programs/`)

| Тест | Описание |
|------|----------|
| **factorial** | Рекурсивное вычисление факториала |
| **fibonacci** | Рекурсивное вычисление чисел Фибоначчи |
| **struct_example** | Работа со структурами и функциями |

## Невалидные тесты парсера

### Синтаксические ошибки (`tests/parser/invalid/syntax_errors/`)

| Тест | Описание | Ожидаемая ошибка |
|------|----------|------------------|
| **missing_semicolon** | Пропущена точка с запятой | `ожидалась ';', получен KW_RETURN` |
| **missing_paren** | Пропущена закрывающая скобка | `ожидался токен RPAREN, получен LBRACE` |
| **missing_brace** | Пропущена закрывающая фигурная скобка | `ожидался токен RBRACE, получен END_OF_FILE` |
| **missing_ident** | Пропущено имя переменной | `ожидалось имя переменной, получен ASSIGN` |
| **invalid_assignment** | Присваивание литералу | `левая часть присваивания должна быть идентификатором` |
| **missing_type** | Объявление переменной без типа | `объявление переменной должно содержать тип` |
| **unclosed_string** | Незакрытый строковый литерал | `Ошибка в L:C: незакрытый строковый литерал` |

## Примеры программ

### `examples/hello.src` - Базовые конструкции
```c
fn main() -> void {
    // This is a comment
    int x = 42;
    float y = 3.14;
    bool flag = true;
    string s = "Hello, World!";
    
    if (x > 0) {
        return;
    } else {
        x = x + 1;
    }
    
    while (x < 10) {
        x = x + 1;
    }
    
    return;
}
```

### `examples/factorial.src` - Рекурсия
```c
fn factorial(n int) -> int {
    if (n <= 1) {
        return 1;
    } else {
        return n * factorial(n - 1);
    }
}

fn main() -> void {
    int result = factorial(5);
    return;
}
```

### `examples/struct.src` - Структуры
```c
struct Point {
    int x;
    int y;
}

struct Rectangle {
    Point topLeft;
    Point bottomRight;
}

fn area(rect Rectangle) -> int {
    int width = rect.bottomRight.x - rect.topLeft.x;
    int height = rect.bottomRight.y - rect.topLeft.y;
    return width * height;
}
```

## Спецификация языка

Полная спецификация языка доступна в файлах:
- `docs/language_spec.md` - лексическая спецификация
- `docs/grammar.md` - грамматика и синтаксис

### Поддерживаемые возможности

| Категория | Токены/Конструкции |
|-----------|-------------------|
| **Ключевые слова** | `fn`, `int`, `float`, `bool`, `string`, `if`, `else`, `while`, `for`, `return`, `true`, `false`, `void`, `struct` |
| **Операторы** | `+`, `-`, `*`, `/`, `%`, `=`, `+=`, `-=`, `*=`, `/=`, `==`, `!=`, `<`, `<=`, `>`, `>=`, `&&`, `||`, `!` |
| **Разделители** | `()`, `{}`, `[]`, `;`, `,`, `.`, `->` |
| **Типы данных** | `int`, `float`, `bool`, `string`, `void`, пользовательские структуры |
| **Литералы** | Целые числа (32-bit), числа с плавающей точкой, строки, булевы значения |
| **Комментарии** | Однострочные (`//`), многострочные (`/* */`) с поддержкой вложенности |

## Структура проекта

```
mikrocompiler/
├── cmd/
│   └── compiler/
│       └── main.go
├── internal/
│   ├── ast/
│   │   ├── ast.go                  # Определения узлов AST
│   │   ├── printer.go               # Pretty printer для AST
│   │   ├── dot_printer.go           # Генератор DOT графов
│   │   └── json_printer.go          # Генератор JSON
│   ├── lexer/
│   │   └── scanner.go               # Лексический анализатор
│   ├── parser/
│   │   └── parser.go                # Синтаксический анализатор
│   └── token/
│       └── token.go                 # Определения токенов
├── tests/
│   ├── lexer/
│   │   ├── valid/                    # Валидные тесты лексера
│   │   └── invalid/                  # Тесты с ошибками лексера
│   ├── parser/
│   │   ├── valid/                     # Валидные тесты парсера
│   │   │   ├── expressions/
│   │   │   ├── statements/
│   │   │   ├── declarations/
│   │   │   └── full_programs/
│   │   └── invalid/                   # Тесты с ошибками парсера
│   │       ├── syntax_errors/
│   └── test_runner/
│       └── run_tests.sh               # Скрипт запуска тестов
├── examples/
│   ├── hello.src                      # Пример с базовыми конструкциями
│   ├── factorial.src                   # Пример с рекурсией
│   └── struct.src                      # Пример со структурами
├── docs/
│   ├── language_spec.md                # Лексическая спецификация
│   └── grammar.md                      # Грамматика языка
├── Makefile                             # Автоматизация сборки
├── go.mod                               # Определение модуля Go
└── README.md                            # Этот файл
```

## Makefile

| Команда | Описание |
|---------|----------|
| `make build` | Сборка компилятора |
| `make run` | Запуск парсера на `examples/factorial.src` |
| `make run-lex` | Запуск лексера на `examples/hello.src` |
| `make test` | Запуск всех тестов |
| `make clean` | Очистка артефактов сборки |
| `make fmt` | Форматирование кода |
| `make help` | Показать все доступные команды |