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

	// board_participants 테이블 생성
	var createSQL string

	if utils.IsPostgres(db) {
		// PostgreSQL - 테이블 존재 여부 확인 및 생성
		createSQL = `
			CREATE TABLE IF NOT EXISTS board_participants (
				board_id INT NOT NULL,
				user_id INT NOT NULL,
				role VARCHAR(20) NOT NULL DEFAULT 'member',
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				
				PRIMARY KEY (board_id, user_id),
				FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			)
		`
	} else if utils.IsMySQL(db) {
		// MySQL - 테이블 존재 여부 확인
		checkSQL := "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'board_participants'"
		var count int
		if err := db.QueryRowContext(context.Background(), checkSQL).Scan(&count); err != nil {
			log.Fatalf("테이블 존재 여부 확인 실패: %v", err)
		}

		if count > 0 {
			log.Println("board_participants 테이블이 이미 존재합니다")
			return
		}

		createSQL = `
			CREATE TABLE board_participants (
				board_id INT NOT NULL,
				user_id INT NOT NULL,
				role VARCHAR(20) NOT NULL DEFAULT 'member',
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				
				PRIMARY KEY (board_id, user_id),
				FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			)
		`
	} else if utils.IsSQLite(db) {
		// SQLite - 테이블 존재 여부 확인
		checkSQL := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='board_participants'"
		var count int
		if err := db.QueryRowContext(context.Background(), checkSQL).Scan(&count); err != nil {
			log.Fatalf("테이블 존재 여부 확인 실패: %v", err)
		}

		if count > 0 {
			log.Println("board_participants 테이블이 이미 존재합니다")
			return
		}

		createSQL = `
			CREATE TABLE board_participants (
				board_id INTEGER NOT NULL,
				user_id INTEGER NOT NULL,
				role TEXT NOT NULL DEFAULT 'member',
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				
				PRIMARY KEY (board_id, user_id),
				FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			)
		`
	}

	// SQL 실행
	_, err = db.ExecContext(context.Background(), createSQL)
	if err != nil {
		// PostgreSQL의 경우 IF NOT EXISTS가 있어서 이미 존재해도 오류가 나지 않지만,
		// 다른 데이터베이스에서 오류가 발생할 수 있음
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate") {
			log.Println("board_participants 테이블이 이미 존재합니다")
		} else {
			log.Fatalf("board_participants 테이블 생성 실패: %v", err)
		}
	} else {
		log.Println("board_participants 테이블이 성공적으로 생성되었습니다")
	}

	// 인덱스 생성 (선택적)
	if utils.IsPostgres(db) {
		indexSQL := `
			CREATE INDEX IF NOT EXISTS idx_board_participants_user 
			ON board_participants(user_id)
		`
		_, err = db.ExecContext(context.Background(), indexSQL)
		if err != nil {
			log.Printf("인덱스 생성 실패: %v", err)
		} else {
			log.Println("인덱스가 성공적으로 생성되었습니다")
		}
	} else if utils.IsMySQL(db) {
		// MySQL에서는 인덱스 존재 여부 확인
		checkIndexSQL := "SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'board_participants' AND index_name = 'idx_board_participants_user'"
		var count int
		if err := db.QueryRowContext(context.Background(), checkIndexSQL).Scan(&count); err != nil {
			log.Printf("인덱스 존재 여부 확인 실패: %v", err)
		} else if count == 0 {
			indexSQL := `
				CREATE INDEX idx_board_participants_user 
				ON board_participants(user_id)
			`
			_, err = db.ExecContext(context.Background(), indexSQL)
			if err != nil {
				log.Printf("인덱스 생성 실패: %v", err)
			} else {
				log.Println("인덱스가 성공적으로 생성되었습니다")
			}
		}
	} else if utils.IsSQLite(db) {
		indexSQL := `
			CREATE INDEX IF NOT EXISTS idx_board_participants_user 
			ON board_participants(user_id)
		`
		_, err = db.ExecContext(context.Background(), indexSQL)
		if err != nil {
			log.Printf("인덱스 생성 실패: %v", err)
		} else {
			log.Println("인덱스가 성공적으로 생성되었습니다")
		}
	}

	// 'group' 타입의 게시판들 확인 (기존에 있다면)
	checkGroupTypeSQL := "SELECT COUNT(*) FROM boards WHERE board_type = 'group'"
	var groupCount int
	err = db.QueryRowContext(context.Background(), checkGroupTypeSQL).Scan(&groupCount)
	if err != nil {
		log.Printf("group 타입 게시판 확인 실패: %v", err)
	} else {
		log.Printf("현재 group 타입의 게시판 수: %d", groupCount)
	}

	log.Println("소모임 게시판을 위한 마이그레이션이 완료되었습니다.")
}
