# Конфигурация проекта
APP_NAME = compiler
MAIN_FILE = ./cmd/compiler/main.go
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


.PHONY: all
all: clean deps build

.PHONY: build
build: $(OUTPUT_DIR)
	$(GO) build $(GOFLAGS) -o $(OUTPUT_FULL) $(MAIN_FILE)
	@echo "Сборка завершена: $(OUTPUT_FULL)"

$(OUTPUT_DIR):
	$(MKDIR) $(OUTPUT_DIR)

.PHONY: run
run: build
	$(OUTPUT_FULL) lex --input examples/hello.src

.PHONY: test
test: build
	cd tests/test_runner && ./run_tests.sh

.PHONY: deps
deps:
	$(GO) mod download
	$(GO) mod verify
	@echo "Зависимости загружены"

.PHONY: clean
clean:
	$(RM) $(OUTPUT_FULL)
	$(RMDIR) $(OUTPUT_DIR)
	$(GO) clean
	@echo "Очистка завершена"

.PHONY: help
help:
	@echo "Доступные цели:"
	@echo "  make build  - Собрать компилятор"
	@echo "  make run    - Запустить компилятор на примере"
	@echo "  make test   - Запустить все тесты"
	@echo "  make clean  - Очистить артефакты сборки"
	@echo "  make deps   - Загрузить зависимости"
	@echo "  make all    - Очистка, зависимости и сборка"

# Специфичные для Windows настройки
ifeq ($(OS),Windows_NT)
# Переопределение для командной строки Windows
run:
	$(OUTPUT_FULL) lex --input examples\hello.src

test:
	cd tests\test_runner && run_tests.bat
endif