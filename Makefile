.PHONY: build run test cover cover-html lint clean
.PHONY: docker-up docker-down docker-reset docker-logs
.PHONY: migrate-up migrate-down migrate-status
.PHONY: dev stop accrual

# ==========================================
# Подхватываем .env файл
# ==========================================
-include .env
export

# ==========================================
# Определение ОС для accrual
# ==========================================
ifeq ($(OS),Windows_NT)
	ACCRUAL_BIN := ./cmd/accrual/accrual_windows_amd64.exe
else
	UNAME_S := $(shell uname -s)
	UNAME_M := $(shell uname -m)
	ifeq ($(UNAME_S),Darwin)
		ifeq ($(UNAME_M),arm64)
			ACCRUAL_BIN := ./cmd/accrual/accrual_darwin_arm64
		else
			ACCRUAL_BIN := ./cmd/accrual/accrual_darwin_amd64
		endif
	else
		ACCRUAL_BIN := ./cmd/accrual/accrual_linux_amd64
	endif
endif

# ==========================================
# Переменные по умолчанию
# ==========================================
APP_NAME ?= loyalty-service
CMD_DIR ?= ./cmd/gophermart
ACCRUAL_ADDR ?= http://localhost:8081
RUN_ADDRESS ?= localhost:8080

# ==========================================
# Сборка
# ==========================================
build:
	go build -o $(APP_NAME) $(CMD_DIR)

# ==========================================
# Запуск
# ==========================================
run:
	go run $(CMD_DIR) -d "$(DATABASE_URI)" -r "$(ACCRUAL_ADDR)" -a "$(RUN_ADDRESS)"

# ==========================================
# Тестирование
# ==========================================
test:
	go test -v ./...

cover:
	go test -cover ./...

cover-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# ==========================================
# Линтер
# ==========================================
lint:
	golangci-lint run ./...

# ==========================================
# Миграции
# ==========================================
migrate-up:
	migrate -path migrations -database "$(DATABASE_URI)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URI)" down

migrate-status:
	migrate -path migrations -database "$(DATABASE_URI)" version

# ==========================================
# Docker (БД)
# ==========================================
docker-up:
	docker compose up -d postgres

docker-down:
	docker compose down

docker-reset:
	docker compose down -v

docker-logs:
	docker compose logs -f postgres

# ==========================================
# Accrual (автоопределение ОС)
# ==========================================
accrual:
	chmod +x $(ACCRUAL_BIN) 2>/dev/null || true
	RUN_ADDRESS=:8081 $(ACCRUAL_BIN)

# ==========================================
# Окружение для разработки
# ==========================================
dev: docker-up
	@echo "========================================="
	@echo "База данных запущена на localhost:5432"
	@echo ""
	@echo "Теперь запусти сервер accrual:"
	@echo "  make accrual"
	@echo ""
	@echo "И приложение в отдельном терминале:"
	@echo "  make run"
	@echo "========================================="

# ==========================================
# Остановка всего
# ==========================================
stop: docker-down
	@echo "Останавливаю accrual..."
	-pkill -f accrual 2>/dev/null || true
	@echo "Всё остановлено"

# ==========================================
# Очистка
# ==========================================
clean:
	rm -f $(APP_NAME)
	rm -f coverage.out