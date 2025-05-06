package main

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gosimple/slug"
)

// Migration 마이그레이션 관리 클래스
type Migration struct {
	sourceDB    *sql.DB
	targetDB    *sql.DB
	config      *Config
	xeParser    *XEParser
	fileHandler *FileHandler
}

// NewMigration Migration 생성자
func NewMigration(sourceDB, targetDB *sql.DB, config *Config) *Migration {
	return &Migration{
		sourceDB:    sourceDB,
		targetDB:    targetDB,
		config:      config,
		xeParser:    NewXEParser(sourceDB, config),
		fileHandler: NewFileHandler(config),
	}
}

// Run 마이그레이션 실행
func (m *Migration) Run() error {
	// 1. 게시판 마이그레이션
	if err := m.migrateBoards(); err != nil {
		return fmt.Errorf("게시판 마이그레이션 실패: %w", err)
	}

	// 스키마만 마이그레이션하는 경우 여기서 종료
	if m.config.SchemaOnly {
		return nil
	}

	return nil
}

// migrateBoards 게시판 마이그레이션
func (m *Migration) migrateBoards() error {
	// XE 게시판 목록 조회
	boards, err := m.xeParser.GetBoards()
	if err != nil {
		return err
	}

	fmt.Printf("총 %d개의 게시판을 발견했습니다.\n", len(boards))

	// 각 게시판 처리
	for i, board := range boards {
		fmt.Printf("[%d/%d] 게시판 마이그레이션 중: %s (모듈ID: %d)\n",
			i+1, len(boards), board.BrowserTitle, board.ModuleSrl)

		if err := m.migrateBoard(board); err != nil {
			fmt.Printf("  오류: %v\n", err)
			// 오류 발생해도 다음 게시판 계속 진행
			continue
		}

		fmt.Printf("  성공: %s 게시판이 마이그레이션되었습니다.\n", board.BrowserTitle)
	}

	return nil
}

// migrateBoard 단일 게시판 마이그레이션
func (m *Migration) migrateBoard(xeBoard XEBoard) error {
	// 1. 게시판 정보 저장
	boardID, err := m.saveBoard(xeBoard)
	if err != nil {
		return err
	}

	// 데이터만 마이그레이션하는 경우 스키마 생성 건너뛰기
	if !m.config.DataOnly {
		// 2. 게시판 테이블 생성
		if err := m.createBoardTable(boardID, xeBoard); err != nil {
			return err
		}
	}

	// 3. 게시물 마이그레이션
	if err := m.migratePosts(boardID, xeBoard); err != nil {
		return err
	}

	return nil
}

