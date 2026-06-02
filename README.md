# MiniCompiler v1.0.0

Компилятор для C-подобного языка программирования (MiniLang), реализованный на Go. 
Генерирует x86-64 ассемблер (NASM) по стандарту System V AMD64 ABI.

## Возможности

- **Лексический анализ** — токенизация исходного кода с отслеживанием позиций
- **Синтаксический анализ** — построение AST с поддержкой структур, массивов, функций
- **Семантический анализ** — проверка типов, областей видимости, сигнатур функций
- **Промежуточное представление (IR)** — три-адресный код с базовыми блоками и CFG
- **Генерация кода x86-64** — NASM ассемблер (System V AMD64 ABI)
- **Оптимизации** — свёртка констант, алгебраические упрощения, удаление мёртвого кода
- **Массивы** — динамическое выделение через malloc/free с автоматической очисткой
- **Внешние функции C** — вызов printf, malloc, free, abort и др.
- **Цветной вывод ошибок** — human-readable, JSON, GCC-совместимые форматы
- **Система предупреждений** — -Wall, -Werror, -Wno

## Быстрый старт

### Требования

- **Go** 1.21+
- **NASM** 2.16+ (для ассемблирования)
- **GCC** (для линковки)
- **Linux** x86-64 (основная платформа)

### Сборка

```bash
git clone https://github.com/kokinf/MiniCompiler
cd MiniCompiler
make build
./bin/compiler --version
```

### Первая программа

```bash
cat > hello.src << 'EOF'
int main() {
    return 42;
}
EOF

# Компиляция
./bin/compiler hello.src -o hello.asm

# Сборка
nasm -f elf64 -o runtime.o src/runtime/runtime.asm
nasm -f elf64 -o hello.o hello.asm
gcc -no-pie -o hello runtime.o hello.o

# Запуск
./hello
echo $?  # 42
```

## Использование

```
compiler [опции] <файл.src>              # Компиляция
compiler [опции] <команда> [аргументы]     # Специальные операции
```

### Основные опции

| Флаг | Описание |
|------|----------|
| `-o <file>` | Выходной файл |
| `-S` | Только ассемблер |
| `-c` | Объектный файл |
| `-E` | Препроцессор (токены) |
| `-v, --verbose` | Подробный вывод |
| `-O, --optimize` | Оптимизации |
| `--ast` | Вывести AST |
| `--ir` | Вывести IR |
| `-Wall` | Все предупреждения |
| `-Werror` | Предупреждения как ошибки |
| `--error-format` | human / json / gcc |
| `--version` | Версия |
| `-h, --help` | Справка |

### Команды

| Команда | Описание |
|---------|----------|
| `lex <file>` | Лексический анализ |
| `parse <file>` | Синтаксический анализ (AST) |
| `check <file>` | Семантический анализ |
| `symbols <file>` | Таблица символов |
| `ir <file>` | Промежуточное представление |
| `compile <file>` | Компиляция в ассемблер |
| `test <type>` | Запуск тестов |

### Примеры использования

```bash
# Компиляция с оптимизациями
compiler -O -v program.src -o program.asm

# Только ассемблер
compiler -S program.src

# Объектный файл
compiler -c program.src -o program.o

# Просмотр токенов
compiler -E program.src

# Просмотр AST
compiler --ast program.src

# Просмотр IR со статистикой
compiler --ir --stats program.src

# Несколько файлов
compiler main.src utils.src -o program

# JSON формат ошибок
compiler --error-format=json program.src

# Строгий режим
compiler -Wall -Werror program.src

# Таблица символов в JSON
compiler symbols --format json program.src
```

## Язык MiniLang

### Типы данных

- `int` — 32-битное целое
- `float` — 64-битное с плавающей точкой
- `bool` — булево (true/false)
- `string` — строковый литерал
- `void` — отсутствие значения
- `struct` — пользовательские структуры
- `T[]` — массив (динамический, через malloc)

### Ключевые слова

`if`, `else`, `while`, `for`, `int`, `float`, `bool`, `string`, `void`, `struct`, `fn`, `extern`, `return`, `true`, `false`

### Операторы

Арифметические: `+`, `-`, `*`, `/`, `%`
Сравнения: `==`, `!=`, `<`, `<=`, `>`, `>=`
Логические: `&&`, `||`, `!`
Присваивания: `=`, `+=`, `-=`, `*=`, `/=`

### Пример программы

```c
int factorial(int n) {
    if (n <= 1) { return 1; }
    return n * factorial(n - 1);
}

int sum_array(int arr[], int size) {
    int sum = 0;
    for (int i = 0; i < size; i = i + 1) {
        sum = sum + arr[i];
    }
    return sum;
}

int main() {
    int arr[5] = {1, 2, 3, 4, 5};
    int fact = factorial(5);
    int sum = sum_array(arr, 5);
    return fact + sum;  // 120 + 15 = 135
}
```

Полная спецификация: [docs/language_spec.md](docs/language_spec.md)

