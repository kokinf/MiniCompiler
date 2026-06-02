#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

COMPILER="./bin/compiler"
RUNTIME="src/runtime/runtime.asm"
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

PASSED=0
FAILED=0
TOTAL=0

run_good_test() {
    local src=$1 expected=$2 category=$3
    local name=$(basename "$src" .src)
    TOTAL=$((TOTAL + 1))
    
    $COMPILER -o "$TEMP_DIR/$name.asm" "$src" 2>/dev/null
    [ $? -ne 0 ] && { echo -e "  ${RED}✗ $name [COMPILE]${NC}"; FAILED=$((FAILED + 1)); return; }
    
    nasm -f elf64 -o "$TEMP_DIR/$name.o" "$TEMP_DIR/$name.asm" 2>/dev/null
    nasm -f elf64 -o "$TEMP_DIR/runtime.o" "$RUNTIME" 2>/dev/null
    gcc -no-pie -o "$TEMP_DIR/$name" "$TEMP_DIR/runtime.o" "$TEMP_DIR/$name.o" 2>/dev/null
    [ $? -ne 0 ] && { echo -e "  ${RED}✗ $name [LINK]${NC}"; FAILED=$((FAILED + 1)); return; }
    
    local result=$("$TEMP_DIR/$name" 2>/dev/null; echo $?)
    [ "$result" = "$expected" ] && {
        echo -e "  ${GREEN}✓ $name (exit: $result)${NC}"; PASSED=$((PASSED + 1))
    } || {
        echo -e "  ${RED}✗ $name (expected: $expected, got: $result)${NC}"; FAILED=$((FAILED + 1))
    }
}

run_bad_test() {
    local src=$1 category=$2
    local name=$(basename "$src" .src)
    TOTAL=$((TOTAL + 1))
    
    $COMPILER -o /dev/null "$src" > /dev/null 2>&1
    local exit_code=$?
    
    if [ $exit_code -ne 0 ]; then
        echo -e "  ${GREEN}✓ $name (error detected)${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "  ${RED}✗ $name (expected error, but compiled OK)${NC}"
        FAILED=$((FAILED + 1))
    fi
}

run_multi_test() {
    local dir=$1 expected=$2
    local name=$(basename "$dir")
    TOTAL=$((TOTAL + 1))
    
    local src_files=$(ls $dir/*.src 2>/dev/null)
    [ -z "$src_files" ] && return
    
    $COMPILER $src_files -o "$TEMP_DIR/${name}_multi" 2>/dev/null
    [ $? -ne 0 ] && { echo -e "  ${RED}✗ $name [COMPILE]${NC}"; FAILED=$((FAILED + 1)); return; }
    
    # Мульти-файловая компиляция создаёт asm и o файлы
    # Нужно слинковать вручную
    local obj_files=""
    for f in $dir/*.src; do
        local n=$(basename "$f" .src)
        [ -f "$n.o" ] && obj_files="$obj_files $n.o"
        [ -f "$n.asm" ] && nasm -f elf64 -o "$n.o" "$n.asm" 2>/dev/null && obj_files="$obj_files $n.o"
    done
    nasm -f elf64 -o "$TEMP_DIR/runtime.o" "$RUNTIME" 2>/dev/null
    gcc -no-pie -o "$TEMP_DIR/$name" "$TEMP_DIR/runtime.o" $obj_files 2>/dev/null
    [ $? -ne 0 ] && { echo -e "  ${RED}✗ $name [LINK]${NC}"; FAILED=$((FAILED + 1)); return; }
    
    local result=$("$TEMP_DIR/$name" 2>/dev/null; echo $?)
    [ "$result" = "$expected" ] && {
        echo -e "  ${GREEN}✓ $name (exit: $result)${NC}"; PASSED=$((PASSED + 1))
    } || {
        echo -e "  ${RED}✗ $name (expected: $expected, got: $result)${NC}"; FAILED=$((FAILED + 1))
    }
}

run_category() {
    local dir=$1 cat_name=$2 type=$3
    [ -d "$dir" ] || return
    [ "$(ls -A $dir/*.src 2>/dev/null)" ] || return
    echo -e "\n  ${CYAN}$cat_name:${NC}"
    for f in $dir/*.src; do
        [ -f "$f" ] || continue
        if [ "$type" = "good" ]; then
            run_good_test "$f" "$(cat ${f%.src}.expected 2>/dev/null || echo 0)" "$cat_name"
        else
            run_bad_test "$f" "$cat_name"
        fi
    done
}

echo -e "${CYAN}╔══════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║         FINAL TEST SUITE v1.0            ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════╝${NC}"

echo -e "\n${YELLOW}Good tests:${NC}"
run_category "tests/final/good/arithmetic" "Arithmetic" "good"
run_category "tests/final/good/control_flow" "Control Flow" "good"
run_category "tests/final/good/functions" "Functions" "good"
run_category "tests/final/good/arrays" "Arrays" "good"
run_category "tests/final/good/integration" "Integration" "good"
run_category "tests/final/regression" "Regression" "good"
run_category "tests/final/benchmarks" "Benchmarks" "good"
run_category "tests/final/demo" "Demo" "good"

echo -e "\n${YELLOW}Multi-file tests:${NC}"
for dir in tests/final/good/multi_file/*/; do
    [ -d "$dir" ] && run_multi_test "$dir" "$(cat ${dir}main.expected 2>/dev/null || echo 0)"
done

echo -e "\n${YELLOW}Bad tests:${NC}"
run_category "tests/final/bad/syntax_errors" "Syntax Errors" "bad"
run_category "tests/final/bad/type_errors" "Type Errors" "bad"
run_category "tests/final/bad/semantic_errors" "Semantic Errors" "bad"
run_category "tests/final/bad/runtime_errors" "Runtime Errors" "bad"

echo -e "\n${CYAN}╔══════════════════════════════════════════╗${NC}"
printf "${CYAN}║  Results: ${GREEN}%d passed${NC}, ${RED}%d failed${NC}, %d total  ║${NC}\n" $PASSED $FAILED $TOTAL
echo -e "${CYAN}╚══════════════════════════════════════════╝${NC}"
[ $FAILED -eq 0 ] && exit 0 || exit 1
