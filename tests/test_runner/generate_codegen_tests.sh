#!/bin/bash

# Скрипт для генерации тестовых программ и expected файлов для кодогенерации
# Запуск: bash generate_codegen_tests.sh

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_ROOT"

echo -e "${YELLOW}Генерация тестов кодогенерации...${NC}"
echo ""

# Создаём директории
mkdir -p tests/codegen/valid/{arithmetic_ops,control_flow,function_calls,integration}

# ============================================================================
# АРИФМЕТИЧЕСКИЕ ОПЕРАЦИИ
# ============================================================================

echo -e "${YELLOW}Арифметические операции:${NC}"

# 1. simple_add - 2 + 3 = 5
cat > tests/codegen/valid/arithmetic_ops/simple_add.src << 'EOF'
fn main() -> int {
    return 2 + 3;
}
EOF
echo "5" > tests/codegen/valid/arithmetic_ops/simple_add.expected
echo -e "  ${GREEN}✓${NC} simple_add (5)"

# 2. subtract - 10 - 3 = 7
cat > tests/codegen/valid/arithmetic_ops/subtract.src << 'EOF'
fn main() -> int {
    return 10 - 3;
}
EOF
echo "7" > tests/codegen/valid/arithmetic_ops/subtract.expected
echo -e "  ${GREEN}✓${NC} subtract (7)"

# 3. multiply - 6 * 7 = 42
cat > tests/codegen/valid/arithmetic_ops/multiply.src << 'EOF'
fn main() -> int {
    return 6 * 7;
}
EOF
echo "42" > tests/codegen/valid/arithmetic_ops/multiply.expected
echo -e "  ${GREEN}✓${NC} multiply (42)"

# 4. divide - 100 / 5 = 20
cat > tests/codegen/valid/arithmetic_ops/divide.src << 'EOF'
fn main() -> int {
    return 100 / 5;
}
EOF
echo "20" > tests/codegen/valid/arithmetic_ops/divide.expected
echo -e "  ${GREEN}✓${NC} divide (20)"

# 5. modulo - 17 % 5 = 2
cat > tests/codegen/valid/arithmetic_ops/modulo.src << 'EOF'
fn main() -> int {
    return 17 % 5;
}
EOF
echo "2" > tests/codegen/valid/arithmetic_ops/modulo.expected
echo -e "  ${GREEN}✓${NC} modulo (2)"

# 6. complex_expr - 2 * 3 + 4 * 5 = 26
cat > tests/codegen/valid/arithmetic_ops/complex_expr.src << 'EOF'
fn main() -> int {
    return 2 * 3 + 4 * 5;
}
EOF
echo "26" > tests/codegen/valid/arithmetic_ops/complex_expr.expected
echo -e "  ${GREEN}✓${NC} complex_expr (26)"

# 7. precedence - (2 + 3) * (4 + 5) = 45
cat > tests/codegen/valid/arithmetic_ops/precedence.src << 'EOF'
fn main() -> int {
    return (2 + 3) * (4 + 5);
}
EOF
echo "45" > tests/codegen/valid/arithmetic_ops/precedence.expected
echo -e "  ${GREEN}✓${NC} precedence (45)"

# 8. negation
cat > tests/codegen/valid/arithmetic_ops/negation.src << 'EOF'
fn main() -> int {
    int x = 42;
    return -x;
}
EOF
echo "214" > tests/codegen/valid/arithmetic_ops/negation.expected
echo -e "  ${GREEN}✓${NC} negation (214)"

# 9. combined arithmetic
cat > tests/codegen/valid/arithmetic_ops/combined.src << 'EOF'
fn main() -> int {
    int a = 10;
    int b = 3;
    return a + b * 2;
}
EOF
echo "16" > tests/codegen/valid/arithmetic_ops/combined.expected
echo -e "  ${GREEN}✓${NC} combined (16)"

# ============================================================================
# УПРАВЛЯЮЩИЕ КОНСТРУКЦИИ
# ============================================================================

echo ""
echo -e "${YELLOW}Управляющие конструкции:${NC}"

# 10. equals
cat > tests/codegen/valid/control_flow/equals.src << 'EOF'
fn main() -> int {
    if (5 == 5) {
        return 1;
    }
    return 0;
}
EOF
echo "1" > tests/codegen/valid/control_flow/equals.expected
echo -e "  ${GREEN}✓${NC} equals (1)"

