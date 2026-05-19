#!/bin/bash

# Скрипт для генерации golden testing файлов (expected output)
# Запускает компилятор на всех тестовых .src файлах и сохраняет вывод как .expected

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Проверка наличия компилятора
COMPILER="./bin/compiler"
if [ ! -f "$COMPILER" ]; then
    echo -e "${YELLOW}Компилятор не найден. Выполняю сборку...${NC}"
    make build
    if [ ! -f "$COMPILER" ]; then
        echo -e "${RED}Ошибка сборки компилятора${NC}"
        exit 1
    fi
fi

# Директории с тестами
TESTS_DIR="tests"

# ============================================================================
# Функция для генерации expected файла
# ============================================================================

generate_expected() {
    local src_file=$1
    local command=$2
    local extra_args=$3
    local expected_file="${src_file%.src}.expected"
    
    echo -n "  $(basename "$src_file") -> $(basename "$expected_file") "
    
    # Формируем команду в зависимости от типа
    local cmd_args=""
    case "$command" in
        "lex")
            # Лексер не поддерживает --format
            cmd_args="$command --input \"$src_file\""
            ;;
        "parse")
            cmd_args="$command --input \"$src_file\" --format text"
            ;;
        "check")
            if [ -n "$extra_args" ]; then
                cmd_args="$command --input \"$src_file\" $extra_args"
            else
                cmd_args="$command --input \"$src_file\""
            fi
            ;;
        "ir")
            if [ -n "$extra_args" ]; then
                cmd_args="$command --input \"$src_file\" $extra_args"
            else
                cmd_args="$command --input \"$src_file\" --format text"
            fi
            ;;
        "symbols")
            cmd_args="$command --input \"$src_file\" --format text"
            ;;
        *)
            cmd_args="$command --input \"$src_file\""
            ;;
    esac
    
    # Выполняем команду с таймаутом 5 секунд
    timeout 5 bash -c "$COMPILER $cmd_args" > "$expected_file" 2>&1
    local exit_code=$?
    
    # Анализируем результат
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}✓${NC}"
    elif [ $exit_code -eq 124 ]; then
        echo -e "${RED}✗ (timeout)${NC}"
        echo "TIMEOUT: compilation took more than 5 seconds" > "$expected_file"
    elif [ $exit_code -eq 1 ]; then
        # Exit code 1 - может быть как ожидаемой ошибкой, так и реальной проблемой
        if [[ "$src_file" == *"/invalid/"* ]]; then
            echo -e "${GREEN}✓ (expected error)${NC}"
        else
            # Проверяем, есть ли в выводе сообщения об ошибках
            if grep -q "Ошибка\|ошибка\|Error\|error" "$expected_file"; then
                echo -e "${RED}✗ (compilation error)${NC}"
            else
                echo -e "${YELLOW}⚠ (exit code 1, check .expected)${NC}"
            fi
        fi
    elif [ $exit_code -eq 2 ]; then
        echo -e "${RED}✗ (exit code 2)${NC}"
        # Сохраняем вывод для диагностики
        if [ ! -s "$expected_file" ]; then
            echo "EXIT CODE 2 (possible usage error or panic)" > "$expected_file"
        fi
    else
        echo -e "${YELLOW}⚠ (exit code: $exit_code)${NC}"
    fi
}

# ============================================================================
# Функция для рекурсивной обработки вложенных директорий
# ============================================================================

process_recursive() {
    local base_dir=$1
    local command=$2
    local extra_args=$3
    
    if [ ! -d "$base_dir" ]; then
        return
    fi
    
    # Проверяем наличие .src файлов в текущей директории
    local has_files=false
    for src_file in "$base_dir"/*.src; do
        if [ -f "$src_file" ]; then
            has_files=true
            break
        fi
    done
    
    # Обрабатываем файлы если они есть
    if [ "$has_files" = true ]; then
        echo ""
        local rel_path="${base_dir#$TESTS_DIR/}"
        echo -e "${YELLOW}${rel_path}:${NC}"
        
        for src_file in "$base_dir"/*.src; do
            if [ -f "$src_file" ]; then
                generate_expected "$src_file" "$command" "$extra_args"
            fi
        done
    fi
    
    # Рекурсивно обрабатываем поддиректории
    for subdir in "$base_dir"/*/; do
        if [ -d "$subdir" ]; then
            process_recursive "$subdir" "$command" "$extra_args"
        fi
    done
}

# ============================================================================
# Очистка старых expected файлов (опционально)
# ============================================================================

if [ "$1" = "--clean" ]; then
    echo -e "${YELLOW}Очистка старых expected файлов...${NC}"
    find "$TESTS_DIR" -name "*.expected" -type f -delete
    find "$TESTS_DIR" -name "*.dot" -type f -delete
    find "$TESTS_DIR" -name "*.json" -type f -delete
    echo -e "${GREEN}Очистка завершена${NC}"
    echo ""
    exit 0
fi

# ============================================================================
# Генерация expected файлов
# ============================================================================

echo -e "${CYAN}=== Генерация golden testing файлов ===${NC}"
echo -e "${YELLOW}Компилятор: $COMPILER${NC}"
echo -e "${YELLOW}Директория: $(pwd)${NC}"

# ----------------------------------------------------------------------------
# Лексер тесты
# ----------------------------------------------------------------------------

echo ""
echo -e "${CYAN}=== ЛЕКСЕР ===${NC}"
process_recursive "$TESTS_DIR/lexer/valid" "lex" ""
process_recursive "$TESTS_DIR/lexer/invalid" "lex" ""

# ----------------------------------------------------------------------------
# Парсер тесты
# ----------------------------------------------------------------------------

