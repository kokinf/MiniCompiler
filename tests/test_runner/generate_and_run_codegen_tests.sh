#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

print_separator() {
    echo -e "${BLUE}========================================${NC}"
}

print_header() {
    echo ""
    print_separator
    echo -e "${BLUE}$1${NC}"
    print_separator
}

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_ROOT"

# Проверка наличия NASM
if ! command -v nasm &> /dev/null; then
    echo -e "${RED}NASM не установлен. Установите его:${NC}"
    echo "  Ubuntu/Debian: sudo apt-get install nasm"
    echo "  macOS: brew install nasm"
    echo "  Windows: скачайте с https://www.nasm.us/"
    exit 1
fi

# Проверка наличия ld
if ! command -v ld &> /dev/null; then
    echo -e "${RED}Линкер ld не найден${NC}"
    exit 1
fi

print_header "ГЕНЕРАЦИЯ ТЕСТОВ КОДОГЕНЕРАЦИИ"

TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

total_tests=0
passed_tests=0
failed_tests=0
failed_tests_list=()

# Создание директории для тестов
mkdir -p tests/codegen/valid/{arithmetic_ops,control_flow,function_calls,integration}
mkdir -p tests/codegen/expected

# ============================================================================
# Генерация тестовых программ
# ============================================================================

