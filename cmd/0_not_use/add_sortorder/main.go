// cmd/migrate/add_sort_order_column.go
package main

import (
	"context"
	"log"
	"strings"

	"github.com/edp1096/go-board/config"
	"github.com/edp1096/go-board/internal/utils"
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

	// boards 테이블에 sort_order 컬럼 추가
	// 데이터베이스 유형에 따라 다른 SQL 실행
	var alterSQL string

	if utils.IsPostgres(db) {
		alterSQL = "ALTER TABLE boards ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0"
	} else if utils.IsMySQL(db) {
		// MySQL에서는 IF NOT EXISTS가 지원되지 않으므로 컬럼 존재 여부 확인
		checkSQL := "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'boards' AND column_name = 'sort_order'"
		var count int
		if err := db.QueryRowContext(context.Background(), checkSQL).Scan(&count); err != nil {
			log.Fatalf("컬럼 존재 여부 확인 실패: %v", err)
		}

		if count == 0 {
			alterSQL = "ALTER TABLE `boards` ADD COLUMN `sort_order` INT NOT NULL DEFAULT 0"
		} else {
			log.Println("boards 테이블에 이미 sort_order 컬럼이 존재합니다")
			return
		}
	} else if utils.IsSQLite(db) {
		// SQLite는 컬럼 존재 여부 확인이 어려우므로 바로 추가
		alterSQL = "ALTER TABLE boards ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0"
	}

	// SQL 실행
	_, err = db.ExecContext(context.Background(), alterSQL)
	if err != nil {
		// SQLite에서 이미 컬럼이 존재하는 경우 duplicate column name 오류 발생
		if utils.IsSQLite(db) && strings.Contains(err.Error(), "duplicate column name") {
			log.Println("boards 테이블에 이미 sort_order 컬럼이 존재합니다")
			return
		}
		log.Fatalf("boards 테이블 수정 실패: %v", err)
	}

	// 기존 레코드들의 sort_order 값 설정 (ID 순서대로)
	updateSQL := ""
	if utils.IsPostgres(db) {
		updateSQL = "UPDATE boards SET sort_order = id"
	} else if utils.IsMySQL(db) {
		updateSQL = "UPDATE `boards` SET `sort_order` = `id`"
	} else if utils.IsSQLite(db) {
		updateSQL = "UPDATE boards SET sort_order = id"
	}

	_, err = db.ExecContext(context.Background(), updateSQL)
	if err != nil {
		log.Printf("sort_order 값 업데이트 실패: %v", err)
	}

	log.Println("boards 테이블에 sort_order 컬럼이 성공적으로 추가되었습니다.")
	log.Println("모든 기존 게시판은 ID 순서대로 정렬되도록 업데이트되었습니다.")
}
