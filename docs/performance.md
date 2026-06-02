# Performance Tuning Guide

## Оптимизации компилятора

MiniCompiler поддерживает несколько уровней оптимизаций. По умолчанию оптимизации отключены.

### Включение оптимизаций

```bash
# Базовые оптимизации
./bin/compiler -O program.src -o program.asm

# С подробным отчётом
./bin/compiler -O -v program.src -o program.asm
```

### Доступные оптимизации IR

| Оптимизация | Описание | Пример | Ускорение |
|-------------|----------|--------|-----------|
| **Constant Folding** | Вычисление констант на этапе компиляции | `3 + 4 → 7` | 5-15% |
| **Algebraic Simplification** | Удаление тривиальных операций | `x + 0 → x`, `x * 1 → x` | 3-8% |
| **Strength Reduction** | Замена дорогих операций | `x * 2 → x + x` | 2-5% |
| **Constant Propagation** | Замена переменных константами | `x = 5; y = x → y = 5` | 5-10% |
| **Dead Code Elimination** | Удаление неиспользуемого кода | `x = 5; y = 6; return x` → удалить `y = 6` | 10-20% |
| **Branch Simplification** | Упрощение условных переходов | `if (true) → безусловный переход` | 3-7% |
| **Jump Optimization** | Сокращение цепочек переходов | `jmp L1; L1: jmp L2 → jmp L2` | 2-5% |

### Просмотр отчёта оптимизаций

```bash
./bin/compiler -O --ir --stats program.src
```

Пример вывода:
```
═══════════════════════════════════
   OPTIMIZATION REPORT
═══════════════════════════════════

Total iterations: 2

Passes executed:
─────────────────────────────────
  ✓ constant_folding      :   3 changes
  ✓ algebraic             :   1 changes
  ✓ dead_code             :   2 changes

Total changes: 6
```

## Советы по написанию эффективного кода

### 1. Используйте локальные переменные вместо повторных вычислений

```c
// Плохо — arr[i] вычисляется 3 раза
for (int i = 0; i < n; i = i + 1) {
    if (arr[i] > 0) {
        sum = sum + arr[i];
    }
}

// Хорошо — загружаем arr[i] один раз
for (int i = 0; i < n; i = i + 1) {
    int val = arr[i];
    if (val > 0) {
        sum = sum + val;
    }
}
```

### 2. Выносите инвариантные вычисления из циклов

```c
// Плохо — size * 2 вычисляется на каждой итерации
for (int i = 0; i < n; i = i + 1) {
    int threshold = size * 2;
    if (arr[i] > threshold) {
        count = count + 1;
    }
}

// Хорошо — вычисляем до цикла
int threshold = size * 2;
for (int i = 0; i < n; i = i + 1) {
    if (arr[i] > threshold) {
        count = count + 1;
    }
}
```

### 3. Используйте while для простых счётчиков

```c
// while иногда эффективнее for
int i = 0;
while (i < n) {
    sum = sum + arr[i];
    i = i + 1;
}
```

### 4. Избегайте избыточных присваиваний

```c
// Плохо
int x = 0;
x = 5;
x = 10;

// Хорошо
int x = 10;
```

### 5. Минимизируйте вызовы функций в циклах

```c
// Плохо — вызов на каждой итерации
for (int i = 0; i < n; i = i + 1) {
    sum = sum + factorial(i);
}

// Хорошо — если факториал не меняется
int fact = factorial(k);
for (int i = 0; i < n; i = i + 1) {
    sum = sum + fact;
}
```

## Память и массивы

### Размер массива

Локальные массивы выделяются через `malloc`. Каждый массив требует:
- **8 байт** на стеке (указатель)
- **N × sizeof(T)** байт в куче
- **free()** при выходе из области видимости

```c
// Память: 8 байт стек + 40 байт куча + free()
int arr[10];  // 10 * 4 = 40 байт в куче

// Память: 8 байт стек + 80 байт куча + free()
double arr[10];  // 10 * 8 = 80 байт в куче
```

### Передача массива в функцию

Передача массива передаёт **указатель** (8 байт), а не копирует весь массив:

```c
// Эффективно — передаётся только указатель
int sum_array(int arr[], int size) {
    int sum = 0;
    for (int i = 0; i < size; i = i + 1) {
        sum = sum + arr[i];
    }
    return sum;
}
```

### Рекурсия vs Итерация

```c
// Рекурсивный факториал (стек: O(n))
int factorial(int n) {
    if (n <= 1) { return 1; }
    return n * factorial(n - 1);
}

// Итеративный факториал (стек: O(1))
int factorial_iter(int n) {
    int result = 1;
    for (int i = 1; i <= n; i = i + 1) {
        result = result * i;
    }
    return result;
}
```

**Рекомендация:** Для глубокой рекурсии (>1000) используйте итеративные версии.

## Измерение производительности

### Время выполнения

```bash
time ./program
```

### Код выхода как результат

```bash
./program
echo $?
```

### Размер сгенерированного кода

```bash
./bin/compiler -v program.src -o program.asm  # показывает размер
wc -l program.asm                              # количество строк
```

### Сравнение с оптимизациями и без

```bash
# Без оптимизаций
./bin/compiler program.src -o program_unopt.asm
nasm -f elf64 -o program_unopt.o program_unopt.asm
gcc -no-pie -o program_unopt runtime.o program_unopt.o
time ./program_unopt

# С оптимизациями
./bin/compiler -O program.src -o program_opt.asm
nasm -f elf64 -o program_opt.o program_opt.asm
gcc -no-pie -o program_opt runtime.o program_opt.o
time ./program_opt
```

## Ограничения

- **Размер стека:** По умолчанию 8 МБ (Linux). Глубокая рекурсия может вызвать stack overflow
- **Массивы:** Каждый локальный массив требует вызов `malloc`/`free`, что добавляет накладные расходы
- **Float операции:** Эмулируются программно (без SSE/AVX), медленнее чем в C
- **Отсутствие инлайнинга:** Вызовы функций всегда через `call` (накладные расходы ~5-10 инструкций)
```

## 2. Auto-generated API docs (GoDoc)

GoDoc генерируется автоматически из комментариев в коде. Добавим базовые комментарии:

```bash
# Сгенерировать GoDoc
go doc ./src/internal/...
```

Или установить `godoc`:

```bash
go install golang.org/x/tools/cmd/godoc@latest
godoc -http=:6060 &
# Открыть http://localhost:6060/pkg/mikrocompiler/
```

Добавим в `Makefile`:

```makefile
.PHONY: docs
docs:
	@echo "$(YELLOW)Генерация документации...$(NC)"
	@echo "GoDoc доступен по команде: go doc ./src/internal/..."
	@echo "Или запустите: godoc -http=:6060"
	@echo "$(GREEN)Документация сгенерирована$(NC)"
```