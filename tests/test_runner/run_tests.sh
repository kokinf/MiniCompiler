#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Сборка компилятора...${NC}"
cd ../..
make build
if [ $? -ne 0 ]; then
    echo -e "${RED}Сборка не удалась${NC}"
    exit 1
fi

cd tests/test_runner

total_tests=0
passed_tests=0
failed_tests=0
failed_tests_list=()

TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Функция для запуска лексер теста
run_lexer_test() {
    local test_type=$1
    local test_name=$2
    local src_file="../lexer/$test_type/$test_name.src"
    local expected_file="../lexer/$test_type/$test_name.expected"
    local temp_output="$TEMP_DIR/${test_name}.out"
    
    echo ""
    echo -e "${YELLOW}Запуск лексер теста: $test_type/$test_name${NC}"
    
    # Проверка существования исходного файла
    if [ ! -f "$src_file" ]; then
        echo -e "${RED}  Исходный файл не найден: $src_file${NC}"
        ((failed_tests++))
        failed_tests_list+=("lexer/$test_type/$test_name (отсутствует исходный файл)")
        ((total_tests++))
        return
    fi
    
    # Запуск компилятора и захват вывода
    ../../bin/compiler lex --input "$src_file" > "$temp_output" 2>&1
    
    # Проверка существования файла с ожидаемым выводом
    if [ ! -f "$expected_file" ]; then
        echo -e "${RED}  Предупреждение: Файл с ожидаемым выводом не найден, показываю фактический вывод:${NC}"
        if [ -f "$temp_output" ]; then
            cat "$temp_output"
        else
            echo "  Вывод не сгенерирован"
        fi
        ((failed_tests++))
        failed_tests_list+=("lexer/$test_type/$test_name (отсутствует ожидаемый вывод)")
        ((total_tests++))
        return
    fi
    
    # Сравнение с ожидаемым выводом
    if diff -u "$expected_file" "$temp_output" > "$TEMP_DIR/${test_name}.diff"; then
        echo -e "${GREEN}  PASSED${NC}"
        ((passed_tests++))
    else
        echo -e "${RED}  FAILED${NC}"
        echo "  Различия:"
        cat "$TEMP_DIR/${test_name}.diff" | sed 's/^/    /'
        ((failed_tests++))
        failed_tests_list+=("lexer/$test_type/$test_name")
    fi
    
    ((total_tests++))
}

# Функция для запуска парсер теста
run_parser_test() {
    local test_category=$1
    local test_subcategory=$2
    local test_name=$3
    local src_file="../parser/$test_category/$test_subcategory/$test_name.src"
    local expected_file="../parser/$test_category/$test_subcategory/$test_name.expected"
    local temp_output="$TEMP_DIR/${test_name}.out"
    
    echo ""
    echo -e "${YELLOW}Запуск парсер теста: $test_category/$test_subcategory/$test_name${NC}"
    
    # Проверка существования исходного файла
    if [ ! -f "$src_file" ]; then
        echo -e "${RED}  Исходный файл не найден: $src_file${NC}"
        ((failed_tests++))
        failed_tests_list+=("parser/$test_category/$test_subcategory/$test_name (отсутствует исходный файл)")
        ((total_tests++))
        return
    fi
    
    # Запуск компилятора и захват вывода
    ../../bin/compiler parse --input "$src_file" > "$temp_output" 2>&1
    
    # Проверка существования файла с ожидаемым выводом
    if [ ! -f "$expected_file" ]; then
        echo -e "${RED}  Предупреждение: Файл с ожидаемым выводом не найден, показываю фактический вывод:${NC}"
        if [ -f "$temp_output" ]; then
            cat "$temp_output"
        else
            echo "  Вывод не сгенерирован"
        fi
        ((failed_tests++))
        failed_tests_list+=("parser/$test_category/$test_subcategory/$test_name (отсутствует ожидаемый вывод)")
        ((total_tests++))
        return
    fi
    
    # Для невалидных тестов проверяем наличие ошибок
    if [ "$test_category" = "invalid" ]; then
        if grep -q "Ошибки парсинга" "$temp_output"; then
            # Сравниваем с ожидаемыми ошибками
            if diff -u "$expected_file" "$temp_output" > "$TEMP_DIR/${test_name}.diff"; then
                echo -e "${GREEN}  PASSED (ожидаемые ошибки)${NC}"
                ((passed_tests++))
            else
                echo -e "${RED}  FAILED (несоответствие ошибок)${NC}"
                cat "$TEMP_DIR/${test_name}.diff" | sed 's/^/    /'
                ((failed_tests++))
                failed_tests_list+=("parser/$test_category/$test_subcategory/$test_name")
            fi
        else
            echo -e "${RED}  FAILED (ожидались ошибки, но парсинг прошел успешно)${NC}"
            ((failed_tests++))
            failed_tests_list+=("parser/$test_category/$test_subcategory/$test_name")
        fi
    else
        # Для валидных тестов сравниваем AST
        if diff -u "$expected_file" "$temp_output" > "$TEMP_DIR/${test_name}.diff"; then
            echo -e "${GREEN}  PASSED${NC}"
            ((passed_tests++))
        else
            echo -e "${RED} FAILED${NC}"
            echo "  Различия:"
            cat "$TEMP_DIR/${test_name}.diff" | sed 's/^/    /'
            ((failed_tests++))
            failed_tests_list+=("parser/$test_category/$test_subcategory/$test_name")
        fi
    fi
    
    ((total_tests++))
}

