# Туториал по MiniLang

## Содержание

1. [Первая программа](#1-первая-программа)
2. [Переменные и типы](#2-переменные-и-типы)
3. [Условные конструкции](#3-условные-конструкции)
4. [Циклы](#4-циклы)
5. [Функции](#5-функции)
6. [Массивы](#6-массивы)
7. [Рекурсия](#7-рекурсия)
8. [Внешние функции](#8-внешние-функции)
9. [Структуры](#9-структуры)
10. [Полная программа](#10-полная-программа)

---

## 1. Первая программа

Создайте файл `hello.src`:

```c
int main() {
    return 42;
}
```

Скомпилируйте и запустите:

```bash
./bin/compiler hello.src -o hello.asm
nasm -f elf64 -o runtime.o src/runtime/runtime.asm
nasm -f elf64 -o hello.o hello.asm
gcc -no-pie -o hello runtime.o hello.o
./hello
echo $?  # 42
```

**Объяснение:**
- `int main()` — главная функция, возвращающая целое число
- `return 42` — возврат значения 42 (код выхода программы)

---

## 2. Переменные и типы

```c
int main() {
    int x = 10;           // целое число (32 бита)
    int y = 20;
    int sum = x + y;      // сложение
    
    int product = x * y;  // умножение
    int diff = y - x;     // вычитание
    int quot = y / x;     // деление
    
    return sum;            // 30
}
```

**Типы данных:**

| Тип | Описание | Пример |
|-----|----------|--------|
| `int` | 32-битное целое | `42`, `-10`, `0` |
| `float` | 64-битное с плавающей точкой | `3.14`, `-0.5` |
| `bool` | Булево значение | `true`, `false` |
| `string` | Строковый литерал | `"Hello"` |
| `void` | Отсутствие значения | Для функций без возврата |

**Операторы:**

| Оператор | Описание | Пример |
|----------|----------|--------|
| `+`, `-`, `*`, `/`, `%` | Арифметика | `x + y` |
| `==`, `!=`, `<`, `<=`, `>`, `>=` | Сравнения | `x == y` |
| `&&`, `||`, `!` | Логические | `x && y` |
| `=`, `+=`, `-=`, `*=`, `/=` | Присваивания | `x += 5` |

---

## 3. Условные конструкции

### If-else

```c
int main() {
    int x = 15;
    
    if (x > 10) {
        return 1;    // x больше 10
    } else {
        return 0;    // x меньше или равно 10
    }
}
```

### Вложенные условия

```c
int main() {
    int score = 85;
    
    if (score >= 90) {
        return 5;    // отлично
    } else {
        if (score >= 75) {
            return 4;    // хорошо
        } else {
            return 3;    // удовлетворительно
        }
    }
}
```

### Логические операторы

```c
int main() {
    int age = 25;
    int has_license = 1;  // 1 = true
    
    if (age >= 18 && has_license) {
        return 1;    // можно водить
    }
    return 0;
}
```

---

## 4. Циклы

### Цикл for

```c
int main() {
    int sum = 0;
    
    for (int i = 1; i <= 5; i = i + 1) {
        sum = sum + i;
    }
    
    return sum;  // 15 (1+2+3+4+5)
}
```

### Цикл while

```c
int main() {
    int i = 0;
    int sum = 0;
    
    while (i < 10) {
        sum = sum + i;
        i = i + 1;
    }
    
    return sum;  // 45 (0+1+...+9)
}
```

### Вложенные циклы

```c
int main() {
    for (int i = 1; i <= 3; i = i + 1) {
        for (int j = 1; j <= 3; j = j + 1) {
            // i*j будет вычислено 9 раз
        }
    }
    return 0;
}
```

---

## 5. Функции

### Объявление и вызов

```c
int add(int a, int b) {
    return a + b;
}

int main() {
    int result = add(10, 20);
    return result;  // 30
}
```

### Несколько параметров

```c
int sum3(int a, int b, int c) {
    return a + b + c;
}

int main() {
    return sum3(1, 2, 3);  // 6
}
```

### Void-функции

```c
void do_nothing() {
    // ничего не делает
}

int main() {
    do_nothing();
    return 0;
}
```

### Функции с условиями

```c
int max(int a, int b) {
    if (a > b) {
        return a;
    }
    return b;
}

int main() {
    return max(10, 20);  // 20
}
```

---

## 6. Массивы

### Объявление и инициализация

```c
int main() {
    int arr[5] = {10, 20, 30, 40, 50};
    
    return arr[0] + arr[4];  // 10 + 50 = 60
}
```

### Доступ к элементам

```c
int main() {
    int arr[3] = {1, 2, 3};
    
    arr[0] = 100;     // изменение элемента
    arr[1] = arr[2];  // копирование
    
    return arr[0];     // 100
}
```

### Массивы в циклах

```c
int main() {
    int arr[5] = {1, 2, 3, 4, 5};
    int sum = 0;
    
    for (int i = 0; i < 5; i = i + 1) {
        sum = sum + arr[i];
    }
    
    return sum;  // 15
}
```

### Передача массива в функцию

```c
int sum_array(int arr[], int size) {
    int sum = 0;
    for (int i = 0; i < size; i = i + 1) {
        sum = sum + arr[i];
    }
    return sum;
}

int main() {
    int arr[4] = {1, 2, 3, 4};
    return sum_array(arr, 4);  // 10
}
```

---

## 7. Рекурсия

### Факториал

```c
int factorial(int n) {
    if (n <= 1) {
        return 1;              // базовый случай
    }
    return n * factorial(n - 1);  // рекурсивный вызов
}

int main() {
    return factorial(5);  // 120
}
```

### Числа Фибоначчи

```c
int fib(int n) {
    if (n <= 1) {
        return n;
    }
    return fib(n - 1) + fib(n - 2);
}

int main() {
    return fib(10);  // 55
}
```

### Алгоритм Евклида (НОД)

```c
int gcd(int a, int b) {
    if (b == 0) {
        return a;
    }
    return gcd(b, a % b);
}

int main() {
    return gcd(48, 18);  // 6
}
```

---

## 8. Внешние функции

### Использование printf

```c
extern int printf(char* format, ...);

int main() {
    printf("Hello, World!\n");
    return 0;
}
```

**Важно:** При использовании внешних функций линкуйтесь с libc:
```bash
gcc -no-pie -o program runtime.o program.o  # libc подключается автоматически
```

### Выделение памяти (вручную)

```c
extern void* malloc(int size);
extern void free(void* ptr);

int main() {
    // MiniLang автоматически управляет памятью массивов,
    // но можно вызвать malloc и вручную
    return 0;
}
```

---

## 9. Структуры

### Объявление структуры

```c
struct Point {
    int x;
    int y;
};
```

### Использование структур

```c
struct Point {
    int x;
    int y;
};

int main() {
    struct Point p;
    p.x = 10;
    p.y = 20;
    return p.x + p.y;  // 30
}
```

**Примечание:** Присваивание полям структур (`p.x = 10`) имеет ограниченную поддержку в текущей версии.

---

## 10. Полная программа

Объединяет все изученные концепции:

```c
// Быстрая сортировка (QuickSort)

void swap(int arr[], int i, int j) {
    int temp = arr[i];
    arr[i] = arr[j];
    arr[j] = temp;
}

int partition(int arr[], int low, int high) {
    int pivot = arr[high];
    int i = low - 1;
    
    for (int j = low; j < high; j = j + 1) {
        if (arr[j] <= pivot) {
            i = i + 1;
            swap(arr, i, j);
        }
    }
    swap(arr, i + 1, high);
    return i + 1;
}

void quicksort(int arr[], int low, int high) {
    if (low < high) {
        int pi = partition(arr, low, high);
        quicksort(arr, low, pi - 1);
        quicksort(arr, pi + 1, high);
    }
}

int main() {
    int arr[8] = {64, 34, 25, 12, 22, 11, 90, 45};
    int size = 8;
    
    quicksort(arr, 0, size - 1);
    
    return arr[0] + arr[size - 1];  // 11 + 90 = 101
}
```

**Результат:** `101`

---

## Отладка программ

### Просмотр токенов
```bash
./bin/compiler -E program.src
```

### Просмотр AST
```bash
./bin/compiler --ast program.src
```

### Просмотр IR
```bash
./bin/compiler --ir --stats program.src
```

### Компиляция с подробным выводом
```bash
./bin/compiler -v program.src -o program.asm
```

### Компиляция с оптимизациями
```bash
./bin/compiler -O -v program.src -o program.asm
```

## Часто встречающиеся ошибки

### Ошибка: `undeclared identifier`
```c
int main() {
    return x;  // Ошибка: x не объявлена
}
```
**Исправление:** Объявите переменную перед использованием:
```c
int main() {
    int x = 42;
    return x;
}
```

### Ошибка: `type mismatch`
```c
int main() {
    int x = "hello";  // Ошибка: строка в int
    return 0;
}
```
**Исправление:** Используйте правильный тип:
```c
int main() {
    int x = 42;
    return x;
}
```

### Ошибка: `argument count mismatch`
```c
int add(int a, int b) { return a + b; }
int main() {
    return add(1);  // Ошибка: ожидалось 2 аргумента
}
```
**Исправление:** Передайте правильное количество аргументов:
```c
return add(1, 2);
```