echo ""
echo -e "${CYAN}=== ПАРСЕР ===${NC}"
process_recursive "$TESTS_DIR/parser/valid" "parse" ""
process_recursive "$TESTS_DIR/parser/invalid" "parse" ""

# ----------------------------------------------------------------------------
# Семантические тесты
# ----------------------------------------------------------------------------

echo ""
echo -e "${CYAN}=== СЕМАНТИЧЕСКИЙ АНАЛИЗ ===${NC}"
process_recursive "$TESTS_DIR/semantic/valid" "check" ""
process_recursive "$TESTS_DIR/semantic/invalid" "check" ""

# ----------------------------------------------------------------------------
# IR тесты
# ----------------------------------------------------------------------------

echo ""
echo -e "${CYAN}=== IR ГЕНЕРАЦИЯ ===${NC}"

if [ -d "$TESTS_DIR/ir/generation" ]; then
    for subdir in "$TESTS_DIR/ir/generation"/*/; do
        if [ -d "$subdir" ]; then
            process_recursive "$subdir" "ir" "--format text"
        fi
    done
fi

echo ""
echo -e "${CYAN}=== IR ВАЛИДАЦИЯ ===${NC}"

if [ -d "$TESTS_DIR/ir/validation" ]; then
    for subdir in "$TESTS_DIR/ir/validation"/*/; do
        if [ -d "$subdir" ]; then
            dir_name=$(basename "$subdir")
            if [[ "$dir_name" == *"optimization"* ]]; then
                process_recursive "$subdir" "ir" "--format text --optimize --stats"
            else
                process_recursive "$subdir" "ir" "--format text"
            fi
        fi
    done
fi

# ============================================================================
# Дополнительно: генерация DOT и JSON для интеграционных тестов IR
# ============================================================================

if [ "$1" = "--all-formats" ] || [ "$2" = "--all-formats" ]; then
    echo ""
    echo -e "${CYAN}=== ГЕНЕРАЦИЯ DOT И JSON ===${NC}"
    
    # IR integration tests
    if [ -d "$TESTS_DIR/ir/generation/integration" ]; then
        echo ""
        echo -e "${YELLOW}ir/integration (DOT & JSON):${NC}"
        for src_file in "$TESTS_DIR/ir/generation/integration"/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                dot_file="${src_file%.src}.dot"
                json_file="${src_file%.src}.json"
                
                echo -n "  $name (dot) "
                timeout 5 $COMPILER ir --input "$src_file" --format dot > "$dot_file" 2>&1
                if [ $? -eq 0 ]; then
                    echo -e "${GREEN}✓${NC}"
                else
                    echo -e "${RED}✗${NC}"
                fi
                
                echo -n "  $name (json) "
                timeout 5 $COMPILER ir --input "$src_file" --format json > "$json_file" 2>&1
                if [ $? -eq 0 ]; then
                    echo -e "${GREEN}✓${NC}"
                else
                    echo -e "${RED}✗${NC}"
                fi
            fi
        done
    fi
    
    # Parser full programs
    if [ -d "$TESTS_DIR/parser/valid/full_programs" ]; then
        echo ""
        echo -e "${YELLOW}parser/full_programs (DOT & JSON):${NC}"
        for src_file in "$TESTS_DIR/parser/valid/full_programs"/*.src; do
            if [ -f "$src_file" ]; then
                name=$(basename "$src_file" .src)
                dot_file="${src_file%.src}.dot"
                json_file="${src_file%.src}.json"
                
                echo -n "  $name (dot) "
                timeout 5 $COMPILER parse --input "$src_file" --format dot > "$dot_file" 2>&1
                if [ $? -eq 0 ]; then
                    echo -e "${GREEN}✓${NC}"
                else
                    echo -e "${RED}✗${NC}"
                fi
                
                echo -n "  $name (json) "
                timeout 5 $COMPILER parse --input "$src_file" --format json > "$json_file" 2>&1
                if [ $? -eq 0 ]; then
                    echo -e "${GREEN}✓${NC}"
                else
                    echo -e "${RED}✗${NC}"
                fi
            fi
        done
    fi
fi

# ============================================================================
# Завершение
# ============================================================================

echo ""
echo -e "${GREEN}=== Генерация golden testing файлов завершена ===${NC}"
echo ""

# Подсчёт результатов
total_expected=$(find "$TESTS_DIR" -name "*.expected" -type f | wc -l)
total_src=$(find "$TESTS_DIR" -name "*.src" -type f | wc -l)

echo -e "${CYAN}Статистика:${NC}"
echo -e "  .src файлов:     ${BLUE}$total_src${NC}"
echo -e "  .expected файлов: ${BLUE}$total_expected${NC}"

if [ $total_src -ne $total_expected ]; then
    echo -e "  ${YELLOW}⚠ Не для всех .src файлов сгенерированы .expected${NC}"
fi

echo ""
echo -e "${YELLOW}Примечания:${NC}"
echo "  • Файлы с ошибками парсинга содержат сообщения об ошибках"
echo "  • Для invalid тестов ошибки ожидаемы"
echo "  • Файлы с таймаутом помечены как TIMEOUT"
echo "  • Проверьте сгенерированные .expected файлы перед коммитом"
echo ""
echo -e "${CYAN}Следующие шаги:${NC}"
echo "  1. Проверьте .expected файлы:"
echo "     git diff tests/"
echo "  2. При необходимости отредактируйте их вручную"
echo "  3. Запустите тесты:"
echo "     make test"
echo ""
echo -e "${YELLOW}Для генерации DOT/JSON файлов:${NC}"
echo "  $0 --all-formats"