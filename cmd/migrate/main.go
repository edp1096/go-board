package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	goboard "github.com/edp1096/go-board"
	"github.com/edp1096/go-board/config"
	"github.com/edp1096/go-board/internal/models"
	"github.com/edp1096/go-board/internal/repository"
	"github.com/edp1096/go-board/internal/service"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "modernc.org/sqlite"
)

// 데이터 마이그레이션을 위한 설정 구조체
type DataMigrationConfig struct {
	SourceDBConfig      *config.Config
	TargetDBConfig      *config.Config
	SourceDB            *bun.DB
	TargetDB            *bun.DB
	BatchSize           int
	SkipTables          []string
	DynamicTablesOnly   bool
	BasicTablesOnly     bool
	IncludeInactive     bool
	VerboseLogging      bool
	DataOnly            bool
	SchemaOnly          bool
	EnableTransactions  bool
	MaxErrorsBeforeExit int
	Errors              []error
}

// 테이블 메타데이터 구조체
type TableMetadata struct {
	Name       string
	Schema     string
	Columns    []ColumnMetadata
	IsDynamic  bool
	SourceRows int
	TargetRows int
}

// 컬럼 메타데이터 구조체
type ColumnMetadata struct {
	Name     string
	Type     string
	Nullable bool
	Default  sql.NullString
}

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

// loadConfig는 환경 파일에서 설정을 로드하고 드라이버를 오버라이드합니다
func loadConfig(envFile string, driverOverride string) (*config.Config, error) {
	// 기존 환경 변수 저장
	oldAppEnv := os.Getenv("APP_ENV")
	oldDBDriver := os.Getenv("DB_DRIVER")

	// 환경 변수 설정
	if envFile != "" && envFile != ".env" {
		// 환경 파일 내용 로드
		content, err := os.ReadFile(envFile)
		if err != nil {
			return nil, fmt.Errorf("환경 파일 읽기 실패: %w", err)
		}

		// 환경 변수 파싱 및 설정
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}

	// 드라이버 오버라이드
	if driverOverride != "" {
		os.Setenv("DB_DRIVER", driverOverride)
	}

	// 설정 로드
	cfg, err := config.Load()

	// 기존 환경 변수 복원
	os.Setenv("APP_ENV", oldAppEnv)
	os.Setenv("DB_DRIVER", oldDBDriver)

	return cfg, err
}

// ensureDatabaseExists는 데이터베이스가 존재하는지 확인하고 없으면 생성합니다
func ensureDatabaseExists(cfg *config.Config) error {
	// 기존 코드 유지
	var db *sql.DB
	var err error

	switch cfg.DBDriver {
	case "postgres":
		// PostgreSQL 관리자 연결 (데이터베이스 생성용) - postgres 데이터베이스에 연결
		connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword)
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return fmt.Errorf("PostgreSQL 관리자 연결 실패: %w", err)
		}
		defer db.Close()

		// 데이터베이스 존재 여부 확인
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", cfg.DBName).Scan(&exists)
		if err != nil {
			return fmt.Errorf("데이터베이스 확인 실패: %w", err)
		}

		// 데이터베이스가 없으면 생성
		if !exists {
			log.Printf("데이터베이스 '%s'가 존재하지 않아 생성합니다", cfg.DBName)

			// 데이터베이스 이름에 특수문자가 있으면 큰따옴표로 감싸줌
			dbNameQuoted := cfg.DBName
			if strings.ContainsAny(dbNameQuoted, "-.") {
				dbNameQuoted = fmt.Sprintf("\"%s\"", dbNameQuoted)
			}

			_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbNameQuoted))
			if err != nil {
				return fmt.Errorf("데이터베이스 생성 실패: %w", err)
			}
			log.Printf("데이터베이스 '%s'가 성공적으로 생성되었습니다", cfg.DBName)
		}

	case "mysql", "mariadb":
		// MySQL/MariaDB 관리자 연결 (데이터베이스 생성용) - 데이터베이스 이름 없이 연결
		connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort)
		db, err = sql.Open("mysql", connStr)
		if err != nil {
			return fmt.Errorf("MySQL 관리자 연결 실패: %w", err)
		}
		defer db.Close()

		// 데이터베이스 존재 여부 확인
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = ?)", cfg.DBName).Scan(&exists)
		if err != nil {
			return fmt.Errorf("데이터베이스 확인 실패: %w", err)
		}

		// 데이터베이스가 없으면 생성
		if !exists {
			log.Printf("데이터베이스 '%s'가 존재하지 않아 생성합니다", cfg.DBName)

			// 데이터베이스 이름에 특수문자가 있으면 백틱으로 감싸줌
			dbNameQuoted := cfg.DBName
			if strings.ContainsAny(dbNameQuoted, "-.") {
				dbNameQuoted = fmt.Sprintf("`%s`", dbNameQuoted)
			}

			_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbNameQuoted))
			if err != nil {
				return fmt.Errorf("데이터베이스 생성 실패: %w", err)
			}
			log.Printf("데이터베이스 '%s'가 성공적으로 생성되었습니다", cfg.DBName)
		}

	case "sqlite":
		// SQLite는 파일 기반이므로 별도의 생성 과정 불필요
		return nil

	default:
		return fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", cfg.DBDriver)
	}

	return nil
}

