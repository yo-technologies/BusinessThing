.PHONY: help build run migrate-up migrate-down proto clean docker-build docker-up docker-down

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	go build -o bin/core-service ./cmd/core-service
	go build -o bin/migrate ./cmd/migrate

run: ## Run the application
	go run ./cmd/core-service/main.go

migrate-up: ## Run database migrations up
	go run ./cmd/migrate/main.go -direction up

migrate-down: ## Run database migrations down
	go run ./cmd/migrate/main.go -direction down

proto: ## Generate protobuf files
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
		--grpc-gateway_opt=generate_unbound_methods=true \
		--validate_out="lang=go,paths=source_relative:." \
		--openapiv2_out=. --openapiv2_opt=allow_merge=true,merge_file_name=core \
		api/core/core.proto

clean: ## Clean build artifacts
	rm -rf bin/

docker-build: ## Build Docker image
	docker-compose build

docker-up: ## Start services with Docker Compose
	docker-compose up -d

docker-down: ## Stop services with Docker Compose
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f core-service

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

lint: ## Run linter
	golangci-lint run

.DEFAULT_GOAL := help
