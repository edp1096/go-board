// cmd/add_ipaddress/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/edp1096/go-board/config"
	"github.com/edp1096/go-board/internal/utils"
)

func main() {
	// // 명령줄 옵션 파싱
	// forceUpdate := flag.Bool("force", false, "컬럼이 이미 존재하더라도 강제로 업데이트")
	// flag.Parse()

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

	// 1-1. 댓글 테이블(comments)에 ip_address 컬럼 추가
	log.Println("댓글 테이블에 IP 주소 컬럼을 추가합니다...")
	ipColumnExistsInComments := false

	var commentAlterSQL string
	if utils.IsPostgres(db) {
		commentAlterSQL = "ALTER TABLE comments ADD COLUMN IF NOT EXISTS ip_address VARCHAR(90)"
	} else if utils.IsMySQL(db) {
		// MySQL에서는 IF NOT EXISTS가 지원되지 않으므로 컬럼 존재 여부 확인
		checkSQL := "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'comments' AND column_name = 'ip_address'"
		var count int
		if err := db.QueryRowContext(context.Background(), checkSQL).Scan(&count); err != nil {
			log.Printf("comments 테이블의 ip_address 컬럼 존재 여부 확인 실패: %v", err)
		} else if count == 0 {
			commentAlterSQL = "ALTER TABLE `comments` ADD COLUMN `ip_address` VARCHAR(90)"
		} else {
			ipColumnExistsInComments = true
			log.Printf("comments 테이블에 이미 ip_address 컬럼이 존재합니다")
		}
	} else if utils.IsSQLite(db) {
		// SQLite는 컬럼 존재 여부 확인이 어려우므로 바로 추가
		commentAlterSQL = "ALTER TABLE comments ADD COLUMN ip_address VARCHAR(90)"
	}

	// 컬럼이 없거나 SQLite인 경우에만 ALTER TABLE 실행
	if !ipColumnExistsInComments && commentAlterSQL != "" {
		_, err := db.ExecContext(context.Background(), commentAlterSQL)
		if err != nil {
			// SQLite에서 이미 컬럼이 존재하는 경우 duplicate column name 오류 발생
			if utils.IsSQLite(db) && strings.Contains(err.Error(), "duplicate column name") {
				log.Printf("comments 테이블에 이미 ip_address 컬럼이 존재합니다")
			} else {
				log.Printf("comments 테이블에 ip_address 컬럼 추가 실패: %v", err)
			}
		} else {
			log.Printf("comments 테이블에 ip_address 컬럼이 추가되었습니다")
		}
	}

	// 1-2. qna답변 테이블(qna_answers)에 ip_address 컬럼 추가
	log.Println("qna답변 테이블에 IP 주소 컬럼을 추가합니다...")
	ipColumnExistsInComments = false

	if utils.IsPostgres(db) {
		commentAlterSQL = "ALTER TABLE qna_answers ADD COLUMN IF NOT EXISTS ip_address VARCHAR(90)"
	} else if utils.IsMySQL(db) {
		// MySQL에서는 IF NOT EXISTS가 지원되지 않으므로 컬럼 존재 여부 확인
		checkSQL := "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'qna_answers' AND column_name = 'ip_address'"
		var count int
		if err := db.QueryRowContext(context.Background(), checkSQL).Scan(&count); err != nil {
			log.Printf("qna_answers 테이블의 ip_address 컬럼 존재 여부 확인 실패: %v", err)
		} else if count == 0 {
			commentAlterSQL = "ALTER TABLE `qna_answers` ADD COLUMN `ip_address` VARCHAR(90)"
		} else {
			ipColumnExistsInComments = true
			log.Printf("qna_answers 테이블에 이미 ip_address 컬럼이 존재합니다")
		}
	} else if utils.IsSQLite(db) {
		// SQLite는 컬럼 존재 여부 확인이 어려우므로 바로 추가
		commentAlterSQL = "ALTER TABLE qna_answers ADD COLUMN ip_address VARCHAR(90)"
	}

	// 컬럼이 없거나 SQLite인 경우에만 ALTER TABLE 실행
	if !ipColumnExistsInComments && commentAlterSQL != "" {
		_, err := db.ExecContext(context.Background(), commentAlterSQL)
		if err != nil {
			// SQLite에서 이미 컬럼이 존재하는 경우 duplicate column name 오류 발생
			if utils.IsSQLite(db) && strings.Contains(err.Error(), "duplicate column name") {
				log.Printf("qna_answers 테이블에 이미 ip_address 컬럼이 존재합니다")
			} else {
				log.Printf("qna_answers 테이블에 ip_address 컬럼 추가 실패: %v", err)
			}
		} else {
			log.Printf("qna_answers 테이블에 ip_address 컬럼이 추가되었습니다")
		}
	}

	// 2. 게시판 테이블 목록 가져오기
	var tableNames []string
	var boardIDs []int64
	var boardTypes []string
	query := "SELECT id, table_name, board_type FROM boards"
	rows, err := db.QueryContext(context.Background(), query)
	if err != nil {
		log.Fatalf("게시판 목록 조회 실패: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		var boardID int64
		var boardType string
		if err := rows.Scan(&boardID, &tableName, &boardType); err != nil {
			log.Printf("행 스캔 실패: %v", err)
			continue
		}

		tableNames = append(tableNames, tableName)
		boardIDs = append(boardIDs, boardID)
		boardTypes = append(boardTypes, boardType)
	}

	// 3. 각 게시판 테이블에 ip_address 컬럼 추가
	for i, tableName := range tableNames {
		boardID := boardIDs[i]
		boardType := boardTypes[i]

		log.Printf("처리 중: %s (ID: %d, Type: %s)", tableName, boardID, boardType)

		ipColumnExists := false

		// ip_address 컬럼 추가
		var ipAlterSQL string
		if utils.IsPostgres(db) {
			ipAlterSQL = fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN IF NOT EXISTS ip_address VARCHAR(90)", tableName)
		} else if utils.IsMySQL(db) {
			// MySQL에서는 IF NOT EXISTS가 지원되지 않으므로 컬럼 존재 여부 확인
			checkSQL := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = '%s' AND column_name = 'ip_address'", tableName)
			var count int
			if err := db.QueryRowContext(context.Background(), checkSQL).Scan(&count); err != nil {
				log.Printf("테이블 %s의 ip_address 컬럼 존재 여부 확인 실패: %v", tableName, err)
				continue
			}

			if count == 0 {
				ipAlterSQL = fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `ip_address` VARCHAR(90)", tableName)
			} else {
				ipColumnExists = true
				log.Printf("테이블 %s에 이미 ip_address 컬럼이 존재합니다", tableName)
			}
		} else if utils.IsSQLite(db) {
			// SQLite는 컬럼 존재 여부 확인이 어려우므로 바로 추가
			ipAlterSQL = fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN ip_address VARCHAR(90)", tableName)
		}

		// 컬럼이 없거나 SQLite인 경우에만 ALTER TABLE 실행
		if !ipColumnExists && ipAlterSQL != "" {
			_, err := db.ExecContext(context.Background(), ipAlterSQL)
			if err != nil {
				// SQLite에서 이미 컬럼이 존재하는 경우 duplicate column name 오류 발생
				if utils.IsSQLite(db) && strings.Contains(err.Error(), "duplicate column name") {
					ipColumnExists = true
					log.Printf("테이블 %s에 이미 ip_address 컬럼이 존재합니다", tableName)
				} else {
					log.Printf("테이블 %s에 ip_address 컬럼 추가 실패: %v", tableName, err)
				}
			} else {
				log.Printf("테이블 %s에 ip_address 컬럼이 추가되었습니다", tableName)
			}
		}
	}

	log.Println("모든 테이블에 IP 주소 컬럼 추가 마이그레이션이 완료되었습니다.")
}
