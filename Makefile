.PHONY: build run dev clean migrate-up migrate-down migrate-status migrate-create setup test


APP_NAME=go-board
MIGRATE_NAME=migrate

ORIGINAL_CGO_USAGE_SET := $(shell go env CGO_ENABLED)
ORIGINAL_OS := $(shell go env GOOS)
ORIGINAL_ARCH := $(shell go env GOARCH)

ifeq ($(OS),Windows_NT)
    APP_NAME=go-board
    MIGRATE_NAME=migrate
    FILE_EXTENSION=exe
    RM_CMD=del /q /f
    APP_RUN=.\bin\$(APP_NAME)
    MIGRATE_RUN=.\bin\$(MIGRATE_NAME)
    DETECTED_OS=windows
else
    APP_NAME=go-board
    MIGRATE_NAME=migrate
    RM_CMD=rm -rf
    APP_RUN=./bin/$(APP_NAME)
    MIGRATE_RUN=./bin/$(MIGRATE_NAME)
    DETECTED_OS=$(shell go env GOOS)
endif

DETECTED_ARCH=$(shell go env GOARCH)


build: cgo-disable build-for-current-platform env-restore

build-for-current-platform:
	@echo "Building for detected platform: $(DETECTED_OS)_$(DETECTED_ARCH)"
ifeq ($(DETECTED_OS),windows)
ifeq ($(DETECTED_ARCH),amd64)
	@$(MAKE) build-windows-amd64
endif
else ifeq ($(DETECTED_OS),linux)
ifeq ($(DETECTED_ARCH),amd64)
	@$(MAKE) build-linux-amd64
else ifeq ($(DETECTED_ARCH),arm64)
	@$(MAKE) build-linux-arm64
endif
else
	@echo "Unknown platform: $(DETECTED_OS)_$(DETECTED_ARCH), falling back to Linux AMD64"
	@$(MAKE) build-linux-amd64
endif


dist: cgo-disable build-linux-amd64 build-linux-arm64 build-windows-amd64 env-restore

cgo-disable:
	@echo "Building application..."
	@go env -w CGO_ENABLED=0

env-restore:
	@go env -w GOOS=$(ORIGINAL_OS)
	@go env -w GOARCH=$(ORIGINAL_ARCH)
	@go env -w CGO_ENABLED=$(ORIGINAL_CGO_USAGE_SET)

build-windows-amd64:
	@go env -w GOARCH=amd64
	@go env -w GOOS=windows
	@go build -ldflags "-w -s" -trimpath -o ./bin/$(APP_NAME)_windows_amd64.$(FILE_EXTENSION) ./cmd
	@go build -ldflags "-w -s" -trimpath -o ./bin/$(MIGRATE_NAME)_windows_amd64.$(FILE_EXTENSION) ./cmd/migrate

build-linux-amd64:
	@go env -w GOOS=linux
	@go env -w GOARCH=amd64
	@go build -ldflags "-w -s" -trimpath -o ./bin/$(APP_NAME)_linux_amd64 ./cmd
	@go build -ldflags "-w -s" -trimpath -o ./bin/$(MIGRATE_NAME)_linux_amd64 ./cmd

build-linux-arm64:
	@go env -w GOOS=linux
	@go env -w GOARCH=arm64
	@go build -ldflags "-w -s" -trimpath -o ./bin/$(APP_NAME)_linux_amd64 ./cmd
	@go build -ldflags "-w -s" -trimpath -o ./bin/$(MIGRATE_NAME)_linux_arm64 ./cmd

dev:
	@echo "Running in development mode..."
	@go run ./cmd/main.go

test:
	@echo "Running tests..."
	@go test ./...


clean:
	@echo "Cleaning up..."
ifeq ($(OS),Windows_NT)
	@if exist .\bin\$(APP_NAME)* $(RM_CMD) .\bin\$(APP_NAME)*
	@if exist .\bin\$(MIGRATE_NAME)* $(RM_CMD) .\bin\$(MIGRATE_NAME)*
else
	@$(RM_CMD) ./bin/$(APP_NAME)* 2>/dev/null || true
	@$(RM_CMD) ./bin/$(MIGRATE_NAME)* 2>/dev/null || true
endif