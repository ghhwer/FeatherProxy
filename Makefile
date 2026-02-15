.PHONY: run build test cov

APP_DIR := app
BINARY := featherproxy

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