// runMigration은 지정된 마이그레이션 작업을 실행합니다
func runMigration(db *sql.DB, driver, migrationsDir, operation, dbName string) error {
	// 기존 코드 유지
	var err error
	switch operation {
	case "up":
		fmt.Println("마이그레이션 적용 중...")
		err = goose.Up(db, migrationsDir)
	case "down":
		fmt.Println("마이그레이션 롤백 중...")
		err = goose.Down(db, migrationsDir)
	case "status":
		fmt.Println("마이그레이션 상태 확인 중...")
		err = goose.Status(db, migrationsDir)
	case "purge":
		// 모든 테이블 삭제
		fmt.Println("모든 테이블 삭제 중...")
		err = dropAllTables(db, driver, dbName)
		if err == nil {
			// 마이그레이션 기록도 삭제
			_, err = db.Exec("DROP TABLE IF EXISTS goose_db_version;")
			if err != nil {
				fmt.Printf("마이그레이션 기록 테이블 삭제 실패: %v\n", err)
			} else {
				fmt.Println("마이그레이션 기록 테이블 삭제됨")
			}
			// 마이그레이션 다시 적용
			fmt.Println("마이그레이션 다시 적용 중...")
			err = goose.Up(db, migrationsDir)
		}
	default:
		return fmt.Errorf("알 수 없는 마이그레이션 작업: %s", operation)
	}

	if err != nil {
		return err
	}

	fmt.Printf("마이그레이션 작업 '%s'가 성공적으로 완료되었습니다\n", operation)
	return nil
}

