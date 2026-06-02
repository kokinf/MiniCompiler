## 0. Подготовка

```bash
# Клонирование и сборка
git clone https://github.com/kokinf/MiniCompiler
cd MiniCompiler
make build

# Проверка
./bin/compiler --version
```

**Ожидаемый вывод:**
```
mikrocompiler 1.0.0
Автор: Fokin Nikita
Сборка: 2024-06-01
Цель: x86_64-linux-gnu

Мини-компилятор языка C-like в x86-64 ассемблер
Поддерживает: переменные, функции, массивы, структуры, указатели
System V AMD64 ABI, NASM синтаксис
```

---

## 1. Справка и флаги

### 1.1 Показать справку

```bash
./bin/compiler --help
```

**Ожидаемый вывод:** список всех опций и команд.

### 1.2 Показать справку с цветом

```bash
./bin/compiler --help --color=always
```

**Ожидаемый вывод:** цветная справка.

---

## 2. Лексический анализ (токенизация)

### 2.1 Простая программа

```bash
cat > demo_hello.src << 'EOF'
int main() {
    return 42;
}
EOF

./bin/compiler -E demo_hello.src
```

**Ожидаемый вывод:**
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

### 2.2 С подробным выводом

```bash
./bin/compiler lex demo_hello.src --verbose
```

---

## 3. Синтаксический анализ (AST)

### 3.1 Текстовый AST

```bash
cat > demo_fact.src << 'EOF'
int factorial(int n) {
    if (n <= 1) { return 1; }
    return n * factorial(n - 1);
}
int main() { return factorial(5); }
EOF

./bin/compiler parse demo_fact.src --format text
```

### 3.3 JSON формат

```bash
./bin/compiler parse demo_fact.src --format json
```

---

## 4. Семантический анализ

### 4.1 Проверка типов

```bash
./bin/compiler check demo_fact.src --verbose
```

### 4.2 С выводом типов

```bash
./bin/compiler check demo_fact.src --show-types
```

### 4.3 Таблица символов

```bash
./bin/compiler symbols demo_fact.src --format text
./bin/compiler symbols demo_fact.src --format json
```

### 4.4 Демонстрация ошибок

```bash
cat > demo_error.src << 'EOF'
int main() {
    int x = "hello";
    return x;
}
EOF

./bin/compiler check demo_error.src
```

**Ожидаемый вывод:** Ошибка E003 (type mismatch).

```bash
cat > demo_error2.src << 'EOF'
int add(int a, int b) { return a + b; }
int main() {
    return add(1, 2, 3);
}
EOF

./bin/compiler check demo_error2.src
```

**Ожидаемый вывод:** Ошибка E004 (argument count mismatch).

---

## 5. Промежуточное представление (IR)

### 5.1 Генерация IR

```bash
./bin/compiler ir demo_fact.src --format text
```

### 5.2 Со статистикой

```bash
./bin/compiler ir demo_fact.src --stats
```

**Ожидаемый вывод:**
```
IR Statistics:
================
Functions:       2
Basic Blocks:    5
Instructions:    12
Temporaries:     6
```

### 5.3 Оптимизированный IR

```bash
./bin/compiler ir demo_fact.src --optimize --stats
```

### 5.4 Control Flow Graph (DOT)

```bash
./bin/compiler ir demo_fact.src --format dot --output demo_cfg.dot
dot -Tpng demo_cfg.dot -o demo_cfg.png
echo "CFG сохранён в demo_cfg.png"
```

### 5.5 IR в JSON

```bash
./bin/compiler ir demo_fact.src --format json
```

---

## 6. Компиляция в ассемблер

### 6.1 Базовая компиляция

```bash
./bin/compiler demo_hello.src -o demo_hello.asm
cat demo_hello.asm
```

### 6.2 Компиляция с оптимизациями

```bash
./bin/compiler -O demo_fact.src -o demo_fact.asm
echo "Размер файла: $(wc -c < demo_fact.asm) байт"
```

### 6.3 Компиляция с подробным выводом

```bash
./bin/compiler -O -v demo_fact.src -o demo_fact_verbose.asm
```

### 6.4 Только ассемблер (флаг -S)

```bash
./bin/compiler -S demo_hello.src -o demo_hello_s.asm
```

### 6.5 Объектный файл (флаг -c)

```bash
./bin/compiler -c demo_hello.src -o demo_hello.o
file demo_hello.o
```

**Ожидаемый вывод:** `ELF 64-bit LSB relocatable, x86-64`

---

## 7. Сборка и запуск

### 7.1 Полный цикл

```bash
# Компиляция
./bin/compiler demo_fact.src -o demo_fact.asm

# Ассемблирование
nasm -f elf64 -o runtime.o src/runtime/runtime.asm
nasm -f elf64 -o demo_fact.o demo_fact.asm

# Линковка
gcc -no-pie -o demo_fact runtime.o demo_fact.o

# Запуск
./demo_fact
echo "Exit code: $?"
```

**Ожидаемый вывод:** `Exit code: 120` (factorial(5) = 120)

### 7.2 Множественные файлы

```bash
cat > demo_add.src << 'EOF'
int add(int a, int b) { return a + b; }
EOF

cat > demo_main.src << 'EOF'
int main() { return add(10, 20); }
EOF

./bin/compiler demo_main.src demo_add.src -o demo_multi
echo "Мульти-файловая компиляция завершена"
```

---

## 8. Форматы ошибок

### 8.1 Human-readable (по умолчанию, с цветом)

