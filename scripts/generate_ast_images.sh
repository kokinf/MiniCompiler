#!/bin/bash

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Проверка наличия graphviz
if ! command -v dot &> /dev/null; then
    echo -e "${RED}Graphviz не установлен. Установите его:${NC}"
    echo "  Ubuntu/Debian: sudo apt-get install graphviz"
    echo "  macOS: brew install graphviz"
    echo "  Windows: скачайте с https://graphviz.org/download/"
    exit 1
fi

# Проверка сборки компилятора
if [ ! -f "./bin/compiler" ]; then
    echo -e "${YELLOW}Сборка компилятора...${NC}"
    make build
fi

echo -e "${YELLOW}Генерация AST изображений...${NC}"

# Создаём директорию для изображений, если её нет
mkdir -p ast_images

# Генерация для всех примеров
for srcfile in examples/*.src; do
    if [ ! -f "$srcfile" ]; then
        continue
    fi
    
    name=$(basename "$srcfile" .src)
    dotfile="ast_images/ast_${name}.dot"
    pngfile="ast_images/ast_${name}.png"
    
    echo "  Обработка: $srcfile"
    
    # Генерация DOT файла
    ./bin/compiler parse --input "$srcfile" --format dot --output "$dotfile"
    
    if [ $? -eq 0 ]; then
        # Конвертация в PNG
        dot -Tpng "$dotfile" -o "$pngfile"
        echo -e "    ${GREEN}✓ Создан: $pngfile${NC}"
    else
        echo -e "    ${RED}✗ Ошибка при парсинге $srcfile${NC}"
    fi
done

# Генерация для тестовых файлов (опционально)
if [ "$1" == "--with-tests" ]; then
    echo -e "\n${YELLOW}Генерация AST для тестовых файлов...${NC}"
    
    for testfile in tests/parser/valid/full_programs/*.src; do
        if [ ! -f "$testfile" ]; then
            continue
        fi
        
        name=$(basename "$testfile" .src)
        dotfile="ast_images/test_${name}.dot"
        pngfile="ast_images/test_${name}.png"
        
        echo "  Обработка: $testfile"
        ./bin/compiler parse --input "$testfile" --format dot --output "$dotfile"
        
        if [ $? -eq 0 ]; then
            dot -Tpng "$dotfile" -o "$pngfile"
            echo -e "    ${GREEN}✓ Создан: $pngfile${NC}"
        fi
    done
fi

echo -e "\n${GREEN}Все изображения сохранены в директории ast_images/${NC}"