// dropAllTables는 데이터베이스의 모든 테이블을 삭제합니다
func dropAllTables(sqlDB *sql.DB, driver string, dbName string) error {
	// 기존 코드 유지
	var query string
	var rows *sql.Rows
	var err error

	switch driver {
	case "postgres":
		// PostgreSQL에서 모든 테이블 목록 가져오기
		query = `
            SELECT tablename FROM pg_tables 
            WHERE schemaname = 'public' AND 
            tablename != 'goose_db_version';
        `
		rows, err = sqlDB.Query(query)
		if err != nil {
			return fmt.Errorf("테이블 목록 조회 실패: %w", err)
		}
		defer rows.Close()

		// 외래 키 제약 조건 비활성화
		_, err = sqlDB.Exec("SET session_replication_role = 'replica';")
		if err != nil {
			return fmt.Errorf("외래 키 제약 비활성화 실패: %w", err)
		}

		// 모든 테이블 삭제
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return fmt.Errorf("테이블 이름 읽기 실패: %w", err)
			}

			if tableName != "goose_db_version" {
				_, err = sqlDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS \"%s\" CASCADE;", tableName))
				if err != nil {
					return fmt.Errorf("테이블 삭제 실패 (%s): %w", tableName, err)
				}
				fmt.Printf("테이블 삭제: %s\n", tableName)
			}
		}

		// 외래 키 제약 조건 다시 활성화
		_, err = sqlDB.Exec("SET session_replication_role = 'origin';")
		if err != nil {
			return fmt.Errorf("외래 키 제약 재활성화 실패: %w", err)
		}

	case "mysql", "mariadb":
		// MySQL/MariaDB에서 모든 테이블 목록 가져오기
		query = fmt.Sprintf("SHOW TABLES FROM `%s` WHERE Tables_in_%s != 'goose_db_version'", dbName, dbName)
		rows, err = sqlDB.Query(query)
		if err != nil {
			return fmt.Errorf("테이블 목록 조회 실패: %w", err)
		}
		defer rows.Close()

		// 외래 키 제약 조건 비활성화
		_, err = sqlDB.Exec("SET FOREIGN_KEY_CHECKS = 0;")
		if err != nil {
			return fmt.Errorf("외래 키 검사 비활성화 실패: %w", err)
		}

		// 모든 테이블 삭제
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return fmt.Errorf("테이블 이름 읽기 실패: %w", err)
			}

			if tableName != "goose_db_version" {
				_, err = sqlDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`;", tableName))
				if err != nil {
					return fmt.Errorf("테이블 삭제 실패 (%s): %w", tableName, err)
				}
				fmt.Printf("테이블 삭제: %s\n", tableName)
			}
		}

		// 외래 키 제약 조건 다시 활성화
		_, err = sqlDB.Exec("SET FOREIGN_KEY_CHECKS = 1;")
		if err != nil {
			return fmt.Errorf("외래 키 검사 재활성화 실패: %w", err)
		}

	case "sqlite":
		// SQLite에서 모든 테이블 목록 가져오기
		query = "SELECT name FROM sqlite_master WHERE type='table' AND name != 'goose_db_version';"
		rows, err = sqlDB.Query(query)
		if err != nil {
			return fmt.Errorf("테이블 목록 조회 실패: %w", err)
		}
		defer rows.Close()

		// 외래 키 제약 조건 비활성화
		_, err = sqlDB.Exec("PRAGMA foreign_keys = OFF;")
		if err != nil {
			return fmt.Errorf("외래 키 제약 비활성화 실패: %w", err)
		}

		// 모든 테이블 삭제
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return fmt.Errorf("테이블 이름 읽기 실패: %w", err)
			}

			if tableName != "goose_db_version" && !strings.HasPrefix(tableName, "sqlite_") {
				_, err = sqlDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS \"%s\";", tableName))
				if err != nil {
					return fmt.Errorf("테이블 삭제 실패 (%s): %w", tableName, err)
				}
				fmt.Printf("테이블 삭제: %s\n", tableName)
			}
		}

		// 외래 키 제약 조건 다시 활성화
		_, err = sqlDB.Exec("PRAGMA foreign_keys = ON;")
		if err != nil {
			return fmt.Errorf("외래 키 제약 재활성화 실패: %w", err)
		}

	default:
		return fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", driver)
	}

	return nil
}

// runDataMigration은 데이터 마이그레이션을 실행합니다
func runDataMigration(config *DataMigrationConfig) error {
	fmt.Println("==========================")
	fmt.Printf("이기종 DB 데이터 마이그레이션 시작\n")
	fmt.Printf("소스: %s (%s)\n", config.SourceDBConfig.DBDriver, config.SourceDBConfig.DBName)
	fmt.Printf("대상: %s (%s)\n", config.TargetDBConfig.DBDriver, config.TargetDBConfig.DBName)
	fmt.Println("==========================")

	// 1. 데이터베이스 연결
	sourceDB, targetDB, err := connectDatabases(config)
	if err != nil {
		return fmt.Errorf("데이터베이스 연결 실패: %w", err)
	}
	defer sourceDB.Close()
	defer targetDB.Close()

	config.SourceDB = sourceDB
	config.TargetDB = targetDB

	// 2. 대상 데이터베이스 스키마 생성 (있을 경우에만)
	if !config.DataOnly {
		if err := migrateSchema(config); err != nil {
			return fmt.Errorf("스키마 마이그레이션 실패: %w", err)
		}
	}

	// 스키마 전용 마이그레이션이면 여기서 종료
	if config.SchemaOnly {
		fmt.Println("스키마 마이그레이션이 완료되었습니다.")
		return nil
	}

	// 3. 게시판 메타데이터 가져오기
	boardService, dynamicBoardService, err := createBoardServices(config)
	if err != nil {
		return fmt.Errorf("서비스 생성 실패: %w", err)
	}

	// 4. 데이터 마이그레이션 실행
	startTime := time.Now()

	// 4.1 기본 테이블 데이터 마이그레이션
	if !config.DynamicTablesOnly {
		if err := migrateBasicTables(config); err != nil {
			return fmt.Errorf("기본 테이블 마이그레이션 실패: %w", err)
		}
	}

	// 4.2 동적 테이블 데이터 마이그레이션
	if !config.BasicTablesOnly {
		if err := migrateDynamicTables(config, boardService, dynamicBoardService); err != nil {
			return fmt.Errorf("동적 테이블 마이그레이션 실패: %w", err)
		}
	}

	// 5. 시퀀스/자동증가 값 복구
	if err := resetSequences(config); err != nil {
		log.Printf("시퀀스 복구 실패: %v (무시하고 계속 진행합니다)", err)
	}

	// 6. 결과 요약
	elapsedTime := time.Since(startTime)
	fmt.Println("==========================")
	fmt.Printf("데이터 마이그레이션 완료 (소요 시간: %s)\n", elapsedTime)

	if len(config.Errors) > 0 {
		fmt.Printf("경고: 마이그레이션 중 %d개의 오류가 발생했습니다\n", len(config.Errors))
		for i, err := range config.Errors {
			if i >= 5 {
				fmt.Printf("추가 %d개 오류 생략...\n", len(config.Errors)-5)
				break
			}
			fmt.Printf("  - %v\n", err)
		}
	} else {
		fmt.Println("마이그레이션이 오류 없이 완료되었습니다.")
	}

	fmt.Println("==========================")

	return nil
}

// connectDatabases는 소스 및 대상 데이터베이스에 연결합니다
func connectDatabases(config *DataMigrationConfig) (*bun.DB, *bun.DB, error) {
	// 소스 데이터베이스가 존재하는지 확인
	if config.SourceDBConfig.DBDriver != "sqlite" {
		if err := ensureDatabaseExists(config.SourceDBConfig); err != nil {
			return nil, nil, fmt.Errorf("소스 데이터베이스 확인 실패: %w", err)
		}
	} else {
		// SQLite 소스 디렉토리 확인
		dir := filepath.Dir(config.SourceDBConfig.DBPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, nil, fmt.Errorf("SQLite 소스 디렉토리 생성 실패: %w", err)
		}

		// 파일 존재 여부 확인
		if _, err := os.Stat(config.SourceDBConfig.DBPath); os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("SQLite 소스 데이터베이스 파일이 존재하지 않습니다: %s", config.SourceDBConfig.DBPath)
		}
	}

	// 대상 데이터베이스 확인/생성
	if config.TargetDBConfig.DBDriver != "sqlite" {
		if err := ensureDatabaseExists(config.TargetDBConfig); err != nil {
			return nil, nil, fmt.Errorf("대상 데이터베이스 확인 실패: %w", err)
		}
	} else {
		// SQLite 대상 디렉토리 확인
		dir := filepath.Dir(config.TargetDBConfig.DBPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, nil, fmt.Errorf("SQLite 대상 디렉토리 생성 실패: %w", err)
		}
	}

	// 소스 데이터베이스 연결
	sourceDB, err := config.ConnectDatabase(config.SourceDBConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("소스 데이터베이스 연결 실패: %w", err)
	}

	// 대상 데이터베이스 연결
	targetDB, err := config.ConnectDatabase(config.TargetDBConfig)
	if err != nil {
		sourceDB.Close()
		return nil, nil, fmt.Errorf("대상 데이터베이스 연결 실패: %w", err)
	}

	return sourceDB, targetDB, nil
}

// ConnectDatabase는 Config 구조체를 사용하여 데이터베이스에 연결합니다
func (c *DataMigrationConfig) ConnectDatabase(cfg *config.Config) (*bun.DB, error) {
	var sqldb *sql.DB
	var err error

	switch cfg.DBDriver {
	case "postgres":
		// PostgreSQL 연결 설정
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
		sqldb, err = sql.Open("postgres", dsn)
		if err != nil {
			return nil, fmt.Errorf("PostgreSQL 연결 실패: %w", err)
		}

	case "mysql", "mariadb":
		// MySQL/MariaDB 연결 설정
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
		sqldb, err = sql.Open("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("MySQL 연결 실패: %w", err)
		}

	case "sqlite":
		// SQLite 연결 설정
		if cfg.DBPath == "" {
			return nil, fmt.Errorf("SQLite 데이터베이스 경로가 설정되지 않았습니다")
		}

		// 디렉토리 존재 확인 및 생성
		dir := filepath.Dir(cfg.DBPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("SQLite 데이터베이스 디렉토리 생성 실패: %w", err)
		}

		// SQLite 연결 - 로킹 타임아웃 추가
		sqldb, err = sql.Open("sqlite", cfg.DBPath+"?_timeout=5000&_journal=WAL&_busy_timeout=5000")
		if err != nil {
			return nil, fmt.Errorf("SQLite 연결 실패: %w", err)
		}

		// SQLite 성능 최적화 설정
		if _, err := sqldb.Exec("PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL; PRAGMA busy_timeout=5000;"); err != nil {
			return nil, fmt.Errorf("SQLite PRAGMA 설정 실패: %w", err)
		}

	default:
		return nil, fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", cfg.DBDriver)
	}

	// 기본 연결 설정
	sqldb.SetMaxOpenConns(25)
	sqldb.SetMaxIdleConns(5)
	sqldb.SetConnMaxLifetime(5 * time.Minute)

	// 연결 테스트
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqldb.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("데이터베이스 연결 확인 실패: %w", err)
	}

	// Bun DB 인스턴스 생성
	var db *bun.DB
	if cfg.DBDriver == "postgres" {
		db = bun.NewDB(sqldb, pgdialect.New())
	} else if cfg.DBDriver == "sqlite" {
		db = bun.NewDB(sqldb, sqlitedialect.New())
	} else {
		db = bun.NewDB(sqldb, mysqldialect.New())
	}

	return db, nil
}

// createBoardServices는 BoardService 및 DynamicBoardService 인스턴스를 생성합니다
func createBoardServices(config *DataMigrationConfig) (service.BoardService, service.DynamicBoardService, error) {
	// 소스 레포지토리 생성
	boardRepo := repository.NewBoardRepository(config.SourceDB)

	// 서비스 생성
	boardService := service.NewBoardService(boardRepo, config.SourceDB)
	dynamicBoardService := service.NewDynamicBoardService(config.SourceDB)

	return boardService, dynamicBoardService, nil
}

// migrateSchema는 스키마 정보를 마이그레이션합니다
func migrateSchema(config *DataMigrationConfig) error {
	fmt.Println("[1/4] 대상 데이터베이스 스키마 생성 중...")

	// 대상 DB에 스키마 생성
	sqlDB := config.TargetDB.DB

	// 마이그레이션 설정
	goose.SetBaseFS(nil)
	if err := goose.SetDialect(config.TargetDBConfig.DBDriver); err != nil {
		return fmt.Errorf("대상 DB 방언 설정 실패: %w", err)
	}

	// 임베디드 마이그레이션 파일 시스템 가져오기
	var migrationFS fs.FS
	var migrationsDir string

	switch config.TargetDBConfig.DBDriver {
	case "postgres":
		subFS, err := fs.Sub(goboard.PostgresMigrationsFS, "migrations/postgres")
		if err != nil {
			return fmt.Errorf("PostgreSQL 마이그레이션 파일 접근 실패: %w", err)
		}
		migrationFS = subFS
		migrationsDir = "."
	case "sqlite":
		subFS, err := fs.Sub(goboard.SQLiteMigrationsFS, "migrations/sqlite")
		if err != nil {
			return fmt.Errorf("SQLite 마이그레이션 파일 접근 실패: %w", err)
		}
		migrationFS = subFS
		migrationsDir = "."
	default:
		subFS, err := fs.Sub(goboard.MysqlMigrationsFS, "migrations/mysql")
		if err != nil {
			return fmt.Errorf("MySQL 마이그레이션 파일 접근 실패: %w", err)
		}
		migrationFS = subFS
		migrationsDir = "."
	}

	goose.SetBaseFS(migrationFS)

	// 마이그레이션 적용 전 데이터베이스 상태 확인
	fmt.Println("마이그레이션 상태 확인 중...")
	if err := goose.Status(sqlDB, migrationsDir); err != nil {
		return fmt.Errorf("마이그레이션 상태 확인 실패: %w", err)
	}

	// 마이그레이션 적용
	fmt.Println("스키마 마이그레이션 적용 중...")
	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		return fmt.Errorf("스키마 마이그레이션 실패: %w", err)
	}

	fmt.Println("기본 스키마 마이그레이션 완료")
	return nil
}

// getBasicTables는 기본 테이블 목록을 반환합니다
func getBasicTables() []string {
	return []string{
		"users",
		"boards",
		"board_fields",
		"board_managers",
		"comments",
		"attachments",
		"qna_answers",
		"qna_question_votes",
		"qna_answer_votes",
		"referrer_stats",
		"system_settings",
	}
}

// migrateBasicTables는 기본 테이블의 데이터를 마이그레이션합니다
func migrateBasicTables(config *DataMigrationConfig) error {
	fmt.Println("[2/4] 기본 테이블 데이터 마이그레이션 중...")

	// 기본 테이블 목록
	basicTables := getBasicTables()

	// 테이블별 데이터 마이그레이션
	for _, tableName := range basicTables {
		// 건너뛸 테이블 확인
		if isTableSkipped(tableName, config.SkipTables) {
			fmt.Printf("  - 테이블 '%s' 건너뛰기 (사용자 설정에 의해 제외)\n", tableName)
			continue
		}

		// 테이블 데이터 복사
		err := migrateTableData(config, tableName)
		if err != nil {
			config.addError(fmt.Errorf("테이블 '%s' 마이그레이션 실패: %w", tableName, err))
			if len(config.Errors) > config.MaxErrorsBeforeExit {
				return fmt.Errorf("오류가 너무 많아 마이그레이션이 중단되었습니다")
			}
		}
	}

	fmt.Println("기본 테이블 데이터 마이그레이션 완료")
	return nil
}

// migrateDynamicTables는 동적 테이블의 데이터를 마이그레이션합니다
func migrateDynamicTables(config *DataMigrationConfig, boardService service.BoardService, dynamicBoardService service.DynamicBoardService) error {
	fmt.Println("[3/4] 동적 테이블 마이그레이션 중...")

	// 모든 게시판 목록 가져오기
	boards, err := boardService.ListBoards(context.Background(), !config.IncludeInactive)
	if err != nil {
		return fmt.Errorf("게시판 목록 가져오기 실패: %w", err)
	}

	fmt.Printf("  총 %d개의 게시판을 마이그레이션합니다\n", len(boards))

	// 게시판별 마이그레이션
	for _, board := range boards {
		// 건너뛸 테이블 확인
		if isTableSkipped(board.TableName, config.SkipTables) {
			fmt.Printf("  - 게시판 '%s' (%s) 건너뛰기 (사용자 설정에 의해 제외)\n", board.Name, board.TableName)
			continue
		}

		fmt.Printf("  - 게시판 '%s' (%s) 마이그레이션 중...\n", board.Name, board.TableName)

		// 게시판 테이블이 대상 DB에 존재하는지 확인
		if err := ensureDynamicTableExists(config, board, dynamicBoardService); err != nil {
			config.addError(fmt.Errorf("게시판 '%s' 테이블 생성 실패: %w", board.Name, err))
			continue
		}

		// 테이블 데이터 복사
		err := migrateTableData(config, board.TableName)
		if err != nil {
			config.addError(fmt.Errorf("게시판 '%s' 데이터 마이그레이션 실패: %w", board.Name, err))
			if len(config.Errors) > config.MaxErrorsBeforeExit {
				return fmt.Errorf("오류가 너무 많아 마이그레이션이 중단되었습니다")
			}
		}
	}

	fmt.Println("동적 테이블 마이그레이션 완료")
	return nil
}

// ensureDynamicTableExists는 대상 DB에 동적 테이블이 존재하는지 확인하고, 없으면 생성합니다
func ensureDynamicTableExists(config *DataMigrationConfig, board *models.Board, dynamicBoardService service.DynamicBoardService) error {
	// 대상 DB에 테이블이 존재하는지 확인
	var exists bool
	ctx := context.Background()

	// DB 종류에 따라 다른 쿼리 사용
	switch config.TargetDBConfig.DBDriver {
	case "postgres":
		err := config.TargetDB.NewSelect().
			TableExpr("information_schema.tables").
			Column("count(*) > 0").
			Where("table_schema = 'public'").
			Where("table_name = ?", board.TableName).
			Scan(ctx, &exists)
		if err != nil {
			return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
		}
	case "mysql", "mariadb":
		err := config.TargetDB.NewSelect().
			TableExpr("information_schema.tables").
			Column("count(*) > 0").
			Where("table_schema = ?", config.TargetDBConfig.DBName).
			Where("table_name = ?", board.TableName).
			Scan(ctx, &exists)
		if err != nil {
			return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
		}
	case "sqlite":
		err := config.TargetDB.NewSelect().
			TableExpr("sqlite_master").
			Column("count(*) > 0").
			Where("type = 'table'").
			Where("name = ?", board.TableName).
			Scan(ctx, &exists)
		if err != nil {
			return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
		}
	default:
		return fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", config.TargetDBConfig.DBDriver)
	}

	// 테이블이 없으면 생성
	if !exists {
		// 필드 목록 가져오기
		var fields []*models.BoardField
		err := config.SourceDB.NewSelect().
			Model(&fields).
			Where("board_id = ?", board.ID).
			Order("sort_order ASC").
			Scan(ctx)
		if err != nil {
			return fmt.Errorf("게시판 필드 가져오기 실패: %w", err)
		}

		// 대상 DB에 테이블 생성
		targetDynamicService := service.NewDynamicBoardService(config.TargetDB)
		if err := targetDynamicService.CreateBoardTable(ctx, board, fields); err != nil {
			return fmt.Errorf("게시판 테이블 생성 실패: %w", err)
		}

		fmt.Printf("    게시판 테이블 '%s' 생성됨\n", board.TableName)
	}

	return nil
}

// migrateTableData는 테이블 데이터를 마이그레이션합니다
func migrateTableData(config *DataMigrationConfig, tableName string) error {
	ctx := context.Background()

	if config.VerboseLogging {
		fmt.Printf("  - 테이블 '%s' 데이터 마이그레이션 중...\n", tableName)
	}

	// 테이블 총 행 수 카운트
	var totalRows int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteTableName(config.SourceDBConfig.DBDriver, tableName))
	err := config.SourceDB.QueryRow(query).Scan(&totalRows)
	if err != nil {
		return fmt.Errorf("행 수 계산 실패: %w", err)
	}

	if totalRows == 0 {
		if config.VerboseLogging {
			fmt.Printf("    테이블 '%s'에 데이터가 없습니다\n", tableName)
		}
		return nil
	}

	// 테이블 구조 가져오기
	sourceColumns, err := getTableColumns(config.SourceDB, config.SourceDBConfig.DBDriver, tableName)
	if err != nil {
		return fmt.Errorf("소스 테이블 구조 가져오기 실패: %w", err)
	}

	targetColumns, err := getTableColumns(config.TargetDB, config.TargetDBConfig.DBDriver, tableName)
	if err != nil {
		return fmt.Errorf("대상 테이블 구조 가져오기 실패: %w", err)
	}

	// 공통 컬럼 찾기
	commonColumns, sourceColumnNames, targetColumnNames := findCommonColumns(sourceColumns, targetColumns)
	if len(commonColumns) == 0 {
		return fmt.Errorf("공통 컬럼이 없습니다")
	}

	// 데이터 배치 처리
	batchSize := config.BatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}

	totalBatches := (totalRows + batchSize - 1) / batchSize
	processedRows := 0

	// 대상 테이블의 기존 데이터 삭제 (필요 시)
	if shouldCleanTableBeforeMigration(tableName) {
		deleteQuery := fmt.Sprintf("DELETE FROM %s", quoteTableName(config.TargetDBConfig.DBDriver, tableName))
		_, err = config.TargetDB.ExecContext(ctx, deleteQuery)
		if err != nil {
			return fmt.Errorf("대상 테이블 데이터 삭제 실패: %w", err)
		}
		if config.VerboseLogging {
			fmt.Printf("    기존 대상 테이블 데이터 삭제됨\n")
		}
	}

	// 트랜잭션 활성화 여부 확인
	useTransaction := config.EnableTransactions && shouldUseTransactionForTable(tableName, totalRows)

	for batchNum := 0; batchNum < totalBatches; batchNum++ {
		offset := batchNum * batchSize
		limit := batchSize

		if config.VerboseLogging && totalBatches > 1 {
			fmt.Printf("    배치 %d/%d 처리 중... (오프셋: %d, 한계: %d)\n", batchNum+1, totalBatches, offset, limit)
		}

		// 소스 데이터 쿼리 구성
		sourceQuery := fmt.Sprintf("SELECT %s FROM %s LIMIT %d OFFSET %d",
			strings.Join(sourceColumnNames, ", "), quoteTableName(config.SourceDBConfig.DBDriver, tableName), limit, offset)

		// 소스 데이터 조회
		sourceRows, err := config.SourceDB.QueryContext(ctx, sourceQuery)
		if err != nil {
			return fmt.Errorf("소스 데이터 조회 실패: %w", err)
		}

		// 트랜잭션 시작 (필요 시)
		var tx *sql.Tx
		if useTransaction {
			tx, err = config.TargetDB.DB.Begin()
			if err != nil {
				sourceRows.Close()
				return fmt.Errorf("트랜잭션 시작 실패: %w", err)
			}
		}

		// 대상 테이블 준비
		// var values []string
		var placeholders []string
		for i := range commonColumns {
			// 데이터베이스별 플레이스홀더 형식
			switch config.TargetDBConfig.DBDriver {
			case "postgres":
				placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
			default:
				placeholders = append(placeholders, "?")
			}
		}

		// 행별 처리
		batchInsertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			quoteTableName(config.TargetDBConfig.DBDriver, tableName),
			strings.Join(targetColumnNames, ", "),
			strings.Join(placeholders, ", "))

		rowsInBatch := 0
		for sourceRows.Next() {
			// 행 데이터 변수 준비
			rowValues := make([]interface{}, len(commonColumns))
			valuePtrs := make([]interface{}, len(commonColumns))
			for i := range rowValues {
				valuePtrs[i] = &rowValues[i]
			}

			// 행 데이터 읽기
			if err := sourceRows.Scan(valuePtrs...); err != nil {
				sourceRows.Close()
				if useTransaction {
					tx.Rollback()
				}
				return fmt.Errorf("행 데이터 읽기 실패: %w", err)
			}

			// 데이터 변환 (필요 시)
			for i, val := range rowValues {
				rowValues[i] = convertValue(val, tableName, commonColumns[i], config.SourceDBConfig.DBDriver, config.TargetDBConfig.DBDriver)
			}

			// 행 삽입
			var err error
			if useTransaction {
				_, err = tx.ExecContext(ctx, batchInsertSQL, rowValues...)
			} else {
				_, err = config.TargetDB.ExecContext(ctx, batchInsertSQL, rowValues...)
			}

			if err != nil {
				// JSON으로 행 데이터 직렬화 (디버깅용)
				rowJSON, _ := json.Marshal(rowValues)
				sourceRows.Close()
				if useTransaction {
					tx.Rollback()
				}
				return fmt.Errorf("행 삽입 실패: %w\n데이터: %s", err, rowJSON)
			}

			rowsInBatch++
		}

		sourceRows.Close()

		// 트랜잭션 커밋 (필요 시)
		if useTransaction {
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("트랜잭션 커밋 실패: %w", err)
			}
		}

		processedRows += rowsInBatch
	}

	if config.VerboseLogging {
		fmt.Printf("    테이블 '%s': %d/%d 행 처리 완료\n", tableName, processedRows, totalRows)
	}

	return nil
}

// getTableColumns는 테이블의 컬럼 정보를 가져옵니다
func getTableColumns(db *bun.DB, driver string, tableName string) ([]ColumnMetadata, error) {
	ctx := context.Background()
	var columns []ColumnMetadata

	switch driver {
	case "postgres":
		// PostgreSQL용 쿼리
		var results []struct {
			ColumnName string         `bun:"column_name"`
			DataType   string         `bun:"data_type"`
			IsNullable string         `bun:"is_nullable"`
			Default    sql.NullString `bun:"column_default"`
		}

		err := db.NewSelect().
			TableExpr("information_schema.columns").
			Column("column_name", "data_type", "is_nullable", "column_default").
			Where("table_schema = 'public'").
			Where("table_name = ?", tableName).
			Order("ordinal_position").
			Scan(ctx, &results)

		if err != nil {
			return nil, fmt.Errorf("컬럼 정보 조회 실패: %w", err)
		}

		for _, r := range results {
			columns = append(columns, ColumnMetadata{
				Name:     r.ColumnName,
				Type:     r.DataType,
				Nullable: r.IsNullable == "YES",
				Default:  r.Default,
			})
		}

	case "mysql", "mariadb":
		// MySQL/MariaDB용 쿼리
		var results []struct {
			ColumnName string         `bun:"COLUMN_NAME"`
			DataType   string         `bun:"DATA_TYPE"`
			IsNullable string         `bun:"IS_NULLABLE"`
			Default    sql.NullString `bun:"COLUMN_DEFAULT"`
		}

		query := db.NewSelect().
			TableExpr("information_schema.columns").
			Column("COLUMN_NAME", "DATA_TYPE", "IS_NULLABLE", "COLUMN_DEFAULT").
			Where("TABLE_SCHEMA = ?", db.DB.Stats().OpenConnections).
			Where("TABLE_NAME = ?", tableName).
			Order("ORDINAL_POSITION")

		err := query.Scan(ctx, &results)
		if err != nil {
			// 데이터베이스 이름 직접 지정 시도
			dbName := ""
			var dbNameRow struct {
				DatabaseName string `bun:"DATABASE()"`
			}
			if err := db.NewSelect().ColumnExpr("DATABASE()").Scan(ctx, &dbNameRow); err == nil {
				dbName = dbNameRow.DatabaseName
			}

			if dbName != "" {
				err = db.NewSelect().
					TableExpr("information_schema.columns").
					Column("COLUMN_NAME", "DATA_TYPE", "IS_NULLABLE", "COLUMN_DEFAULT").
					Where("TABLE_SCHEMA = ?", dbName).
					Where("TABLE_NAME = ?", tableName).
					Order("ORDINAL_POSITION").
					Scan(ctx, &results)
			}

			if err != nil {
				return nil, fmt.Errorf("컬럼 정보 조회 실패: %w", err)
			}
		}

		for _, r := range results {
			columns = append(columns, ColumnMetadata{
				Name:     r.ColumnName,
				Type:     r.DataType,
				Nullable: r.IsNullable == "YES",
				Default:  r.Default,
			})
		}

	case "sqlite":
		// SQLite용 쿼리
		// var results []struct {
		// 	Name    string         `bun:"name"`
		// 	Type    string         `bun:"type"`
		// 	NotNull int            `bun:"notnull"`
		// 	Default sql.NullString `bun:"dflt_value"`
		// }

		err := db.QueryRow(fmt.Sprintf("PRAGMA table_info(%s)", tableName)).Scan()
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
		}

		rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", tableName))
		if err != nil {
			return nil, fmt.Errorf("컬럼 정보 조회 실패: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var cid int
			var name, typeName string
			var notNull int
			var dfltValue sql.NullString
			var pk int

			if err := rows.Scan(&cid, &name, &typeName, &notNull, &dfltValue, &pk); err != nil {
				return nil, fmt.Errorf("컬럼 정보 스캔 실패: %w", err)
			}

			columns = append(columns, ColumnMetadata{
				Name:     name,
				Type:     typeName,
				Nullable: notNull == 0,
				Default:  dfltValue,
			})
		}

	default:
		return nil, fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", driver)
	}

	return columns, nil
}

// findCommonColumns는 두 테이블의 공통 컬럼을 찾습니다
func findCommonColumns(sourceColumns, targetColumns []ColumnMetadata) ([]ColumnMetadata, []string, []string) {
	var common []ColumnMetadata
	var sourceNames []string
	var targetNames []string

	sourceMap := make(map[string]ColumnMetadata)
	for _, col := range sourceColumns {
		sourceMap[strings.ToLower(col.Name)] = col
	}

	targetMap := make(map[string]ColumnMetadata)
	for _, col := range targetColumns {
		targetMap[strings.ToLower(col.Name)] = col
	}

	// 소스 컬럼을 기준으로 공통 컬럼 찾기
	for _, sourceCol := range sourceColumns {
		lowerName := strings.ToLower(sourceCol.Name)
		if targetCol, ok := targetMap[lowerName]; ok {
			common = append(common, sourceCol)
			sourceNames = append(sourceNames, quoteColumnName(sourceCol.Name))
			targetNames = append(targetNames, quoteColumnName(targetCol.Name))
		}
	}

	return common, sourceNames, targetNames
}

// quoteTableName은 데이터베이스 드라이버에 따라 테이블 이름을 인용 부호로 묶습니다
func quoteTableName(driver, tableName string) string {
	switch driver {
	case "postgres":
		return fmt.Sprintf("\"%s\"", tableName)
	case "mysql", "mariadb":
		return fmt.Sprintf("`%s`", tableName)
	default:
		return tableName
	}
}

// quoteColumnName은 데이터베이스 드라이버에 따라 컬럼 이름을 인용 부호로 묶습니다
func quoteColumnName(columnName string) string {
	return columnName
}

// convertValue는 소스 DB의 값을 대상 DB에 맞게 변환합니다
func convertValue(val interface{}, tableName string, column ColumnMetadata, sourceDriver, targetDriver string) interface{} {
	if val == nil {
		return nil
	}

	// 불리언 타입 변환
	if sourceDriver == "sqlite" && targetDriver != "sqlite" {
		// SQLite의 불리언은 0/1 정수로 저장됨
		if column.Type == "INTEGER" || column.Type == "BOOLEAN" || column.Type == "TINYINT" {
			if i, ok := val.(int64); ok && (i == 0 || i == 1) {
				if targetDriver == "postgres" {
					return i == 1
				}
			}
		}
	} else if sourceDriver != "sqlite" && targetDriver == "sqlite" {
		// PostgreSQL/MySQL의 불리언을 SQLite의 0/1로 변환
		if b, ok := val.(bool); ok {
			if b {
				return 1
			}
			return 0
		}
	}

	// MySQL에서 PostgreSQL로 날짜/시간 변환
	if (sourceDriver == "mysql" || sourceDriver == "mariadb") && targetDriver == "postgres" {
		if s, ok := val.(string); ok && strings.Contains(column.Type, "time") {
			// MySQL은 가끔 문자열로 날짜 반환
			t, err := time.Parse("2006-01-02 15:04:05", s)
			if err == nil {
				return t
			}
		}
	}

	// PostgreSQL에서 MySQL로 날짜/시간 변환
	if sourceDriver == "postgres" && (targetDriver == "mysql" || targetDriver == "mariadb") {
		if t, ok := val.(time.Time); ok {
			return t.Format("2006-01-02 15:04:05")
		}
	}

	return val
}

// shouldUseTransactionForTable은 해당 테이블에 트랜잭션을 사용해야 하는지 결정합니다
func shouldUseTransactionForTable(tableName string, totalRows int) bool {
	// 대규모 테이블은 트랜잭션 없이 처리
	if totalRows > 10000 {
		return false
	}

	// 빈번하게 접근되는 테이블은 트랜잭션 없이 처리 (개별 삽입)
	highTrafficTables := map[string]bool{
		"referrer_stats": true,
	}

	return !highTrafficTables[tableName]
}

// shouldCleanTableBeforeMigration은 마이그레이션 전에 테이블을 정리해야 하는지 결정합니다
func shouldCleanTableBeforeMigration(tableName string) bool {
	// 이력 데이터는 정리하지 않음
	historyTables := map[string]bool{
		"referrer_stats": true,
	}

	return !historyTables[tableName]
}

// isTableSkipped는 주어진 테이블이 건너뛰기 목록에 있는지 확인합니다
func isTableSkipped(tableName string, skipTables []string) bool {
	for _, skipTable := range skipTables {
		if skipTable == tableName {
			return true
		}
	}
	return false
}

// resetSequences는 ID 시퀀스/자동증가 값을 재설정합니다
func resetSequences(config *DataMigrationConfig) error {
	fmt.Println("[4/4] 시퀀스/자동증가 값 재설정 중...")

	if config.TargetDBConfig.DBDriver == "postgres" {
		// PostgreSQL 시퀀스 재설정
		tables := getBasicTables()
		for _, tableName := range tables {
			// 시퀀스 재설정이 필요한 테이블만 처리
			if !needsSequenceReset(tableName) {
				continue
			}

			resetSQL := fmt.Sprintf(
				"SELECT setval(pg_get_serial_sequence('%s', 'id'), COALESCE((SELECT MAX(id) FROM \"%s\"), 1));",
				tableName, tableName)

			_, err := config.TargetDB.ExecContext(context.Background(), resetSQL)
			if err != nil {
				return fmt.Errorf("테이블 '%s' 시퀀스 재설정 실패: %w", tableName, err)
			}

			if config.VerboseLogging {
				fmt.Printf("  - 테이블 '%s' 시퀀스 재설정 완료\n", tableName)
			}
		}
	} else if config.TargetDBConfig.DBDriver == "mysql" || config.TargetDBConfig.DBDriver == "mariadb" {
		// MySQL/MariaDB 자동증가 값 재설정
		tables := getBasicTables()
		for _, tableName := range tables {
			// 자동증가 재설정이 필요한 테이블만 처리
			if !needsSequenceReset(tableName) {
				continue
			}

			// 현재 최대 ID 값 조회
			var maxID int64
			err := config.TargetDB.NewSelect().
				TableExpr(tableName).
				ColumnExpr("COALESCE(MAX(id), 0) as max_id").
				Scan(context.Background(), &maxID)

			if err != nil {
				return fmt.Errorf("테이블 '%s' 최대 ID 조회 실패: %w", tableName, err)
			}

			// 자동증가 값 재설정
			if maxID > 0 {
				resetSQL := fmt.Sprintf("ALTER TABLE `%s` AUTO_INCREMENT = %d;", tableName, maxID+1)
				_, err := config.TargetDB.ExecContext(context.Background(), resetSQL)
				if err != nil {
					return fmt.Errorf("테이블 '%s' 자동증가 값 재설정 실패: %w", tableName, err)
				}

				if config.VerboseLogging {
					fmt.Printf("  - 테이블 '%s' 자동증가 값 재설정 완료 (ID: %d)\n", tableName, maxID+1)
				}
			}
		}
	}

	// SQLite는 특별한 재설정이 필요 없음

	fmt.Println("시퀀스/자동증가 값 재설정 완료")
	return nil
}

// needsSequenceReset은 테이블에 시퀀스 재설정이 필요한지 확인합니다
func needsSequenceReset(tableName string) bool {
	// 시퀀스 재설정이 필요 없는 테이블
	noResetTables := map[string]bool{
		"goose_db_version": true,
		"referrer_stats":   true,
	}

	return !noResetTables[tableName]
}

// addError는 마이그레이션 오류 목록에 오류를 추가합니다
func (c *DataMigrationConfig) addError(err error) {
	if c.VerboseLogging {
		log.Printf("오류: %v", err)
	}
	c.Errors = append(c.Errors, err)
}
