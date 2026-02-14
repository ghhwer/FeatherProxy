.PHONY: run build

APP_DIR := app
BINARY := featherproxy

run:
	cd $(APP_DIR) && go run main.go

build:
	cd $(APP_DIR) && go build -o $(BINARY) .
	@echo "Binary: $(APP_DIR)/$(BINARY)"
