DEBUG=true
SERVER_ADDRESS=127.0.0.1:3000

# 데이터베이스 설정 (PostgreSQL)
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=root
DB_PASSWORD=pgsql
DB_NAME=go_board

# 보안 설정
JWT_SECRET=your_very_long_and_secure_jwt_secret_key
SESSION_SECRET=your_very_long_and_secure_session_secret_key
COOKIE_SECURE=false
COOKIE_HTTP_ONLY=true