package main

import (
	"database/sql"
	"embed"
	"flag"
	"fmt"
	goboard "go-board"
	"go-board/config"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	// 명령행 인자 파싱
	driver := flag.String("driver", "", "데이터베이스 드라이버 (mysql 또는 postgres)")
	operation := flag.String("op", "up", "마이그레이션 작업 (up, down, status, redo, version)")
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

	// 데이터베이스 연결 가져오기
	sqlDB := database.DB.DB

	// 마이그레이션 디렉토리 준비
	migrationsDir, err := prepareMigrations(cfg.DBDriver)
	if err != nil {
		log.Fatalf("마이그레이션 파일 준비 실패: %v", err)
	}

	// 임시 디렉토리가 생성된 경우, 함수 종료 시 정리
	if strings.HasPrefix(migrationsDir, os.TempDir()) {
		defer os.RemoveAll(migrationsDir)
	}

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

// prepareMigrations는 마이그레이션 파일을 준비합니다
func prepareMigrations(driver string) (string, error) {
	// 먼저 실제 디렉토리가 있는지 확인
	var realDir string
	if driver == "postgres" {
		realDir = "./migrations/postgres"
	} else {
		realDir = "./migrations/mysql"
	}

	// 실제 디렉토리가 있으면 그것을 사용
	if _, err := os.Stat(realDir); err == nil {
		fmt.Printf("실제 마이그레이션 디렉토리 사용: %s\n", realDir)
		return realDir, nil
	}

	// 임시 디렉토리 생성
	tmpDir, err := os.MkdirTemp("", "migrations")
	if err != nil {
		return "", fmt.Errorf("임시 디렉토리 생성 실패: %w", err)
	}

	fmt.Printf("임베디드 마이그레이션에서 임시 디렉토리 생성: %s\n", tmpDir)

	// 임베디드 마이그레이션 파일 추출
	if driver == "postgres" {
		if err := extractMigrations(goboard.PostgresMigrationsFS, "migrations/postgres", tmpDir); err != nil {
			os.RemoveAll(tmpDir)
			return "", err
		}
	} else {
		if err := extractMigrations(goboard.MysqlMigrationsFS, "migrations/mysql", tmpDir); err != nil {
			os.RemoveAll(tmpDir)
			return "", err
		}
	}

	return tmpDir, nil
}

// extractMigrations는 임베디드 파일시스템에서 마이그레이션 파일을 추출합니다
func extractMigrations(embeddedFS embed.FS, basePath, targetDir string) error {
	// 파일 시스템에서 마이그레이션 디렉토리 접근
	entries, err := embeddedFS.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("마이그레이션 디렉토리 읽기 실패: %w", err)
	}

	// 모든 파일 순회
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// SQL 파일만 처리
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		// 파일 내용 읽기
		filePath := filepath.Join(basePath, name)
		content, err := embeddedFS.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("파일 읽기 실패: %w", err)
		}

		// 대상 경로 생성
		targetPath := filepath.Join(targetDir, name)

		// 파일 쓰기
		if err := os.WriteFile(targetPath, content, 0644); err != nil {
			return fmt.Errorf("파일 쓰기 실패: %w", err)
		}

		fmt.Printf("마이그레이션 파일 추출: %s\n", name)
	}

	return nil
}

// runMigration은 지정된 마이그레이션 작업을 실행합니다
func runMigration(db *sql.DB, driver, migrationsDir, operation, newMigration string) error {
	// Goose 디렉터리 설정
	goose.SetBaseFS(nil)

	// Goose 방언 설정
	if driver == "postgres" {
		if err := goose.SetDialect("postgres"); err != nil {
			return fmt.Errorf("PostgreSQL 방언 설정 실패: %w", err)
		}
	} else {
		if err := goose.SetDialect("mysql"); err != nil {
			return fmt.Errorf("MySQL 방언 설정 실패: %w", err)
		}
	}

	// Goose 명령 실행
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
		if newMigration == "" {
			return fmt.Errorf("새 마이그레이션 이름을 지정해야 합니다 (-name 플래그 사용)")
		}
		fmt.Printf("새 마이그레이션 생성 중: %s\n", newMigration)
		err = goose.Create(db, migrationsDir, newMigration, "sql")
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