// saveBoard 게시판 정보 저장
func (m *Migration) saveBoard(xeBoard XEBoard) (int64, error) {
	// 슬러그 생성
	boardSlug := slug.Make(xeBoard.Mid)

	// 테이블명 생성
	tableName := fmt.Sprintf("board_%s", boardSlug)

	// 게시판 타입 결정 (모든 사용자 접근 가능 = 일반 게시판, 그 외 = 소모임 게시판)
	boardType := "normal"
	if !xeBoard.IsPublic {
		boardType = "group"
	}

	// 기존 게시판 확인
	var existingID int64
	err := m.targetDB.QueryRow(
		"SELECT id FROM boards WHERE slug = ?",
		boardSlug,
	).Scan(&existingID)

	if err == nil {
		// 이미 존재하는 게시판
		if m.config.Verbose {
			fmt.Printf("  게시판 '%s'(ID: %d)는 이미 존재합니다. 기존 ID를 사용합니다.\n",
				xeBoard.BrowserTitle, existingID)
		}
		return existingID, nil
	} else if err != sql.ErrNoRows {
		// 다른 오류 발생
		return 0, fmt.Errorf("기존 게시판 확인 실패: %w", err)
	}

	// 게시판 저장
	query := `
		INSERT INTO boards (
			name, slug, description, board_type, table_name, 
			active, comments_enabled, votes_enabled, allow_anonymous, allow_private, 
			sort_order, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	res, err := m.targetDB.Exec(
		query,
		xeBoard.BrowserTitle,
		boardSlug,
		xeBoard.Description,
		boardType,
		tableName,
		true,              // active
		true,              // comments_enabled
		true,              // votes_enabled
		xeBoard.IsPublic,  // allow_anonymous
		false,             // allow_private
		xeBoard.ModuleSrl, // sort_order
		xeBoard.CreatedAt, // created_at
		now,               // updated_at
	)
	if err != nil {
		return 0, fmt.Errorf("게시판 저장 실패: %w", err)
	}

	boardID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("게시판 ID 조회 실패: %w", err)
	}

	// 게시판 필드 추가 (기본 필드만 추가)
	if err := m.addBoardFields(boardID); err != nil {
		return 0, fmt.Errorf("게시판 필드 추가 실패: %w", err)
	}

	// 소모임 게시판인 경우 기본 모더레이터 추가
	if boardType == "group" {
		if err := m.addBoardModerator(boardID); err != nil {
			// 오류가 발생해도 계속 진행
			fmt.Printf("  경고: 소모임 모더레이터 추가 실패: %v\n", err)
		}
	}

	return boardID, nil
}

// addBoardFields 게시판 필드 추가
func (m *Migration) addBoardFields(boardID int64) error {
	// 게시판 타입 조회
	var boardType string
	err := m.targetDB.QueryRow("SELECT board_type FROM boards WHERE id = ?", boardID).Scan(&boardType)
	if err != nil {
		return fmt.Errorf("게시판 타입 조회 실패: %w", err)
	}

	// 게시판 타입에 따른 필드 설정
	var fields []struct {
		Name        string
		ColumnName  string
		DisplayName string
		FieldType   string
		Required    bool
		Sortable    bool
		Searchable  bool
		SortOrder   int
	}

	// 기본 필드 (태그)
	fields = append(fields, struct {
		Name        string
		ColumnName  string
		DisplayName string
		FieldType   string
		Required    bool
		Sortable    bool
		Searchable  bool
		SortOrder   int
	}{"tags", "tags", "태그", "text", false, false, true, 1})

	// 파일 첨부 필드 추가 (게시판 타입에 따라 다르게 설정)
	if boardType == "gallery" {
		// 갤러리 게시판은 이미지 필수
		fields = append(fields, struct {
			Name        string
			ColumnName  string
			DisplayName string
			FieldType   string
			Required    bool
			Sortable    bool
			Searchable  bool
			SortOrder   int
		}{"file", "file", "이미지", "file", true, false, false, 2})
	} else {
		// 일반 게시판은 파일 첨부 선택
		fields = append(fields, struct {
			Name        string
			ColumnName  string
			DisplayName string
			FieldType   string
			Required    bool
			Sortable    bool
			Searchable  bool
			SortOrder   int
		}{"file", "file", "첨부 파일", "file", false, false, false, 2})
	}

	// Q&A 게시판인 경우 상태 필드 추가
	if boardType == "qna" {
		fields = append(fields, struct {
			Name        string
			ColumnName  string
			DisplayName string
			FieldType   string
			Required    bool
			Sortable    bool
			Searchable  bool
			SortOrder   int
		}{"status", "status", "상태", "select", true, true, true, 3})
	}

	// 필드 추가
	for _, field := range fields {
		// 기존 필드 확인
		var count int
		err := m.targetDB.QueryRow(
			"SELECT COUNT(*) FROM board_fields WHERE board_id = ? AND name = ?",
			boardID, field.Name,
		).Scan(&count)

		if err != nil {
			return fmt.Errorf("필드 확인 실패 (%s): %w", field.Name, err)
		}

		// 이미 존재하면 건너뛰기
		if count > 0 {
			continue
		}

		query := `
            INSERT INTO board_fields (
                board_id, name, column_name, display_name, field_type, 
                required, sortable, searchable, sort_order, created_at, updated_at
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `

		now := time.Now()
		_, err = m.targetDB.Exec(
			query,
			boardID,
			field.Name,
			field.ColumnName,
			field.DisplayName,
			field.FieldType,
			field.Required,
			field.Sortable,
			field.Searchable,
			field.SortOrder,
			now,
			now,
		)
		if err != nil {
			return fmt.Errorf("필드 추가 실패 (%s): %w", field.Name, err)
		}
	}

	return nil
}

// addBoardModerator 소모임 게시판 모더레이터 추가
func (m *Migration) addBoardModerator(boardID int64) error {
	// 이미 존재하는지 확인
	var count int
	err := m.targetDB.QueryRow(
		"SELECT COUNT(*) FROM board_participants WHERE board_id = ? AND user_id = ?",
		boardID, 1,
	).Scan(&count)

	if err != nil {
		return fmt.Errorf("참여자 확인 실패: %w", err)
	}

	// 이미 존재하면 건너뛰기
	if count > 0 {
		return nil
	}

	// 관리자(ID=1)를 기본 모더레이터로 추가
	query := `
		INSERT INTO board_participants (
			board_id, user_id, role, created_at
		) VALUES (?, ?, ?, ?)
	`

	_, err = m.targetDB.Exec(
		query,
		boardID,
		1,           // user_id (관리자)
		"moderator", // role
		time.Now(),  // created_at
	)

	return err
}

// createBoardTable 게시판 테이블 생성
func (m *Migration) createBoardTable(boardID int64, xeBoard XEBoard) error {
	// 게시판 정보 조회
	var tableName, boardType string
	query := "SELECT table_name, board_type FROM boards WHERE id = ?"
	err := m.targetDB.QueryRow(query, boardID).Scan(&tableName, &boardType)
	if err != nil {
		return fmt.Errorf("게시판 정보 조회 실패: %w", err)
	}

	// 이미 테이블이 존재하는지 확인
	var tableExists bool
	var checkQuery string

	switch {
	case strings.HasPrefix(m.config.TargetDriver, "mysql"):
		checkQuery = "SELECT 1 FROM information_schema.tables WHERE table_schema = ? AND table_name = ?"
		err = m.targetDB.QueryRow(checkQuery, m.config.TargetDB, tableName).Scan(&tableExists)
	case strings.HasPrefix(m.config.TargetDriver, "postgres"):
		checkQuery = "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = ?)"
		err = m.targetDB.QueryRow(checkQuery, tableName).Scan(&tableExists)
	case strings.HasPrefix(m.config.TargetDriver, "sqlite"):
		checkQuery = "SELECT 1 FROM sqlite_master WHERE type='table' AND name=?"
		err = m.targetDB.QueryRow(checkQuery, tableName).Scan(&tableExists)
	default:
		return fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", m.config.TargetDriver)
	}

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("테이블 존재 여부 확인 실패: %w", err)
	}

	// 테이블이 이미 존재하면 건너뛰기
	if tableExists {
		return nil
	}

	// 기본 컬럼 정의 - 데이터베이스 타입별로 다른 구문 사용
	var columns []string

	if strings.HasPrefix(m.config.TargetDriver, "sqlite") {
		// SQLite용 컬럼 정의
		columns = []string{
			"id INTEGER PRIMARY KEY AUTOINCREMENT",
			"title TEXT NOT NULL",
			"content TEXT NOT NULL",
			"user_id INTEGER NOT NULL",
			"view_count INTEGER NOT NULL DEFAULT 0",
			"comment_count INTEGER NOT NULL DEFAULT 0",
			"like_count INTEGER NOT NULL DEFAULT 0",
			"dislike_count INTEGER NOT NULL DEFAULT 0",
			"is_private INTEGER NOT NULL DEFAULT 0",
			"created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP",
			"updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP",
			"tags TEXT",
			"file TEXT", // 파일 첨부 필드 추가
		}
	} else {
		var contentType string

		switch {
		case strings.HasPrefix(m.config.TargetDriver, "mysql"):
			contentType = "MEDIUMTEXT"
		case strings.HasPrefix(m.config.TargetDriver, "postgres"):
			contentType = "TEXT"
		default:
			return fmt.Errorf("지원하지 않는 데이터베이스 드라이버: %s", m.config.TargetDriver)
		}

		// MySQL/PostgreSQL용 컬럼 정의
		columns = []string{
			"id INT AUTO_INCREMENT PRIMARY KEY",
			"title VARCHAR(200) NOT NULL",
			// "content TEXT NOT NULL",
			fmt.Sprintf("content %s NOT NULL", contentType),
			"user_id INT NOT NULL",
			"view_count INT NOT NULL DEFAULT 0",
			"comment_count INT NOT NULL DEFAULT 0",
			"like_count INT NOT NULL DEFAULT 0",
			"dislike_count INT NOT NULL DEFAULT 0",
			"is_private BOOLEAN NOT NULL DEFAULT 0",
			"created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP",
			"updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP",
			"tags VARCHAR(255)",
			"file VARCHAR(255)", // 파일 첨부 필드 추가
		}
	}

	// 게시판 타입에 따라 추가 필드 정의
	if boardType == "qna" {
		columns = append(columns, "status VARCHAR(20) DEFAULT 'open'")
		columns = append(columns, "best_answer_id INT DEFAULT NULL")
	}

	// 테이블 생성 쿼리
	createQuery := fmt.Sprintf("CREATE TABLE %s (%s)", tableName, strings.Join(columns, ", "))

	// 테이블 생성
	_, err = m.targetDB.Exec(createQuery)
	if err != nil {
		return fmt.Errorf("테이블 생성 실패: %w", err)
	}

	return nil
}

// migratePosts 게시물 마이그레이션
func (m *Migration) migratePosts(boardID int64, xeBoard XEBoard) error {
	// 게시판 정보 조회
	var tableName string
	query := "SELECT table_name FROM boards WHERE id = ?"
	err := m.targetDB.QueryRow(query, boardID).Scan(&tableName)
	if err != nil {
		return fmt.Errorf("게시판 정보 조회 실패: %w", err)
	}

	// 총 게시물 수 조회
	totalCount, err := m.xeParser.GetTotalDocumentCount(xeBoard.ModuleSrl)
	if err != nil {
		return fmt.Errorf("게시물 개수 조회 실패: %w", err)
	}

	fmt.Printf("  게시판 '%s'에서 총 %d개의 게시물을 발견했습니다.\n", xeBoard.BrowserTitle, totalCount)

	// 배치 단위로 게시물 마이그레이션
	batchSize := m.config.BatchSize
	for offset := 0; offset < totalCount; offset += batchSize {
		// 현재 배치 크기 계산
		currentBatchSize := batchSize
		if offset+currentBatchSize > totalCount {
			currentBatchSize = totalCount - offset
		}

		fmt.Printf("  게시물 마이그레이션 중: %d-%d / %d\n", offset+1, offset+currentBatchSize, totalCount)

		// XE 게시물 조회
		documents, err := m.xeParser.GetDocuments(xeBoard.ModuleSrl, offset, currentBatchSize)
		if err != nil {
			return fmt.Errorf("게시물 조회 실패: %w", err)
		}

		// 각 게시물 마이그레이션
		for _, doc := range documents {
			if err := m.migratePost(boardID, tableName, doc); err != nil {
				fmt.Printf("    게시물 마이그레이션 실패 (ID: %d): %v\n", doc.DocumentSrl, err)
				// 오류 발생해도 다음 게시물 계속 진행
				continue
			}
		}
	}

	return nil
}

// migratePost 단일 게시물 마이그레이션
func (m *Migration) migratePost(boardID int64, tableName string, xeDoc XEDocument) error {
	// 트랜잭션 시작
	tx, err := m.targetDB.Begin()
	if err != nil {
		return fmt.Errorf("트랜잭션 시작 실패: %w", err)
	}
	defer tx.Rollback()

	// 이미 존재하는 게시물인지 확인
	var existingID int64
	err = tx.QueryRow(
		fmt.Sprintf("SELECT id FROM %s WHERE id = ?", tableName),
		xeDoc.DocumentSrl,
	).Scan(&existingID)

	if err == nil {
		// 이미 존재하는 게시물
		if m.config.Verbose {
			fmt.Printf("    게시물 ID %d는 이미 존재합니다. 건너뜁니다.\n", xeDoc.DocumentSrl)
		}
		return nil
	} else if err != sql.ErrNoRows {
		// 다른 오류 발생
		return fmt.Errorf("기존 게시물 확인 실패: %w", err)
	}

	// 1. 게시물 저장
	postQuery := fmt.Sprintf(`
		INSERT INTO %s (
			id, title, content, user_id, view_count, comment_count, 
			like_count, dislike_count, is_private, created_at, updated_at, tags
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tableName)

	// 사용자 ID는 1로 고정 (관리자)
	userID := 1

	_, err = tx.Exec(
		postQuery,
		xeDoc.DocumentSrl, // id (XE 문서 번호 유지)
		xeDoc.Title,
		xeDoc.Content,
		userID,
		xeDoc.ViewCount,
		xeDoc.CommentCount,
		xeDoc.Voted,
		xeDoc.Blamed,
		xeDoc.IsSecret,
		xeDoc.RegDate,
		xeDoc.UpdatedAt,
		xeDoc.Tags,
	)
	if err != nil {
		return fmt.Errorf("게시물 저장 실패: %w", err)
	}

	// 게시물 ID 사용
	postID := xeDoc.DocumentSrl

	// 2. 댓글 마이그레이션
	if xeDoc.CommentCount > 0 {
		if err := m.migrateComments(tx, boardID, postID, xeDoc); err != nil {
			return fmt.Errorf("댓글 마이그레이션 실패: %w", err)
		}
	}

	// 3. 첨부파일 마이그레이션
	if m.config.MigrateFiles {
		if err := m.migrateFiles(tx, boardID, postID, xeDoc); err != nil {
			return fmt.Errorf("첨부파일 마이그레이션 실패: %w", err)
		}
	}

	// 트랜잭션 커밋
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("트랜잭션 커밋 실패: %w", err)
	}

	return nil
}

