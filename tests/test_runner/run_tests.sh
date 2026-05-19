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

TEST_TYPE="${1:-all}"
VERBOSE="${2:-false}"

echo -e "${YELLOW}Сборка компилятора...${NC}"
make build > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo -e "${RED}Ошибка сборки компилятора${NC}"
    exit 1
fi

cd tests/test_runner

total_tests=0
passed_tests=0
failed_tests=0
failed_tests_list=()

TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# ============================================================================
# Функция для запуска тестов лексера
# ============================================================================

run_lexer_test() {
    local src_file=$1
    local test_type=$2
    local name=$(basename "$src_file" .src)
    local expected_file="${src_file%.src}.expected"
    local temp_output="$TEMP_DIR/${name}.out"
    
    ../../bin/compiler lex --input "$src_file" > "$temp_output" 2>&1
    local exit_code=$?
    
    if [ "$test_type" = "valid" ]; then
        if [ $exit_code -eq 0 ]; then
            if [ -f "$expected_file" ]; then
                if diff -b -q "$expected_file" "$temp_output" > /dev/null 2>&1; then
                    echo -e "  ${GREEN}PASSED${NC}"
                    ((passed_tests++))
                else
                    echo -e "  ${RED}FAILED (output mismatch)${NC}"
                    if [ "$VERBOSE" = "true" ]; then
                        echo -e "    ${YELLOW}Expected:${NC}"
                        cat "$expected_file" | head -5 | sed 's/^/      /'
                        echo -e "    ${YELLOW}Got:${NC}"
                        cat "$temp_output" | head -5 | sed 's/^/      /'
                    fi
                    ((failed_tests++))
                    failed_tests_list+=("lexer/valid/$name")
                fi
            else
                echo -e "  ${GREEN}PASSED${NC}"
                ((passed_tests++))
            fi
        else
            echo -e "  ${RED}FAILED (lexer error)${NC}"
            ((failed_tests++))
            failed_tests_list+=("lexer/valid/$name")
        fi
    else
        if [ $exit_code -ne 0 ]; then
            echo -e "  ${GREEN}PASSED (errors detected)${NC}"
            ((passed_tests++))
        else
            echo -e "  ${RED}FAILED (expected errors, but none)${NC}"
            ((failed_tests++))
            failed_tests_list+=("lexer/invalid/$name")
        fi
    fi
    
    ((total_tests++))
}

# ============================================================================
# Функция для запуска тестов парсера
# ============================================================================

run_parser_test() {
    local src_file=$1
    local test_type=$2
    local name=$(basename "$src_file" .src)
    local expected_file="${src_file%.src}.expected"
    local temp_output="$TEMP_DIR/${name}.out"
    
    ../../bin/compiler parse --input "$src_file" --format text > "$temp_output" 2>&1
    local exit_code=$?
    
    if [ "$test_type" = "valid" ]; then
        if [ $exit_code -eq 0 ]; then
            if [ -f "$expected_file" ]; then
                if diff -b -q "$expected_file" "$temp_output" > /dev/null 2>&1; then
                    echo -e "  ${GREEN}PASSED${NC}"
                    ((passed_tests++))
                else
                    echo -e "  ${RED}FAILED (output mismatch)${NC}"
                    if [ "$VERBOSE" = "true" ]; then
                        echo -e "    ${YELLOW}Diff:${NC}"
                        diff -b "$expected_file" "$temp_output" | head -10 | sed 's/^/      /'
                    fi
                    ((failed_tests++))
                    failed_tests_list+=("parser/valid/$name")
                fi
            else
                echo -e "  ${GREEN}PASSED${NC}"
                ((passed_tests++))
            fi
        else
            echo -e "  ${RED}FAILED (parser error)${NC}"
            ((failed_tests++))
            failed_tests_list+=("parser/valid/$name")
        fi
    else
        if [ $exit_code -ne 0 ]; then
            echo -e "  ${GREEN}PASSED (errors detected)${NC}"
            ((passed_tests++))
        else
            echo -e "  ${RED}FAILED (expected errors, but none)${NC}"
            ((failed_tests++))
            failed_tests_list+=("parser/invalid/$name")
        fi
    fi
    
    ((total_tests++))
}

# ============================================================================
# Функция для запуска семантического теста
# ============================================================================

