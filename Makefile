.PHONY: build run dev clean migrate-up migrate-down setup test

# 설정 변수
APP_NAME=go-board
DB_DRIVER=mysql
DB_CONN=root:@tcp(localhost:13306)/go_board?parseTime=true

# 운영체제 확인
ifeq ($(OS),Windows_NT)
    APP_NAME=go-board.exe
    RM_CMD=del /q /f
    APP_RUN=.\bin\$(APP_NAME)
else
    APP_NAME=go-board
    RM_CMD=rm -rf
    APP_RUN=./bin/$(APP_NAME)
endif

# 빌드
build:
	@echo "Building application..."
	@go build -o ./bin/$(APP_NAME) ./cmd

# 실행
run: build
	@echo "Running application..."
	@$(APP_RUN)

# 개발 모드 (go run 사용)
dev:
	@echo "Running in development mode..."
	@go run ./cmd/main.go

# 마이그레이션 적용
migrate-up: ## 데이터베이스 마이그레이션 적용 (Goose 필요)
	@echo "Applying migrations..."
ifeq ($(filter postgres,$(MAKECMDGOALS)),postgres)
ifeq ($(OS),Windows_NT)
	@goose -dir migrations/postgres postgres "user=root password=pgsql dbname=go_board sslmode=disable" up
else
	@DB_DRIVER=postgres DB_CONN="user=root password=pgsql dbname=go_board sslmode=disable" goose -dir migrations/postgres up
endif
else ifeq ($(filter mysql,$(MAKECMDGOALS)),mysql)
ifeq ($(OS),Windows_NT)
	@goose -dir migrations/mysql mysql "root:@tcp(localhost:13306)/go_board?parseTime=true" up
else
	@DB_DRIVER=mysql DB_CONN="root:@tcp(localhost:13306)/go_board?parseTime=true" goose -dir migrations/mysql up
endif
else
	@goose -dir migrations $(DB_DRIVER) "$(DB_CONN)" up
endif

# 마이그레이션 롤백
migrate-down: ## 데이터베이스 마이그레이션 롤백 (Goose 필요)
	@echo "Rolling back migrations..."
ifeq ($(filter postgres,$(MAKECMDGOALS)),postgres)
ifeq ($(OS),Windows_NT)
	@goose -dir migrations/postgres postgres "user=root password=pgsql dbname=go_board sslmode=disable" down
else
	@DB_DRIVER=postgres DB_CONN="user=root password=pgsql dbname=go_board sslmode=disable" goose -dir migrations/postgres down
endif
else ifeq ($(filter mysql,$(MAKECMDGOALS)),mysql)
ifeq ($(OS),Windows_NT)
	@goose -dir migrations/mysql mysql "root:@tcp(localhost:13306)/go_board?parseTime=true" down
else
	@DB_DRIVER=mysql DB_CONN="root:@tcp(localhost:13306)/go_board?parseTime=true" goose -dir migrations/mysql down
endif
else
	@goose -dir migrations $(DB_DRIVER) "$(DB_CONN)" down
endif

# 더미 타겟
postgres mysql:
	@:

# 테스트
test:
	@echo "Running tests..."
	@go test ./...

# 의존성 설치
setup:
	@echo "Setting up project..."
	@go mod download
	@echo "Setup complete!"

# 정리
clean:
	@echo "Cleaning up..."
ifeq ($(OS),Windows_NT)
	@if exist .\bin $(RM_CMD) .\bin\*
else
	@$(RM_CMD) ./bin 2>/dev/null || true
endif