## Примеры работы

### Лексический анализ

```bash
./bin/compiler lex --input examples/hello.src
```

```
1:1 KW_INT "int"
1:5 IDENTIFIER "main"
1:9 LPAREN "("
1:10 RPAREN ")"
1:12 LBRACE "{"
2:5 KW_RETURN "return"
2:12 INT_LITERAL "42" 42
2:14 SEMICOLON ";"
3:1 RBRACE "}"
4:1 END_OF_FILE ""
```

### Синтаксический анализ (AST)

```bash
./bin/compiler parse --input examples/factorial.src
```

```
Program:
  FunctionDecl: factorial -> int [line 1]:
    Parameters:
      int n
    Body:
      IfStmt:
        Condition: (n <= 1)
        Then: Return: 1
        Else: Return: (n * factorial((n - 1)))
```

### Семантический анализ

```bash
./bin/compiler check --input examples/factorial.src --verbose
```

```
Symbol Table:
global scope (level 0):
  factorial: function -> int (line 1)
  main: function -> int (line 8)
```

### Промежуточное представление (IR)

```bash
./bin/compiler ir --input examples/factorial.src --optimize --stats
```

```
function factorial: int (int n)
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

IR Statistics: Functions: 1, Blocks: 3, Instructions: 7
```

### Оптимизации

| Оптимизация | Описание | Пример |
|-------------|----------|--------|
| Свёртка констант | Вычисление на этапе компиляции | `3 + 4 → 7` |
| Алгебраические упрощения | Удаление тривиальных операций | `x + 0 → x`, `x * 1 → x` |
| Strength reduction | Замена дорогих операций | `x * 2 → x + x` |
| Dead code elimination | Удаление неиспользуемого кода | Неиспользуемые присваивания |
| Jump chaining | Упрощение переходов | `jmp L1; L1: jmp L2 → jmp L2` |
| Constant propagation | Распространение констант | `x = 5; y = x → y = 5` |

## Тестирование

```bash
make test           # все тесты
make test-lexer     # только лексер
make test-parser    # только парсер
make test-semantic  # только семантика
make test-codegen   # только кодогенерация
make test-final     # финальные тесты (80+)
```

### Категории тестов

| Категория | Валидных | Невалидных | Всего |
|-----------|----------|------------|-------|
| Лексер | 7 | 5 | 12 |
| Парсер | 12 | 6 | 18 |
| Семантика | 13 | 18 | 31 |
| IR | 18 | — | 18 |
| Codegen | 12 | — | 12 |
| Финальные | 70+ | 7 | 80+ |

## Документация

- [Спецификация языка](docs/language_spec.md) — полная грамматика, типы, семантика
- [Руководство разработчика](docs/developer.md) — архитектура, API, расширение
- [Туториал](docs/tutorial.md) — пошаговое руководство
- [История изменений](CHANGELOG.md)
- Man-страница: `man mikrocompiler` (после `make install-man`)

## Установка

```bash
# Установка компилятора
sudo make install

# Установка man-страницы
sudo make install-man

# Проверка
mikrocompiler --help
man mikrocompiler
```

## Структура проекта

```
mikrocompiler/
├── src/
│   ├── cmd/compiler/main.go     # CLI (точка входа)
│   ├── internal/
│   │   ├── lexer/scanner.go     # Лексический анализатор
│   │   ├── parser/parser.go     # Синтаксический анализатор
│   │   ├── ast/                 # AST (узлы, принтеры)
│   │   ├── semantic/            # Семантический анализ
│   │   ├── ir/                  # IR (генератор, оптимизатор)
│   │   ├── codegen/             # Генератор x86-64
│   │   ├── token/token.go       # Токены
│   │   └── utils/               # Утилиты (цвета, ошибки)
│   ├── runtime/runtime.asm      # Рантайм (print_int, read_int, exit)
│   └── libc/stdlib.h            # Заголовки libc
├── tests/
│   ├── lexer/                   # Тесты лексера
│   ├── parser/                  # Тесты парсера
│   ├── semantic/                # Семантические тесты
│   ├── ir/                      # IR тесты
│   ├── codegen/                 # Тесты кодогенерации
│   ├── control_flow/            # Control flow тесты
│   ├── final/                   # Финальные тесты (80+)
│   └── test_runner/             # Скрипты запуска
├── examples/                    # Примеры программ
├── docs/                        # Документация
├── Makefile                     # Сборка и тестирование
├── mikrocompiler.1              # Man-страница
├── go.mod
├── LICENSE
└── README.md
```

## Makefile

| Команда | Описание |
|---------|----------|
| `make build` | Сборка компилятора |
| `make demo` | Демонстрация QuickSort |
| `make test` | Все тесты |
| `make test-final` | Финальные тесты (80+) |
| `make install` | Установка в /usr/local/bin |
| `make install-man` | Установка man-страницы |
| `make dist` | Создание дистрибутива |
| `make clean` | Очистка |
| `make help` | Все команды |

