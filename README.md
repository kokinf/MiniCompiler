
# MiniCompiler

Простой компилятор для C-подобного языка программирования, реализованный на Go. Проект включает лексический, синтаксический, семантический анализ и генерацию промежуточного представления (IR)

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
make

# Сборка для Windows
go build -o bin/compiler.exe ./src/cmd/compiler
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
dot -Tpng ast.dot -o ast.png

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

### 3. Семантический анализ (проверка типов и областей видимости)

```bash
# Базовый семантический анализ
./bin/compiler check --input examples/factorial.src

# Подробный вывод с таблицей символов
./bin/compiler check --input examples/factorial.src --verbose

# Вывод с аннотациями типов выражений
./bin/compiler check --input examples/factorial.src --show-types

# Вывод таблицы символов
./bin/compiler symbols --input examples/factorial.src

# Вывод таблицы символов в JSON
./bin/compiler symbols --input examples/factorial.src --format json
```

**Пример вывода таблицы символов:**
```
Symbol Table:
global scope (level 0):
  factorial: function function -> int (line 1)
  main: function function -> int (line 8)
```

**Пример вывода семантических ошибок:**
```
type_mismatch: cannot assign float to int
  --> line 2, column 9
  |
  = in function main
  = cannot assign float to int
```

### 4. Генерация промежуточного представления (IR)

```bash
# Генерация IR в текстовом формате
./bin/compiler ir --input examples/factorial.src

# Генерация оптимизированного IR со статистикой
./bin/compiler ir --input examples/factorial.src --optimize --stats

# Сохранение IR в файл
./bin/compiler ir --input examples/factorial.src --output factorial.ir

# Генерация Control Flow Graph (DOT формат)
./bin/compiler ir --input examples/factorial.src --format dot --output cfg.dot
dot -Tpng cfg.dot -o cfg.png

# Генерация IR в JSON
./bin/compiler ir --input examples/factorial.src --format json --output ir.json

# Вывод статистики IR
./bin/compiler ir --input examples/factorial.src --stats
```

**Пример вывода IR (текстовый формат):**
```
function factorial: int (int n)
  ; Locals:

  entry:
    %t1 = CMP_LE n, 1
    JUMP_IF %t1, if_then_1
    JUMP if_else_2

  if_then_1:
    RETURN 1

  if_else_2:
    %t2 = SUB n, 1
    PARAM 0, %t2
    %t3 = CALL factorial(%t2)
    %t4 = MUL n, %t3
    RETURN %t4
```

**Пример вывода статистики IR:**
```
IR Statistics:
================

Functions:       1
Basic Blocks:    3
Instructions:    7
Temporaries:     4
Phi Nodes:       0
Def-Use Chains:  4

Instruction breakdown:
  CMP_LE     : 1
  JUMP_IF    : 1
  JUMP       : 1
  RETURN     : 2
  SUB        : 1
  PARAM      : 1
  CALL       : 1
  MUL        : 1
```

**Пример оптимизации IR:**
```
# До оптимизации:
  %t1 = ADD x, 0
  %t2 = MUL %t1, 1
  %t3 = CMP_GT %t2, 5
  JUMP_IF %t3, L1
  JUMP L2

# После оптимизации:
  %t1 = CMP_GT x, 5    # x + 0 → x, x * 1 → x
  JUMP_IF %t1, L1
  JUMP L2
```

#### Оптимизации IR

Компилятор поддерживает следующие оптимизации:

| Оптимизация | Описание | Пример |
|-------------|----------|--------|
| **Алгебраические упрощения** | Удаление тривиальных операций | `x + 0 → x`, `x * 1 → x` |
| **Свёртка констант** | Вычисление константных выражений | `3 + 4 → 7` |
| **Strength reduction** | Замена дорогих операций | `x * 2 → x + x` |
| **Dead code elimination** | Удаление неиспользуемого кода | `MOVE x, x → удалить` |
| **Jump chaining** | Упрощение цепочек переходов | `JUMP L1; L1: JUMP L2 → JUMP L2` |

## Тестирование

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

## Валидные тесты семантического анализа

### Совместимость типов (`tests/semantic/valid/`)

| Тест | Описание |
|------|----------|
| **01_simple** | Простая программа с возвратом значения |
| **02_variables** | Объявление и использование переменных |
| **03_arithmetic** | Арифметические операции с int |
| **04_float** | Операции с float и преобразование типов |
| **05_logical** | Логические операции с bool |
| **06_comparisons** | Операции сравнения |
| **07_mixed** | Смешанные операции (int + float) |
| **08_if_else** | Условные конструкции с возвратом |
| **09_recursion** | Рекурсивная функция |
| **11_call** | Вызов функции с аргументами |
| **12_factorial** | Полная программа с факториалом |
| **13_fibonacci** | Полная программа с числами Фибоначчи |

