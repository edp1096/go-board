# README.md
# 동적 게시판 시스템

이 프로젝트는 Go Fiber와 Alpine.js를 활용한 동적 게시판 시스템입니다.

## 기능

- 회원 관리 기능
- 동적 게시판 생성 기능
- Alpine.js를 활용한 하이브리드 렌더링
- MariaDB, PostgreSQL 지원
- 관리자 페이지

## 시작하기

### 요구사항

- Go 1.18+
- PostgreSQL 또는 MariaDB
- Node.js (개발 시)

### 설치

1. 저장소 클론하기
   ```
   git clone https://github.com/your-username/dynamic-board.git
   cd dynamic-board
   ```

2. 의존성 설치
   ```
   go mod download
   ```

3. 환경 설정
   ```
   cp .env.example .env
   ```
   `.env` 파일을 수정하여 데이터베이스 연결 정보 등을 설정하세요.

4. 마이그레이션 적용
   ```
   make migrate-up
   ```

5. 관리자 계정 생성
   ```
   make create-admin
   ```

6. 애플리케이션 실행
   ```
   make run
   ```

### 개발 모드

Air를 사용한 라이브 리로딩 개발 환경 설정:

1. Air 설치
   ```
   go install github.com/cosmtrek/air@latest
   ```

2. 개발 모드로 실행
   ```
   make dev
   ```

## 구성

- `cmd`: 애플리케이션 진입점
- `config`: 설정 관리
- `internal`: 내부 패키지
  - `models`: 데이터 모델
  - `handlers`: HTTP 핸들러
  - `middleware`: 미들웨어
  - `repository`: 데이터 액세스
  - `service`: 비즈니스 로직
- `migrations`: 데이터베이스 마이그레이션
- `web`: 웹 리소스
  - `templates`: HTML 템플릿
  - `static`: 정적 파일 (CSS, JS 등)

## 라이선스

MIT