echo -e "${YELLOW}ТЕСТЫ ЛЕКСЕРА${NC}"

echo ""
echo "VALID TESTS"
if [ -d "../lexer/valid" ]; then
    for src_file in ../lexer/valid/*.src; do
        if [ -f "$src_file" ]; then
            test_name=$(basename "$src_file" .src)
            run_lexer_test "valid" "$test_name"
        fi
    done
else
    echo "Директория ../lexer/valid не найдена"
fi

echo ""
echo "INVALID TESTS"
if [ -d "../lexer/invalid" ]; then
    for src_file in ../lexer/invalid/*.src; do
        if [ -f "$src_file" ]; then
            test_name=$(basename "$src_file" .src)
            run_lexer_test "invalid" "$test_name"
        fi
    done
else
    echo "Директория ../lexer/invalid не найдена"
fi

echo -e "${YELLOW}ТЕСТЫ ПАРСЕРА${NC}"

# VALID - Expressions
echo ""
echo "VALID EXPRESSIONS"
if [ -d "../parser/valid/expressions" ]; then
    for src_file in ../parser/valid/expressions/*.src; do
        if [ -f "$src_file" ]; then
            test_name=$(basename "$src_file" .src)
            run_parser_test "valid" "expressions" "$test_name"
        fi
    done
else
    echo "Директория ../parser/valid/expressions не найдена"
fi

# VALID - Statements
echo ""
echo "VALID STATEMENTS"
if [ -d "../parser/valid/statements" ]; then
    for src_file in ../parser/valid/statements/*.src; do
        if [ -f "$src_file" ]; then
            test_name=$(basename "$src_file" .src)
            run_parser_test "valid" "statements" "$test_name"
        fi
    done
else
    echo "Директория ../parser/valid/statements не найдена"
fi

# VALID - Declarations
echo ""
echo "VALID DECLARATIONS"
if [ -d "../parser/valid/declarations" ]; then
    for src_file in ../parser/valid/declarations/*.src; do
        if [ -f "$src_file" ]; then
            test_name=$(basename "$src_file" .src)
            run_parser_test "valid" "declarations" "$test_name"
        fi
    done
else
    echo "Директория ../parser/valid/declarations не найдена"
fi

# VALID - Full Programs
echo ""
echo "VALID FULL PROGRAMS"
if [ -d "../parser/valid/full_programs" ]; then
    for src_file in ../parser/valid/full_programs/*.src; do
        if [ -f "$src_file" ]; then
            test_name=$(basename "$src_file" .src)
            run_parser_test "valid" "full_programs" "$test_name"
        fi
    done
else
    echo "Директория ../parser/valid/full_programs не найдена"
fi

# INVALID - Syntax Errors
echo ""
echo "INVALID SYNTAX ERRORS"
if [ -d "../parser/invalid/syntax_errors" ]; then
    for src_file in ../parser/invalid/syntax_errors/*.src; do
        if [ -f "$src_file" ]; then
            test_name=$(basename "$src_file" .src)
            run_parser_test "invalid" "syntax_errors" "$test_name"
        fi
    done
else
    echo "Директория ../parser/invalid/syntax_errors не найдена"
fi

# INVALID - Type Errors (если есть)
echo ""
echo "INVALID TYPE ERRORS"
if [ -d "../parser/invalid/type_errors" ]; then
    for src_file in ../parser/invalid/type_errors/*.src; do
        if [ -f "$src_file" ]; then
            test_name=$(basename "$src_file" .src)
            run_parser_test "invalid" "type_errors" "$test_name"
        fi
    done
else
    echo "Директория ../parser/invalid/type_errors не найдена"
fi

# INVALID - Semantic Errors (если есть)
echo ""
echo "INVALID SEMANTIC ERRORS"
if [ -d "../parser/invalid/semantic_errors" ]; then
    for src_file in ../parser/invalid/semantic_errors/*.src; do
        if [ -f "$src_file" ]; then
            test_name=$(basename "$src_file" .src)
            run_parser_test "invalid" "semantic_errors" "$test_name"
        fi
    done
else
    echo "Директория ../parser/invalid/semantic_errors не найдена"
fi

echo -e "${YELLOW}ИТОГИ ТЕСТИРОВАНИЯ${NC}"
echo ""
echo "Total tests:  $total_tests"
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
    echo -e "${GREEN}All tests PASSED${NC}"
    exit 0
else
    echo -e "${RED}Некоторые тесты не пройдены${NC}"
    exit 1
fi