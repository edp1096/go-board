// cmd/add_likecount/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/edp1096/go-board/config"
	"github.com/edp1096/go-board/internal/utils"
)

func main() {
	// 명령줄 옵션 파싱
	forceUpdate := flag.Bool("force", false, "컬럼이 이미 존재하더라도 좋아요/싫어요 수를 강제로 업데이트")
	flag.Parse()

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

		// QnA 게시판은 이미 처리되므로 제외
		if boardType == "qna" {
			log.Printf("QnA 게시판 %s는 이미 좋아요/싫어요 필드가 있으므로 건너뜁니다", tableName)
			continue
		}

		tableNames = append(tableNames, tableName)
		boardIDs = append(boardIDs, boardID)
		boardTypes = append(boardTypes, boardType)
	}

	// 각 게시판 테이블에 like_count와 dislike_count 컬럼 추가
	for i, tableName := range tableNames {
		boardID := boardIDs[i]
		boardType := boardTypes[i]

		log.Printf("처리 중: %s (ID: %d, Type: %s)", tableName, boardID, boardType)

		likeColumnExists := false
		dislikeColumnExists := false

		// 1. like_count 컬럼 추가
		var likeAlterSQL string
		if utils.IsPostgres(db) {
			likeAlterSQL = fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN IF NOT EXISTS like_count INTEGER NOT NULL DEFAULT 0", tableName)
		} else if utils.IsMySQL(db) {
			// MySQL에서는 IF NOT EXISTS가 지원되지 않으므로 컬럼 존재 여부 확인
			checkSQL := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = '%s' AND column_name = 'like_count'", tableName)
			var count int
			if err := db.QueryRowContext(context.Background(), checkSQL).Scan(&count); err != nil {
				log.Printf("like_count 컬럼 존재 여부 확인 실패: %v", err)
				continue
			}

			if count == 0 {
				likeAlterSQL = fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `like_count` INT NOT NULL DEFAULT 0", tableName)
			} else {
				likeColumnExists = true
				log.Printf("테이블 %s에 이미 like_count 컬럼이 존재합니다", tableName)
			}
		} else if utils.IsSQLite(db) {
			// SQLite는 컬럼 존재 여부 확인이 어려우므로 바로 추가
			likeAlterSQL = fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN like_count INTEGER NOT NULL DEFAULT 0", tableName)
		}

		// 컬럼이 없거나 SQLite인 경우에만 ALTER TABLE 실행
		if !likeColumnExists && likeAlterSQL != "" {
			_, err := db.ExecContext(context.Background(), likeAlterSQL)
			if err != nil {
				// SQLite에서 이미 컬럼이 존재하는 경우 duplicate column name 오류 발생
				if utils.IsSQLite(db) && strings.Contains(err.Error(), "duplicate column name") {
					likeColumnExists = true
					log.Printf("테이블 %s에 이미 like_count 컬럼이 존재합니다", tableName)
				} else {
					log.Printf("테이블 %s에 like_count 컬럼 추가 실패: %v", tableName, err)
					continue
				}
			} else {
				log.Printf("테이블 %s에 like_count 컬럼이 추가되었습니다", tableName)
			}
		}

		// 2. dislike_count 컬럼 추가
		var dislikeAlterSQL string
		if utils.IsPostgres(db) {
			dislikeAlterSQL = fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN IF NOT EXISTS dislike_count INTEGER NOT NULL DEFAULT 0", tableName)
		} else if utils.IsMySQL(db) {
			// MySQL에서는 IF NOT EXISTS가 지원되지 않으므로 컬럼 존재 여부 확인
			checkSQL := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = '%s' AND column_name = 'dislike_count'", tableName)
			var count int
			if err := db.QueryRowContext(context.Background(), checkSQL).Scan(&count); err != nil {
				log.Printf("dislike_count 컬럼 존재 여부 확인 실패: %v", err)
				continue
			}

			if count == 0 {
				dislikeAlterSQL = fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `dislike_count` INT NOT NULL DEFAULT 0", tableName)
			} else {
				dislikeColumnExists = true
				log.Printf("테이블 %s에 이미 dislike_count 컬럼이 존재합니다", tableName)
			}
		} else if utils.IsSQLite(db) {
			// SQLite는 컬럼 존재 여부 확인이 어려우므로 바로 추가
			dislikeAlterSQL = fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN dislike_count INTEGER NOT NULL DEFAULT 0", tableName)
		}

		// 컬럼이 없거나 SQLite인 경우에만 ALTER TABLE 실행
		if !dislikeColumnExists && dislikeAlterSQL != "" {
			_, err := db.ExecContext(context.Background(), dislikeAlterSQL)
			if err != nil {
				// SQLite에서 이미 컬럼이 존재하는 경우 duplicate column name 오류 발생
				if utils.IsSQLite(db) && strings.Contains(err.Error(), "duplicate column name") {
					dislikeColumnExists = true
					log.Printf("테이블 %s에 이미 dislike_count 컬럼이 존재합니다", tableName)
				} else {
					log.Printf("테이블 %s에 dislike_count 컬럼 추가 실패: %v", tableName, err)
					continue
				}
			} else {
				log.Printf("테이블 %s에 dislike_count 컬럼이 추가되었습니다", tableName)
			}
		}

		// 좋아요/싫어요 수 업데이트는 force 옵션이 있거나 새로 컬럼을 추가한 경우에만 실행
		if !likeColumnExists || !dislikeColumnExists || *forceUpdate {
			log.Printf("테이블 %s의 좋아요/싫어요 수를 업데이트합니다", tableName)

			// 각 게시물의 좋아요/싫어요 수 계산 및 업데이트
			// 1. 게시물 ID 목록 조회
			var postQuery string
			if utils.IsPostgres(db) {
				postQuery = fmt.Sprintf("SELECT id FROM \"%s\"", tableName)
			} else {
				postQuery = fmt.Sprintf("SELECT id FROM `%s`", tableName)
			}

			postRows, err := db.QueryContext(context.Background(), postQuery)
			if err != nil {
				log.Printf("테이블 %s의 게시물 목록 조회 실패: %v", tableName, err)
				continue
			}

			var postIDs []int64
			for postRows.Next() {
				var postID int64
				if err := postRows.Scan(&postID); err != nil {
					log.Printf("게시물 ID 스캔 실패: %v", err)
					continue
				}
				postIDs = append(postIDs, postID)
			}
			postRows.Close()

			// 2. 각 게시물의 좋아요/싫어요 수 업데이트
			for _, postID := range postIDs {
				// 좋아요 수 조회
				likeCountQuery := "SELECT COUNT(*) FROM post_votes WHERE board_id = ? AND post_id = ? AND value = 1"
				var likeCount int
				err := db.QueryRowContext(context.Background(), likeCountQuery, boardID, postID).Scan(&likeCount)
				if err != nil {
					log.Printf("게시물 ID %d의 좋아요 수 조회 실패: %v", postID, err)
					continue
				}

				// 싫어요 수 조회
				dislikeCountQuery := "SELECT COUNT(*) FROM post_votes WHERE board_id = ? AND post_id = ? AND value = -1"
				var dislikeCount int
				err = db.QueryRowContext(context.Background(), dislikeCountQuery, boardID, postID).Scan(&dislikeCount)
				if err != nil {
					log.Printf("게시물 ID %d의 싫어요 수 조회 실패: %v", postID, err)
					continue
				}

				// 좋아요/싫어요 수 업데이트
				var updateQuery string
				if utils.IsPostgres(db) {
					updateQuery = fmt.Sprintf("UPDATE \"%s\" SET like_count = ?, dislike_count = ? WHERE id = ?", tableName)
				} else {
					updateQuery = fmt.Sprintf("UPDATE `%s` SET like_count = ?, dislike_count = ? WHERE id = ?", tableName)
				}

				_, err = db.ExecContext(context.Background(), updateQuery, likeCount, dislikeCount, postID)
				if err != nil {
					log.Printf("게시물 ID %d의 좋아요/싫어요 수 업데이트 실패: %v", postID, err)
					continue
				}

				log.Printf("게시물 ID %d의 좋아요 수(%d)/싫어요 수(%d)가 업데이트되었습니다", postID, likeCount, dislikeCount)
			}
		} else {
			log.Printf("테이블 %s의 좋아요/싫어요 수 업데이트를 건너뜁니다 (force 옵션을 사용하면 강제 업데이트 가능)", tableName)
		}
	}

	log.Println("모든 게시판 테이블에 좋아요/싫어요 수 추가 마이그레이션이 완료되었습니다.")
}
