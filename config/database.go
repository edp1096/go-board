// config/database.go
package config

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// Database는 데이터베이스 연결 관련 설정을 관리하는 구조체
type Database struct {
	DB     *bun.DB
	Config *Config
}

// ConnectDatabase는 설정에 따라 데이터베이스에 연결하고 Bun DB 인스턴스를 반환
func ConnectDatabase(cfg *Config) (*bun.DB, error) {
	var sqldb *sql.DB
	var err error

	switch cfg.DBDriver {
	case "postgres":
		// PostgreSQL 연결 설정
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
		sqldb = sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	case "mysql", "mariadb":
		// MySQL/MariaDB 연결 설정
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
		sqldb, err = sql.Open("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("MySQL 연결 실패: %w", err)
		}
	default:
		return nil, fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", cfg.DBDriver)
	}

	// 기본 연결 설정
	sqldb.SetMaxOpenConns(25)
	sqldb.SetMaxIdleConns(5)
	sqldb.SetConnMaxLifetime(5 * time.Minute)

	// 연결 테스트 - 컨텍스트 제한 시간 설정
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqldb.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("데이터베이스 연결 확인 실패: %w", err)
	}

	// Bun DB 인스턴스 생성
	var db *bun.DB
	if cfg.DBDriver == "postgres" {
		db = bun.NewDB(sqldb, pgdialect.New())
	} else {
		db = bun.NewDB(sqldb, mysqldialect.New())
	}

	return db, nil
}

// NewDatabase는 Database 인스턴스를 생성하고 반환
func NewDatabase(cfg *Config) (*Database, error) {
	db, err := ConnectDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &Database{
		DB:     db,
		Config: cfg,
	}, nil
}

// Close는 데이터베이스 연결을 안전하게 종료
func (d *Database) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

// IsPostgres는 현재 드라이버가 PostgreSQL인지 확인
func (d *Database) IsPostgres() bool {
	return d.Config.DBDriver == "postgres"
}

// IsMySQLOrMariaDB는 현재 드라이버가 MySQL/MariaDB인지 확인
func (d *Database) IsMySQLOrMariaDB() bool {
	return d.Config.DBDriver == "mysql" || d.Config.DBDriver == "mariadb"
}
