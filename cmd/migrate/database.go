package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/edp1096/toy-board/config"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
)

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
			fmt.Printf("데이터베이스 '%s'가 존재하지 않아 생성합니다\n", cfg.DBName)

			// 데이터베이스 이름에 특수문자가 있으면 큰따옴표로 감싸줌
			dbNameQuoted := cfg.DBName
			if strings.ContainsAny(dbNameQuoted, "-.") {
				dbNameQuoted = fmt.Sprintf("\"%s\"", dbNameQuoted)
			}

			_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbNameQuoted))
			if err != nil {
				return fmt.Errorf("데이터베이스 생성 실패: %w", err)
			}
			fmt.Printf("데이터베이스 '%s'가 성공적으로 생성되었습니다\n", cfg.DBName)
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
			fmt.Printf("데이터베이스 '%s'가 존재하지 않아 생성합니다\n", cfg.DBName)

			// 데이터베이스 이름에 특수문자가 있으면 백틱으로 감싸줌
			dbNameQuoted := cfg.DBName
			if strings.ContainsAny(dbNameQuoted, "-.") {
				dbNameQuoted = fmt.Sprintf("`%s`", dbNameQuoted)
			}

			_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbNameQuoted))
			if err != nil {
				return fmt.Errorf("데이터베이스 생성 실패: %w", err)
			}
			fmt.Printf("데이터베이스 '%s'가 성공적으로 생성되었습니다\n", cfg.DBName)
		}

	case "sqlite":
		// SQLite는 파일 기반이므로 별도의 생성 과정 불필요
		return nil

	default:
		return fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", cfg.DBDriver)
	}

	return nil
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
