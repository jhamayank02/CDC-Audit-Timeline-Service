include .env
export

APP_NAME=cdc_audit_timeline_service

.PHONY: up down restart logs build run test lint fmt

DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
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
	go run main.go

# make test
test:
	go test ./...

# make lint
lint:
	golangci-lint run

# make fmt
fmt:
	go fmt ./...

# make build
build:
	go build

# DB commands
# make migrate-create name=create_user_table
migrate-create:
	goose -dir $(DB_MIGRATIONS_DIR) create $(name) sql

# make migrate-up
migrate-up:
	goose -dir $(DB_MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_URL)" up

# make migrate-down
migrate-down:
	goose -dir $(DB_MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_URL)" down