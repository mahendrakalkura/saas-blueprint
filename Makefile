.PHONY: help build up down restart logs clean test migrate-up migrate-down migrate-create seed

# Default target
help:
	@echo "Available commands:"
	@echo "  make build          - Build all Docker images"
	@echo "  make up             - Start all services"
	@echo "  make down           - Stop all services"
	@echo "  make restart        - Restart all services"
	@echo "  make logs           - View logs from all services"
	@echo "  make logs-backend   - View backend logs"
	@echo "  make logs-frontend  - View frontend logs"
	@echo "  make logs-db        - View database logs"
	@echo "  make clean          - Remove all containers, volumes, and images"
	@echo "  make test           - Run all tests"
	@echo "  make test-backend   - Run backend tests"
	@echo "  make test-frontend  - Run frontend tests"
	@echo "  make migrate-up     - Run database migrations"
	@echo "  make migrate-down   - Rollback last migration"
	@echo "  make migrate-create - Create new migration (name=migration_name)"
	@echo "  make seed           - Seed database with test data"
	@echo "  make sqlc           - Generate sqlc code"
	@echo "  make lint           - Run linters"
	@echo "  make fmt            - Format code"
	@echo "  make shell-backend  - Open shell in backend container"
	@echo "  make shell-frontend - Open shell in frontend container"
	@echo "  make shell-db       - Open psql shell in database"

# Docker commands
build:
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

restart:
	docker-compose restart

logs:
	docker-compose logs -f

logs-backend:
	docker-compose logs -f backend

logs-frontend:
	docker-compose logs -f frontend

logs-db:
	docker-compose logs -f db

clean:
	docker-compose down -v --rmi all

# Testing
test: test-backend test-frontend

test-backend:
	docker-compose exec backend go test -v ./...

test-frontend:
	docker-compose exec frontend npm test

# Database migrations
migrate-up:
	docker-compose exec backend goose -dir db/migrations postgres "host=db port=5432 user=$(shell grep DB_USER .env | cut -d '=' -f2) password=$(shell grep DB_PASSWORD .env | cut -d '=' -f2) dbname=$(shell grep DB_NAME .env | cut -d '=' -f2) sslmode=disable" up

migrate-down:
	docker-compose exec backend goose -dir db/migrations postgres "host=db port=5432 user=$(shell grep DB_USER .env | cut -d '=' -f2) password=$(shell grep DB_PASSWORD .env | cut -d '=' -f2) dbname=$(shell grep DB_NAME .env | cut -d '=' -f2) sslmode=disable" down

migrate-create:
ifndef name
	@echo "Error: migration name required. Usage: make migrate-create name=create_users_table"
else
	docker-compose exec backend goose -dir db/migrations create $(name) sql
endif

seed:
	docker-compose exec backend go run scripts/seed.go

# Code generation
sqlc:
	docker-compose exec backend sqlc generate

# Linting and formatting
lint: lint-backend lint-frontend

lint-backend:
	docker-compose exec backend golangci-lint run

lint-frontend:
	docker-compose exec frontend npm run lint

fmt:
	docker-compose exec backend go fmt ./...
	docker-compose exec frontend npm run lint -- --fix

# Shell access
shell-backend:
	docker-compose exec backend sh

shell-frontend:
	docker-compose exec frontend sh

shell-db:
	docker-compose exec db psql -U $(shell grep DB_USER .env | cut -d '=' -f2) -d $(shell grep DB_NAME .env | cut -d '=' -f2)

# Development helpers
install-backend:
	cd backend && go mod download

install-frontend:
	cd frontend && npm install

dev:
	@echo "Starting development environment..."
	@make up
	@echo "Services started! Frontend: http://localhost:3000, Backend: http://localhost:8080"
