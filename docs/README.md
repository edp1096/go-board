# Go-Board

커스텀 필드를 지원하는 동적 게시판 시스템

## 주요 기능

- 관리자 정의 커스텀 필드 게시판
- 다양한 입력 타입 지원: 텍스트, 숫자, 날짜, 체크박스, 드롭다운
- 사용자 인증 및 권한 관리
- PostgreSQL, MySQL/MariaDB 지원

## 빠른 시작

### 요구사항
- Go 1.24+
- PostgreSQL 또는 MySQL/MariaDB

### 설치 및 실행

1. **환경 설정** - `.env` 파일 생성:
```
# 기본 설정
DEBUG=true
SERVER_ADDRESS=:3300

# 데이터베이스 (PostgreSQL)
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=root
DB_PASSWORD=pgsql
DB_NAME=go_board

# 보안
JWT_SECRET=secure_jwt_key_here
SESSION_SECRET=secure_session_key_here
```

2. **데이터베이스 마이그레이션**:
```bash
# PostgreSQL
make migrate-up postgres

# MySQL
make migrate-up mysql
```

3. **실행**:
```bash
make dev    # 개발 모드
# 또는
go build -o go-board ./cmd
./go-board  # 프로덕션
```

### 관리자 계정
초기 관리자: ID `admin`, 비밀번호 `admin`

## 디렉토리 구조

```
cmd/          - 애플리케이션 진입점
config/       - 설정 코드
internal/     - 메인 코드
  handlers/   - HTTP 핸들러
  middleware/ - 미들웨어
  models/     - 데이터 모델
  repository/ - 데이터 액세스
  service/    - 비즈니스 로직
  utils/      - 유틸리티
migrations/   - DB 마이그레이션
web/          - 프론트엔드
```
