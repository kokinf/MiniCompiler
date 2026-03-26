# Конфигурация проекта
APP_NAME = compiler
MAIN_FILE = ./src/cmd/compiler/main.go
OUTPUT_DIR = ./bin
OUTPUT = $(OUTPUT_DIR)/$(APP_NAME)

# Команды Go
GO = go
GOFLAGS = -ldflags="-s -w"
GOTESTFLAGS = -v

# Определение ОС
UNAME_S := $(shell uname -s 2>/dev/null || echo Windows)
ifeq ($(UNAME_S),Linux)
	EXT = 
	RM = rm -f
	RMDIR = rm -rf
	MKDIR = mkdir -p
	SEP = /
else
	EXT = .exe
	RM = del /Q /F
	RMDIR = rmdir /S /Q
	MKDIR = mkdir
	SEP = \\
endif

# Полный путь к выходному файлу с расширением
OUTPUT_FULL = $(OUTPUT)$(EXT)

# Цвета для вывода (только для Unix)
ifeq ($(UNAME_S),Linux)
	GREEN = \033[0;32m
	RED = \033[0;31m
	YELLOW = \033[1;33m
	BLUE = \033[0;34m
	NC = \033[0m
else
	GREEN = 
	RED = 
	YELLOW = 
	BLUE = 
	NC = 
endif

# ============================================================================
# Основные цели
# ============================================================================

.PHONY: all
all: clean deps build
	@echo -e "$(GREEN)Проект успешно собран!$(NC)"

.PHONY: build
build: $(OUTPUT_DIR)
	@echo -e "$(YELLOW)Сборка компилятора...$(NC)"
	$(GO) build $(GOFLAGS) -o $(OUTPUT_FULL) $(MAIN_FILE)
	@echo -e "$(GREEN)Сборка завершена: $(OUTPUT_FULL)$(NC)"

$(OUTPUT_DIR):
	$(MKDIR) $(OUTPUT_DIR)

.PHONY: run
run: build
	@echo -e "$(YELLOW)Запуск парсера на examples/factorial.src...$(NC)"
	$(OUTPUT_FULL) parse --input examples/factorial.src

.PHONY: run-lex
run-lex: build
	@echo -e "$(YELLOW)Запуск лексера на examples/hello.src...$(NC)"
	$(OUTPUT_FULL) lex --input examples/hello.src

.PHONY: run-parse
run-parse: build
	@echo -e "$(YELLOW)Запуск парсера на examples/struct.src...$(NC)"
	$(OUTPUT_FULL) parse --input examples/struct.src

.PHONY: check
check: build
	@echo -e "$(YELLOW)Запуск семантического анализа на examples/factorial.src...$(NC)"
	$(OUTPUT_FULL) check --input examples/factorial.src --verbose

.PHONY: check-types
check-types: build
	@echo -e "$(YELLOW)Запуск семантического анализа с выводом типов...$(NC)"
	$(OUTPUT_FULL) check --input examples/factorial.src --verbose --show-types

.PHONY: symbols
symbols: build
	@echo -e "$(YELLOW)Вывод таблицы символов для examples/factorial.src...$(NC)"
	$(OUTPUT_FULL) symbols --input examples/factorial.src

.PHONY: symbols-json
symbols-json: build
	@echo -e "$(YELLOW)Вывод таблицы символов в JSON формате...$(NC)"
	$(OUTPUT_FULL) symbols --input examples/factorial.src --format json

# ============================================================================
# Тестирование
# ============================================================================

.PHONY: test
test: build
	@echo -e "$(YELLOW)Запуск всех тестов (лексер + парсер + семантический)...$(NC)"
	@cd tests/test_runner && ./run_tests.sh all

.PHONY: test-lexer
test-lexer: build
	@echo -e "$(YELLOW)Запуск тестов лексера...$(NC)"
	@cd tests/test_runner && ./run_tests.sh lexer

.PHONY: test-parser
test-parser: build
	@echo -e "$(YELLOW)Запуск тестов парсера...$(NC)"
	@cd tests/test_runner && ./run_tests.sh parser

.PHONY: test-semantic
test-semantic: build
	@echo -e "$(YELLOW)Запуск семантических тестов...$(NC)"
	@cd tests/test_runner && ./run_tests.sh semantic

.PHONY: test-valid
test-valid: build
	@echo -e "$(YELLOW)Запуск валидных тестов...$(NC)"
	@cd tests/test_runner && ./run_tests.sh valid

.PHONY: test-invalid
test-invalid: build
	@echo -e "$(YELLOW)Запуск невалидных тестов...$(NC)"
	@cd tests/test_runner && ./run_tests.sh invalid

# ============================================================================
# Управление зависимостями
# ============================================================================

.PHONY: deps
deps:
	@echo -e "$(YELLOW)Загрузка зависимостей...$(NC)"
	$(GO) mod download
	$(GO) mod verify
	@echo -e "$(GREEN)Зависимости загружены$(NC)"

.PHONY: tidy
tidy:
	@echo -e "$(YELLOW)Очистка зависимостей...$(NC)"
	$(GO) mod tidy
	@echo -e "$(GREEN)Зависимости очищены$(NC)"

# ============================================================================
# Очистка
# ============================================================================

.PHONY: clean
clean:
	@echo -e "$(YELLOW)Очистка артефактов сборки...$(NC)"
	$(RM) $(OUTPUT_FULL)
	$(RMDIR) $(OUTPUT_DIR)
	$(GO) clean
	@echo -e "$(GREEN)Очистка завершена$(NC)"

.PHONY: clean-all
clean-all: clean
	@echo -e "$(YELLOW)Очистка кэша Go...$(NC)"
	$(GO) clean -cache -modcache -i -r
	@echo -e "$(GREEN)Полная очистка завершена$(NC)"

# ============================================================================
# Качество кода
# ============================================================================

.PHONY: fmt
fmt:
	@echo -e "$(YELLOW)Форматирование кода...$(NC)"
	$(GO) fmt ./...
	@echo -e "$(GREEN)Форматирование завершено$(NC)"

.PHONY: vet
vet:
	@echo -e "$(YELLOW)Проверка кода статическим анализатором...$(NC)"
	$(GO) vet ./...
	@echo -e "$(GREEN)Проверка завершена$(NC)"

.PHONY: lint
lint: vet
	@echo -e "$(YELLOW)Запуск линтера...$(NC)"
	@which golint > /dev/null && golint ./... || echo "golint не установлен"

# ============================================================================
# Генерация AST изображений
# ============================================================================

.PHONY: dot-to-png
dot-to-png:
	@echo -e "$(YELLOW)Конвертация DOT в PNG...$(NC)"
	@for dotfile in *.dot; do \
		if [ -f "$$dotfile" ]; then \
			pngfile="$${dotfile%.dot}.png"; \
			echo "  $$dotfile -> $$pngfile"; \
			dot -Tpng "$$dotfile" -o "$$pngfile"; \
		fi \
	done
	@echo -e "$(GREEN)Конвертация завершена$(NC)"

.PHONY: generate-ast-png
generate-ast-png: build
	@echo -e "$(YELLOW)Генерация AST для examples/factorial.src...$(NC)"
	$(OUTPUT_FULL) parse --input examples/factorial.src --format dot --output ast.dot
	dot -Tpng ast.dot -o ast.png
	@echo -e "$(GREEN)AST сохранён в ast.png$(NC)"

.PHONY: generate-all-asts
generate-all-asts: build
	@echo -e "$(YELLOW)Генерация AST для всех примеров...$(NC)"
	@for srcfile in examples/*.src; do \
		name=$$(basename "$$srcfile" .src); \
		echo "  $$srcfile -> ast_$$name.dot"; \
		$(OUTPUT_FULL) parse --input "$$srcfile" --format dot --output "ast_$$name.dot"; \
		dot -Tpng "ast_$$name.dot" -o "ast_$$name.png"; \
		echo "  ast_$$name.png создан"; \
	done
	@echo -e "$(GREEN)Все AST сгенерированы$(NC)"

# ============================================================================
# Примеры и демонстрация
# ============================================================================

.PHONY: demo-lexer
demo-lexer: build
	@echo -e "$(BLUE)=== Демонстрация лексера ===$(NC)"
	@echo -e "$(YELLOW)Файл: examples/hello.src$(NC)"
	$(OUTPUT_FULL) lex --input examples/hello.src

.PHONY: demo-parser
demo-parser: build
	@echo -e "$(BLUE)=== Демонстрация парсера ===$(NC)"
	@echo -e "$(YELLOW)Файл: examples/factorial.src$(NC)"
	$(OUTPUT_FULL) parse --input examples/factorial.src

.PHONY: demo-semantic
demo-semantic: build
	@echo -e "$(BLUE)=== Демонстрация семантического анализа ===$(NC)"
	@echo -e "$(YELLOW)Файл: examples/factorial.src$(NC)"
	$(OUTPUT_FULL) check --input examples/factorial.src --verbose --show-types

.PHONY: demo-all
demo-all: demo-lexer demo-parser demo-semantic
	@echo -e "$(GREEN)Все демонстрации завершены$(NC)"

# ============================================================================
# Установка и запуск
# ============================================================================

.PHONY: install
install: build
	@echo -e "$(YELLOW)Установка компилятора в GOPATH/bin...$(NC)"
	$(GO) install $(MAIN_FILE)
	@echo -e "$(GREEN)Установка завершена$(NC)"

.PHONY: dev
dev: build
	@echo -e "$(YELLOW)Запуск в режиме разработки...$(NC)"
	$(OUTPUT_FULL) check --input examples/factorial.src --verbose --show-types

# ============================================================================
# Справка
# ============================================================================

.PHONY: help
help:
	@echo -e "$(BLUE)MiniCompiler - Доступные цели$(NC)"
	@echo ""
	@echo -e "$(YELLOW)Сборка и запуск:$(NC)"
	@echo "  make build           - Собрать компилятор"
	@echo "  make run              - Запустить парсер на examples/factorial.src"
	@echo "  make run-lex          - Запустить лексер на examples/hello.src"
	@echo "  make run-parse        - Запустить парсер на examples/struct.src"
	@echo "  make check            - Запустить семантический анализ"
	@echo "  make check-types      - Семантический анализ с выводом типов"
	@echo "  make symbols          - Вывести таблицу символов"
	@echo "  make install          - Установить в GOPATH/bin"
	@echo "  make dev              - Запуск в режиме разработки"
	@echo ""
	@echo -e "$(YELLOW)Тестирование:$(NC)"
	@echo "  make test             - Запустить все тесты"
	@echo "  make test-lexer       - Запустить только тесты лексера"
	@echo "  make test-parser      - Запустить только тесты парсера"
	@echo "  make test-semantic    - Запустить только семантические тесты"
	@echo "  make test-valid       - Запустить только валидные тесты"
	@echo "  make test-invalid     - Запустить только невалидные тесты"
	@echo ""
	@echo -e "$(YELLOW)Генерация изображений AST:$(NC)"
	@echo "  make generate-ast-png   - Сгенерировать PNG для factorial.src"
	@echo "  make generate-all-asts  - Сгенерировать PNG для всех примеров"
	@echo ""
	@echo -e "$(YELLOW)Качество кода:$(NC)"
	@echo "  make fmt              - Форматировать код"
	@echo "  make vet              - Запустить статический анализатор"
	@echo "  make lint             - Запустить линтер (требуется golint)"
	@echo ""
	@echo -e "$(YELLOW)Управление зависимостями:$(NC)"
	@echo "  make deps             - Загрузить зависимости"
	@echo "  make tidy             - Очистить зависимости"
	@echo ""
	@echo -e "$(YELLOW)Очистка:$(NC)"
	@echo "  make clean            - Очистить артефакты сборки"
	@echo "  make clean-all        - Полная очистка (включая кэш Go)"
	@echo ""
	@echo -e "$(YELLOW)Демонстрация:$(NC)"
	@echo "  make demo-lexer       - Показать работу лексера"
	@echo "  make demo-parser      - Показать работу парсера"
	@echo "  make demo-semantic    - Показать работу семантического анализатора"
	@echo "  make demo-all         - Показать все этапы"

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

test:
	cd tests\test_runner && run_tests.bat

test-lexer:
	cd tests\test_runner && run_tests.bat lexer

test-parser:
	cd tests\test_runner && run_tests.bat parser

test-semantic:
	cd tests\test_runner && run_tests.bat semantic
endif