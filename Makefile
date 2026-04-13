export
BINARY=engine

.PHONY: help models swagger run docker-push docker-login

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

models: ## Generate SQLBoiler models
	@echo "Generating models"
	@sqlboiler psql

swagger: ## Generate Swagger documentation
	@echo "Generating swagger"
	@swag init -g cmd/server/main.go --parseVendor
	@echo "Fixing swagger docs (removing deprecated LeftDelim/RightDelim)..."
	@sed -i '' '/LeftDelim:/d' docs/docs.go
	@sed -i '' '/RightDelim:/d' docs/docs.go

run: swagger ## Run the identity service
	@echo "Running the application"
	@go run cmd/server/main.go

docker-push: ## Build and push image to Harbor registry
	@echo "Building and pushing image to Harbor registry"
	@./scripts/build-api.sh build-push

docker-login: ## Login to Harbor registry
	@./scripts/build-api.sh login

.DEFAULT_GOAL := help