generate_test_programs() {
    echo -e "\n${YELLOW}Генерация тестовых программ...${NC}"
    
    # 1. Простое сложение
    cat > tests/codegen/valid/arithmetic_ops/simple_add.src << 'EOF'
fn main() -> int {
    return 2 + 3;
}
EOF
    echo "5" > tests/codegen/expected/simple_add.expected
    echo -e "  ${GREEN}✓${NC} simple_add"
    
    # 2. Вычитание
    cat > tests/codegen/valid/arithmetic_ops/subtract.src << 'EOF'
fn main() -> int {
    return 10 - 3;
}
EOF
    echo "7" > tests/codegen/expected/subtract.expected
    echo -e "  ${GREEN}✓${NC} subtract"
    
    # 3. Умножение
    cat > tests/codegen/valid/arithmetic_ops/multiply.src << 'EOF'
fn main() -> int {
    return 6 * 7;
}
EOF
    echo "42" > tests/codegen/expected/multiply.expected
    echo -e "  ${GREEN}✓${NC} multiply"
    
    # 4. Деление
    cat > tests/codegen/valid/arithmetic_ops/divide.src << 'EOF'
fn main() -> int {
    return 100 / 5;
}
EOF
    echo "20" > tests/codegen/expected/divide.expected
    echo -e "  ${GREEN}✓${NC} divide"
    
    # 5. Остаток от деления
    cat > tests/codegen/valid/arithmetic_ops/modulo.src << 'EOF'
fn main() -> int {
    return 17 % 5;
}
EOF
    echo "2" > tests/codegen/expected/modulo.expected
    echo -e "  ${GREEN}✓${NC} modulo"
    
    # 6. Сложное выражение
    cat > tests/codegen/valid/arithmetic_ops/complex_expr.src << 'EOF'
fn main() -> int {
    return 2 * 3 + 4 * 5;
}
EOF
    echo "26" > tests/codegen/expected/complex_expr.expected
    echo -e "  ${GREEN}✓${NC} complex_expr"
    
    # 7. Приоритет операций
    cat > tests/codegen/valid/arithmetic_ops/precedence.src << 'EOF'
fn main() -> int {
    return (2 + 3) * (4 + 5);
}
EOF
    echo "45" > tests/codegen/expected/precedence.expected
    echo -e "  ${GREEN}✓${NC} precedence"
    
    # 8. Сравнения: равно
    cat > tests/codegen/valid/control_flow/equals.src << 'EOF'
fn main() -> int {
    if (5 == 5) {
        return 1;
    }
    return 0;
}
EOF
    echo "1" > tests/codegen/expected/equals.expected
    echo -e "  ${GREEN}✓${NC} equals"
    
    # 9. Сравнения: не равно
    cat > tests/codegen/valid/control_flow/not_equals.src << 'EOF'
fn main() -> int {
    if (5 != 3) {
        return 1;
    }
    return 0;
}
EOF
    echo "1" > tests/codegen/expected/not_equals.expected
    echo -e "  ${GREEN}✓${NC} not_equals"
    
    # 10. Сравнения: меньше
    cat > tests/codegen/valid/control_flow/less_than.src << 'EOF'
fn main() -> int {
    if (3 < 5) {
        return 1;
    }
    return 0;
}
EOF
    echo "1" > tests/codegen/expected/less_than.expected
    echo -e "  ${GREEN}✓${NC} less_than"
    
    # 11. Сравнения: больше
    cat > tests/codegen/valid/control_flow/greater_than.src << 'EOF'
fn main() -> int {
    if (10 > 5) {
        return 1;
    }
    return 0;
}
EOF
    echo "1" > tests/codegen/expected/greater_than.expected
    echo -e "  ${GREEN}✓${NC} greater_than"
    
    # 12. If-else конструкция
    cat > tests/codegen/valid/control_flow/if_else.src << 'EOF'
fn main() -> int {
    if (10 > 5) {
        return 1;
    } else {
        return 0;
    }
}
EOF
    echo "1" > tests/codegen/expected/if_else.expected
    echo -e "  ${GREEN}✓${NC} if_else"
    
    # 13. Вложенные if
    cat > tests/codegen/valid/control_flow/nested_if.src << 'EOF'
fn main() -> int {
    int x = 10;
    if (x > 5) {
        if (x < 20) {
            return 1;
        }
    }
    return 0;
}
EOF
    echo "1" > tests/codegen/expected/nested_if.expected
    echo -e "  ${GREEN}✓${NC} nested_if"
    
    # 14. Цикл while
    cat > tests/codegen/valid/control_flow/while_loop.src << 'EOF'
fn main() -> int {
    int sum = 0;
    int i = 1;
    while (i <= 10) {
        sum = sum + i;
        i = i + 1;
    }
    return sum;
}
EOF
    echo "55" > tests/codegen/expected/while_loop.expected
    echo -e "  ${GREEN}✓${NC} while_loop"
    
    # 15. Цикл for
    cat > tests/codegen/valid/control_flow/for_loop.src << 'EOF'
fn main() -> int {
    int sum = 0;
    int i = 0;
    for (i = 1; i <= 10; i = i + 1) {
        sum = sum + i;
    }
    return sum;
}
EOF
    echo "55" > tests/codegen/expected/for_loop.expected
    echo -e "  ${GREEN}✓${NC} for_loop"
    
    # 16. Функция без параметров
    cat > tests/codegen/valid/function_calls/simple_func.src << 'EOF'
fn get_number() -> int {
    return 42;
}

fn main() -> int {
    return get_number();
}
EOF
    echo "42" > tests/codegen/expected/simple_func.expected
    echo -e "  ${GREEN}✓${NC} simple_func"
    
    # 17. Функция с параметрами
    cat > tests/codegen/valid/function_calls/function_params.src << 'EOF'
fn add(a int, b int) -> int {
    return a + b;
}

fn main() -> int {
    return add(10, 20);
}
EOF
    echo "30" > tests/codegen/expected/function_params.expected
    echo -e "  ${GREEN}✓${NC} function_params"
    
    # 18. Функция с несколькими параметрами
    cat > tests/codegen/valid/function_calls/multiple_params.src << 'EOF'
fn multiply(a int, b int, c int) -> int {
    return a * b * c;
}

fn main() -> int {
    return multiply(2, 3, 4);
}
EOF
    echo "24" > tests/codegen/expected/multiple_params.expected
    echo -e "  ${GREEN}✓${NC} multiple_params"
    
    # 19. Рекурсивная функция
    cat > tests/codegen/valid/function_calls/recursive.src << 'EOF'
fn factorial(n int) -> int {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

fn main() -> int {
    return factorial(5);
}
EOF
    echo "120" > tests/codegen/expected/recursive.expected
    echo -e "  ${GREEN}✓${NC} recursive"
    
    # 20. Числа Фибоначчи
    cat > tests/codegen/valid/function_calls/fibonacci.src << 'EOF'
fn fibonacci(n int) -> int {
    if (n <= 1) {
        return n;
    }
    return fibonacci(n - 1) + fibonacci(n - 2);
}

fn main() -> int {
    return fibonacci(10);
}
EOF
    echo "55" > tests/codegen/expected/fibonacci.expected
    echo -e "  ${GREEN}✓${NC} fibonacci"
    
    # 21. Логические операции
    cat > tests/codegen/valid/control_flow/logical_and.src << 'EOF'
fn main() -> int {
    if ((5 > 3) && (10 > 5)) {
        return 1;
    }
    return 0;
}
EOF
    echo "1" > tests/codegen/expected/logical_and.expected
    echo -e "  ${GREEN}✓${NC} logical_and"
    
    # 22. Логическое ИЛИ
    cat > tests/codegen/valid/control_flow/logical_or.src << 'EOF'
fn main() -> int {
    if ((5 < 3) || (10 > 5)) {
        return 1;
    }
    return 0;
}
EOF
    echo "1" > tests/codegen/expected/logical_or.expected
    echo -e "  ${GREEN}✓${NC} logical_or"
    
    # 23. Логическое НЕ
    cat > tests/codegen/valid/control_flow/logical_not.src << 'EOF'
fn main() -> int {
    if (!(5 < 3)) {
        return 1;
    }
    return 0;
}
EOF
    echo "1" > tests/codegen/expected/logical_not.expected
    echo -e "  ${GREEN}✓${NC} logical_not"
    
    # 24. Комплексный тест - GCD
    cat > tests/codegen/valid/integration/gcd.src << 'EOF'
fn gcd(a int, b int) -> int {
    while (b != 0) {
        int temp = b;
        b = a % b;
        a = temp;
    }
    return a;
}

fn main() -> int {
    return gcd(48, 18);
}
EOF
    echo "6" > tests/codegen/expected/gcd.expected
    echo -e "  ${GREEN}✓${NC} gcd"
    
    # 25. Проверка на простоту
    cat > tests/codegen/valid/integration/is_prime.src << 'EOF'
fn is_prime(n int) -> int {
    if (n <= 1) {
        return 0;
    }
    int i = 2;
    while (i * i <= n) {
        if (n % i == 0) {
            return 0;
        }
        i = i + 1;
    }
    return 1;
}

fn main() -> int {
    return is_prime(17);
}
EOF
    echo "1" > tests/codegen/expected/is_prime.expected
    echo -e "  ${GREEN}✓${NC} is_prime"
    
    echo ""
    echo -e "${GREEN}Сгенерировано 25 тестовых программ${NC}"
}

# ============================================================================
# Запуск тестов
# ============================================================================

run_codegen_test() {
    local src_file=$1
    local name=$(basename "$src_file" .src)
    local expected_file="tests/codegen/expected/${name}.expected"
    local asm_file="$TEMP_DIR/${name}.asm"
    local obj_file="$TEMP_DIR/${name}.o"
    local exec_file="$TEMP_DIR/${name}"
    local output_file="$TEMP_DIR/${name}.out"
    
    # Компиляция в ассемблер
    ./bin/compiler compile --input "$src_file" --output "$asm_file" > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        echo -e "    ${RED}FAILED (compilation error)${NC}"
        echo -e "    ${YELLOW}Source: $src_file${NC}"
        ((failed_tests++))
        failed_tests_list+=("codegen/$name")
        ((total_tests++))
        return
    fi
    
    # Ассемблирование
    nasm -f elf64 -o "$obj_file" "$asm_file" 2>&1
    if [ $? -ne 0 ]; then
        echo -e "    ${RED}FAILED (assembly error)${NC}"
        echo -e "    ${YELLOW}ASM file: $asm_file${NC}"
        ((failed_tests++))
        failed_tests_list+=("codegen/$name")
        ((total_tests++))
        return
    fi
    
    # Линковка
    ld -o "$exec_file" runtime.o "$obj_file" 2>&1
    if [ $? -ne 0 ]; then
        echo -e "    ${RED}FAILED (linker error)${NC}"
        echo -e "    ${YELLOW}Object file: $obj_file${NC}"
        ((failed_tests++))
        failed_tests_list+=("codegen/$name")
        ((total_tests++))
        return
    fi
    
    # Исполнение
    timeout 5 "$exec_file" > "$output_file" 2>&1
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        echo -e "    ${RED}FAILED (runtime error, exit code: $exit_code)${NC}"
        ((failed_tests++))
        failed_tests_list+=("codegen/$name")
        ((total_tests++))
        return
    fi
    
    # Проверка результата
    if [ -f "$expected_file" ]; then
        local expected=$(cat "$expected_file")
        local actual=$exit_code
        
        if [ "$expected" = "$actual" ]; then
            echo -e "    ${GREEN}PASSED${NC} (result: $actual)"
            ((passed_tests++))
        else
            echo -e "    ${RED}FAILED (expected: $expected, got: $actual)${NC}"
            ((failed_tests++))
            failed_tests_list+=("codegen/$name")
        fi
    else
        echo -e "    ${YELLOW}SKIPPED (no expected file)${NC}"
    fi
    
    ((total_tests++))
}

# ============================================================================
# Главная функция
# ============================================================================

main() {
    # Генерация тестовых программ
    generate_test_programs
    
    # Сборка компилятора
    print_header "СБОРКА КОМПИЛЯТОРА"
    make build > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        echo -e "${RED}Ошибка сборки компилятора${NC}"
        exit 1
    fi
    echo -e "${GREEN}Компилятор собран${NC}"
    
    # Сборка рантайм библиотеки
    print_header "СБОРКА РАНТАЙМ БИБЛИОТЕКИ"
    nasm -f elf64 -o runtime.o src/runtime/runtime.asm
    if [ $? -ne 0 ]; then
        echo -e "${RED}Ошибка сборки рантайм библиотеки${NC}"
        exit 1
    fi
    echo -e "${GREEN}Рантайм библиотека собрана${NC}"
    
    # Запуск тестов
    print_header "ЗАПУСК ТЕСТОВ КОДОГЕНЕРАЦИИ"
    
    # Арифметические операции
    if [ -d "tests/codegen/valid/arithmetic_ops" ]; then
        echo ""
        echo -e "${CYAN}Арифметические операции:${NC}"
        for src_file in tests/codegen/valid/arithmetic_ops/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "    $name "
                run_codegen_test "$src_file"
            fi
        done
    fi
    
    # Управляющие конструкции
    if [ -d "tests/codegen/valid/control_flow" ]; then
        echo ""
        echo -e "${CYAN}Управляющие конструкции:${NC}"
        for src_file in tests/codegen/valid/control_flow/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "    $name "
                run_codegen_test "$src_file"
            fi
        done
    fi
    
    # Вызовы функций
    if [ -d "tests/codegen/valid/function_calls" ]; then
        echo ""
        echo -e "${CYAN}Вызовы функций:${NC}"
        for src_file in tests/codegen/valid/function_calls/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "    $name "
                run_codegen_test "$src_file"
            fi
        done
    fi
    
    # Интеграционные тесты
    if [ -d "tests/codegen/valid/integration" ]; then
        echo ""
        echo -e "${CYAN}Интеграционные тесты:${NC}"
        for src_file in tests/codegen/valid/integration/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "    $name "
                run_codegen_test "$src_file"
            fi
        done
    fi
    
    # Итоги
    print_header "ИТОГИ ТЕСТИРОВАНИЯ КОДОГЕНЕРАЦИИ"
    
    echo ""
    echo -e "Всего тестов: ${BLUE}$total_tests${NC}"
    echo -e "Пройдено:     ${GREEN}$passed_tests${NC}"
    echo -e "Провалено:    ${RED}$failed_tests${NC}"
    
    if [ ${#failed_tests_list[@]} -gt 0 ]; then
        echo ""
        echo -e "${RED}Не пройденные тесты:${NC}"
        for test in "${failed_tests_list[@]}"; do
            echo -e "${RED}  • $test${NC}"
        done
        echo ""
        echo -e "${YELLOW}Подсказка: Для отладки используйте:${NC}"
        echo "  ./bin/compiler compile --input tests/codegen/valid/arithmetic_ops/simple_add.src"
        exit 1
    fi
    
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}Все тесты кодогенерации пройдены успешно!${NC}"
    echo -e "${GREEN}========================================${NC}"
    
    # Очистка временных файлов
    rm -f runtime.o
}

# Запуск
main "$@"