## Невалидные тесты семантического анализа

Расположены в: `tests/semantic/invalid/`

| Тест | Описание | Ожидаемая ошибка |
|------|----------|------------------|
| **01_undeclared_var** | Использование необъявленной переменной | `identifier 'x' not declared` |
| **02_undeclared_expr** | Необъявленная переменная в выражении | `identifier 'unknown' not declared` |
| **03_undeclared_func** | Вызов необъявленной функции | `identifier 'unknown_func' not declared` |
| **04_float_to_int** | Присваивание float в int | `cannot assign float to int` |
| **05_bool_to_int** | Присваивание bool в int | `cannot assign bool to int` |
| **06_string_to_int** | Присваивание string в int | `cannot assign string to int` |
| **07_bool_arithmetic** | bool в арифметической операции | `operator + requires numeric operands` |
| **08_if_condition** | Неbool условие в if | `if condition must be bool` |
| **09_while_condition** | Неbool условие в while | `while condition must be bool` |
| **10_duplicate_var** | Повторное объявление переменной | `variable 'x' already declared` |
| **11_duplicate_func** | Повторное объявление функции | `function 'foo' already declared` |
| **12_arg_count** | Неправильное количество аргументов | `expected 2 arguments, got 1` |
| **13_arg_type** | Неправильный тип аргумента | `argument 2: expected int, got float` |
| **14_call_non_func** | Вызов не функции | `'x' is not a function` |
| **15_return_type** | Неправильный тип возврата | `cannot return float, expected int` |
| **16_return_in_void** | Возврат значения из void функции | `cannot return int, expected void` |
| **17_missing_return** | Отсутствие return в не-void функции | `function must return a value` |
| **18_scope** | Использование переменной после блока | `identifier 'x' not declared` |

## IR тесты

Расположены в: `tests/ir/`

### Генерация (`tests/ir/generation/`)

| Категория | Тест | Описание |
|-----------|------|----------|
| **expressions** | test_arithmetic | Арифметические выражения |
| | test_comparison | Операторы сравнения |
| | test_literals | Литералы разных типов |
| | test_logical | Логические операции |
| | test_unary | Унарные операции |
| **control_flow** | test_simple_if | Условный оператор if |
| | test_if_else | If-else конструкция |
| | test_nested_if | Вложенные условия |
| | test_while_loop | Цикл while |
| | test_for_loop | Цикл for |
| | test_empty_for | Бесконечный цикл for |
| **functions** | test_simple_call | Вызов функции |
| | test_recursive | Рекурсивная функция |
| | test_multiple_params | Функция с параметрами |
| | test_void_function | Void-функция |
| **integration** | test_factorial | Факториал |
| | test_fibonacci | Числа Фибоначчи |
| | test_gcd | Алгоритм Евклида |
| | test_prime | Проверка на простоту |

### Валидация (`tests/ir/validation/`)

| Категория | Тест | Описание |
|-----------|------|----------|
| **structural** | test_empty_function | Пустая функция |
| | test_multiple_returns | Множественные возвраты |
| | test_nested_blocks | Вложенные блоки |
| **type_consistency** | test_mixed_types | Смешанные типы |
| **optimization** | test_algebraic | Алгебраические упрощения |
| | test_constant_folding | Свёртка констант |

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
| **Операторы** | `+`, `-`, `*`, `/`, `%`, `=`, `+=`, `-=`, `*=`, `/=`, `==`, `!=`, `<`, `<=`, `>`, `>=`, `&&`, `\|\|`, `!` |
| **Разделители** | `()`, `{}`, `[]`, `;`, `,`, `.`, `->` |
| **Типы данных** | `int`, `float`, `bool`, `string`, `void`, пользовательские структуры |
| **Литералы** | Целые числа (32-bit), числа с плавающей точкой, строки, булевы значения |
| **Комментарии** | Однострочные (`//`), многострочные (`/* */`) с поддержкой вложенности |

### Правила семантического анализа

| Проверка | Описание |
|----------|----------|
| **Совместимость типов** | `int` может быть неявно преобразован в `float`, обратное требует явного приведения |
| **Области видимости** | Поддержка глобальной, функциональной и блочной областей видимости |
| **Проверка возврата** | Не-void функции должны иметь `return` во всех ветках |
| **Проверка аргументов** | Количество и типы аргументов должны соответствовать объявлению функции |
| **Условия** | Выражения в `if` и `while` должны иметь тип `bool` |

### IR инструкции

