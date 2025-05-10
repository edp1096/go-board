// cmd/migrate/add_is_private_column.go
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/edp1096/toy-board/config"
	"github.com/edp1096/toy-board/internal/utils"
)

func main() {
	// 설정 로드
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("설정을 로드할 수 없습니다: %v", err)
	}

	// 데이터베이스 연결
	database, err := config.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("데이터베이스에 연결할 수 없습니다: %v", err)
	}
	defer database.Close()
	db := database.DB

	// 게시판 테이블 목록 가져오기
	var tableNames []string
	query := "SELECT table_name FROM boards"
	rows, err := db.QueryContext(context.Background(), query)
	if err != nil {
		log.Fatalf("게시판 목록 조회 실패: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Printf("행 스캔 실패: %v", err)
			continue
		}
		tableNames = append(tableNames, tableName)
	}

	// 데이터베이스 유형에 따라 다른 SQL 실행
	for _, tableName := range tableNames {
		var alterSQL string

		if utils.IsPostgres(db) {
			alterSQL = fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN IF NOT EXISTS is_private BOOLEAN NOT NULL DEFAULT false", tableName)
		} else if utils.IsMySQL(db) {
			// MySQL에서는 IF NOT EXISTS가 지원되지 않으므로 컬럼 존재 여부 확인
			checkSQL := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = '%s' AND column_name = 'is_private'", tableName)
			var count int
			if err := db.QueryRowContext(context.Background(), checkSQL).Scan(&count); err != nil {
				log.Printf("컬럼 존재 여부 확인 실패: %v", err)
				continue
			}

			if count == 0 {
				alterSQL = fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `is_private` BOOLEAN NOT NULL DEFAULT false", tableName)
			} else {
				log.Printf("테이블 %s에 이미 is_private 컬럼이 존재합니다", tableName)
				continue
			}
		} else if utils.IsSQLite(db) {
			// SQLite는 컬럼 존재 여부 확인이 어려우므로 바로 추가
			alterSQL = fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN is_private INTEGER NOT NULL DEFAULT 0", tableName)
		}

		// SQL 실행
		_, err := db.ExecContext(context.Background(), alterSQL)
		if err != nil {
			// SQLite에서 이미 컬럼이 존재하는 경우 duplicate column name 오류 발생
			if utils.IsSQLite(db) && strings.Contains(err.Error(), "duplicate column name") {
				log.Printf("테이블 %s에 이미 is_private 컬럼이 존재합니다", tableName)
				continue
			}
			log.Printf("테이블 %s 수정 실패: %v", tableName, err)
		} else {
			log.Printf("테이블 %s에 is_private 컬럼이 추가되었습니다", tableName)
		}
	}

	log.Println("모든 게시판 테이블 마이그레이션이 완료되었습니다.")
}
