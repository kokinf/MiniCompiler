# Конфигурация проекта
APP_NAME = compiler
MAIN_FILE = ./src/cmd/compiler/main.go
OUTPUT_DIR = ./bin
OUTPUT = $(OUTPUT_DIR)/$(APP_NAME)

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

# Полный путь к выходному файлу с расширением
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

.PHONY: run
run: build
	@echo "$(YELLOW)Запуск парсера на examples/factorial.src...$(NC)"
	$(OUTPUT_FULL) parse --input examples/factorial.src

.PHONY: run-lex
run-lex: build
	@echo "$(YELLOW)Запуск лексера на examples/hello.src...$(NC)"
	$(OUTPUT_FULL) lex --input examples/hello.src

.PHONY: run-parse
run-parse: build
	@echo "$(YELLOW)Запуск парсера на examples/struct.src...$(NC)"
	$(OUTPUT_FULL) parse --input examples/struct.src

.PHONY: check
check: build
	@echo "$(YELLOW)Запуск семантического анализа на examples/factorial.src...$(NC)"
	$(OUTPUT_FULL) check --input examples/factorial.src --verbose

.PHONY: check-types
check-types: build
	@echo "$(YELLOW)Запуск семантического анализа с выводом типов...$(NC)"
	$(OUTPUT_FULL) check --input examples/factorial.src --verbose --show-types

.PHONY: symbols
symbols: build
	@echo "$(YELLOW)Вывод таблицы символов для examples/factorial.src...$(NC)"
	$(OUTPUT_FULL) symbols --input examples/factorial.src

.PHONY: symbols-json
symbols-json: build
	@echo "$(YELLOW)Вывод таблицы символов в JSON формате...$(NC)"
	$(OUTPUT_FULL) symbols --input examples/factorial.src --format json

# ============================================================================
# IR (Intermediate Representation) цели
# ============================================================================

.PHONY: ir
ir: build
	@echo "$(YELLOW)Генерация IR для examples/factorial.src...$(NC)"
	$(OUTPUT_FULL) ir --input examples/factorial.src

.PHONY: ir-opt
ir-opt: build
	@echo "$(YELLOW)Генерация оптимизированного IR...$(NC)"
	$(OUTPUT_FULL) ir --input examples/factorial.src --optimize --stats

.PHONY: ir-dot
ir-dot: build
	@echo "$(YELLOW)Генерация CFG в DOT формате...$(NC)"
	$(OUTPUT_FULL) ir --input examples/factorial.src --format dot --output cfg.dot
	@if command -v dot > /dev/null; then \
		dot -Tpng cfg.dot -o cfg.png; \
		echo "$(GREEN)CFG сохранён в cfg.png$(NC)"; \
	else \
		echo "$(YELLOW)Graphviz не установлен, DOT файл сохранён в cfg.dot$(NC)"; \
	fi

.PHONY: ir-json
ir-json: build
	@echo "$(YELLOW)Генерация IR в JSON...$(NC)"
	$(OUTPUT_FULL) ir --input examples/factorial.src --format json

.PHONY: ir-stats
ir-stats: build
	@echo "$(YELLOW)Статистика IR для examples/factorial.src...$(NC)"
	$(OUTPUT_FULL) ir --input examples/factorial.src --stats

# ============================================================================
# Компиляция в ассемблер (Sprint 7)
# ============================================================================

.PHONY: compile
compile: build
	@echo "$(YELLOW)Компиляция factorial.src в ассемблер...$(NC)"
	$(OUTPUT_FULL) compile --input examples/factorial.src --output build/program.asm
	@echo "$(GREEN)Ассемблерный код: build/program.asm$(NC)"

.PHONY: compile-opt
compile-opt: build
	@echo "$(YELLOW)Компиляция с оптимизациями factorial.src...$(NC)"
	$(OUTPUT_FULL) compile --input examples/factorial.src --output build/program_opt.asm --optimize
	@echo "$(GREEN)Оптимизированный ассемблерный код: build/program_opt.asm$(NC)"