# 11. not_equals
cat > tests/codegen/valid/control_flow/not_equals.src << 'EOF'
fn main() -> int {
    if (5 != 3) {
        return 1;
    }
    return 0;
}
EOF
echo "1" > tests/codegen/valid/control_flow/not_equals.expected
echo -e "  ${GREEN}✓${NC} not_equals (1)"

# 12. less_than
cat > tests/codegen/valid/control_flow/less_than.src << 'EOF'
fn main() -> int {
    if (3 < 5) {
        return 1;
    }
    return 0;
}
EOF
echo "1" > tests/codegen/valid/control_flow/less_than.expected
echo -e "  ${GREEN}✓${NC} less_than (1)"

# 13. greater_than
cat > tests/codegen/valid/control_flow/greater_than.src << 'EOF'
fn main() -> int {
    if (10 > 5) {
        return 1;
    }
    return 0;
}
EOF
echo "1" > tests/codegen/valid/control_flow/greater_than.expected
echo -e "  ${GREEN}✓${NC} greater_than (1)"

# 14. less_equal
cat > tests/codegen/valid/control_flow/less_equal.src << 'EOF'
fn main() -> int {
    if (5 <= 5) {
        return 1;
    }
    return 0;
}
EOF
echo "1" > tests/codegen/valid/control_flow/less_equal.expected
echo -e "  ${GREEN}✓${NC} less_equal (1)"

# 15. greater_equal
cat > tests/codegen/valid/control_flow/greater_equal.src << 'EOF'
fn main() -> int {
    if (10 >= 5) {
        return 1;
    }
    return 0;
}
EOF
echo "1" > tests/codegen/valid/control_flow/greater_equal.expected
echo -e "  ${GREEN}✓${NC} greater_equal (1)"

# 16. if_else
cat > tests/codegen/valid/control_flow/if_else.src << 'EOF'
fn main() -> int {
    if (10 > 5) {
        return 1;
    } else {
        return 0;
    }
}
EOF
echo "1" > tests/codegen/valid/control_flow/if_else.expected
echo -e "  ${GREEN}✓${NC} if_else (1)"

# 17. if_else_false
cat > tests/codegen/valid/control_flow/if_else_false.src << 'EOF'
fn main() -> int {
    if (10 < 5) {
        return 1;
    } else {
        return 0;
    }
}
EOF
echo "0" > tests/codegen/valid/control_flow/if_else_false.expected
echo -e "  ${GREEN}✓${NC} if_else_false (0)"

# 18. nested_if
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
echo "1" > tests/codegen/valid/control_flow/nested_if.expected
echo -e "  ${GREEN}✓${NC} nested_if (1)"

# 19. logical_and
cat > tests/codegen/valid/control_flow/logical_and.src << 'EOF'
fn main() -> int {
    if ((5 > 3) && (10 > 5)) {
        return 1;
    }
    return 0;
}
EOF
echo "1" > tests/codegen/valid/control_flow/logical_and.expected
echo -e "  ${GREEN}✓${NC} logical_and (1)"

# 20. logical_or
cat > tests/codegen/valid/control_flow/logical_or.src << 'EOF'
fn main() -> int {
    if ((5 < 3) || (10 > 5)) {
        return 1;
    }
    return 0;
}
EOF
echo "1" > tests/codegen/valid/control_flow/logical_or.expected
echo -e "  ${GREEN}✓${NC} logical_or (1)"

# 21. logical_not
cat > tests/codegen/valid/control_flow/logical_not.src << 'EOF'
fn main() -> int {
    if (!(5 < 3)) {
        return 1;
    }
    return 0;
}
EOF
echo "1" > tests/codegen/valid/control_flow/logical_not.expected
echo -e "  ${GREEN}✓${NC} logical_not (1)"

# 22. while_loop - сумма 1..10 = 55
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
echo "55" > tests/codegen/valid/control_flow/while_loop.expected
echo -e "  ${GREEN}✓${NC} while_loop (55)"

# 23. while_simple
cat > tests/codegen/valid/control_flow/while_simple.src << 'EOF'
fn main() -> int {
    int i = 0;
    while (i < 5) {
        i = i + 1;
    }
    return i;
}
EOF
echo "5" > tests/codegen/valid/control_flow/while_simple.expected
echo -e "  ${GREEN}✓${NC} while_simple (5)"

# 24. for_loop - сумма 1..10 = 55
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
echo "55" > tests/codegen/valid/control_flow/for_loop.expected
echo -e "  ${GREEN}✓${NC} for_loop (55)"

