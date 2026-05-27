#!/bin/bash
set -e

BASE_DIR="tests/control_flow/valid"
EXPECTED_DIR="tests/control_flow/expected"

mkdir -p "$BASE_DIR/conditionals"
mkdir -p "$BASE_DIR/loops"
mkdir -p "$BASE_DIR/logical_ops"
mkdir -p "$BASE_DIR/complex_expressions"
mkdir -p "$EXPECTED_DIR/conditionals"
mkdir -p "$EXPECTED_DIR/loops"
mkdir -p "$EXPECTED_DIR/logical_ops"
mkdir -p "$EXPECTED_DIR/complex_expressions"

# -------------------------------------------------------------------
# conditionals
# -------------------------------------------------------------------

cat > "$BASE_DIR/conditionals/simple_if_true.src" << 'EOF'
fn main() -> int {
    int x = 1;
    if (x > 0) { return 1; }
    return 0;
}
EOF
echo "1" > "$EXPECTED_DIR/conditionals/simple_if_true.expected"

cat > "$BASE_DIR/conditionals/simple_if_false.src" << 'EOF'
fn main() -> int {
    int x = 0;
    if (x > 0) { return 1; }
    return 0;
}
EOF
echo "0" > "$EXPECTED_DIR/conditionals/simple_if_false.expected"

cat > "$BASE_DIR/conditionals/if_else_true.src" << 'EOF'
fn main() -> int {
    int x = 1;
    if (x > 0) { return 10; } else { return 20; }
}
EOF
echo "10" > "$EXPECTED_DIR/conditionals/if_else_true.expected"

cat > "$BASE_DIR/conditionals/if_else_false.src" << 'EOF'
fn main() -> int {
    int x = 0;
    if (x > 0) { return 10; } else { return 20; }
}
EOF
echo "20" > "$EXPECTED_DIR/conditionals/if_else_false.expected"

cat > "$BASE_DIR/conditionals/nested_if.src" << 'EOF'
fn main() -> int {
    int x = 15;
    if (x > 10) {
        if (x < 20) { return 1; } else { return 2; }
    } else { return 3; }
}
EOF
echo "1" > "$EXPECTED_DIR/conditionals/nested_if.expected"

# -------------------------------------------------------------------
# loops
# -------------------------------------------------------------------

cat > "$BASE_DIR/loops/while_zero.src" << 'EOF'
fn main() -> int {
    int i = 0;
    int flag = 0;
    while (flag) { i = i + 1; }
    return 42;
}
EOF
echo "42" > "$EXPECTED_DIR/loops/while_zero.expected"

cat > "$BASE_DIR/loops/while_sum.src" << 'EOF'
fn main() -> int {
    int i = 0;
    int sum = 0;
    while (i < 10) { sum = sum + i; i = i + 1; }
    return sum;
}
EOF
echo "45" > "$EXPECTED_DIR/loops/while_sum.expected"

cat > "$BASE_DIR/loops/for_sum.src" << 'EOF'
fn main() -> int {
    int result = 0;
    int i = 0;
    for (i = 0; i < 5; i = i + 1) { result = result + i; }
    return result;
}
EOF
echo "10" > "$EXPECTED_DIR/loops/for_sum.expected"

cat > "$BASE_DIR/loops/nested_loop.src" << 'EOF'
fn main() -> int {
    int i = 0;
    int count = 0;
    while (i < 3) {
        int j = 0;
        while (j < 2) { count = count + 1; j = j + 1; }
        i = i + 1;
    }
    return count;
}
EOF
echo "6" > "$EXPECTED_DIR/loops/nested_loop.expected"

# -------------------------------------------------------------------
# logical_ops
# -------------------------------------------------------------------

cat > "$BASE_DIR/logical_ops/short_circuit_and.src" << 'EOF'
fn main() -> int {
    int a = 0;
    int b = 5;
    if (a != 0 && b / a > 2) { return 1; } else { return 2; }
}
EOF
echo "2" > "$EXPECTED_DIR/logical_ops/short_circuit_and.expected"

cat > "$BASE_DIR/logical_ops/short_circuit_or.src" << 'EOF'
fn main() -> int {
    int a = 10;
    if (a > 0 || a > 100) { return 1; } else { return 2; }
}
EOF
echo "1" > "$EXPECTED_DIR/logical_ops/short_circuit_or.expected"

cat > "$BASE_DIR/logical_ops/truth_tables.src" << 'EOF'
fn main() -> int {
    int result = 0;
    int t = 1;
    int f = 0;
    if (t == 1 && t == 1) { result = result + 1; }
    if (t == 1 && f == 1) { result = result + 2; }
    if (f == 1 && t == 1) { result = result + 4; }
    if (f == 1 && f == 1) { result = result + 8; }
    if (t == 1 || t == 1) { result = result + 16; }
    if (t == 1 || f == 1) { result = result + 32; }
    if (f == 1 || t == 1) { result = result + 64; }
    return result;
}
EOF
echo "113" > "$EXPECTED_DIR/logical_ops/truth_tables.expected"

cat > "$BASE_DIR/logical_ops/mixed_logical.src" << 'EOF'
fn main() -> int {
    int a = 1;
    int b = 0;
    if (a == 1 && b == 0) { return 1; }
    return 0;
}
EOF
echo "1" > "$EXPECTED_DIR/logical_ops/mixed_logical.expected"

# -------------------------------------------------------------------
# complex_expressions
# -------------------------------------------------------------------

cat > "$BASE_DIR/complex_expressions/precedence.src" << 'EOF'
fn main() -> int {
    return 2 + 3 * 4;
}
EOF
echo "14" > "$EXPECTED_DIR/complex_expressions/precedence.expected"

cat > "$BASE_DIR/complex_expressions/associativity.src" << 'EOF'
fn main() -> int {
    return 10 - 5 - 2;
}
EOF
echo "3" > "$EXPECTED_DIR/complex_expressions/associativity.expected"

cat > "$BASE_DIR/complex_expressions/mixed_type.src" << 'EOF'
fn main() -> int {
    int x = 5;
    int y = 2;
    int z = x * y;
    return z + 2;
}
EOF
echo "12" > "$EXPECTED_DIR/complex_expressions/mixed_type.expected"

echo "Тесты сгенерированы в $BASE_DIR"
echo "Expected файлы в $EXPECTED_DIR"