```bash
./bin/compiler demo_error.src --color=always
```

### 8.2 JSON формат

```bash
./bin/compiler --error-format=json demo_error.src 2>&1
```

### 8.3 GCC-совместимый формат

```bash
./bin/compiler --error-format=gcc demo_error.src 2>&1
```

---

## 9. Предупреждения

### 9.1 Все предупреждения

```bash
./bin/compiler -Wall demo_hello.src
```

### 9.2 Предупреждения как ошибки

```bash
./bin/compiler -Wall -Werror demo_hello.src
```

---

## 10. Финальная демо-программа

### 10.1 Исходный код

```bash
cat examples/demo_final.srс
```

### 10.2 Компиляция с оптимизациями

```bash
./bin/compiler -O -v examples/demo_final.srс -o demo_final.asm
```

### 10.3 Сборка

```bash
mkdir -p build
nasm -f elf64 -o build/runtime.o src/runtime/runtime.asm
nasm -f elf64 -o build/demo_final.o demo_final.asm
gcc -no-pie -o build/demo_final build/runtime.o build/demo_final.o
```

### 10.4 Запуск

```bash
./build/demo_final
```

**Ожидаемый вывод:**
```
╔══════════════════════════════════╗
║   MiniCompiler FINAL DEMO       ║
╚══════════════════════════════════╝

--- Arithmetic ---
Arithmetic sum 1..10 = 55
Is 17 prime? = 1
Is 100 prime? = 0

--- Recursion ---
Factorial(6) = 720
Fibonacci(10) = 55
GCD(48, 18) = 6

--- Arrays ---
Original array: [64, 34, 25, 12, 22, 11, 90, 45]
Sum = 303, Max = 90
Sorted array: [11, 12, 22, 25, 34, 45, 64, 90]
Sum = 303, Min = 11, Max = 90

--- Verification ---
All tests passed!
```

### 10.5 Проверка результата

```bash
./build/demo_final
EXIT=$?
if [ $EXIT -eq 0 ]; then
    echo " Демонстрация успешна!"
else
    echo " Ошибка! Exit code: $EXIT"
fi
```

---

## 11. Просмотр статистики

### 11.1 Статистика IR

```bash
./bin/compiler ir examples/demo_final.srс --stats --optimize
```

### 11.2 Размер сгенерированного кода

```bash
./bin/compiler -O examples/demo_final.srс -o /tmp/demo_final.asm
echo "Строк ассемблера: $(wc -l < /tmp/demo_final.asm)"
echo "Байт ассемблера: $(wc -c < /tmp/demo_final.asm)"
```

### 11.3 Сравнение с оптимизациями и без

```bash
# Без оптимизаций
./bin/compiler examples/demo_final.srс -o /tmp/demo_unopt.asm
echo "Без оптимизаций: $(wc -l < /tmp/demo_unopt.asm) строк"

# С оптимизациями
./bin/compiler -O examples/demo_final.srс -o /tmp/demo_opt.asm
echo "С оптимизациями: $(wc -l < /tmp/demo_opt.asm) строк"
```

---

## 12. Очистка

```bash
# Удалить временные файлы
rm -f demo_*.src demo_*.asm demo_*.o demo_*.dot demo_*.png demo_*.json
rm -f runtime.o
rm -rf build/

echo "Очистка завершена"
```

---

## Полный скрипт для автоматического демо

```bash
#!/bin/bash
# demo_full.sh — полная демонстрация MiniCompiler

echo "╔══════════════════════════════════════════════════╗"
echo "║     MiniCompiler v1.0 — Full Demo               ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""

# 1. Версия
echo "=== 1. Version ==="
./bin/compiler --version
echo ""

# 2. Лексер
echo "=== 2. Lexer ==="
echo 'int main() { return 42; }' > /tmp/demo.src
./bin/compiler -E /tmp/demo.src
echo ""

# 3. Парсер
echo "=== 3. Parser (AST) ==="
./bin/compiler parse /tmp/demo.src --format text
echo ""

# 4. Семантика
echo "=== 4. Semantic ==="
./bin/compiler check /tmp/demo.src
echo ""

# 5. IR
echo "=== 5. IR ==="
./bin/compiler ir /tmp/demo.src --stats
echo ""

# 6. Компиляция
echo "=== 6. Compilation ==="
./bin/compiler -O -v /tmp/demo.src -o /tmp/demo.asm
echo ""

# 7. Сборка и запуск
echo "=== 7. Build & Run ==="
nasm -f elf64 -o /tmp/runtime.o src/runtime/runtime.asm
nasm -f elf64 -o /tmp/demo.o /tmp/demo.asm
gcc -no-pie -o /tmp/demo /tmp/runtime.o /tmp/demo.o
/tmp/demo
echo "Exit code: $?"
echo ""

# 8. Ошибки
echo "=== 8. Error Demo ==="
echo 'int main() { int x = "hello"; return x; }' > /tmp/err.src
./bin/compiler check /tmp/err.src
echo ""

# 9. JSON ошибки
echo "=== 9. JSON Error Format ==="
./bin/compiler --error-format=json /tmp/err.src 2>&1
echo ""

# 10. Очистка
rm -f /tmp/demo.src /tmp/demo.asm /tmp/demo.o /tmp/runtime.o /tmp/demo /tmp/err.src

echo "╔══════════════════════════════════════════════════╗"
echo "║     Demo completed successfully!                ║"
echo "╚══════════════════════════════════════════════════╝"
```
