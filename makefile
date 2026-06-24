include .env
export

APP_NAME=cdc_audit_timeline_service

.PHONY: up down restart logs build run test integration-test lint fmt

DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
TEST_DB_URL=postgres://$(TEST_DB_USER):$(TEST_DB_PASSWORD)@$(TEST_DB_HOST):$(TEST_DB_PORT)/$(TEST_DB_NAME)?sslmode=disable
DB_MIGRATIONS_DIR=./db/migrations
DB_DRIVER=postgres

# Docker commands
# make up
up:
	docker compose up -d

# make down
down:
	docker compose down

# make restart
restart:
	docker compose down
	docker compose up -d

# make logs
logs:
	docker compose logs -f

# make rebuild
rebuild:
	docker compose up --build

# Go commands
# make run
run:
	go run ./cmd/api

# make run-consumer
run-consumer:
	go run ./cmd/consumer

# make test
test:
	go test ./...

# make integration-test
integration-test:
	go test -p 1 -tags integration ./...

# make lint
lint:
	golangci-lint run

# make fmt
fmt:
	go fmt ./...

# make build
build:
	go build -o bin/api ./cmd/api
	go build -o bin/consumer ./cmd/consumer

# DB commands
# make migrate-create name=create_user_table
migrate-create:
	goose -dir $(DB_MIGRATIONS_DIR) create $(name) sql

# make migrate-up-test
migrate-up-test:
	goose -dir $(DB_MIGRATIONS_DIR) $(DB_DRIVER) "$(TEST_DB_URL)" up

# make migrate-up
migrate-up:
	goose -dir $(DB_MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_URL)" up

# make migrate-down
migrate-down:
	goose -dir $(DB_MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_URL)" down
