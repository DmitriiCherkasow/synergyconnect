.PHONY: help run build migrate test lint docker-up docker-down clean

help: ## Показать помощь
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## Запустить сервер
	go run cmd/api/main.go

build: ## Собрать бинарный файл
	go build -o bin/api cmd/api/main.go

migrate-up: ## Применить миграции
	migrate -path migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up

migrate-down: ## Откатить миграции
	migrate -path migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" down

test: ## Запустить тесты
	go test -v -cover ./...

test-coverage: ## Запустить тесты с покрытием
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Запустить линтер
	golangci-lint run ./...

fmt: ## Форматировать код
	go fmt ./...

tidy: ## Очистить и обновить зависимости
	go mod tidy
	go mod verify

docker-up: ## Запустить Docker Compose
	docker-compose up -d

docker-down: ## Остановить Docker Compose
	docker-compose down

docker-logs: ## Посмотреть логи Docker
	docker-compose logs -f

clean: ## Очистить бинарные файлы
	rm -rf bin/
	rm -rf coverage.html coverage.out

# Установка зависимостей (будет использоваться в CI)
install-deps:
	go mod download

# Все команды для CI (будет использовать GitHub Actions)
ci: tidy lint test build