.PHONY: build-asm
build-asm:
	@echo "$(YELLOW)Сборка исполняемого файла...$(NC)"
	@if [ -f build/program.asm ]; then \
		nasm -f elf64 -o build/program.o build/program.asm; \
		nasm -f elf64 -o build/runtime.o src/runtime/runtime.asm; \
		gcc -no-pie -o build/program build/runtime.o build/program.o; \
		echo "$(GREEN)Исполняемый файл: build/program$(NC)"; \
	else \
		echo "$(RED)Сначала выполните: make compile$(NC)"; \
	fi

.PHONY: build-asm-opt
build-asm-opt:
	@echo "$(YELLOW)Сборка оптимизированного исполняемого файла...$(NC)"
	@if [ -f build/program_opt.asm ]; then \
		nasm -f elf64 -o build/program_opt.o build/program_opt.asm; \
		nasm -f elf64 -o build/runtime.o src/runtime/runtime.asm; \
		gcc -no-pie -o build/program_opt build/runtime.o build/program_opt.o; \
		echo "$(GREEN)Исполняемый файл: build/program_opt$(NC)"; \
	else \
		echo "$(RED)Сначала выполните: make compile-opt$(NC)"; \
	fi

# ============================================================================
# Демонстрационная программа (Sprint 7)
# ============================================================================

.PHONY: demo
demo: build
	@echo "$(YELLOW)╔══════════════════════════════════════════╗$(NC)"
	@echo "$(YELLOW)║  MiniCompiler Sprint 7 Demo: Quicksort   ║$(NC)"
	@echo "$(YELLOW)╚══════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(CYAN)1. Компиляция quicksort.src...$(NC)"
	$(OUTPUT_FULL) compile --input examples/quicksort.src --output build/quicksort.asm --optimize
	@echo ""
	@echo "$(CYAN)2. Ассемблирование...$(NC)"
	nasm -f elf64 -o build/quicksort.o build/quicksort.asm
	nasm -f elf64 -o build/runtime.o src/runtime/runtime.asm
	@echo ""
	@echo "$(CYAN)3. Линковка с libc...$(NC)"
	gcc -no-pie -o build/quicksort build/runtime.o build/quicksort.o
	@echo ""
	@echo "$(CYAN)4. Запуск программы...$(NC)"
	@echo "$(GREEN)──────────────────────────────────────────$(NC)"
	./build/quicksort
	@echo "$(GREEN)──────────────────────────────────────────$(NC)"
	@echo ""
	@echo "$(GREEN)Демонстрация завершена успешно!$(NC)"

.PHONY: demo-verbose
demo-verbose: build
	@echo "$(YELLOW)Демонстрация с подробным выводом...$(NC)"
	@echo ""
	@echo "$(CYAN)Исходный код:$(NC)"
	@cat examples/quicksort.src
	@echo ""
	@echo "$(CYAN)Генерация IR...$(NC)"
	$(OUTPUT_FULL) ir --input examples/quicksort.src --stats
	@echo ""
	@echo "$(CYAN)Компиляция...$(NC)"
	$(OUTPUT_FULL) compile --input examples/quicksort.src --output build/quicksort.asm --optimize
	@echo ""
	@echo "$(CYAN)Сгенерированный ассемблер (первые 60 строк):$(NC)"
	@head -60 build/quicksort.asm
	@echo "..."
	@echo ""
	@echo "$(CYAN)Сборка и запуск...$(NC)"
	nasm -f elf64 -o build/quicksort.o build/quicksort.asm
	nasm -f elf64 -o build/runtime.o src/runtime/runtime.asm
	gcc -no-pie -o build/quicksort build/runtime.o build/quicksort.o
	@echo "$(GREEN)Вывод программы:$(NC)"
	./build/quicksort

# ============================================================================
# Тестирование
# ============================================================================

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

.PHONY: test-control-flow-verbose
test-control-flow-verbose: build
	@echo "$(YELLOW)Запуск тестов Control Flow (verbose)...$(NC)"
	@if [ -f tests/test_runner/run_tests.sh ]; then \
		cd tests/test_runner && bash run_tests.sh control-flow true; \
	fi

.PHONY: test-go
test-go:
	@echo "$(YELLOW)Запуск Go unit тестов...$(NC)"
	$(GO) test ./src/internal/... -v

.PHONY: test-all
test-all: build test-go
	@echo "$(YELLOW)Запуск integration тестов...$(NC)"
	@if [ -f tests/test_runner/run_tests.sh ]; then \
		cd tests/test_runner && bash run_tests.sh all; \
	fi

