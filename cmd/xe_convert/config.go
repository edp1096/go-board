package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config 환경 설정 구조체
type Config struct {
	// 소스 데이터베이스 (XE) 설정
	SourceDriver string
	SourceHost   string
	SourcePort   int
	SourceDB     string
	SourceUser   string
	SourcePass   string
	SourcePath   string // SQLite 사용시

	// 대상 데이터베이스 (Go-Board) 설정
	TargetDriver string
	TargetHost   string
	TargetPort   int
	TargetDB     string
	TargetUser   string
	TargetPass   string
	TargetPath   string // SQLite 사용시

	// 기타 설정
	XEPrefix         string
	XEUploadPath     string
	TargetUploadPath string
	BatchSize        int
	MigrateFiles     bool
	SchemaOnly       bool
	DataOnly         bool
	Verbose          bool
}

// LoadConfig 명령행 인자와 환경 파일에서 설정 로드
func LoadConfig() (*Config, error) {
	config := &Config{}

	// 명령행 인자 파싱
	flag.StringVar(&config.SourceDriver, "source-driver", "mysql", "소스 데이터베이스 드라이버 (mysql, postgres, sqlite)")
	flag.StringVar(&config.SourceHost, "source-host", "localhost", "소스 데이터베이스 호스트")
	flag.IntVar(&config.SourcePort, "source-port", 3306, "소스 데이터베이스 포트")
	flag.StringVar(&config.SourceDB, "source-db", "", "소스 데이터베이스 이름")
	flag.StringVar(&config.SourceUser, "source-user", "root", "소스 데이터베이스 사용자")
	flag.StringVar(&config.SourcePass, "source-pass", "", "소스 데이터베이스 비밀번호")
	flag.StringVar(&config.SourcePath, "source-path", "", "소스 SQLite 데이터베이스 경로")

	flag.StringVar(&config.XEPrefix, "prefix", "xe_", "XE 테이블 접두사")
	flag.StringVar(&config.XEUploadPath, "upload-path", "./files", "XE 업로드 파일 경로")
	flag.StringVar(&config.TargetUploadPath, "target-upload-path", "./uploads", "Go-Board 업로드 파일 경로")

	envPath := flag.String("env", ".env", "환경 설정 파일 경로")
	flag.IntVar(&config.BatchSize, "batch-size", 500, "데이터 처리 배치 크기")
	flag.BoolVar(&config.MigrateFiles, "migrate-files", true, "첨부파일 마이그레이션 여부")
	flag.BoolVar(&config.SchemaOnly, "schema-only", false, "스키마만 마이그레이션")
	flag.BoolVar(&config.DataOnly, "data-only", false, "데이터만 마이그레이션")
	flag.BoolVar(&config.Verbose, "verbose", true, "자세한 로그 출력")

	flag.Parse()

	// 환경 변수 파일 로드
	if _, err := os.Stat(*envPath); err == nil {
		if err := godotenv.Load(*envPath); err != nil {
			return nil, fmt.Errorf("환경 설정 파일 로드 실패: %w", err)
		}
	}

	// 대상 데이터베이스 설정 로드
	config.TargetDriver = os.Getenv("DB_DRIVER")
	config.TargetHost = os.Getenv("DB_HOST")
	if port := os.Getenv("DB_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &config.TargetPort)
	} else {
		config.TargetPort = 3306 // 기본값
	}
	config.TargetDB = os.Getenv("DB_NAME")
	config.TargetUser = os.Getenv("DB_USER")
	config.TargetPass = os.Getenv("DB_PASS")
	config.TargetPath = os.Getenv("DB_PATH") // SQLite 사용시

	// 필수 설정 검증
	if config.SourceDriver == "sqlite" {
		if config.SourcePath == "" {
			return nil, fmt.Errorf("SQLite 소스 데이터베이스 경로를 지정해야 합니다")
		}
	} else {
		if config.SourceDB == "" {
			return nil, fmt.Errorf("소스 데이터베이스 이름을 지정해야 합니다")
		}
	}

	if config.TargetDriver == "sqlite" {
		if config.TargetPath == "" {
			return nil, fmt.Errorf("SQLite 대상 데이터베이스 경로를 지정해야 합니다")
		}
	} else {
		if config.TargetDB == "" {
			return nil, fmt.Errorf("대상 데이터베이스 이름을 지정해야 합니다")
		}
	}

	return config, nil
}

// GetSourceDSN 소스 데이터베이스 DSN 생성
func (c *Config) GetSourceDSN() string {
	switch c.SourceDriver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
			c.SourceUser, c.SourcePass, c.SourceHost, c.SourcePort, c.SourceDB)
	case "postgres":
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			c.SourceUser, c.SourcePass, c.SourceHost, c.SourcePort, c.SourceDB)
	case "sqlite":
		return c.SourcePath
	default:
		return ""
	}
}

// GetTargetDSN 대상 데이터베이스 DSN 생성
func (c *Config) GetTargetDSN() string {
	switch c.TargetDriver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
			c.TargetUser, c.TargetPass, c.TargetHost, c.TargetPort, c.TargetDB)
	case "postgres":
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			c.TargetUser, c.TargetPass, c.TargetHost, c.TargetPort, c.TargetDB)
	case "sqlite":
		return c.TargetPath
	default:
		return ""
	}
}
