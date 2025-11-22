.env:
	@touch .env

include .env
export

.PHONY: all

all: help

# --- Application Commands ---
.PHONY: run migrate-up migrate-down migrate-create

run: .env
	@echo "Trying to start application..."
	@go run cmd/Avito-review/main.go

migrate-up: .env
	@echo "Applying database migrations..."
	@migrate -path=${MIGRATIONS_PATH} -database "${PG_URL}" -verbose up

migrate-down: .env
	@echo "Reverting database migrations..."
	@migrate -path=${MIGRATIONS_PATH} -database "${PG_URL}" -verbose down

migrate-create: .env
	@echo "Creating new migration file..."
	@migrate create -ext=sql -dir=${MIGRATIONS_PATH} -seq init

# --- Code Generation ---
.PHONY: generate-dto generate-swagger

generate-dto:
	@echo "Generating DTO types from OpenAPI spec..."
	@oapi-codegen -generate types -package dto -o internal/dto/types.gen.go api/openapi.yml

generate-swagger:
	@echo "Generating Swagger documentation..."
	@swag init --generalInfo internal/controller/http/router.go --output ./docs --parseDependency --parseInternal

help:
	@echo "Available commands:"
	@echo ""
	@echo "Application:"
	@echo "  make run              - Start the application"
	@echo "  make migrate-up       - Apply database migrations"
	@echo "  make migrate-down     - Revert database migrations"
	@echo "  make migrate-create   - Create new migration file"
	@echo ""
	@echo "Code Generation:"
	@echo "  make generate-dto     - Generate DTO types"
	@echo "  make generate-swagger - Generate Swagger docs"
	@echo ""
	@echo "Utilities:"
	@echo "  make help             - Show this help"