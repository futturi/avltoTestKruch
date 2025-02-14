DOCKER_COMPOSE_FILE = docker-compose.test.yaml
PKG = ./...

.PHONY: all test unit integration lint coverage clean run

all: test

test: unit

unit:
	@echo "Запуск unit тестов..."
	go test -v -short $(PKG)

integration:
	@echo "Останавливаем предыдущую тестовую БД (если запущена)..."
	-docker compose -f $(DOCKER_COMPOSE_FILE) down
	@echo "Поднимаем тестовую БД через docker compose..."
	docker compose -f $(DOCKER_COMPOSE_FILE) up -d || { echo "Ошибка при поднятии контейнера"; exit 1; }
	@echo "Ожидаем, пока база данных будет готова..."
	@sleep 10
	@echo "Запуск интеграционных тестов..."
	go test -v -tags=integration $(PKG)
	@echo "Останавливаем тестовую БД..."
	docker compose -f $(DOCKER_COMPOSE_FILE) down

lint:
	@echo "Запуск golangci-lint..."
	golangci-lint run

coverage:
	@echo "Запуск тестов с покрытием..."
	go test -coverprofile=coverage.out $(PKG)
	go tool cover -html=coverage.out -o coverage.html
	@echo "Отчёт о покрытии сгенерирован в файле coverage.html"

clean:
	@echo "Очистка кэшей..."
	go clean -cache -testcache -modcache

run:
	@echo "Запуск..."
	docker compose up