# ============================================================================
# Golden testing
# ============================================================================

.PHONY: golden-generate
golden-generate: build
	@echo "$(YELLOW)Генерация golden testing файлов...$(NC)"
	@bash scripts/generate_golden_ir.sh
	@echo "$(GREEN)Golden testing файлы сгенерированы$(NC)"

.PHONY: golden-update
golden-update: build
	@echo "$(YELLOW)Обновление golden testing файлов...$(NC)"
	@bash scripts/update_golden.sh update
	@echo "$(GREEN)Golden testing файлы обновлены$(NC)"

.PHONY: golden-check
golden-check: build
	@echo "$(YELLOW)Проверка golden testing файлов...$(NC)"
	@bash scripts/update_golden.sh check verbose
	@echo "$(GREEN)Проверка завершена$(NC)"

.PHONY: golden-validate
golden-validate:
	@echo "$(YELLOW)Валидация структуры golden testing файлов...$(NC)"
	@bash scripts/validate_golden.sh
	@echo "$(GREEN)Валидация завершена$(NC)"

.PHONY: golden-clean
golden-clean:
	@echo "$(YELLOW)Очистка golden testing файлов...$(NC)"
	@bash scripts/generate_golden_ir.sh --clean
	@echo "$(GREEN)Golden testing файлы удалены$(NC)"

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
	$(GO) clean
	@echo "$(GREEN)Очистка завершена$(NC)"

.PHONY: clean-all
clean-all: clean
	@echo "$(YELLOW)Очистка кэша Go...$(NC)"
	$(GO) clean -cache -modcache -i -r
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
# Генерация AST изображений
# ============================================================================

.PHONY: generate-ast-png
generate-ast-png: build
	@echo "$(YELLOW)Генерация AST для examples/factorial.src...$(NC)"
	$(OUTPUT_FULL) parse --input examples/factorial.src --format dot --output ast.dot
	@if command -v dot > /dev/null; then \
		dot -Tpng ast.dot -o ast.png; \
		echo "$(GREEN)AST сохранён в ast.png$(NC)"; \
	fi