# 25. for_simple
cat > tests/codegen/valid/control_flow/for_simple.src << 'EOF'
fn main() -> int {
    int i = 0;
    for (i = 0; i < 5; i = i + 1) {
    }
    return i;
}
EOF
echo "5" > tests/codegen/valid/control_flow/for_simple.expected
echo -e "  ${GREEN}✓${NC} for_simple (5)"

# ============================================================================
# ВЫЗОВЫ ФУНКЦИЙ
# ============================================================================

echo ""
echo -e "${YELLOW}Вызовы функций:${NC}"

# 26. simple_func
cat > tests/codegen/valid/function_calls/simple_func.src << 'EOF'
fn get_number() -> int {
    return 42;
}

fn main() -> int {
    return get_number();
}
EOF
echo "42" > tests/codegen/valid/function_calls/simple_func.expected
echo -e "  ${GREEN}✓${NC} simple_func (42)"

# 27. function_params
cat > tests/codegen/valid/function_calls/function_params.src << 'EOF'
fn add(a int, b int) -> int {
    return a + b;
}

fn main() -> int {
    return add(10, 20);
}
EOF
echo "30" > tests/codegen/valid/function_calls/function_params.expected
echo -e "  ${GREEN}✓${NC} function_params (30)"

# 28. multiple_params
cat > tests/codegen/valid/function_calls/multiple_params.src << 'EOF'
fn multiply(a int, b int, c int) -> int {
    return a * b * c;
}

fn main() -> int {
    return multiply(2, 3, 4);
}
EOF
echo "24" > tests/codegen/valid/function_calls/multiple_params.expected
echo -e "  ${GREEN}✓${NC} multiple_params (24)"

# 29. recursive - factorial(5) = 120
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
echo "120" > tests/codegen/valid/function_calls/recursive.expected
echo -e "  ${GREEN}✓${NC} recursive (120)"

# 30. fibonacci - fib(10) = 55
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
echo "55" > tests/codegen/valid/function_calls/fibonacci.expected
echo -e "  ${GREEN}✓${NC} fibonacci (55)"

# 31. void_func
cat > tests/codegen/valid/function_calls/void_func.src << 'EOF'
fn set_value() -> int {
    return 99;
}

fn main() -> int {
    return set_value();
}
EOF
echo "99" > tests/codegen/valid/function_calls/void_func.expected
echo -e "  ${GREEN}✓${NC} void_func (99)"

# 32. nested_call
cat > tests/codegen/valid/function_calls/nested_call.src << 'EOF'
fn add(a int, b int) -> int {
    return a + b;
}

fn double(x int) -> int {
    return add(x, x);
}

fn main() -> int {
    return double(21);
}
EOF
echo "42" > tests/codegen/valid/function_calls/nested_call.expected
echo -e "  ${GREEN}✓${NC} nested_call (42)"

# ============================================================================
# ИНТЕГРАЦИОННЫЕ ТЕСТЫ
# ============================================================================

echo ""
echo -e "${YELLOW}Интеграционные тесты:${NC}"

# 33. gcd - НОД(48, 18) = 6
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
echo "6" > tests/codegen/valid/integration/gcd.expected
echo -e "  ${GREEN}✓${NC} gcd (6)"

# 34. is_prime - 17 простое
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
echo "1" > tests/codegen/valid/integration/is_prime.expected
echo -e "  ${GREEN}✓${NC} is_prime (1)"

# 35. sum_of_squares
cat > tests/codegen/valid/integration/sum_of_squares.src << 'EOF'
fn square(x int) -> int {
    return x * x;
}

fn sum_squares(n int) -> int {
    int sum = 0;
    int i = 1;
    while (i <= n) {
        sum = sum + square(i);
        i = i + 1;
    }
    return sum;
}

fn main() -> int {
    return sum_squares(5);
}
EOF
echo "55" > tests/codegen/valid/integration/sum_of_squares.expected
echo -e "  ${GREEN}✓${NC} sum_of_squares (55)"

# 36. power
cat > tests/codegen/valid/integration/power.src << 'EOF'
fn power(base int, exp int) -> int {
    int result = 1;
    int i = 0;
    while (i < exp) {
        result = result * base;
        i = i + 1;
    }
    return result;
}

fn main() -> int {
    return power(2, 8);
}
EOF
echo "256" > tests/codegen/valid/integration/power.expected
echo -e "  ${GREEN}✓${NC} power (256)"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Сгенерировано 36 тестовых программ с expected файлами${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "Для запуска тестов выполните:"
echo -e "  ${YELLOW}make test${NC}"
echo ""
echo -e "Для запуска только тестов кодогенерации:"
echo -e "  ${YELLOW}./tests/test_runner/run_tests.sh codegen${NC}"