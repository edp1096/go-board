package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	goboard "github.com/edp1096/go-board"
	"github.com/edp1096/go-board/config"
	"github.com/pressly/goose/v3"
)

func main() {
	// 명령행 인자 파싱
	sourceDriver := flag.String("source-driver", "", "소스 데이터베이스 드라이버 (mysql, postgres, sqlite)")
	targetDriver := flag.String("target-driver", "", "대상 데이터베이스 드라이버 (mysql, postgres, sqlite)")
	sourceEnvFile := flag.String("source-env", ".env", "소스 환경 설정 파일")
	targetEnvFile := flag.String("target-env", "", "대상 환경 설정 파일 (기본값: 소스와 동일)")
	operation := flag.String("op", "up", "마이그레이션 작업 (up, down, status, purge, data-migrate)")
	batchSize := flag.Int("batch-size", 1000, "데이터 마이그레이션 배치 크기")
	skipTables := flag.String("skip-tables", "", "마이그레이션 시 건너뛸 테이블 목록 (쉼표로 구분)")
	dynamicOnly := flag.Bool("dynamic-only", false, "동적 테이블만 마이그레이션")
	basicOnly := flag.Bool("basic-only", false, "기본 테이블만 마이그레이션")
	includeInactive := flag.Bool("include-inactive", true, "비활성 게시판 포함 여부")
	verboseLogging := flag.Bool("verbose", true, "상세 로깅 여부")
	dataOnly := flag.Bool("data-only", false, "데이터만 마이그레이션 (스키마 마이그레이션 없음)")
	schemaOnly := flag.Bool("schema-only", false, "스키마만 마이그레이션 (데이터 마이그레이션 없음)")
	disableTransactions := flag.Bool("disable-transactions", false, "트랜잭션 비활성화 (대용량 데이터에 유용)")
	maxErrors := flag.Int("max-errors", 10, "마이그레이션 중단 전 최대 오류 수")

	flag.Parse()

	// 환경 설정 로드
	sourceConfig, err := loadConfig(*sourceEnvFile, *sourceDriver)
	if err != nil {
		log.Fatalf("소스 설정을 로드할 수 없습니다: %v", err)
	}

	targetEnvPath := *targetEnvFile
	if targetEnvPath == "" {
		targetEnvPath = *sourceEnvFile
	}

	// 대상 데이터베이스 설정
	targetConfig, err := loadConfig(targetEnvPath, *targetDriver)
	if err != nil {
		log.Fatalf("대상 설정을 로드할 수 없습니다: %v", err)
	}

	// 데이터 마이그레이션인 경우 소스와 대상 DB 모두 설정
	if *operation == "data-migrate" {
		if *sourceDriver == "" || *targetDriver == "" {
			log.Fatalf("데이터 마이그레이션에는 소스와 대상 드라이버가 모두 필요합니다")
		}

		if *sourceDriver == *targetDriver && sourceConfig.DBName == targetConfig.DBName &&
			sourceConfig.DBHost == targetConfig.DBHost && sourceConfig.DBPort == targetConfig.DBPort {
			log.Fatalf("소스와 대상 데이터베이스가 동일합니다")
		}

		// 데이터 마이그레이션 구성
		migConfig := &DataMigrationConfig{
			SourceDBConfig:      sourceConfig,
			TargetDBConfig:      targetConfig,
			BatchSize:           *batchSize,
			SkipTables:          []string{},
			DynamicTablesOnly:   *dynamicOnly,
			BasicTablesOnly:     *basicOnly,
			IncludeInactive:     *includeInactive,
			VerboseLogging:      *verboseLogging,
			DataOnly:            *dataOnly,
			SchemaOnly:          *schemaOnly,
			EnableTransactions:  !*disableTransactions,
			MaxErrorsBeforeExit: *maxErrors,
			Errors:              []error{},
		}

		// 건너뛸 테이블 목록 파싱
		if *skipTables != "" {
			migConfig.SkipTables = strings.Split(*skipTables, ",")
			for i, table := range migConfig.SkipTables {
				migConfig.SkipTables[i] = strings.TrimSpace(table)
			}
		}

		// 데이터 마이그레이션 실행
		err = runDataMigration(migConfig)
		if err != nil {
			log.Fatalf("데이터 마이그레이션 실패: %v", err)
		}

		log.Println("데이터 마이그레이션이 성공적으로 완료되었습니다")
		return
	}

	// 일반 스키마 마이그레이션 로직 (기존 코드)
	// SQLite가 아닌 경우에만 데이터베이스 존재 여부 확인
	if sourceConfig.DBDriver != "sqlite" {
		// 데이터베이스 존재 여부 확인 및 없으면 생성
		if err := ensureDatabaseExists(sourceConfig); err != nil {
			log.Fatalf("데이터베이스 생성 실패: %v", err)
		}
	} else {
		// SQLite인 경우 디렉토리 확인
		dir := filepath.Dir(sourceConfig.DBPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("SQLite 데이터베이스 디렉토리 생성 실패: %v", err)
		}
	}

	// 데이터베이스 연결
	database, err := config.NewDatabase(sourceConfig)
	if err != nil {
		log.Fatalf("데이터베이스에 연결할 수 없습니다: %v", err)
	}
	defer database.Close()

	fmt.Printf("데이터베이스 '%s'가 준비되었습니다\n", sourceConfig.DBName)

	// SQL DB 인스턴스 가져오기
	sqlDB := database.DB.DB

	// 마이그레이션 설정
	goose.SetBaseFS(nil) // 기본 파일시스템 설정 초기화

	// 방언 설정
	if err := goose.SetDialect(sourceConfig.DBDriver); err != nil {
		log.Fatalf("데이터베이스 방언 설정 실패: %v", err)
	}

	// 임베디드 마이그레이션 파일 시스템 가져오기
	var migrationFS fs.FS
	var migrationsDir string

	switch sourceConfig.DBDriver {
	case "postgres":
		// PostgreSQL 마이그레이션 설정
		subFS, err := fs.Sub(goboard.PostgresMigrationsFS, "migrations/postgres")
		if err != nil {
			log.Fatalf("PostgreSQL 마이그레이션 파일 접근 실패: %v", err)
		}
		migrationFS = subFS
		migrationsDir = "."
	case "sqlite":
		// SQLite 마이그레이션 설정
		subFS, err := fs.Sub(goboard.SQLiteMigrationsFS, "migrations/sqlite")
		if err != nil {
			log.Fatalf("SQLite 마이그레이션 파일 접근 실패: %v", err)
		}
		migrationFS = subFS
		migrationsDir = "."
	default:
		// MySQL 마이그레이션 설정
		subFS, err := fs.Sub(goboard.MysqlMigrationsFS, "migrations/mysql")
		if err != nil {
			log.Fatalf("MySQL 마이그레이션 파일 접근 실패: %v", err)
		}
		migrationFS = subFS
		migrationsDir = "."
	}

	// 마이그레이션 파일 시스템 설정
	goose.SetBaseFS(migrationFS)

	// 마이그레이션 실행
	if err := runMigration(sqlDB, sourceConfig.DBDriver, migrationsDir, *operation, sourceConfig.DBName); err != nil {
		log.Fatalf("마이그레이션 실패: %v", err)
	}
}
