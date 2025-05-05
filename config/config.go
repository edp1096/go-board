package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Environment 타입 정의
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvTest        Environment = "test"
	EnvProduction  Environment = "production"
)

// Config 구조체 정의
type Config struct {
	SiteName        string // 사이트 이름
	LogoPath        string // 로고 이미지 경로
	LogoDisplayMode string // 로고 표시 모드 (text, image, both)
	Environment     Environment
	Debug           bool
	RequireSetup    bool
	ServerAddress   string
	DBDriver        string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBPath          string // SQLite 파일 경로
	JWTSecret       string
	SessionSecret   string
	CookieSecure    bool
	CookieHTTPOnly  bool
	TemplateDir     string
	StaticDir       string

	// 파일 업로드 관련 설정
	UploadDir          string // 업로드 디렉토리 경로
	MaxUploadSize      int64  // 일반 파일 업로드 최대 크기 (바이트)
	MaxImageUploadSize int64  // 이미지 업로드 최대 크기 (바이트)
	MaxBodyLimit       int    // HTTP 요청 본문 최대 크기 (바이트)
}

// Load 함수는 환경에 따라 config를 로드하고 반환
func Load() (*Config, error) {
	// 환경 변수 설정 확인
	env := Environment(strings.ToLower(os.Getenv("APP_ENV")))
	if env == "" {
		env = EnvDevelopment
	}

	// 유효한 환경 값인지 확인
	if env != EnvDevelopment && env != EnvTest && env != EnvProduction {
		return nil, fmt.Errorf("유효하지 않은 환경 변수: %s", env)
	}

	// 기본 .env 파일과 환경별 .env 파일 로드
	_ = godotenv.Load()
	envFile := fmt.Sprintf(".env.%s", env)
	_ = godotenv.Load(envFile)

	// 사이트 이름
	siteName := getEnvWithDefault("SITE_NAME", "게시판")

	// 로고
	logoPath := getEnvWithDefault("LOGO_PATH", "/static/images/logo.png")
	logoDisplayMode := getEnvWithDefault("LOGO_DISPLAY_MODE", "both") // 기본값: 모두 표시

	// 디버그 모드 설정 (환경변수 기반)
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
	if dbDriver != "postgres" && dbDriver != "mysql" && dbDriver != "sqlite" {
		return nil, fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s (지원: postgres, mysql, sqlite)", dbDriver)
	}

	// SQLite 데이터베이스 경로 설정
	dbPath := os.Getenv("DB_PATH")
	if dbDriver == "sqlite" && dbPath == "" {
		// 기본 파일 위치 - data 디렉토리 만들기
		dbPath = "./data/go_board.db"
		os.MkdirAll("./data", 0755)
	}

	// JWT 시크릿 키 확인
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		if env == EnvProduction {
			return nil, fmt.Errorf("환경 변수 JWT_SECRET이 설정되지 않았습니다")
		}
		// 개발/테스트 환경에서는 기본값 사용
		jwtSecret = "dev-jwt-secret-key"
	}

	// 세션 시크릿 키 확인
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		if env == EnvProduction {
			return nil, fmt.Errorf("환경 변수 SESSION_SECRET이 설정되지 않았습니다")
		}
		// 개발/테스트 환경에서는 기본값 사용
		sessionSecret = "dev-session-secret-key"
	}

	// 디렉토리 경로 설정
	templateDir := os.Getenv("TEMPLATE_DIR")
	if templateDir == "" {
		templateDir = "./web/templates"
	}

	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./web/static"
	}

	// 디렉토리 경로 정규화
	templateDir = filepath.Clean(templateDir)
	staticDir = filepath.Clean(staticDir)

	// // 디렉토리 존재 여부 확인
	// _, errTemplateDir := os.Stat(templateDir)
	// _, errStaticDir := os.Stat(staticDir)

	// // 디렉토리가 없으면 오류 표시
	// if os.IsNotExist(errTemplateDir) || os.IsNotExist(errStaticDir) {
	// 	log.Printf("경고: 템플릿 디렉토리 또는 정적 파일 디렉토리가 존재하지 않습니다.")
	// }

	// 초기 설정 필요 여부 설정
	requireSetup := os.Getenv("REQUIRE_SETUP") == "true"
	// 환경 변수가 설정되어 있지 않으면 기본값은 true (자동 감지)
	if os.Getenv("REQUIRE_SETUP") == "" {
		requireSetup = true
	}

	// 업로드 디렉토리 설정
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	// 파일 업로드 크기 설정
	maxUploadSizeKB, err := strconv.ParseInt(getEnvWithDefault("MAX_UPLOAD_SIZE",
		strconv.Itoa(DefaultUploadSizeKB)), 10, 64)
	if err != nil {
		maxUploadSizeKB = DefaultUploadSizeKB
	}

	// 이미지 업로드 크기 설정
	maxImageUploadSizeKB, err := strconv.ParseInt(getEnvWithDefault("MAX_IMAGE_UPLOAD_SIZE",
		strconv.Itoa(DefaultImageUploadSizeKB)), 10, 64)
	if err != nil {
		maxImageUploadSizeKB = DefaultImageUploadSizeKB
	}

	// HTTP 요청 본문 최대 크기 설정
	maxBodyLimitKB, err := strconv.Atoi(getEnvWithDefault("MAX_BODY_LIMIT", strconv.Itoa(DefaultBodyLimitKB)))
	if err != nil {
		maxBodyLimitKB = DefaultBodyLimitKB
	}

	return &Config{
		SiteName:           siteName,
		LogoPath:           logoPath,
		LogoDisplayMode:    logoDisplayMode,
		Environment:        env,
		Debug:              debug,
		RequireSetup:       requireSetup,
		ServerAddress:      serverAddress,
		DBDriver:           dbDriver,
		DBHost:             getEnvWithDefault("DB_HOST", "localhost"),
		DBPort:             getEnvWithDefault("DB_PORT", "5432"),
		DBUser:             getEnvWithDefault("DB_USER", "postgres"),
		DBPassword:         os.Getenv("DB_PASSWORD"),
		DBName:             getEnvWithDefault("DB_NAME", "go_board"),
		DBPath:             dbPath,
		JWTSecret:          jwtSecret,
		SessionSecret:      sessionSecret,
		CookieSecure:       os.Getenv("COOKIE_SECURE") == "true" || env == EnvProduction,
		CookieHTTPOnly:     os.Getenv("COOKIE_HTTP_ONLY") != "false",
		TemplateDir:        templateDir,
		StaticDir:          staticDir,
		UploadDir:          uploadDir,
		MaxUploadSize:      maxUploadSizeKB * BytesPerKB,
		MaxImageUploadSize: maxImageUploadSizeKB * BytesPerKB,
		MaxBodyLimit:       maxBodyLimitKB * BytesPerKB,
	}, nil
}

// 환경 변수가 없을 경우 기본값 사용
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// 현재 환경이 개발 환경인지 확인
func (c *Config) IsDevelopment() bool {
	return c.Environment == EnvDevelopment
}

// 현재 환경이 테스트 환경인지 확인
func (c *Config) IsTest() bool {
	return c.Environment == EnvTest
}

// 현재 환경이 운영 환경인지 확인
func (c *Config) IsProduction() bool {
	return c.Environment == EnvProduction
}
