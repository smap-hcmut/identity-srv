export
BINARY=engine

models:
	@echo "Generating models"
	@sqlboiler psql

swagger:
	@echo "Generating swagger"
	@swag init -g cmd/api/main.go
	@echo "Fixing swagger docs (removing deprecated LeftDelim/RightDelim)..."
	@sed -i '' '/LeftDelim:/d' docs/docs.go
	@sed -i '' '/RightDelim:/d' docs/docs.go

run-api:
	@echo "Generating swagger"
	@swag init -g cmd/api/main.go --parseVendor
	@sed -i '' '/LeftDelim:/d' docs/docs.go
	@sed -i '' '/RightDelim:/d' docs/docs.go
	@echo "Running the application"
	@go run cmd/api/main.go

run-consumer:
	@echo "Running the consumer"
	@go run cmd/consumer/main.go

build-docker-compose:
	@echo "make models first"
	@make models
	@echo "Building docker compose"
	docker compose up --build -d

# Docker build targets (Zot registry)
docker-push:
	@echo "Building and pushing API image to Zot registry"
	@./scripts/build-api.sh build-push

docker-login:
	@./scripts/build-api.sh login

# Consumer Docker build targets (Zot registry)
consumer-push:
	@echo "Building and pushing Consumer image to Zot registry"
	@./scripts/build-consumer.sh build-push

consumer-login:
	@./scripts/build-consumer.sh login

# Show all available targets
help:
	@echo "Available targets:"
	@echo ""
	@echo "Development:"
	@echo "  models              - Generate SQLBoiler models"
	@echo "  swagger             - Generate Swagger documentation"
	@echo "  run-api             - Run API server locally"
	@echo "  run-consumer        - Run consumer locally"
	@echo "  build-docker-compose - Build with docker-compose"
	@echo ""
	@echo "Docker (Zot registry at 172.16.21.10:5000):"
	@echo "  docker-push         - Build & push API to Zot"
	@echo "  docker-login        - Login to Zot registry"
	@echo "  consumer-push       - Build & push Consumer to Zot"
	@echo "  consumer-login      - Login to Zot registry"
	@echo ""
	@echo "Environment Variables:"
	@echo "  ZOT_REGISTRY   Registry URL     (default: 172.16.21.10:5000)"
	@echo "  ZOT_USERNAME   Registry user    (default: tantai)"
	@echo "  ZOT_PASSWORD   Registry pass    (will prompt if empty)"
	@echo "  PLATFORM       Target platform  (default: linux/amd64)"

.PHONY: models swagger run-api run-consumer build-docker-compose \
        docker-push docker-login \
        consumer-push consumer-login \