// migrateComments 게시물 댓글 마이그레이션
func (m *Migration) migrateComments(tx *sql.Tx, boardID, postID int64, xeDoc XEDocument) error {
	// // XE 댓글 조회
	// comments, err := m.xeParser.GetComments(xeDoc.DocumentSrl)
	// if err != nil {
	// 	return fmt.Errorf("댓글 조회 실패: %w", err)
	// }

	fmt.Printf("    댓글 마이그레이션 시작 (게시물 ID: %d, 댓글 수: %d)\n",
		xeDoc.DocumentSrl, xeDoc.CommentCount)

	// XE 댓글 조회
	comments, err := m.xeParser.GetComments(xeDoc.DocumentSrl)
	if err != nil {
		fmt.Printf("    댓글 조회 실패: %v\n", err)
		return fmt.Errorf("댓글 조회 실패: %w", err)
	}

	fmt.Printf("    총 %d개의 댓글을 찾았습니다.\n", len(comments))

	if len(comments) == 0 {
		return nil
	}

	// 댓글 ID 매핑 테이블 (XE 댓글 ID -> Go-Board 댓글 ID)
	commentIDMap := make(map[int64]int64)

	// 먼저 최상위 댓글부터 처리
	for _, comment := range comments {
		// 부모 댓글이 아닌 경우 건너뛰기
		if comment.ParentSrl != 0 {
			continue
		}

		// 댓글 저장
		commentID, err := m.saveComment(tx, boardID, postID, comment, nil)
		if err != nil {
			return fmt.Errorf("최상위 댓글 저장 실패 (ID: %d): %w", comment.CommentSrl, err)
		}

		// 댓글 ID 매핑 저장
		commentIDMap[comment.CommentSrl] = commentID
	}

	// 그 다음 하위 댓글 처리
	for _, comment := range comments {
		// 최상위 댓글인 경우 건너뛰기
		if comment.ParentSrl == 0 {
			continue
		}

		// 부모 댓글 ID 찾기
		parentID, exists := commentIDMap[comment.ParentSrl]
		if !exists {
			// 부모 댓글을 찾을 수 없는 경우, 일반 댓글로 처리
			parentID = 0
		}

		// 답글 저장
		var parentIDPtr *int64
		if parentID > 0 {
			parentIDPtr = &parentID
		}

		commentID, err := m.saveComment(tx, boardID, postID, comment, parentIDPtr)
		if err != nil {
			return fmt.Errorf("답글 저장 실패 (ID: %d): %w", comment.CommentSrl, err)
		}

		// 댓글 ID 매핑 저장
		commentIDMap[comment.CommentSrl] = commentID
	}

	return nil
}

