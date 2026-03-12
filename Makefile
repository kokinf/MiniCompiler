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
else
	EXT = .exe
	RM = del /Q /F
	RMDIR = rmdir /S /Q
	MKDIR = mkdir
endif

# Полный путь к выходному файлу с расширением
OUTPUT_FULL = $(OUTPUT)$(EXT)

# Цвета для вывода
GREEN = \033[0;32m
RED = \033[0;31m
YELLOW = \033[1;33m
NC = \033[0m

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

.PHONY: test
test: build
	@echo -e "$(YELLOW)Запуск всех тестов...$(NC)"
	@cd tests/test_runner && ./run_tests.sh

.PHONY: test-lexer
test-lexer: build
	@echo -e "$(YELLOW)Запуск тестов лексера...$(NC)"
	@cd tests/test_runner && ./run_tests.sh lexer

.PHONY: test-parser
test-parser: build
	@echo -e "$(YELLOW)Запуск тестов парсера...$(NC)"
	@cd tests/test_runner && ./run_tests.sh parser

.PHONY: test-valid
test-valid: build
	@echo -e "$(YELLOW)Запуск валидных тестов...$(NC)"
	@cd tests/test_runner && ./run_tests.sh valid

.PHONY: test-invalid
test-invalid: build
	@echo -e "$(YELLOW)Запуск невалидных тестов...$(NC)"
	@cd tests/test_runner && ./run_tests.sh invalid

.PHONY: deps
deps:
	@echo -e "$(YELLOW)Загрузка зависимостей...$(NC)"
	$(GO) mod download
	$(GO) mod verify
	@echo -e "$(GREEN)Зависимости загружены$(NC)"

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

.PHONY: help
help:
	@echo "Доступные цели:"
	@echo ""
	@echo "Сборка:"
	@echo "  make build        - Собрать компилятор"
	@echo "  make clean        - Очистить артефакты сборки"
	@echo "  make clean-all    - Полная очистка (включая кэш Go)"
	@echo "  make deps         - Загрузить зависимости"
	@echo "  make all          - Очистка, зависимости и сборка"
	@echo ""
	@echo "Запуск:"
	@echo "  make run          - Запустить парсер на examples/factorial.src"
	@echo "  make run-lex      - Запустить лексер на examples/hello.src"
	@echo "  make run-parse    - Запустить парсер на examples/struct.src"
	@echo ""
	@echo "Тестирование:"
	@echo "  make test         - Запустить все тесты (лексер + парсер)"
	@echo "  make test-lexer   - Запустить только тесты лексера"
	@echo "  make test-parser  - Запустить только тесты парсера"
	@echo "  make test-valid   - Запустить только валидные тесты"
	@echo "  make test-invalid - Запустить только невалидные тесты"
	@echo ""
	@echo "Качество кода:"
	@echo "  make fmt          - Форматировать код"
	@echo "  make vet          - Запустить статический анализатор"
	@echo "  make lint         - Запустить линтер (требуется golint)"

# Специфичные для Windows настройки
ifeq ($(OS),Windows_NT)
run:
	$(OUTPUT_FULL) parse --input examples\factorial.src

run-lex:
	$(OUTPUT_FULL) lex --input examples\hello.src

run-parse:
	$(OUTPUT_FULL) parse --input examples\struct.src

test:
	cd tests\test_runner && run_tests.bat

test-lexer:
	cd tests\test_runner && run_tests.bat lexer

test-parser:
	cd tests\test_runner && run_tests.bat parser
endif