.PHONY: generate-all-asts
generate-all-asts: build
	@echo "$(YELLOW)Генерация AST для всех примеров...$(NC)"
	@for srcfile in examples/*.src; do \
		name=$$(basename "$$srcfile" .src); \
		echo "  $$srcfile -> ast_$$name.dot"; \
		$(OUTPUT_FULL) parse --input "$$srcfile" --format dot --output "ast_$$name.dot"; \
		if command -v dot > /dev/null; then \
			dot -Tpng "ast_$$name.dot" -o "ast_$$name.png"; \
			echo "  ast_$$name.png создан"; \
		fi \
	done
	@echo "$(GREEN)Все AST сгенерированы$(NC)"

# ============================================================================
# Установка и запуск
# ============================================================================

.PHONY: install
install: build
	@echo "$(YELLOW)Установка компилятора в GOPATH/bin...$(NC)"
	$(GO) install $(MAIN_FILE)
	@echo "$(GREEN)Установка завершена$(NC)"

.PHONY: dev
dev: build
	@echo "$(YELLOW)Запуск в режиме разработки...$(NC)"
	$(OUTPUT_FULL) ir --input examples/factorial.src --optimize --stats

# ============================================================================
# Справка
# ============================================================================

.PHONY: help
help:
	@echo "$(BLUE)MiniCompiler - Доступные цели$(NC)"
	@echo ""
	@echo "$(YELLOW)Сборка и запуск:$(NC)"
	@echo "  make build           - Собрать компилятор"
	@echo "  make run             - Запустить парсер на examples/factorial.src"
	@echo "  make run-lex         - Запустить лексер на examples/hello.src"
	@echo "  make run-parse       - Запустить парсер на examples/struct.src"
	@echo "  make check           - Запустить семантический анализ"
	@echo "  make check-types     - Семантический анализ с выводом типов"
	@echo "  make symbols         - Вывести таблицу символов"
	@echo "  make compile         - Скомпилировать в ассемблер"
	@echo "  make compile-opt     - Скомпилировать с оптимизациями"
	@echo "  make build-asm       - Собрать исполняемый файл"
	@echo "  make build-asm-opt   - Собрать оптимизированный исполняемый файл"
	@echo "  make install         - Установить в GOPATH/bin"
	@echo "  make dev             - Запуск в режиме разработки"
	@echo ""
	@echo "$(YELLOW)Демонстрация (Sprint 7):$(NC)"
	@echo "  make demo            - Запустить демо (Quicksort)"
	@echo "  make demo-verbose    - Демо с подробным выводом"
	@echo ""
	@echo "$(YELLOW)Генерация IR:$(NC)"
	@echo "  make ir              - Сгенерировать IR для factorial.src"
	@echo "  make ir-opt          - Сгенерировать оптимизированный IR"
	@echo "  make ir-dot          - Сгенерировать CFG в PNG"
	@echo "  make ir-json         - Сгенерировать IR в JSON"
	@echo "  make ir-stats        - Показать статистику IR"
	@echo ""
	@echo "$(YELLOW)Тестирование:$(NC)"
	@echo "  make test                  - Запустить все тесты"
	@echo "  make test-lexer            - Запустить только тесты лексера"
	@echo "  make test-parser           - Запустить только тесты парсера"
	@echo "  make test-semantic         - Запустить только семантические тесты"
	@echo "  make test-ir               - Запустить тесты IR"
	@echo "  make test-codegen          - Запустить тесты кодогенерации"
	@echo "  make test-control-flow     - Тесты Control Flow"
	@echo "  make test-control-flow-verbose - Тесты Control Flow подробно"
	@echo "  make test-go               - Запустить Go unit тесты"
	@echo "  make test-all              - Запустить все тесты (unit + integration)"
	@echo ""
	@echo "$(YELLOW)Golden testing:$(NC)"
	@echo "  make golden-generate - Сгенерировать expected файлы"
	@echo "  make golden-update   - Обновить expected файлы"
	@echo "  make golden-check    - Проверить expected файлы"
	@echo "  make golden-validate - Валидировать структуру"
	@echo "  make golden-clean    - Удалить expected файлы"
	@echo ""
	@echo "$(YELLOW)Генерация изображений AST:$(NC)"
	@echo "  make generate-ast-png   - Сгенерировать PNG для factorial.src"
	@echo "  make generate-all-asts  - Сгенерировать PNG для всех примеров"
	@echo ""
	@echo "$(YELLOW)Качество кода:$(NC)"
	@echo "  make fmt              - Форматировать код"
	@echo "  make vet              - Запустить статический анализатор"
	@echo "  make lint             - Запустить линтер"
	@echo ""
	@echo "$(YELLOW)Управление зависимостями:$(NC)"
	@echo "  make deps             - Загрузить зависимости"
	@echo "  make tidy             - Очистить зависимости"
	@echo ""
	@echo "$(YELLOW)Очистка:$(NC)"
	@echo "  make clean            - Очистить артефакты сборки"
	@echo "  make clean-all        - Полная очистка (включая кэш Go)"

# ============================================================================
# Специфичные для Windows настройки
# ============================================================================

ifeq ($(OS),Windows_NT)
run:
	$(OUTPUT_FULL) parse --input examples\factorial.src

run-lex:
	$(OUTPUT_FULL) lex --input examples\hello.src

run-parse:
	$(OUTPUT_FULL) parse --input examples\struct.src

check:
	$(OUTPUT_FULL) check --input examples\factorial.src --verbose

symbols:
	$(OUTPUT_FULL) symbols --input examples\factorial.src

ir:
	$(OUTPUT_FULL) ir --input examples\factorial.src

ir-opt:
	$(OUTPUT_FULL) ir --input examples\factorial.src --optimize --stats

compile:
	$(OUTPUT_FULL) compile --input examples\factorial.src --output build\program.asm

test:
	@if exist tests\test_runner\run_tests.bat ( \
		cd tests\test_runner && run_tests.bat \
	) else ( \
		echo Тестовый раннер не найден \
	)

test-lexer:
	@if exist tests\test_runner\run_tests.bat ( \
		cd tests\test_runner && run_tests.bat lexer \
	)

test-parser:
	@if exist tests\test_runner\run_tests.bat ( \
		cd tests\test_runner && run_tests.bat parser \
	)

test-semantic:
	@if exist tests\test_runner\run_tests.bat ( \
		cd tests\test_runner && run_tests.bat semantic \
	)
endif