// saveComment 댓글 저장
func (m *Migration) saveComment(tx *sql.Tx, boardID, postID int64, xeComment XEComment, parentID *int64) (int64, error) {
	// 이미 존재하는 댓글인지 확인
	var existingID int64
	err := tx.QueryRow(
		"SELECT id FROM comments WHERE id = ?",
		xeComment.CommentSrl,
	).Scan(&existingID)

	if err == nil {
		// 이미 존재하는 댓글
		if m.config.Verbose {
			fmt.Printf("    댓글 ID %d는 이미 존재합니다. 기존 ID를 사용합니다.\n", xeComment.CommentSrl)
		}
		return existingID, nil
	} else if err != sql.ErrNoRows {
		// 다른 오류 발생
		return 0, fmt.Errorf("기존 댓글 확인 실패: %w", err)
	}

	// 댓글 저장
	query := `
		INSERT INTO comments (
			id, post_id, board_id, user_id, content, parent_id, 
			like_count, dislike_count, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// 사용자 ID는 1로 고정 (관리자)
	userID := 1

	_, err = tx.Exec(
		query,
		xeComment.CommentSrl, // id (XE 댓글 번호 유지)
		postID,
		boardID,
		userID,
		xeComment.Content,
		parentID,
		xeComment.Voted,
		xeComment.Blamed,
		xeComment.RegDate,
		xeComment.RegDate,
	)
	if err != nil {
		return 0, fmt.Errorf("댓글 저장 실패: %w", err)
	}

	return xeComment.CommentSrl, nil
}

// migrateFiles 첨부파일 마이그레이션
func (m *Migration) migrateFiles(tx *sql.Tx, boardID, postID int64, xeDoc XEDocument) error {
	// XE 첨부파일 조회
	files, err := m.xeParser.GetFiles(xeDoc.DocumentSrl)
	if err != nil {
		return fmt.Errorf("첨부파일 조회 실패: %w", err)
	}

	if len(files) == 0 {
		return nil
	}

	// 최대 첨부파일 ID 조회 (ID 충돌 방지)
	var maxAttachmentID int64
	err = tx.QueryRow("SELECT COALESCE(MAX(id), 0) FROM attachments").Scan(&maxAttachmentID)
	if err != nil {
		fmt.Printf("  경고: 최대 첨부파일 ID 조회 실패: %v\n", err)
		// 조회 실패 시 기본값 설정
		maxAttachmentID = 100000 // 안전한 큰 값으로 시작
	}

	// 게시물 ID보다 충분히 큰 값으로 설정
	if maxAttachmentID <= postID {
		maxAttachmentID = postID + 1000 // 충분한 간격 설정
	}

	for _, file := range files {
		// 이미 존재하는 첨부파일인지 확인
		var existingID int64
		err := tx.QueryRow(
			"SELECT id FROM attachments WHERE id = ?",
			file.FileSrl,
		).Scan(&existingID)

		if err == nil {
			// 이미 존재하는 첨부파일
			if m.config.Verbose {
				fmt.Printf("    첨부파일 ID %d는 이미 존재합니다. 건너뜁니다.\n", file.FileSrl)
			}
			continue
		} else if err != sql.ErrNoRows {
			// 다른 오류 발생
			return fmt.Errorf("기존 첨부파일 확인 실패: %w", err)
		}

		// 파일 처리
		urlPath, err := m.fileHandler.ProcessFile(file, boardID, postID)
		if err != nil {
			fmt.Printf("    첨부파일 처리 실패 (%s): %v\n", file.OriginName, err)
			// 오류 발생해도 다음 파일 계속 진행
			continue
		}

		// 고유한 첨부파일 ID 생성
		maxAttachmentID++
		attachmentID := maxAttachmentID

		// 파일 메타데이터 저장 - 새로운 ID 사용
		if err := m.saveAttachment(tx, boardID, postID, file, urlPath, attachmentID); err != nil {
			return fmt.Errorf("첨부파일 메타데이터 저장 실패: %w", err)
		}
	}

	return nil
}

// saveAttachment 첨부파일 메타데이터 저장
func (m *Migration) saveAttachment(tx *sql.Tx, boardID, postID int64, xeFile XEFile, urlPath string, attachmentID int64) error {
	// 파일 경로 및 이름 처리
	fileName := filepath.Base(urlPath)

	// 전체 경로(uploads부터 파일명까지 포함)
	filePath := filepath.Join("uploads", "boards", strconv.FormatInt(boardID, 10), "posts", strconv.FormatInt(postID, 10), "attachments", fileName)

	// 윈도우 경로를 URL 경로로 변환
	filePath = strings.ReplaceAll(filePath, "\\", "/")

	// MIME 타입 추론
	mimeType := getMimeType(fileName)

	// 썸네일 URL 생성 (이미지인 경우)
	var thumbnailURL string
	if xeFile.IsImage {
		thumbnailURL = filepath.Join("/uploads", "boards", strconv.FormatInt(boardID, 10), "posts", strconv.FormatInt(postID, 10), "attachments", "thumbs", fileName)
		thumbnailURL = strings.ReplaceAll(thumbnailURL, "\\", "/")
	}

	// 다운로드 URL 생성
	downloadURL := filepath.Join("/uploads", "boards", strconv.FormatInt(boardID, 10), "posts", strconv.FormatInt(postID, 10), "attachments", fileName)
	downloadURL = strings.ReplaceAll(downloadURL, "\\", "/")

	// 첨부파일 저장
	query := `
        INSERT INTO attachments (
            id, board_id, post_id, user_id, file_name, storage_name, 
            file_path, file_size, mime_type, is_image, download_url, 
            thumbnail_url, download_count, created_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	// 사용자 ID는 1로 고정 (관리자)
	userID := 1

	_, err := tx.Exec(
		query,
		attachmentID, // 새로 생성된 ID 사용
		boardID,
		postID,
		userID,
		xeFile.OriginName,
		fileName,
		filePath,
		xeFile.FileSize,
		mimeType,
		xeFile.IsImage,
		downloadURL,
		thumbnailURL,
		xeFile.DownloadCount,
		xeFile.RegDate,
	)

	return err
}

// getMimeType 파일 확장자에 따른 MIME 타입 추론
func getMimeType(fileName string) string {
	ext := getLowerExtension(fileName)

	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".txt":  "text/plain",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".zip":  "application/zip",
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}

	return "application/octet-stream"
}

// getLowerExtension은 이미 xe_parser.go에 정의되어 있음
