package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"

	goboard "github.com/edp1096/go-board"
	"github.com/pressly/goose/v3"
)

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

	// MySQL/MariaDB인 경우 특별 처리
	if config.TargetDBConfig.DBDriver == "mysql" || config.TargetDBConfig.DBDriver == "mariadb" {
		// MySQL 마이그레이션 특별 처리
		if err := applyMySQLMigrations(sqlDB, migrationFS, migrationsDir); err != nil {
			return fmt.Errorf("MySQL 스키마 마이그레이션 실패: %w", err)
		}
	} else {
		// 기존 방식 사용
		if err := goose.Up(sqlDB, migrationsDir); err != nil {
			return fmt.Errorf("스키마 마이그레이션 실패: %w", err)
		}
	}

	fmt.Println("기본 스키마 마이그레이션 완료")
	return nil
}

// applyMySQLMigrations는 MySQL/MariaDB용 마이그레이션을 적용합니다
func applyMySQLMigrations(db *sql.DB, migrationFS fs.FS, migrationsDir string) error {
	// 현재 적용된 마이그레이션 버전 확인
	current, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("현재 마이그레이션 버전 확인 실패: %w", err)
	}

	// 직접 파일 목록 가져오기
	entries, err := fs.ReadDir(migrationFS, migrationsDir)
	if err != nil {
		return fmt.Errorf("마이그레이션 파일 목록 가져오기 실패: %w", err)
	}

	// 버전별로 정렬
	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}
	sort.Strings(migrationFiles)

	// 각 마이그레이션 파일 처리
	for _, fileName := range migrationFiles {
		// 파일명에서 버전 추출 (예: "001_create_tables.sql" -> 1)
		versionStr := strings.Split(fileName, "_")[0]
		version, err := strconv.ParseInt(versionStr, 10, 64)
		if err != nil {
			return fmt.Errorf("마이그레이션 버전 파싱 실패: %w", err)
		}

		// 이미 적용된 마이그레이션은 건너뛰기
		if version <= current {
			continue
		}

		fmt.Printf("마이그레이션 적용 중: %s\n", fileName)

		// 파일 내용 읽기
		content, err := fs.ReadFile(migrationFS, path.Join(migrationsDir, fileName))
		if err != nil {
			return fmt.Errorf("마이그레이션 파일 읽기 실패: %w", err)
		}

		// 쿼리 분리하기
		sqlStr := string(content)
		statements := strings.Split(sqlStr, ";")

		// 트랜잭션 시작
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("트랜잭션 시작 실패: %w", err)
		}

		// 개별 쿼리 실행
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}

			if _, err := tx.Exec(stmt); err != nil {
				tx.Rollback()
				return fmt.Errorf("SQL 실행 실패 (%s): %w", stmt, err)
			}
		}

		// 마이그레이션 버전 기록
		versionSQL := fmt.Sprintf("INSERT INTO goose_db_version (version_id, is_applied) VALUES (%d, true)", version)
		if _, err := tx.Exec(versionSQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("마이그레이션 버전 기록 실패: %w", err)
		}

		// 트랜잭션 커밋
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("트랜잭션 커밋 실패: %w", err)
		}

		fmt.Printf("마이그레이션 성공: %s\n", fileName)
	}

	return nil
}
