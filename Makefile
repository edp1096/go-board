.PHONY: build run dev clean migrate-up migrate-down help setup-dirs setup

# 설정 변수
APP_NAME=dynamic-board
DB_DRIVER=mysql
DB_CONN=root:@tcp(localhost:13306)/dynamic_board?parseTime=true

# 기본 목표
.DEFAULT_GOAL := help

# 빌드
build: ## 애플리케이션 빌드
	@echo "Building $(APP_NAME)..."
	@go build -o ./bin/$(APP_NAME).exe ./cmd

# 실행
run: build ## 애플리케이션 실행
	@echo "Running $(APP_NAME)..."
	@.\bin\$(APP_NAME).exe

# 개발 모드로 실행 (Air 활용)
dev: ## 개발 모드로 실행 (Air 필요)
	@echo "Running in development mode..."
	@air

# 마이그레이션 적용
migrate-up: ## 데이터베이스 마이그레이션 적용 (Goose 필요)
	@echo "Applying migrations..."
	@goose -dir migrations $(DB_DRIVER) "$(DB_CONN)" up

# 마이그레이션 롤백
migrate-down: ## 데이터베이스 마이그레이션 롤백 (Goose 필요)
	@echo "Rolling back migrations..."
	@goose -dir migrations $(DB_DRIVER) "$(DB_CONN)" down

# 초기 관리자 계정 생성
create-admin: ## 관리자 계정 생성
	@echo "Creating admin user..."
	@go run cmd/admin/create_admin.go

# 정리 (Windows 호환)
clean: ## 빌드 파일 정리
	@echo "Cleaning..."
	@if exist .\bin rmdir /s /q .\bin

# 디렉토리 생성 (Windows 호환)
setup-dirs: ## 필요한 디렉토리 구조 생성
	@echo "Creating directory structure..."
	@if not exist .\bin mkdir .\bin
	@if not exist .\web\static\css mkdir .\web\static\css
	@if not exist .\web\static\js\pages mkdir .\web\static\js\pages
	@if not exist .\web\templates\layouts mkdir .\web\templates\layouts
	@if not exist .\web\templates\partials mkdir .\web\templates\partials
	@if not exist .\web\templates\auth mkdir .\web\templates\auth
	@if not exist .\web\templates\board mkdir .\web\templates\board
	@if not exist .\web\templates\admin mkdir .\web\templates\admin

# 초기 설정 (모든 준비 단계 실행)
setup: setup-dirs ## 전체 프로젝트 초기 설정
	@echo "Setting up the project..."
	@go mod download
	@echo "Creating database..."
	@echo "CREATE DATABASE IF NOT EXISTS dynamic_board;" | mysql -u root -h localhost -P 13306
	@echo "Done! Don't forget to configure your .env file."

# 도움말
help: ## 도움말 표시
	@powershell -Command "Get-Content Makefile | ForEach-Object { if ($$_ -match '^[a-zA-Z_-]+:.*?## (.*)') { Write-Host $$_[0].Split(':')[0] -ForegroundColor Cyan -NoNewline; Write-Host (' - ' + $$Matches[1]) } }"