package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	goboard "go-board"
	"go-board/config"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	// 명령행 인자 파싱
	driver := flag.String("driver", "", "데이터베이스 드라이버 (mysql 또는 postgres)")
	operation := flag.String("op", "up", "마이그레이션 작업 (up, down, reset, status, create, redo, version)")
	newMigration := flag.String("name", "", "새 마이그레이션 이름 (create 명령에만 사용)")
	flag.Parse()

	// 환경 설정 로드
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("설정을 로드할 수 없습니다: %v", err)
	}

	// 드라이버 설정 오버라이드 (플래그가 제공된 경우)
	if *driver != "" {
		cfg.DBDriver = *driver
	}

	// 데이터베이스 존재 여부 확인 및 없으면 생성
	if err := ensureDatabaseExists(cfg); err != nil {
		log.Fatalf("데이터베이스 생성 실패: %v", err)
	}

	// 데이터베이스 연결
	database, err := config.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("데이터베이스에 연결할 수 없습니다: %v", err)
	}
	defer database.Close()

	fmt.Printf("데이터베이스 '%s'가 준비되었습니다\n", cfg.DBName)

	// SQL DB 인스턴스 가져오기
	sqlDB := database.DB.DB

	// 마이그레이션 설정
	goose.SetBaseFS(nil) // 기본 파일시스템 설정 초기화

	// 방언 설정
	if err := goose.SetDialect(cfg.DBDriver); err != nil {
		log.Fatalf("데이터베이스 방언 설정 실패: %v", err)
	}

	// 임베디드 마이그레이션 파일 시스템 가져오기
	var migrationFS fs.FS
	var migrationsDir string
	if cfg.DBDriver == "postgres" {
		// PostgreSQL 마이그레이션 설정
		subFS, err := fs.Sub(goboard.PostgresMigrationsFS, "migrations/postgres")
		if err != nil {
			log.Fatalf("PostgreSQL 마이그레이션 파일 접근 실패: %v", err)
		}
		migrationFS = subFS
		migrationsDir = "."
	} else {
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
	if err := runMigration(sqlDB, cfg.DBDriver, migrationsDir, *operation, *newMigration); err != nil {
		log.Fatalf("마이그레이션 실패: %v", err)
	}
}

// ensureDatabaseExists는 데이터베이스가 존재하는지 확인하고 없으면 생성합니다
func ensureDatabaseExists(cfg *config.Config) error {
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

	default:
		return fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", cfg.DBDriver)
	}

	return nil
}

// runMigration은 지정된 마이그레이션 작업을 실행합니다
func runMigration(db *sql.DB, driver, migrationsDir, operation, newMigration string) error {
	// 마이그레이션 작업 실행
	var err error
	switch operation {
	case "up":
		fmt.Println("마이그레이션 적용 중...")
		err = goose.Up(db, migrationsDir)
	case "down":
		fmt.Println("마이그레이션 롤백 중...")
		err = goose.Down(db, migrationsDir)
	case "reset":
		fmt.Println("마이그레이션 초기화 중...")
		err = goose.Reset(db, migrationsDir)
	case "status":
		fmt.Println("마이그레이션 상태 확인 중...")
		err = goose.Status(db, migrationsDir)
	case "create":
		// create는 임베디드 파일 시스템에서 직접 수행할 수 없으므로
		// 임시 디렉토리에 생성해야 합니다.
		if newMigration == "" {
			return fmt.Errorf("새 마이그레이션 이름을 지정해야 합니다 (-name 플래그 사용)")
		}

		// 임시 디렉토리 생성
		tmpDir, err := os.MkdirTemp("", "migrations")
		if err != nil {
			return fmt.Errorf("임시 디렉토리 생성 실패: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		fmt.Printf("새 마이그레이션 생성 중: %s (임시 위치: %s)\n", newMigration, tmpDir)

		// 임시 디렉토리로 변경 (goose.Create가 현재 디렉토리 기준으로 동작)
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("현재 작업 디렉토리 확인 실패: %w", err)
		}

		if err := os.Chdir(tmpDir); err != nil {
			return fmt.Errorf("작업 디렉토리 변경 실패: %w", err)
		}

		// 작업 완료 후 원래 디렉토리로 돌아가기
		defer os.Chdir(wd)

		// 마이그레이션 파일 생성
		err = goose.Create(db, ".", newMigration, "sql")
		if err != nil {
			return fmt.Errorf("마이그레이션 생성 실패: %w", err)
		}

		// 생성된 파일 경로 확인
		upFile := filepath.Join(tmpDir, fmt.Sprintf("%s.sql", newMigration))

		// SQL 파일 내용 읽기
		upContent, err := os.ReadFile(upFile)
		if err != nil {
			return fmt.Errorf("생성된 마이그레이션 파일 읽기 실패: %w", err)
		}

		// 마이그레이션 디렉토리 확인 및 생성
		var targetDir string
		if driver == "postgres" {
			targetDir = "migrations/postgres"
		} else {
			targetDir = "migrations/mysql"
		}

		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("마이그레이션 디렉토리 생성 실패: %w", err)
			}
		}

		// 새 마이그레이션 파일 경로
		targetFile := filepath.Join(targetDir, filepath.Base(upFile))

		// 임시 파일을 실제 마이그레이션 디렉토리로 복사
		if err := os.WriteFile(targetFile, upContent, 0644); err != nil {
			return fmt.Errorf("마이그레이션 파일 복사 실패: %w", err)
		}

		fmt.Printf("새 마이그레이션 파일이 생성되었습니다: %s\n", targetFile)

	case "redo":
		fmt.Println("마지막 마이그레이션 재실행 중...")
		err = goose.Redo(db, migrationsDir)
	case "version":
		fmt.Println("현재 마이그레이션 버전 확인 중...")
		err = goose.Version(db, migrationsDir)
	default:
		return fmt.Errorf("알 수 없는 마이그레이션 작업: %s", operation)
	}

	if err != nil {
		return err
	}

	fmt.Printf("마이그레이션 작업 '%s'가 성공적으로 완료되었습니다\n", operation)
	return nil
}