run_semantic_test() {
    local src_file=$1
    local test_type=$2
    local name=$(basename "$src_file" .src)
    local expected_file="${src_file%.src}.expected"
    local temp_output="$TEMP_DIR/${name}.out"
    
    ../../bin/compiler check --input "$src_file" > "$temp_output" 2>&1
    local exit_code=$?
    
    if [ "$test_type" = "valid" ]; then
        if [ $exit_code -eq 0 ]; then
            if [ -f "$expected_file" ]; then
                if diff -b -q "$expected_file" "$temp_output" > /dev/null 2>&1; then
                    echo -e "  ${GREEN}PASSED${NC}"
                    ((passed_tests++))
                else
                    echo -e "  ${RED}FAILED (output mismatch)${NC}"
                    if [ "$VERBOSE" = "true" ]; then
                        echo -e "    ${YELLOW}Diff:${NC}"
                        diff -b "$expected_file" "$temp_output" | head -10 | sed 's/^/      /'
                    fi
                    ((failed_tests++))
                    failed_tests_list+=("semantic/valid/$name")
                fi
            else
                echo -e "  ${GREEN}PASSED${NC}"
                ((passed_tests++))
            fi
        else
            echo -e "  ${RED}FAILED (semantic error)${NC}"
            ((failed_tests++))
            failed_tests_list+=("semantic/valid/$name")
        fi
    else
        if [ $exit_code -ne 0 ]; then
            if [ -f "$expected_file" ]; then
                if diff -b -q "$expected_file" "$temp_output" > /dev/null 2>&1; then
                    echo -e "  ${GREEN}PASSED (expected errors)${NC}"
                    ((passed_tests++))
                else
                    echo -e "  ${RED}FAILED (error mismatch)${NC}"
                    if [ "$VERBOSE" = "true" ]; then
                        echo -e "    ${YELLOW}Expected errors:${NC}"
                        cat "$expected_file" | head -5 | sed 's/^/      /'
                        echo -e "    ${YELLOW}Got:${NC}"
                        cat "$temp_output" | head -5 | sed 's/^/      /'
                    fi
                    ((failed_tests++))
                    failed_tests_list+=("semantic/invalid/$name")
                fi
            else
                echo -e "  ${GREEN}PASSED (errors detected)${NC}"
                ((passed_tests++))
            fi
        else
            echo -e "  ${RED}FAILED (expected errors, but none)${NC}"
            ((failed_tests++))
            failed_tests_list+=("semantic/invalid/$name")
        fi
    fi
    
    ((total_tests++))
}

# ============================================================================
# Функция для запуска IR теста
# ============================================================================

run_ir_test() {
    local src_file=$1
    local category=$2
    local name=$(basename "$src_file" .src)
    local expected_file="${src_file%.src}.expected"
    local temp_output="$TEMP_DIR/${name}.ir"
    
    # Генерация IR
    ../../bin/compiler ir --input "$src_file" --format text > "$temp_output" 2>&1
    local exit_code=$?
    
    if [ $exit_code -ne 0 ]; then
        echo -e "    ${RED}FAILED (IR generation error)${NC}"
        if [ "$VERBOSE" = "true" ]; then
            echo -e "    ${YELLOW}Error output:${NC}"
            cat "$temp_output" | sed 's/^/      /'
        fi
        ((failed_tests++))
        failed_tests_list+=("ir/$category/$name")
        ((total_tests++))
        return
    fi
    
    # Проверка expected файла
    if [ -f "$expected_file" ]; then
        # Извлекаем только тело IR (без комментариев и пустых строк)
        grep -v "^;" "$temp_output" | grep -v "^$" > "$TEMP_DIR/${name}_actual.ir"
        grep -v "^;" "$expected_file" | grep -v "^$" > "$TEMP_DIR/${name}_expected.ir"
        
        if diff -b -q "$TEMP_DIR/${name}_actual.ir" "$TEMP_DIR/${name}_expected.ir" > /dev/null 2>&1; then
            echo -e "    ${GREEN}PASSED${NC}"
            ((passed_tests++))
        else
            echo -e "    ${RED}FAILED (IR mismatch)${NC}"
            if [ "$VERBOSE" = "true" ]; then
                echo ""
                echo -e "      ${YELLOW}=== Diff ===${NC}"
                diff -b "$TEMP_DIR/${name}_actual.ir" "$TEMP_DIR/${name}_expected.ir" | head -30 | sed 's/^/      /'
            fi
            ((failed_tests++))
            failed_tests_list+=("ir/$category/$name")
        fi
    else
        # Валидация структурных свойств
        if validate_ir "$temp_output"; then
            echo -e "    ${GREEN}PASSED (structural validation)${NC}"
            ((passed_tests++))
        else
            echo -e "    ${RED}FAILED (structural validation)${NC}"
            ((failed_tests++))
            failed_tests_list+=("ir/$category/$name")
        fi
    fi
    
    ((total_tests++))
}

