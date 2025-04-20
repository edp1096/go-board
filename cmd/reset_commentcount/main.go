// cmd/migrate_comments/main.go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/edp1096/go-board/config"
	"github.com/edp1096/go-board/internal/repository"
	"github.com/edp1096/go-board/internal/utils"
)

func main() {
	// 환경 설정 로드
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

	// 게시판 저장소 생성
	boardRepo := repository.NewBoardRepository(db)

	// 게시판 목록 조회
	boards, err := boardRepo.List(context.Background(), false)
	if err != nil {
		log.Fatalf("게시판 목록 조회 실패: %v", err)
	}

	// 각 게시판별로 댓글 수 업데이트
	for _, board := range boards {
		fmt.Printf("게시판 '%s' 처리 중...\n", board.Name)

		// 게시물 목록 조회 쿼리
		var postsQuery string
		if utils.IsPostgres(db) {
			postsQuery = fmt.Sprintf("SELECT id FROM \"%s\"", board.TableName)
		} else {
			postsQuery = fmt.Sprintf("SELECT id FROM `%s`", board.TableName)
		}

		var postIDs []int64
		rows, err := db.Query(postsQuery)
		if err != nil {
			log.Printf("게시물 목록 조회 실패 (게시판 '%s'): %v\n", board.Name, err)
			continue
		}

		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				log.Printf("게시물 ID 스캔 실패: %v\n", err)
				continue
			}
			postIDs = append(postIDs, id)
		}
		rows.Close()

		// 각 게시물별로 댓글 수 업데이트
		for _, postID := range postIDs {
			var count int
			countQuery := "SELECT COUNT(*) FROM comments WHERE board_id = ? AND post_id = ?"
			err := db.QueryRow(countQuery, board.ID, postID).Scan(&count)
			if err != nil {
				log.Printf("댓글 수 조회 실패 (게시판 '%s', 게시물 %d): %v\n", board.Name, postID, err)
				continue
			}

			var updateQuery string
			if utils.IsPostgres(db) {
				updateQuery = fmt.Sprintf("UPDATE \"%s\" SET comment_count = ? WHERE id = ?", board.TableName)
			} else {
				updateQuery = fmt.Sprintf("UPDATE `%s` SET comment_count = ? WHERE id = ?", board.TableName)
			}

			_, err = db.Exec(updateQuery, count, postID)
			if err != nil {
				log.Printf("댓글 수 업데이트 실패 (게시판 '%s', 게시물 %d): %v\n", board.Name, postID, err)
				continue
			}

			fmt.Printf("게시물 ID %d의 댓글 수를 %d로 업데이트했습니다.\n", postID, count)
		}
	}

	fmt.Println("모든 게시물의 댓글 수 업데이트가 완료되었습니다.")
}
