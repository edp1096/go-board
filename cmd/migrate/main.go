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

	goboard "github.com/edp1096/go-board"
	"github.com/edp1096/go-board/config"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite" // SQLite 드라이버 추가
)

func main() {
	// 명령행 인자 파싱
	driver := flag.String("driver", "", "데이터베이스 드라이버 (mysql, postgres, sqlite)")
	operation := flag.String("op", "up", "마이그레이션 작업 (up, down, status, purge)")
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

	// SQLite가 아닌 경우에만 데이터베이스 존재 여부 확인
	if cfg.DBDriver != "sqlite" {
		// 데이터베이스 존재 여부 확인 및 없으면 생성
		if err := ensureDatabaseExists(cfg); err != nil {
			log.Fatalf("데이터베이스 생성 실패: %v", err)
		}
	} else {
		// SQLite인 경우 디렉토리 확인
		dir := filepath.Dir(cfg.DBPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("SQLite 데이터베이스 디렉토리 생성 실패: %v", err)
		}
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

	switch cfg.DBDriver {
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
	if err := runMigration(sqlDB, cfg.DBDriver, migrationsDir, *operation, cfg.DBName); err != nil {
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
	// 마이그레이션 작업 실행
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