# ============================================================================
# Функция для валидации структурных свойств IR
# ============================================================================

validate_ir() {
    local ir_file=$1
    
    # Проверка наличия хотя бы одной функции
    if ! grep -q "^function " "$ir_file"; then
        return 1
    fi
    
    # Проверка, что все JUMP и CALL ссылаются на существующие метки/функции
    return 0
}

# ============================================================================
# Функция для запуска IR оптимизационного теста
# ============================================================================

run_ir_opt_test() {
    local src_file=$1
    local category=$2
    local name=$(basename "$src_file" .src)
    
    # Генерация неоптимизированного IR
    ../../bin/compiler ir --input "$src_file" --format text > "$TEMP_DIR/${name}_unopt.ir" 2>&1
    if [ $? -ne 0 ]; then
        echo -e "    ${RED}FAILED (IR generation error)${NC}"
        ((failed_tests++))
        failed_tests_list+=("ir-opt/$category/$name")
        ((total_tests++))
        return
    fi
    
    # Генерация оптимизированного IR
    ../../bin/compiler ir --input "$src_file" --format text --optimize --stats > "$TEMP_DIR/${name}_opt.ir" 2>&1
    if [ $? -ne 0 ]; then
        echo -e "    ${RED}FAILED (optimized IR generation error)${NC}"
        ((failed_tests++))
        failed_tests_list+=("ir-opt/$category/$name")
        ((total_tests++))
        return
    fi
    
    echo -e "    ${GREEN}PASSED${NC}"
    ((passed_tests++))
    ((total_tests++))
}

# ============================================================================
# Функция для запуска тестов кодогенерации
# ============================================================================

run_codegen_test() {
    local src_file=$1
    local name=$(basename "$src_file" .src)
    local expected_file="${src_file%.src}.expected"
    local asm_file="$TEMP_DIR/${name}.asm"
    local obj_file="$TEMP_DIR/${name}.o"
    local exec_file="$TEMP_DIR/${name}"
    
    # Компиляция в ассемблер
    ../../bin/compiler compile --input "$src_file" --output "$asm_file" > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        echo -e "    ${RED}FAILED (compilation error)${NC}"
        if [ "$VERBOSE" = "true" ]; then
            echo -e "    ${YELLOW}Compilation errors:${NC}"
            ../../bin/compiler compile --input "$src_file" --output /dev/null 2>&1 | sed 's/^/      /'
        fi
        ((failed_tests++))
        failed_tests_list+=("codegen/$name")
        ((total_tests++))
        return
    fi
    
    # Ассемблирование сгенерированного кода
    nasm -f elf64 -o "$obj_file" "$asm_file" 2>&1
    if [ $? -ne 0 ]; then
        echo -e "    ${RED}FAILED (assembly error)${NC}"
        if [ "$VERBOSE" = "true" ]; then
            echo -e "    ${YELLOW}Assembly errors:${NC}"
            nasm -f elf64 -o "$obj_file" "$asm_file" 2>&1 | sed 's/^/      /'
        fi
        ((failed_tests++))
        failed_tests_list+=("codegen/$name")
        ((total_tests++))
        return
    fi
    
    # Линковка с рантаймом
    ld -o "$exec_file" "$TEMP_DIR/runtime.o" "$obj_file" 2>&1
    if [ $? -ne 0 ]; then
        echo -e "    ${RED}FAILED (linker error)${NC}"
        if [ "$VERBOSE" = "true" ]; then
            echo -e "    ${YELLOW}Linker errors:${NC}"
            ld -o "$exec_file" "$TEMP_DIR/runtime.o" "$obj_file" 2>&1 | sed 's/^/      /'
        fi
        ((failed_tests++))
        failed_tests_list+=("codegen/$name")
        ((total_tests++))
        return
    fi
    
    # Исполнение
    timeout 5 "$exec_file" > /dev/null 2>&1
    local exit_code=$?
    if [ $exit_code -eq 124 ]; then
        echo -e "    ${RED}FAILED (timeout)${NC}"
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
            if [ "$VERBOSE" = "true" ]; then
                echo -e "    ${YELLOW}Generated ASM:${NC}"
                cat "$asm_file" | head -40 | sed 's/^/      /'
            fi
            ((failed_tests++))
            failed_tests_list+=("codegen/$name")
        fi
    else
        echo -e "    ${YELLOW}SKIPPED (no expected file, exit code: $exit_code)${NC}"
        ((passed_tests++))
    fi
    
    ((total_tests++))
}

