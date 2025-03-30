.PHONY: build run dev clean migrate-up migrate-down migrate-status migrate-create setup test

# 설정 변수
APP_NAME=go-board
MIGRATE_NAME=migrate
DB_DRIVER=mysql

# 운영체제 확인
ifeq ($(OS),Windows_NT)
    APP_NAME=go-board.exe
    MIGRATE_NAME=migrate.exe
    RM_CMD=del /q /f
    APP_RUN=.\bin\$(APP_NAME)
    MIGRATE_RUN=.\bin\$(MIGRATE_NAME)
else
    APP_NAME=go-board
    MIGRATE_NAME=migrate
    RM_CMD=rm -rf
    APP_RUN=./bin/$(APP_NAME)
    MIGRATE_RUN=./bin/$(MIGRATE_NAME)
endif

# 빌드
build: build-app build-migrate

# 앱 빌드
build-app:
	@echo "Building application..."
	@go build -o ./bin/$(APP_NAME) ./cmd

# 마이그레이터 빌드
build-migrate:
	@echo "Building migrator..."
	@go build -o ./bin/$(MIGRATE_NAME) ./cmd/migrate

# 실행
run: build
	@echo "Running application..."
	@$(APP_RUN)

# 개발 모드 (go run 사용)
dev:
	@echo "Running in development mode..."
	@go run ./cmd/main.go

# 마이그레이션 적용
migrate-up: build-migrate ## 데이터베이스 마이그레이션 적용
	@echo "Applying migrations..."
ifeq ($(filter postgres,$(MAKECMDGOALS)),postgres)
	@$(MIGRATE_RUN) -driver postgres -op up
else ifeq ($(filter mysql,$(MAKECMDGOALS)),mysql)
	@$(MIGRATE_RUN) -driver mysql -op up
else
	@$(MIGRATE_RUN) -driver $(DB_DRIVER) -op up
endif

# 마이그레이션 롤백
migrate-down: build-migrate ## 데이터베이스 마이그레이션 롤백
	@echo "Rolling back migrations..."
ifeq ($(filter postgres,$(MAKECMDGOALS)),postgres)
	@$(MIGRATE_RUN) -driver postgres -op down
else ifeq ($(filter mysql,$(MAKECMDGOALS)),mysql)
	@$(MIGRATE_RUN) -driver mysql -op down
else
	@$(MIGRATE_RUN) -driver $(DB_DRIVER) -op down
endif

# 마이그레이션 상태
migrate-status: build-migrate ## 데이터베이스 마이그레이션 상태 확인
	@echo "Checking migration status..."
ifeq ($(filter postgres,$(MAKECMDGOALS)),postgres)
	@$(MIGRATE_RUN) -driver postgres -op status
else ifeq ($(filter mysql,$(MAKECMDGOALS)),mysql)
	@$(MIGRATE_RUN) -driver mysql -op status
else
	@$(MIGRATE_RUN) -driver $(DB_DRIVER) -op status
endif

# 새 마이그레이션 생성
migrate-create: build-migrate ## 새 마이그레이션 생성
	@echo "Creating new migration..."
	@[ "${name}" ] || ( echo "Error: name parameter is required. Use: make migrate-create name=your_migration_name"; exit 1 )
ifeq ($(filter postgres,$(MAKECMDGOALS)),postgres)
	@$(MIGRATE_RUN) -driver postgres -op create -name $(name)
else ifeq ($(filter mysql,$(MAKECMDGOALS)),mysql)
	@$(MIGRATE_RUN) -driver mysql -op create -name $(name)
else
	@$(MIGRATE_RUN) -driver $(DB_DRIVER) -op create -name $(name)
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
	@if exist .\bin\$(APP_NAME) $(RM_CMD) .\bin\$(APP_NAME)
	@if exist .\bin\$(MIGRATE_NAME) $(RM_CMD) .\bin\$(MIGRATE_NAME)
else
	@$(RM_CMD) ./bin/$(APP_NAME) 2>/dev/null || true
	@$(RM_CMD) ./bin/$(MIGRATE_NAME) 2>/dev/null || true
endif