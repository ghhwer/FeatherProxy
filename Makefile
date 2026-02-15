.PHONY: run build test cov docker-build docker-run docker-publish

APP_DIR := app
BINARY := featherproxy
DOCKER_IMAGE ?= featherproxy:latest

run:
	cd $(APP_DIR) && go run main.go

build:
	cd $(APP_DIR) && go build -o $(BINARY) .
	@echo "Binary: $(APP_DIR)/$(BINARY)"

test:
	cd $(APP_DIR) && go test -v ./...

cov:
	cd $(APP_DIR) && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: $(APP_DIR)/coverage.html"

# Docker: build image (default tag featherproxy:latest)
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Docker: run container (publishes UI on 4545; mount app/ for .env and SQLite data if needed)
docker-run:
	docker run --rm -p 4545:4545 $(DOCKER_RUN_ARGS) $(DOCKER_IMAGE)

# Docker: build, tag as IMAGE, and push. Example: make docker-publish IMAGE=ghcr.io/myorg/featherproxy:v1.0
docker-publish: docker-build
	@if [ -z "$(IMAGE)" ]; then echo "Usage: make docker-publish IMAGE=<registry>/<repo>:<tag>"; exit 1; fi
	docker tag $(DOCKER_IMAGE) $(IMAGE)
	docker push $(IMAGE)