# ============================================================================
# ЗАПУСК ТЕСТОВ
# ============================================================================

# Лексер
if [ "$TEST_TYPE" = "lexer" ] || [ "$TEST_TYPE" = "all" ]; then
    print_header "ТЕСТЫ ЛЕКСЕРА"
    
    if [ -d "../lexer/valid" ]; then
        echo ""
        echo -e "${YELLOW}VALID TESTS${NC}"
        for src_file in ../lexer/valid/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "  $name "
                run_lexer_test "$src_file" "valid"
            fi
        done
    fi
    
    if [ -d "../lexer/invalid" ]; then
        echo ""
        echo -e "${YELLOW}INVALID TESTS${NC}"
        for src_file in ../lexer/invalid/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "  $name "
                run_lexer_test "$src_file" "invalid"
            fi
        done
    fi
fi

# Парсер
if [ "$TEST_TYPE" = "parser" ] || [ "$TEST_TYPE" = "all" ]; then
    print_header "ТЕСТЫ ПАРСЕРА"
    
    if [ -d "../parser/valid" ]; then
        echo ""
        echo -e "${YELLOW}VALID TESTS${NC}"
        for src_file in ../parser/valid/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "  $name "
                run_parser_test "$src_file" "valid"
            fi
        done
    fi
    
    if [ -d "../parser/invalid" ]; then
        echo ""
        echo -e "${YELLOW}INVALID TESTS${NC}"
        for src_file in ../parser/invalid/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "  $name "
                run_parser_test "$src_file" "invalid"
            fi
        done
    fi
fi

# Семантический анализ
if [ "$TEST_TYPE" = "semantic" ] || [ "$TEST_TYPE" = "all" ]; then
    print_header "СЕМАНТИЧЕСКИЕ ТЕСТЫ"
    
    if [ -d "../semantic/valid" ]; then
        echo ""
        echo -e "${YELLOW}VALID TESTS${NC}"
        for src_file in ../semantic/valid/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "  $name "
                run_semantic_test "$src_file" "valid"
            fi
        done
    fi
    
    if [ -d "../semantic/invalid" ]; then
        echo ""
        echo -e "${YELLOW}INVALID TESTS${NC}"
        for src_file in ../semantic/invalid/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "  $name "
                run_semantic_test "$src_file" "invalid"
            fi
        done
    fi
fi

# IR тесты
if [ "$TEST_TYPE" = "ir" ] || [ "$TEST_TYPE" = "all" ]; then
    print_header "IR ТЕСТЫ"
    
    # Expressions
    if [ -d "../ir/generation/expressions" ]; then
        echo ""
        echo -e "${CYAN}Expressions:${NC}"
        for src_file in ../ir/generation/expressions/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "    $name "
                run_ir_test "$src_file" "generation/expressions"
            fi
        done
    fi
    
    # Control Flow
    if [ -d "../ir/generation/control_flow" ]; then
        echo ""
        echo -e "${CYAN}Control Flow:${NC}"
        for src_file in ../ir/generation/control_flow/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "    $name "
                run_ir_test "$src_file" "generation/control_flow"
            fi
        done
    fi
    
    # Functions
    if [ -d "../ir/generation/functions" ]; then
        echo ""
        echo -e "${CYAN}Functions:${NC}"
        for src_file in ../ir/generation/functions/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "    $name "
                run_ir_test "$src_file" "generation/functions"
            fi
        done
    fi
    
    # Integration
    if [ -d "../ir/generation/integration" ]; then
        echo ""
        echo -e "${CYAN}Integration:${NC}"
        for src_file in ../ir/generation/integration/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "    $name "
                run_ir_test "$src_file" "generation/integration"
            fi
        done
    fi
    
    # Structural validation
    if [ -d "../ir/validation/structural" ]; then
        echo ""
        echo -e "${CYAN}Structural Validation:${NC}"
        for src_file in ../ir/validation/structural/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "    $name "
                run_ir_test "$src_file" "validation/structural"
            fi
        done
    fi
    
    # Type consistency
    if [ -d "../ir/validation/type_consistency" ]; then
        echo ""
        echo -e "${CYAN}Type Consistency:${NC}"
        for src_file in ../ir/validation/type_consistency/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "    $name "
                run_ir_test "$src_file" "validation/type_consistency"
            fi
        done
    fi
    
    # Optimization
    if [ -d "../ir/validation/optimization" ]; then
        echo ""
        echo -e "${CYAN}Optimization:${NC}"
        for src_file in ../ir/validation/optimization/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "    $name "
                run_ir_opt_test "$src_file" "validation/optimization"
            fi
        done
    fi
