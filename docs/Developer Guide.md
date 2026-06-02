# Руководство разработчика (Developer Guide)

## Архитектура компилятора

MiniCompiler следует классической архитектуре компилятора с несколькими проходами:

```
Исходный код (.src)
    │
    ▼
[Lexer] ──► Токены
    │
    ▼
[Parser] ──► AST (Abstract Syntax Tree)
    │
    ▼
[Semantic Analyzer] ──► Декорированный AST + Таблица символов
    │
    ▼
[IR Generator] ──► IR (Intermediate Representation)
    │
    ▼
[Optimizer] ──► Оптимизированный IR
    │
    ▼
[Code Generator] ──► x86-64 Assembly (NASM)
    │
    ▼
[nasm + gcc] ──► Executable
```

## Структура пакетов

| Пакет | Назначение | Ключевые файлы |
|-------|-----------|----------------|
| `cmd/compiler` | CLI, точка входа | `main.go` |
| `internal/lexer` | Лексический анализ | `scanner.go` |
| `internal/parser` | Синтаксический анализ | `parser.go` |
| `internal/ast` | Узлы AST и принтеры | `ast.go`, `printer.go`, `dot_printer.go`, `json_printer.go` |
| `internal/token` | Определения токенов | `token.go` |
| `internal/semantic` | Семантический анализ | `analyzer.go`, `symbol_table.go`, `type_system.go`, `types.go`, `errors.go` |
| `internal/ir` | Промежуточное представление | `generator.go`, `instruction.go`, `basic_block.go`, `function.go`, `operand.go`, `optimizer.go`, `program.go`, `printer.go` |
| `internal/codegen` | Генерация x86-64 кода | `x86_generator.go`, `expression_generator.go`, `control_flow_generator.go`, `array_generator.go`, `external_calls.go`, `types.go`, `register_allocator.go`, `label_manager.go`, `optimization_passes.go` |
| `internal/utils` | Утилиты | `colors.go`, `warnings.go`, `error_formatter.go`, `suggestions.go` |
| `libc` | Парсер заголовков libc | `stdlib_parser.go` |

## Поток данных

### 1. Лексический анализ (Lexer)

**Вход:** строка исходного кода
**Выход:** слайс токенов `[]token.Token`

```go
scanner := lexer.NewScanner(source)
for !scanner.IsAtEnd() {
    tok := scanner.NextToken()
    // tok.Type, tok.Lexeme, tok.Line, tok.Column
}
```

### 2. Синтаксический анализ (Parser)

**Вход:** слайс токенов
**Выход:** `*ast.ProgramNode`

```go
parser := parser.NewParser(tokens, source)
program := parser.Parse()
// program.Declarations: []DeclarationNode (FunctionDecl, StructDecl, VarDecl, ExternFuncDecl)
```

**AST узлы:**
- `ExpressionNode` — `IdentifierNode`, `LiteralExprNode`, `BinaryExprNode`, `UnaryExprNode`, `CallExprNode`, `AssignmentExprNode`, `IndexExprNode`, `ArrayLiteralExprNode`
- `StatementNode` — `BlockStmtNode`, `IfStmtNode`, `WhileStmtNode`, `ForStmtNode`, `ReturnStmtNode`, `ExprStmtNode`
- `DeclarationNode` — `FunctionDeclNode`, `StructDeclNode`, `VarDeclNode`, `ExternFuncDeclNode`

### 3. Семантический анализ (Semantic)

**Вход:** `*ast.ProgramNode`
**Выход:** `*SymbolTable`, `*ErrorCollector`, декорированный AST

```go
analyzer := semantic.NewSemanticAnalyzer()
symbolTable, errors, decoratedAST := analyzer.Analyze(program)
```

**Проверки:**
- Объявления до использования
- Совместимость типов при присваивании
- Количество и типы аргументов функций
- Тип возвращаемого значения
- Тип условия в if/while/for
- Дубликаты объявлений в одной области видимости

### 4. Генерация IR

**Вход:** декорированный AST + таблица символов
**Выход:** `*ir.Program`

```go
irGenerator := ir.NewIRGenerator(symbolTable, typeSystem)
irProgram := irGenerator.Generate(decoratedAST)
```

**IR инструкции:**
- Арифметика: `ADD`, `SUB`, `MUL`, `DIV`, `MOD`, `NEG`
- Логика: `AND`, `OR`, `NOT`
- Сравнения: `CMP_EQ`, `CMP_NE`, `CMP_LT`, `CMP_LE`, `CMP_GT`, `CMP_GE`
- Память: `LOAD`, `STORE`, `ALLOCA`, `GEP`
- Управление: `JUMP`, `JUMP_IF`, `JUMP_IF_NOT`, `LABEL`, `PHI`
- Функции: `CALL`, `RETURN`, `PARAM`
- Данные: `MOVE`

### 5. Оптимизация IR

**Вход:** `*ir.Program`
**Выход:** оптимизированный `*ir.Program`

```go
optimizer := ir.NewPeepholeOptimizer(irProgram)
optimizer.Optimize()
// optimizer.GetOptimizationReport()
```

