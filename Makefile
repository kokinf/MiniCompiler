# Конфигурация проекта
APP_NAME = compiler
MAIN_FILE = ./src/cmd/compiler/main.go
OUTPUT_DIR = ./bin
OUTPUT = $(OUTPUT_DIR)/$(APP_NAME)
VERSION = 1.0.0

# Команды Go
GO = go
GOFLAGS = -ldflags="-s -w"

# Определение ОС
UNAME_S := $(shell uname -s 2>/dev/null || echo Windows)
ifeq ($(UNAME_S),Linux)
	EXT = 
	RM = rm -f
	RMDIR = rm -rf
	MKDIR = mkdir -p
else ifeq ($(UNAME_S),Darwin)
	EXT = 
	RM = rm -f
	RMDIR = rm -rf
	MKDIR = mkdir -p
else
	EXT = .exe
	RM = del /Q /F 2>nul || true
	RMDIR = rmdir /S /Q 2>nul || true
	MKDIR = mkdir 2>nul || true
endif

OUTPUT_FULL = $(OUTPUT)$(EXT)

# Цвета для вывода
ifeq ($(UNAME_S),Linux)
	GREEN = \033[0;32m
	RED = \033[0;31m
	YELLOW = \033[1;33m
	BLUE = \033[0;34m
	CYAN = \033[0;36m
	NC = \033[0m
else ifeq ($(UNAME_S),Darwin)
	GREEN = \033[0;32m
	RED = \033[0;31m
	YELLOW = \033[1;33m
	BLUE = \033[0;34m
	CYAN = \033[0;36m
	NC = \033[0m
else
	GREEN = 
	RED = 
	YELLOW = 
	BLUE = 
	CYAN = 
	NC = 
endif

# ============================================================================
# Основные цели
# ============================================================================

.PHONY: all
all: clean deps build
	@echo "$(GREEN)Проект успешно собран!$(NC)"

.PHONY: build
build: $(OUTPUT_DIR)
	@echo "$(YELLOW)Сборка компилятора...$(NC)"
	$(GO) build $(GOFLAGS) -o $(OUTPUT_FULL) $(MAIN_FILE)
	@echo "$(GREEN)Сборка завершена: $(OUTPUT_FULL)$(NC)"

$(OUTPUT_DIR):
	$(MKDIR) $(OUTPUT_DIR)

# ============================================================================
# Демонстрация Sprint 8
# ============================================================================

.PHONY: demo
demo: build
	@echo "$(GREEN)╔══════════════════════════════════════════╗$(NC)"
	@echo "$(GREEN)║   MiniCompiler v$(VERSION) - Демонстрация  ║$(NC)"
	@echo "$(GREEN)╚══════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(CYAN)1. Компиляция quicksort.src...$(NC)"
	$(OUTPUT_FULL) compile --input examples/demo_quicksort.src --output build/demo_quicksort.asm --optimize
	@echo ""
	@echo "$(CYAN)2. Ассемблирование...$(NC)"
	nasm -f elf64 -o build/demo_quicksort.o build/demo_quicksort.asm
	nasm -f elf64 -o build/runtime.o src/runtime/runtime.asm
	@echo ""
	@echo "$(CYAN)3. Линковка с libc...$(NC)"
	gcc -no-pie -o build/demo_quicksort build/runtime.o build/demo_quicksort.o
	@echo ""
	@echo "$(CYAN)4. Запуск программы...$(NC)"
	@echo "$(GREEN)──────────────────────────────────────────$(NC)"
	./build/demo_quicksort
	@echo "$(GREEN)──────────────────────────────────────────$(NC)"
	@echo ""
	@echo "$(GREEN)Демонстрация завершена успешно!$(NC)"

.PHONY: demo-verbose
demo-verbose: build
	@echo "$(YELLOW)Демонстрация с подробным выводом...$(NC)"
	@echo ""
	@echo "$(CYAN)Исходный код:$(NC)"
	@cat examples/demo_quicksort.src
	@echo ""
	@echo "$(CYAN)Компиляция с подробным выводом...$(NC)"
	$(OUTPUT_FULL) compile --input examples/demo_quicksort.src --output build/demo_quicksort.asm --optimize --verbose
	@echo ""
	@echo "$(CYAN)Сгенерированный ассемблер (первые 60 строк):$(NC)"
	@head -60 build/demo_quicksort.asm
	@echo "..."
	@echo ""
	@echo "$(CYAN)Сборка и запуск...$(NC)"
	nasm -f elf64 -o build/demo_quicksort.o build/demo_quicksort.asm
	nasm -f elf64 -o build/runtime.o src/runtime/runtime.asm
	gcc -no-pie -o build/demo_quicksort build/runtime.o build/demo_quicksort.o
	@echo "$(GREEN)Вывод программы:$(NC)"
	./build/demo_quicksort

.PHONY: demo-complex
demo-complex: build
	@echo "$(GREEN)╔══════════════════════════════════════════╗$(NC)"
	@echo "$(GREEN)║   Комплексная демонстрация                ║$(NC)"
	@echo "$(GREEN)╚══════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(CYAN)1. Компиляция demo_complex.src...$(NC)"
	$(OUTPUT_FULL) examples/demo_complex.src -o build/demo_complex.asm
	@echo ""
	@echo "$(CYAN)2. Ассемблирование...$(NC)"
	nasm -f elf64 -o build/demo_complex.o build/demo_complex.asm
	nasm -f elf64 -o build/runtime.o src/runtime/runtime.asm
	@echo ""
	@echo "$(CYAN)3. Линковка...$(NC)"
	gcc -no-pie -o build/demo_complex build/runtime.o build/demo_complex.o
	@echo ""
	@echo "$(CYAN)4. Запуск...$(NC)"
	@echo "$(GREEN)──────────────────────────────────────────$(NC)"
	./build/demo_complex
	@echo "$(GREEN)──────────────────────────────────────────$(NC)"
	@echo ""
	@echo "$(GREEN)Демонстрация завершена!$(NC)"

# ============================================================================
# Компиляция примеров
# ============================================================================

.PHONY: examples
examples: build
	@echo "$(YELLOW)Компиляция всех примеров...$(NC)"
	@for file in examples/*.src; do \
		name=$$(basename "$$file" .src); \
		echo "  $$file -> build/$$name.asm"; \
		$(OUTPUT_FULL) compile --input "$$file" --output "build/$$name.asm"; \
	done
	@echo "$(GREEN)Все примеры скомпилированы$(NC)"

# ============================================================================
# Тестирование
# ============================================================================

.PHONY: install-man
install-man:
	@echo "$(YELLOW)Установка man page...$(NC)"
	sudo mkdir -p /usr/local/share/man/man1
	sudo cp mikrocompiler.1 /usr/local/share/man/man1/
	sudo mandb
	@echo "$(GREEN)Man page установлена. Используйте: man mikrocompiler$(NC)"

.PHONY: test-final
test-final: build
	@echo "$(YELLOW)Запуск финальных тестов...$(NC)"
	@bash tests/run_final_tests.sh

.PHONY: test
test: build
	@echo "$(YELLOW)Запуск всех тестов...$(NC)"
	@if [ -f tests/test_runner/run_tests.sh ]; then \
		cd tests/test_runner && bash run_tests.sh all; \
	else \
		echo "$(RED)Тестовый раннер не найден$(NC)"; \
	fi

.PHONY: test-lexer
test-lexer: build
	@echo "$(YELLOW)Запуск тестов лексера...$(NC)"
	@if [ -f tests/test_runner/run_tests.sh ]; then \
		cd tests/test_runner && bash run_tests.sh lexer; \
	fi

.PHONY: test-parser
test-parser: build
	@echo "$(YELLOW)Запуск тестов парсера...$(NC)"
	@if [ -f tests/test_runner/run_tests.sh ]; then \
		cd tests/test_runner && bash run_tests.sh parser; \
	fi

.PHONY: test-semantic
test-semantic: build
	@echo "$(YELLOW)Запуск семантических тестов...$(NC)"
	@if [ -f tests/test_runner/run_tests.sh ]; then \
		cd tests/test_runner && bash run_tests.sh semantic; \
	fi

.PHONY: test-ir
test-ir: build
	@echo "$(YELLOW)Запуск тестов IR...$(NC)"
	@if [ -f tests/test_runner/run_tests.sh ]; then \
		cd tests/test_runner && bash run_tests.sh ir; \
	fi

.PHONY: test-codegen
test-codegen: build
	@echo "$(YELLOW)Запуск тестов кодогенерации...$(NC)"
	@if [ -f tests/test_runner/run_tests.sh ]; then \
		cd tests/test_runner && bash run_tests.sh codegen; \
	fi

.PHONY: test-control-flow
test-control-flow: build
	@echo "$(YELLOW)Запуск тестов Control Flow...$(NC)"
	@if [ -f tests/test_runner/run_tests.sh ]; then \
		cd tests/test_runner && bash run_tests.sh control-flow; \
	fi

.PHONY: test-all
test-all: build
	@echo "$(YELLOW)Запуск всех тестов...$(NC)"
	$(GO) test ./src/internal/... -v
	@if [ -f tests/test_runner/run_tests.sh ]; then \
		cd tests/test_runner && bash run_tests.sh all; \
	fi

.PHONY: test-report
test-report: build
	@echo "$(YELLOW)Запуск тестов с отчётом...$(NC)"
	@bash tests/test_runner/run_tests.sh all 2>&1 | tee test_report.log
	@echo "$(GREEN)Отчёт сохранён в test_report.log$(NC)"

# ============================================================================
# Информация о компиляторе
# ============================================================================

.PHONY: version
version: build
	$(OUTPUT_FULL) --version

.PHONY: help-compiler
help-compiler: build
	$(OUTPUT_FULL) --help

# ============================================================================
# Управление зависимостями
# ============================================================================

.PHONY: deps
deps:
	@echo "$(YELLOW)Загрузка зависимостей...$(NC)"
	$(GO) mod download
	$(GO) mod verify
	@echo "$(GREEN)Зависимости загружены$(NC)"

.PHONY: tidy
tidy:
	@echo "$(YELLOW)Очистка зависимостей...$(NC)"
	$(GO) mod tidy
	@echo "$(GREEN)Зависимости очищены$(NC)"

# ============================================================================
# Очистка
# ============================================================================

.PHONY: clean
clean:
	@echo "$(YELLOW)Очистка артефактов сборки...$(NC)"
	-$(RM) $(OUTPUT_FULL)
	-$(RMDIR) $(OUTPUT_DIR)
	-$(RMDIR) build
	-$(RM) cfg.dot cfg.png ast.dot ast.png *.ir
	-$(RM) test_report.log
	$(GO) clean
	@echo "$(GREEN)Очистка завершена$(NC)"

.PHONY: clean-all
clean-all: clean
	@echo "$(YELLOW)Очистка кэша Go...$(NC)"
	$(GO) clean -cache -modcache -i -r
	@echo "$(GREEN)Полная очистка завершена$(NC)"

.PHONY: distclean
distclean: clean
	@echo "$(YELLOW)Полная очистка...$(NC)"
	$(GO) clean -cache -modcache -testcache
	-$(RM) -rf tests/golden/expected
	@echo "$(GREEN)Полная очистка завершена$(NC)"

# ============================================================================
# Качество кода
# ============================================================================

.PHONY: fmt
fmt:
	@echo "$(YELLOW)Форматирование кода...$(NC)"
	$(GO) fmt ./...
	@echo "$(GREEN)Форматирование завершено$(NC)"

.PHONY: vet
vet:
	@echo "$(YELLOW)Проверка кода статическим анализатором...$(NC)"
	$(GO) vet ./...
	@echo "$(GREEN)Проверка завершена$(NC)"

.PHONY: lint
lint: vet
	@echo "$(YELLOW)Запуск линтера...$(NC)"
	@which golint > /dev/null && golint ./... || echo "golint не установлен"

# ============================================================================
# Установка и распространение
# ============================================================================

.PHONY: install
install: build
	@echo "$(YELLOW)Установка компилятора в /usr/local/bin...$(NC)"
	sudo cp $(OUTPUT_FULL) /usr/local/bin/mikrocompiler
	@echo "$(GREEN)Установка завершена!$(NC)"
	@echo "Теперь можно использовать: mikrocompiler --help"

.PHONY: uninstall
uninstall:
	@echo "$(YELLOW)Удаление компилятора...$(NC)"
	sudo rm -f /usr/local/bin/mikrocompiler
	@echo "$(GREEN)Удаление завершено$(NC)"

.PHONY: dist
dist: build
	@echo "$(YELLOW)Создание дистрибутива...$(NC)"
	@mkdir -p dist/mikrocompiler-$(VERSION)
	@cp $(OUTPUT_FULL) dist/mikrocompiler-$(VERSION)/
	@cp src/runtime/runtime.asm dist/mikrocompiler-$(VERSION)/
	@cp src/libc/stdlib.h dist/mikrocompiler-$(VERSION)/
	@cp -r examples dist/mikrocompiler-$(VERSION)/
	@cp README.md dist/mikrocompiler-$(VERSION)/
	@cp Makefile dist/mikrocompiler-$(VERSION)/
	@cd dist && tar czf mikrocompiler-$(VERSION).tar.gz mikrocompiler-$(VERSION)
	@echo "$(GREEN)Дистрибутив создан: dist/mikrocompiler-$(VERSION).tar.gz$(NC)"
	@echo "Размер: $$(du -h dist/mikrocompiler-$(VERSION).tar.gz | cut -f1)"

# ============================================================================
# Генерация AST/CFG изображений
# ============================================================================

.PHONY: ast
ast: build
	@echo "$(YELLOW)Генерация AST для examples/demo_quicksort.src...$(NC)"
	$(OUTPUT_FULL) parse --input examples/demo_quicksort.src --format dot --output ast.dot
	@if command -v dot > /dev/null; then \
		dot -Tpng ast.dot -o ast.png; \
		echo "$(GREEN)AST сохранён в ast.png$(NC)"; \
	else \
		echo "$(YELLOW)Graphviz не установлен, DOT файл сохранён в ast.dot$(NC)"; \
	fi

.PHONY: cfg
cfg: build
	@echo "$(YELLOW)Генерация CFG для examples/demo_quicksort.src...$(NC)"
	$(OUTPUT_FULL) ir --input examples/demo_quicksort.src --format dot --output cfg.dot
	@if command -v dot > /dev/null; then \
		dot -Tpng cfg.dot -o cfg.png; \
		echo "$(GREEN)CFG сохранён в cfg.png$(NC)"; \
	else \
		echo "$(YELLOW)Graphviz не установлен, DOT файл сохранён в cfg.dot$(NC)"; \
	fi

# ============================================================================
# Справка
# ============================================================================

.PHONY: help
help:
	@echo "$(BLUE)MiniCompiler v$(VERSION) - Доступные цели$(NC)"
	@echo ""
	@echo "$(YELLOW)Сборка и запуск:$(NC)"
	@echo "  make build           - Собрать компилятор"
	@echo "  make clean           - Очистить артефакты сборки"
	@echo "  make version         - Показать версию"
	@echo "  make help-compiler   - Справка по компилятору"
	@echo ""
	@echo "$(YELLOW)Демонстрация:$(NC)"
	@echo "  make demo            - Запустить демо (QuickSort)"
	@echo "  make demo-verbose    - Демо с подробным выводом"
	@echo "  make demo-complex    - Комплексная демонстрация"
	@echo ""
	@echo "$(YELLOW)Компиляция примеров:$(NC)"
	@echo "  make examples        - Скомпилировать все примеры"
	@echo ""
	@echo "$(YELLOW)Тестирование:$(NC)"
	@echo "  make test            - Запустить все тесты"
	@echo "  make test-lexer      - Тесты лексера"
	@echo "  make test-parser     - Тесты парсера"
	@echo "  make test-semantic   - Семантические тесты"
	@echo "  make test-ir         - Тесты IR"
	@echo "  make test-codegen    - Тесты кодогенерации"
	@echo "  make test-all        - Все тесты (unit + integration)"
	@echo "  make test-report     - Тесты с отчётом"
	@echo ""
	@echo "$(YELLOW)Качество кода:$(NC)"
	@echo "  make fmt             - Форматировать код"
	@echo "  make vet             - Статический анализатор"
	@echo "  make lint            - Линтер"
	@echo ""
	@echo "$(YELLOW)Управление зависимостями:$(NC)"
	@echo "  make deps            - Загрузить зависимости"
	@echo "  make tidy            - Очистить зависимости"
	@echo ""
	@echo "$(YELLOW)Установка и распространение:$(NC)"
	@echo "  make install         - Установить в /usr/local/bin"
	@echo "  make uninstall       - Удалить из системы"
	@echo "  make dist            - Создать дистрибутив"
	@echo ""
	@echo "$(YELLOW)Визуализация:$(NC)"
	@echo "  make ast             - Сгенерировать AST (PNG)"
	@echo "  make cfg             - Сгенерировать CFG (PNG)"
	@echo ""
	@echo "$(YELLOW)Очистка:$(NC)"
	@echo "  make clean           - Очистить артефакты"
	@echo "  make clean-all       - Очистить всё (включая кэш)"
	@echo "  make distclean       - Полная очистка"