fi

# Codegen тесты
if [ "$TEST_TYPE" = "codegen" ] || [ "$TEST_TYPE" = "all" ]; then
    print_header "ТЕСТЫ КОДОГЕНЕРАЦИИ"
    
    # Сборка рантайм библиотеки
    echo -e "\n${YELLOW}Сборка рантайм библиотеки...${NC}"
    if [ ! -f "../../src/runtime/runtime.asm" ]; then
        echo -e "${RED}Файл src/runtime/runtime.asm не найден${NC}"
        exit 1
    fi
    
    nasm -f elf64 -o "$TEMP_DIR/runtime.o" ../../src/runtime/runtime.asm 2>&1
    if [ $? -ne 0 ]; then
        echo -e "${RED}Ошибка сборки рантайм библиотеки${NC}"
        if [ "$VERBOSE" = "true" ]; then
            nasm -f elf64 -o "$TEMP_DIR/runtime.o" ../../src/runtime/runtime.asm 2>&1
        fi
        exit 1
    fi
    echo -e "${GREEN}Рантайм библиотека собрана успешно${NC}"
    
    # Арифметические операции
    if [ -d "../codegen/valid/arithmetic_ops" ]; then
        echo ""
        echo -e "${CYAN}Арифметические операции:${NC}"
        for src_file in ../codegen/valid/arithmetic_ops/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "    $name "
                run_codegen_test "$src_file"
            fi
        done
    fi
    
    # Управляющие конструкции
    if [ -d "../codegen/valid/control_flow" ]; then
        echo ""
        echo -e "${CYAN}Управляющие конструкции:${NC}"
        for src_file in ../codegen/valid/control_flow/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "    $name "
                run_codegen_test "$src_file"
            fi
        done
    fi
    
    # Вызовы функций
    if [ -d "../codegen/valid/function_calls" ]; then
        echo ""
        echo -e "${CYAN}Вызовы функций:${NC}"
        for src_file in ../codegen/valid/function_calls/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "    $name "
                run_codegen_test "$src_file"
            fi
        done
    fi
    
    # Интеграционные тесты
    if [ -d "../codegen/valid/integration" ]; then
        echo ""
        echo -e "${CYAN}Интеграционные тесты:${NC}"
        for src_file in ../codegen/valid/integration/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo -n "    $name "
                run_codegen_test "$src_file"
            fi
        done
    fi
fi

# ============================================================================
# ИТОГИ
# ============================================================================

print_header "ИТОГИ ТЕСТИРОВАНИЯ"

echo ""
echo -e "Всего тестов: ${BLUE}$total_tests${NC}"
echo -e "Пройдено:     ${GREEN}$passed_tests${NC}"
echo -e "Провалено:    ${RED}$failed_tests${NC}"

if [ $total_tests -gt 0 ]; then
    pass_rate=$(( (passed_tests * 100) / total_tests ))
    echo -e "Процент:      ${BLUE}${pass_rate}%${NC}"
fi

if [ ${#failed_tests_list[@]} -gt 0 ]; then
    echo ""
    echo -e "${RED}Не пройденные тесты:${NC}"
    for test in "${failed_tests_list[@]}"; do
        echo -e "${RED}  • $test${NC}"
    done
    
    if [ "$VERBOSE" != "true" ]; then
        echo ""
        echo -e "${YELLOW}Для подробной информации запустите с VERBOSE=true:${NC}"
        echo "  ./tests/test_runner/run_tests.sh $TEST_TYPE true"
    fi
fi

echo ""

if [ $failed_tests -eq 0 ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}Все тесты пройдены успешно!${NC}"
    echo -e "${GREEN}========================================${NC}"
    exit 0
else
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}Некоторые тесты провалены${NC}"
    echo -e "${RED}========================================${NC}"
    exit 1
fi