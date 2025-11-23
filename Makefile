.PHONY: build run test docker-build docker-run compose-up compose-down

# Сборка бинарного файла
build:
	go build -o bin/pr-reviewer-service ./cmd/api/main.go

# Запуск приложения локально (с PostgreSQL, Redis)
run:
	go run ./cmd/api/main.go

# Запуск unit-тестов
test:
	go test ./...

# Сборка Docker-образа
docker-build:
	docker build -t pr-reviewer-service .

# Запуск Docker-контейнера (без инфраструктуры)
docker-run:
	docker run -p 8080:8080 pr-reviewer-service

# Запуск всей инфраструктуры через docker-compose
compose-up:
	docker-compose up -d

# Остановка всей инфраструктуры
compose-down:
	docker-compose down

# Остановка всей инфраструктуры с удалением volumes
compose-down-v:
	docker-compose down -v

# Просмотр логов
logs:
	docker-compose logs -f api

# Запуск миграций вручную (если нужно)
migrate-up:
	docker-compose exec postgres psql -U postgres -d reviewer_db -c "\dt"

# Запуск нагрузочного теста basic
loadtest-basic:
	docker run --network=host -i grafana/k6 run - < ./test/loadtests/basic.js

# Запуск нагрузочного теста edge_cases
loadtest-edge:
	docker run --network=host -i grafana/k6 run - < ./test/loadtests/edge_cases.js

# Запуск нагрузочного теста scenario
loadtest-scenario:
	docker run --network=host -i grafana/k6 run - < ./test/loadtests/scenario.js

# Запуск линтера
lint:
	golangci-lint run

# Запуск линтера с автоисправлением
lint-fix:
	golangci-lint run --fix

# Комплексный запуск нагрузочного тестирования
loadtest-all: compose-up
	@echo "Ожидание запуска сервисов..."
	@sleep 10
	@echo "Запуск нагрузочных тестов..."
	docker run --network=host -i grafana/k6 run - < ./test/loadtests/basic.js
	docker run --network=host -i grafana/k6 run - < ./test/loadtests/edge_cases.js
	docker run --network=host -i grafana/k6 run - < ./test/loadtests/scenario.js