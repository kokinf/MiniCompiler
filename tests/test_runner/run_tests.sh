#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

echo -e "${YELLOW}Сборка компилятора...${NC}"
make build > /dev/null 2>&1

cd tests/test_runner

total_tests=0
passed_tests=0
failed_tests=0
failed_tests_list=()

TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# ============================================================================
# Функция для запуска семантического теста
# ============================================================================

run_semantic_test() {
    local src_file=$1
    local name=$(basename "$src_file" .src)
    local expected_file="${src_file%.src}.expected"
    local temp_output="$TEMP_DIR/${name}.out"
    
    ../../bin/compiler check --input "$src_file" > "$temp_output" 2>&1
    local exit_code=$?
    
    if [ "$2" = "valid" ]; then
        if [ $exit_code -eq 0 ]; then
            if [ -f "$expected_file" ]; then
                if diff -b -q "$expected_file" "$temp_output" > /dev/null 2>&1; then
                    echo -e "${GREEN}  PASSED${NC}"
                    ((passed_tests++))
                else
                    echo -e "${RED}  FAILED (output mismatch)${NC}"
                    ((failed_tests++))
                    failed_tests_list+=("$name")
                fi
            else
                echo -e "${GREEN}  PASSED${NC}"
                ((passed_tests++))
            fi
        else
            echo -e "${RED}  FAILED (semantic error)${NC}"
            ((failed_tests++))
            failed_tests_list+=("$name")
        fi
    else
        if [ $exit_code -ne 0 ]; then
            if [ -f "$expected_file" ]; then
                if diff -b -q "$expected_file" "$temp_output" > /dev/null 2>&1; then
                    echo -e "${GREEN}  PASSED (expected errors)${NC}"
                    ((passed_tests++))
                else
                    echo -e "${RED}  FAILED (error mismatch)${NC}"
                    ((failed_tests++))
                    failed_tests_list+=("$name")
                fi
            else
                echo -e "${GREEN}  PASSED (errors detected)${NC}"
                ((passed_tests++))
            fi
        else
            echo -e "${RED}  FAILED (expected errors, but none)${NC}"
            ((failed_tests++))
            failed_tests_list+=("$name")
        fi
    fi
    
    ((total_tests++))
}

# ============================================================================
# ЗАПУСК СЕМАНТИЧЕСКИХ ТЕСТОВ
# ============================================================================

if [ "$TEST_TYPE" = "semantic" ] || [ "$TEST_TYPE" = "all" ]; then
    
    print_header "СЕМАНТИЧЕСКИЕ ТЕСТЫ"
    
    # VALID TESTS
    echo ""
    echo -e "${YELLOW}VALID TESTS${NC}"
    if [ -d "../semantic/valid" ]; then
        for src_file in ../semantic/valid/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo ""
                echo -e "${YELLOW}Запуск: $name${NC}"
                run_semantic_test "$src_file" "valid"
            fi
        done
    else
        echo "Директория ../semantic/valid не найдена"
    fi
    
    # INVALID TESTS
    echo ""
    echo -e "${YELLOW}INVALID TESTS${NC}"
    if [ -d "../semantic/invalid" ]; then
        for src_file in ../semantic/invalid/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                echo ""
                echo -e "${YELLOW}Запуск: $name${NC}"
                run_semantic_test "$src_file" "invalid"
            fi
        done
    else
        echo "Директория ../semantic/invalid не найдена"
    fi
    
    # ============================================================================
    # ИТОГИ
    # ============================================================================
    
    print_header "ИТОГИ ТЕСТИРОВАНИЯ"
    
    echo ""
    echo "Total semantic tests: $total_tests"
    echo -e "${GREEN}PASSED:       $passed_tests${NC}"
    echo -e "${RED}FAILED:       $failed_tests${NC}"
    
    if [ ${#failed_tests_list[@]} -gt 0 ]; then
        echo ""
        echo -e "${RED}Не пройденные тесты:${NC}"
        for test in "${failed_tests_list[@]}"; do
            echo -e "${RED}  • $test${NC}"
        done
    fi
    
    echo ""
    
    if [ $failed_tests -eq 0 ]; then
        echo -e "${GREEN}All semantic tests PASSED!${NC}"
        exit 0
    else
        echo -e "${RED}Some semantic tests FAILED${NC}"
        exit 1
    fi
    
else
    echo "Использование: ./run_tests.sh [semantic|all]"
    exit 1
fi