| Категория | Инструкции |
|-----------|-----------|
| **Арифметические** | `ADD`, `SUB`, `MUL`, `DIV`, `MOD`, `NEG` |
| **Логические** | `AND`, `OR`, `NOT`, `XOR` |
| **Сравнения** | `CMP_EQ`, `CMP_NE`, `CMP_LT`, `CMP_LE`, `CMP_GT`, `CMP_GE` |
| **Память** | `LOAD`, `STORE`, `ALLOCA`, `GEP` |
| **Управление** | `JUMP`, `JUMP_IF`, `JUMP_IF_NOT`, `LABEL`, `PHI` |
| **Функции** | `CALL`, `RETURN`, `PARAM` |
| **Перемещение** | `MOVE` |

## Структура проекта

```
mikrocompiler/
├── cmd/
│   └── compiler/
│       └── main.go                      # Точка входа
├── internal/
│   ├── ast/
│   │   ├── ast.go                       # Определения узлов AST
│   │   ├── printer.go                   # Pretty printer для AST
│   │   ├── dot_printer.go               # Генератор DOT графов
│   │   └── json_printer.go              # Генератор JSON
│   ├── lexer/
│   │   └── scanner.go                   # Лексический анализатор
│   ├── parser/
│   │   ├── parser.go                    # Синтаксический анализатор
│   │   └── grammar.txt                  # Формальная грамматика
│   ├── semantic/
│   │   ├── analyzer.go                  # Семантический анализатор
│   │   ├── symbol_table.go              # Таблица символов
│   │   ├── type_system.go               # Система типов
│   │   ├── errors.go                    # Ошибки семантики
│   │   ├── types.go                     # Определения типов
│   │   └── semantic_test.go             # Unit тесты семантики
│   ├── ir/
│   │   ├── generator.go                 # Генератор IR
│   │   ├── instruction.go               # IR инструкции
│   │   ├── basic_block.go               # Базовые блоки CFG
│   │   ├── function.go                  # Функции в IR
│   │   ├── operand.go                   # Операнды IR
│   │   ├── optimizer.go                 # Peephole оптимизатор
│   │   ├── printer.go                   # Вывод IR (text/dot/json)
│   │   ├── program.go                   # IR программа
│   │   └── ir_test.go                   # Unit тесты IR
│   └── token/
│       └── token.go                     # Определения токенов
├── tests/
│   ├── lexer/
│   │   ├── valid/                       # Валидные тесты лексера
│   │   └── invalid/                     # Тесты с ошибками лексера
│   ├── parser/
│   │   ├── valid/                       # Валидные тесты парсера
│   │   │   ├── expressions/
│   │   │   ├── statements/
│   │   │   ├── declarations/
│   │   │   └── full_programs/
│   │   └── invalid/                     # Тесты с ошибками парсера
│   │       └── syntax_errors/
│   ├── semantic/
│   │   ├── valid/                       # Валидные семантические тесты
│   │   └── invalid/                     # Тесты с семантическими ошибками
│   ├── ir/
│   │   ├── generation/                  # Тесты генерации IR
│   │   │   ├── expressions/
│   │   │   ├── control_flow/
│   │   │   ├── functions/
│   │   │   └── integration/
│   │   └── validation/                  # Тесты валидации IR
│   │       ├── structural/
│   │       ├── type_consistency/
│   │       └── optimization/
│   └── test_runner/
│       └── run_tests.sh                 # Скрипт запуска тестов
├── examples/
│   ├── hello.src                        # Пример с базовыми конструкциями
│   ├── factorial.src                    # Пример с рекурсией
│   └── struct.src                       # Пример со структурами
├── docs/
│   ├── language_spec.md                 # Лексическая спецификация
│   └── grammar.md                       # Грамматика языка
├── Makefile                             # Автоматизация сборки
├── go.mod
└── README.md
```

## Makefile

| Команда | Описание |
|---------|----------|
| `make build` | Сборка компилятора |
| `make run` | Запуск парсера на `examples/factorial.src` |
| `make run-lex` | Запуск лексера на `examples/hello.src` |
| `make run-parse` | Запуск парсера на `examples/struct.src` |
| `make check` | Семантический анализ на `examples/factorial.src` |
| `make check-types` | Семантический анализ с выводом типов |
| `make symbols` | Вывод таблицы символов |
| `make ir` | Генерация IR |
| `make ir-dot` | Генерация CFG в PNG |
| `make ir-json` | Генерация IR в JSON |
| `make ir-stats` | Статистика IR |
| `make test` | Запуск всех тестов |
| `make clean` | Очистка артефактов сборки |
| `make help` | Показать все доступные команды |
