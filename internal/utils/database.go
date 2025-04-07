// internal/utils/database.go
package utils

import "github.com/uptrace/bun"

// IsPostgres는 데이터베이스가 PostgreSQL인지 확인합니다
func IsPostgres(db *bun.DB) bool {
	dialectName := db.Dialect().Name()
	return dialectName.String() == "pg" || dialectName.String() == "postgres"
}

// IsSQLite는 데이터베이스가 SQLite인지 확인합니다
func IsSQLite(db *bun.DB) bool {
	dialectName := db.Dialect().Name()
	return dialectName.String() == "sqlite" || dialectName.String() == "sqlite3"
}

// IsMySQL은 데이터베이스가 MySQL 또는 MariaDB인지 확인합니다
func IsMySQL(db *bun.DB) bool {
	dialectName := db.Dialect().Name()
	return dialectName.String() == "mysql" || dialectName.String() == "mariadb"
}