**Проходы оптимизации:**
1. `constant_folding` — свёртка константных выражений (`3 + 4 → 7`)
2. `algebraic` — алгебраические упрощения (`x + 0 → x`)
3. `strength_reduction` — понижение силы операций (`x * 2 → x + x`)
4. `constant_propagation` — распространение констант
5. `branch_simplification` — упрощение условных переходов
6. `dead_allocation` — удаление неиспользуемых malloc/free
7. `dead_code` — удаление мёртвого кода
8. `jump_optimization` — оптимизация цепочек переходов
9. `empty_block` — удаление пустых базовых блоков

### 6. Генерация кода x86-64

**Вход:** `*ir.Program`
**Выход:** строка с NASM ассемблером

```go
codeGenerator := codegen.NewX86Generator(irProgram, symbolTable, typeSystem)
asmCode := codeGenerator.Generate()
```

**Компоненты кодогенератора:**
- `X86Generator` — основной генератор
- `ExpressionGenerator` — арифметика, сравнения, логика
- `ControlFlowGenerator` — условные/безусловные переходы, циклы
- `ArrayGenerator` — операции с массивами (GEP, malloc/free)
- `ExternalCallGenerator` — вызовы внешних функций (printf, malloc)
- `LabelManager` — управление метками
- `StackFrame` — управление стековым кадром
- `ABIInfo` — System V AMD64 ABI

**Соглашения ABI:**
- Параметры: `rdi`, `rsi`, `rdx`, `rcx`, `r8`, `r9` (целые), `xmm0-xmm7` (float)
- Возврат: `rax` (64-bit), `eax` (32-bit)
- Стек: выравнивание на 16 байт перед `call`
- Callee-saved: `rbx`, `r12-r15`, `rbp`, `rsp`
- Caller-saved: `rax`, `rcx`, `rdx`, `rsi`, `rdi`, `r8-r11`

## Добавление новых возможностей

### Добавление нового типа выражения

1. **AST:** Добавить узел в `internal/ast/ast.go`:
```go
type NewExprNode struct {
    Token token.Token
    // поля...
    TypeAnnotation *TypeAnnotation
}
func (n *NewExprNode) expressionNode() {}
func (n *NewExprNode) Accept(v Visitor) interface{} { return v.VisitNewExpr(n) }
```

2. **Visitor:** Добавить метод в интерфейс `Visitor`:
```go
VisitNewExpr(node *NewExprNode) interface{}
```

3. **Parser:** Добавить разбор в `internal/parser/parser.go`

4. **Semantic:** Добавить проверку в `internal/semantic/analyzer.go`

5. **IR:** Добавить генерацию в `internal/ir/generator.go`

6. **Codegen:** Добавить генерацию кода в `internal/codegen/`

7. **Принтеры:** Обновить `PrettyPrinter`, `DOTPrinter`, `JSONPrinter`

### Добавление новой IR инструкции

1. **Opcode:** Добавить в `internal/ir/instruction.go`:
```go
const (
    OpNew Opcode = iota
    // ...
)
var opcodeNames = map[Opcode]string{
    OpNew: "NEW",
    // ...
}
```

2. **IR Generator:** Использовать новую инструкцию

3. **Codegen:** Добавить обработку в `control_flow_generator.go`

### Добавление оптимизационного прохода

1. **Optimizer:** Добавить в `internal/ir/optimizer.go`:
```go
// В NewPeepholeOptimizer:
{Name: "my_new_opt", Description: "Описание", Enabled: true}
```

2. **Метод оптимизации:**
```go
func (o *PeepholeOptimizer) myNewOptimization(fn *Function) bool {
    changed := false
    for _, block := range fn.Blocks {
        // логика оптимизации
    }
    return changed
}
```

## Отладка

### Просмотр IR

```bash
./bin/compiler --ir program.src              # текстовый IR
./bin/compiler --ir --stats program.src       # IR + статистика
./bin/compiler --ir --format dot program.src  # DOT для Graphviz
./bin/compiler --ir --format json program.src # JSON
```

### Просмотр AST

```bash
./bin/compiler --ast program.src              # текстовый AST
./bin/compiler parse --format dot program.src # DOT для Graphviz
./bin/compiler parse --format json program.src # JSON
```

### Трассировка компиляции

```bash
./bin/compiler -v program.src    # подробный вывод всех этапов
./bin/compiler -O -v program.src # с оптимизациями
```

### Профилирование

```bash
go test -cpuprofile=cpu.prof -memprofile=mem.prof ./src/internal/...
go tool pprof cpu.prof
```

## Расширение для новых архитектур

Для добавления новой целевой архитектуры:

1. Создать пакет `internal/codegen/arm64/`
2. Реализовать интерфейс `CodeGenerator`:
```go
type CodeGenerator interface {
    Generate() string
    GenerateFunction(fn *ir.Function)
    // ...
}
```
3. Добавить в `runCompiler` поддержку `--target arm64`

## Тестирование

### Unit тесты

```bash
go test ./src/internal/... -v
go test ./src/internal/ir/... -v -run TestOptimizer
```

### Интеграционные тесты

```bash
make test-final          # все финальные тесты
bash tests/run_final_tests.sh  # прямой запуск
```

### Создание нового теста

```bash
# Валидный тест
cat > tests/final/good/category/my_test.src << 'EOF'
int main() {
    return 42;
}
EOF
echo "42" > tests/final/good/category/my_test.expected

# Невалидный тест
cat > tests/final/bad/syntax_errors/my_test.src << 'EOF'
int main() {
    return 42  // нет точки с запятой
}
EOF
```