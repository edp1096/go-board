// config/config.go
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Debug          bool
	ServerAddress  string
	DBDriver       string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	JWTSecret      string
	SessionSecret  string
	CookieSecure   bool
	CookieHTTPOnly bool
}

func Load() (*Config, error) {
	// .env 파일 로드 (없으면 무시)
	_ = godotenv.Load()

	debug := os.Getenv("DEBUG") == "true"

	// 서버 주소 설정
	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = ":3000"
	}

	// 데이터베이스 드라이버 확인
	dbDriver := os.Getenv("DB_DRIVER")
	if dbDriver == "" {
		dbDriver = "postgres" // 기본 드라이버는 PostgreSQL
	}

	// 지원하는 드라이버 확인
	if dbDriver != "postgres" && dbDriver != "mysql" {
		return nil, fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s (지원: postgres, mysql)", dbDriver)
	}

	// JWT 시크릿 키 확인
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("환경 변수 JWT_SECRET이 설정되지 않았습니다")
	}

	// 세션 시크릿 키 확인
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		return nil, fmt.Errorf("환경 변수 SESSION_SECRET이 설정되지 않았습니다")
	}

	return &Config{
		Debug:          debug,
		ServerAddress:  serverAddress,
		DBDriver:       dbDriver,
		DBHost:         os.Getenv("DB_HOST"),
		DBPort:         os.Getenv("DB_PORT"),
		DBUser:         os.Getenv("DB_USER"),
		DBPassword:     os.Getenv("DB_PASSWORD"),
		DBName:         os.Getenv("DB_NAME"),
		JWTSecret:      jwtSecret,
		SessionSecret:  sessionSecret,
		CookieSecure:   os.Getenv("COOKIE_SECURE") == "true",
		CookieHTTPOnly: os.Getenv("COOKIE_HTTP_ONLY") != "false",
	}, nil
}
