#!/bin/bash

echo "Сборка компилятора..."
cd ../..
make build
if [ $? -ne 0 ]; then
    echo "Сборка не удалась"
    exit 1
fi

cd tests/test_runner

total_tests=0
passed_tests=0
failed_tests=0
failed_tests_list=()

TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Функция для запуска теста
run_test() {
    local test_type=$1
    local test_name=$2
    local src_file="../lexer/$test_type/$test_name.src"
    local expected_file="../lexer/$test_type/$test_name.expected"
    local temp_output="$TEMP_DIR/${test_name}.out"
    
    echo ""
    echo "Запуск теста: $test_type/$test_name"
    
    # Проверка существования исходного файла
    if [ ! -f "$src_file" ]; then
        echo "  Исходный файл не найден: $src_file"
        ((failed_tests++))
        failed_tests_list+=("$test_type/$test_name (отсутствует исходный файл)")
        ((total_tests++))
        return
    fi
    
    # Запуск компилятора и захват вывода
    ../../bin/compiler lex --input "$src_file" > "$temp_output" 2>&1
    
    # Проверка успешности выполнения компилятора
    if [ $? -ne 0 ]; then
        echo "  Выполнение компилятора не удалось"
    fi
    
    # Проверка существования файла с ожидаемым выводом
    if [ ! -f "$expected_file" ]; then
        echo "  Предупреждение: Файл с ожидаемым выводом не найден, показываю фактический вывод:"
        if [ -f "$temp_output" ]; then
            cat "$temp_output"
        else
            echo "  Вывод не сгенерирован"
        fi
        ((failed_tests++))
        failed_tests_list+=("$test_type/$test_name (отсутствует ожидаемый вывод)")
        ((total_tests++))
        return
    fi
    
    # Сравнение с ожидаемым выводом
    if diff -u "$expected_file" "$temp_output" > "$TEMP_DIR/${test_name}.diff"; then
        echo "test PASSED"
        ((passed_tests++))
    else
        echo "test FAILED"
        echo "  Различия:"
        cat "$TEMP_DIR/${test_name}.diff" | sed 's/^/    /'
        ((failed_tests++))
        failed_tests_list+=("$test_type/$test_name")
    fi
    
    ((total_tests++))
}

echo ""
echo "VALID TESTS"

# Проверяем наличие файлов .src в папке valid
if [ -d "../lexer/valid" ]; then
    for src_file in ../lexer/valid/*.src; do
        if [ -f "$src_file" ]; then
            test_name=$(basename "$src_file" .src)
            run_test "valid" "$test_name"
        fi
    done
else
    echo "Директория ../lexer/valid не найдена"
fi

# Запуск всех невалидных тестов
echo ""
echo "INVALID TESTS"

# Проверяем наличие файлов .src в папке invalid
if [ -d "../lexer/invalid" ]; then
    for src_file in ../lexer/invalid/*.src; do
        if [ -f "$src_file" ]; then
            test_name=$(basename "$src_file" .src)
            run_test "invalid" "$test_name"
        fi
    done
else
    echo "Директория ../lexer/invalid не найдена"
fi

# Вывод итоговой сводки
echo ""
echo "Total tests:"
echo "  Total tests:  $total_tests"
echo "  PASSED:       $passed_tests"
echo "  FAILED:    $failed_tests"

if [ ${#failed_tests_list[@]} -gt 0 ]; then
    echo ""
    echo "Не пройденные тесты:"
    for test in "${failed_tests_list[@]}"; do
        echo "  • $test"
    done
fi

echo ""

if [ $failed_tests -eq 0 ]; then
    echo "All tests PASSED"
    exit 0
else
    echo "Некоторые тесты не пройдены"
    exit 1
fi