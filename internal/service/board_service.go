// internal/service/board_service.go
package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-board/internal/models"
	"go-board/internal/repository"
	"go-board/internal/utils"
	"path/filepath"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/uptrace/bun"
)

var (
	ErrBoardNotFound = errors.New("게시판을 찾을 수 없음")
	ErrInvalidBoard  = errors.New("유효하지 않은 게시판")
	ErrPostNotFound  = errors.New("게시물을 찾을 수 없음")
)

type BoardService interface {
	CreateBoard(ctx context.Context, board *models.Board) error
	GetBoardByID(ctx context.Context, id int64) (*models.Board, error)
	GetBoardBySlug(ctx context.Context, slug string) (*models.Board, error)
	UpdateBoard(ctx context.Context, board *models.Board) error
	DeleteBoard(ctx context.Context, id int64) error
	ListBoards(ctx context.Context, onlyActive bool) ([]*models.Board, error)

	// 매니저 관련 메서드
	GetBoardManagers(ctx context.Context, boardID int64) ([]*models.User, error)
	AddBoardManager(ctx context.Context, boardID, userID int64) error
	RemoveBoardManager(ctx context.Context, boardID, userID int64) error
	RemoveAllBoardManagers(ctx context.Context, boardID int64) error
	IsBoardManager(ctx context.Context, boardID, userID int64) (bool, error)

	// 게시판 필드 관련
	AddBoardField(ctx context.Context, boardID int64, field *models.BoardField) error
	UpdateBoardField(ctx context.Context, field *models.BoardField) error
	DeleteBoardField(ctx context.Context, id int64) error

	// 게시물 관련
	CreatePost(ctx context.Context, boardID int64, post *models.DynamicPost) error
	GetPost(ctx context.Context, boardID int64, postID int64) (*models.DynamicPost, error)
	UpdatePost(ctx context.Context, boardID int64, post *models.DynamicPost) error
	DeletePost(ctx context.Context, boardID int64, postID int64) error
	ListPosts(ctx context.Context, boardID int64, page, pageSize int, sortField, sortDir string) ([]*models.DynamicPost, int, error)
	SearchPosts(ctx context.Context, boardID int64, query string, page, pageSize int) ([]*models.DynamicPost, int, error)

	// 썸네일 관련
	GetPostThumbnails(ctx context.Context, boardID int64, postIDs []int64) (map[int64]string, error)

	// Q&A 관련
	SearchPostsWithStatus(ctx context.Context, boardID int64, query, status string, page, pageSize int) ([]*models.DynamicPost, int, error)
}

type boardService struct {
	boardRepo repository.BoardRepository
	db        *bun.DB
}

func NewBoardService(boardRepo repository.BoardRepository, db *bun.DB) BoardService {
	return &boardService{
		boardRepo: boardRepo,
		db:        db,
	}
}

func (s *boardService) CreateBoard(ctx context.Context, board *models.Board) error {
	// 슬러그가 없으면 생성
	if board.Slug == "" {
		board.Slug = slug.Make(board.Name)
	}

	// 테이블명이 없으면 생성
	if board.TableName == "" {
		// 테이블명 생성 (알파벳+숫자만 포함)
		// board.TableName = fmt.Sprintf("board_%s", slug.Make(board.Name))
		board.TableName = fmt.Sprintf("board_%s", board.Slug)
	}

	// 생성 시간 설정
	now := time.Now()
	board.CreatedAt = now
	board.UpdatedAt = now

	return s.boardRepo.Create(ctx, board)
}

func (s *boardService) GetBoardByID(ctx context.Context, id int64) (*models.Board, error) {
	return s.boardRepo.GetByID(ctx, id)
}

func (s *boardService) GetBoardBySlug(ctx context.Context, slug string) (*models.Board, error) {
	return s.boardRepo.GetBySlug(ctx, slug)
}

func (s *boardService) UpdateBoard(ctx context.Context, board *models.Board) error {
	board.UpdatedAt = time.Now()
	return s.boardRepo.Update(ctx, board)
}

func (s *boardService) DeleteBoard(ctx context.Context, id int64) error {
	return s.boardRepo.Delete(ctx, id)
}

func (s *boardService) ListBoards(ctx context.Context, onlyActive bool) ([]*models.Board, error) {
	return s.boardRepo.List(ctx, onlyActive)
}

// 매니저 관련
func (s *boardService) GetBoardManagers(ctx context.Context, boardID int64) ([]*models.User, error) {
	var managers []*models.User
	err := s.db.NewSelect().
		Model(&managers).
		Join("JOIN board_managers AS bm ON bm.user_id = u.id").
		Where("bm.board_id = ?", boardID).
		Scan(ctx)

	return managers, err
}

func (s *boardService) AddBoardManager(ctx context.Context, boardID, userID int64) error {
	manager := &models.BoardManager{
		BoardID:   boardID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	_, err := s.db.NewInsert().
		Model(manager).
		Exec(ctx)

	return err
}

func (s *boardService) RemoveBoardManager(ctx context.Context, boardID, userID int64) error {
	_, err := s.db.NewDelete().
		Model((*models.BoardManager)(nil)).
		Where("board_id = ? AND user_id = ?", boardID, userID).
		Exec(ctx)

	return err
}

// BoardService 인터페이스에 메서드 추가
func (s *boardService) RemoveAllBoardManagers(ctx context.Context, boardID int64) error {
	_, err := s.db.NewDelete().
		Model((*models.BoardManager)(nil)).
		Where("board_id = ?", boardID).
		Exec(ctx)

	return err
}

func (s *boardService) IsBoardManager(ctx context.Context, boardID, userID int64) (bool, error) {
	exists, err := s.db.NewSelect().
		Model((*models.BoardManager)(nil)).
		Where("board_id = ? AND user_id = ?", boardID, userID).
		Exists(ctx)

	return exists, err
}

// 게시판 필드 관련

func (s *boardService) AddBoardField(ctx context.Context, boardID int64, field *models.BoardField) error {
	// 필드 컬럼명이 없으면 생성
	if field.ColumnName == "" {
		field.ColumnName = slug.Make(field.Name)
	}

	field.BoardID = boardID
	field.CreatedAt = time.Now()
	field.UpdatedAt = time.Now()

	return s.boardRepo.CreateField(ctx, field)
}

func (s *boardService) UpdateBoardField(ctx context.Context, field *models.BoardField) error {
	field.UpdatedAt = time.Now()
	return s.boardRepo.UpdateField(ctx, field)
}

func (s *boardService) DeleteBoardField(ctx context.Context, id int64) error {
	return s.boardRepo.DeleteField(ctx, id)
}

// 게시물 관련

func (s *boardService) CreatePost(ctx context.Context, boardID int64, post *models.DynamicPost) error {
	// 게시판 정보 조회
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return ErrBoardNotFound
	}

	// 값 맵 생성
	values := map[string]any{
		"title":      post.Title,
		"content":    post.Content,
		"user_id":    post.UserID,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	// 동적 필드 추가
	for _, field := range post.Fields {
		values[field.ColumnName] = field.Value
	}

	var insertErr error

	if utils.IsPostgres(s.db) {
		// PostgreSQL에서는 RETURNING 구문 사용
		var id int64
		tableName := board.TableName

		insertErr = s.db.NewInsert().
			Model(&values).
			Table(tableName).
			Returning("id").
			Scan(ctx, &id)

		if insertErr == nil {
			post.ID = id
		}
	} else {
		// MySQL/MariaDB에서는 LastInsertId 사용
		var res sql.Result
		res, insertErr = s.db.NewInsert().
			Model(&values).
			Table(board.TableName).
			Exec(ctx)

		if insertErr == nil {
			// 생성된 ID 반환
			var id int64
			id, insertErr = res.LastInsertId()
			if insertErr == nil {
				post.ID = id
			}
		}
	}

	if insertErr != nil {
		return fmt.Errorf("게시물 생성 실패: %w", insertErr)
	}

	return nil
}

func (s *boardService) GetPost(ctx context.Context, boardID int64, postID int64) (*models.DynamicPost, error) {
	// 게시판 정보 조회
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, ErrBoardNotFound
	}

	// 게시판 필드 조회
	fields, err := s.boardRepo.GetFieldsByBoardID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	// 쿼리 빌더 초기화
	// 테이블명을 적절한 구분자로 감싸고 별칭은 AS 키워드로 구분
	var query *bun.SelectQuery
	if utils.IsPostgres(s.db) {
		query = s.db.NewSelect().
			TableExpr(fmt.Sprintf("\"%s\" AS p", board.TableName)).
			Column("p.*").
			ColumnExpr("u.username").
			Join("LEFT JOIN users AS u ON u.id = p.user_id").
			Where("p.id = ?", postID)
	} else {
		query = s.db.NewSelect().
			TableExpr(fmt.Sprintf("`%s` AS p", board.TableName)).
			Column("p.*").
			ColumnExpr("u.username").
			Join("LEFT JOIN users AS u ON u.id = p.user_id").
			Where("p.id = ?", postID)
	}

	// 쿼리 실행
	var row map[string]any
	err = query.Scan(ctx, &row)
	if err != nil {
		return nil, ErrPostNotFound
	}

	// 조회수 증가
	if utils.IsPostgres(s.db) {
		s.db.NewUpdate().
			Table(board.TableName).
			Set("view_count = view_count + 1").
			Where("id = ?", postID).
			Exec(ctx)
	} else {
		s.db.NewUpdate().
			Table(board.TableName).
			Set("view_count = view_count + 1").
			Where("id = ?", postID).
			Exec(ctx)
	}

	// ID 값 확인
	if row["id"] == nil {
		return nil, ErrPostNotFound
	}

	// 유틸리티 함수를 사용한 타입 변환
	viewCount := utils.InterfaceToInt(row["view_count"])

	// DynamicPost 객체 생성
	post := &models.DynamicPost{
		ID:        postID,
		Title:     utils.InterfaceToString(row["title"]),
		Content:   utils.InterfaceToString(row["content"]),
		UserID:    utils.InterfaceToInt64(row["user_id"]),
		Username:  utils.InterfaceToString(row["username"]),
		ViewCount: viewCount + 1, // 방금 증가한 조회수 반영
		CreatedAt: utils.InterfaceToTime(row["created_at"], time.Now()),
		UpdatedAt: utils.InterfaceToTime(row["updated_at"], time.Now()),
		Fields:    make(map[string]models.DynamicField),
		RawData:   row,
	}

	// 동적 필드 처리
	for _, field := range fields {
		if val, ok := row[field.ColumnName]; ok {
			post.Fields[field.Name] = models.DynamicField{
				Name:       field.Name,
				ColumnName: field.ColumnName,
				Value:      val,
				FieldType:  field.FieldType,
				Required:   field.Required,
			}
		}
	}

	return post, nil
}

func (s *boardService) UpdatePost(ctx context.Context, boardID int64, post *models.DynamicPost) error {
	// 게시판 정보 조회
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return ErrBoardNotFound
	}

	// 기본 필드 설정
	values := map[string]any{
		"title":      post.Title,
		"content":    post.Content,
		"updated_at": time.Now(),
	}

	// 동적 필드 추가
	for _, field := range post.Fields {
		values[field.ColumnName] = field.Value
	}

	// Model 메서드를 사용하여 업데이트
	tableName := board.TableName

	_, err = s.db.NewUpdate().
		Model(&values).
		Table(tableName).
		Where("id = ?", post.ID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("게시물 업데이트 실패: %w", err)
	}

	return nil
}

func (s *boardService) DeletePost(ctx context.Context, boardID int64, postID int64) error {
	// 게시판 정보 조회
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return ErrBoardNotFound
	}

	// 쿼리 실행
	tableName := board.TableName

	_, err = s.db.NewDelete().
		Table(tableName).
		Where("id = ?", postID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("게시물 삭제 실패: %w", err)
	}

	return nil
}

func (s *boardService) ListPosts(ctx context.Context, boardID int64, page, pageSize int, sortField, sortDir string) ([]*models.DynamicPost, int, error) {
	// 게시판 정보 조회
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, 0, ErrBoardNotFound
	}

	// 게시판 필드 조회
	fields, err := s.boardRepo.GetFieldsByBoardID(ctx, boardID)
	if err != nil {
		return nil, 0, err
	}

	// 정렬 필드 확인 및 기본값 설정
	if sortField == "" {
		sortField = "created_at"
	}
	if sortDir == "" {
		sortDir = "DESC"
	}

	// 페이지네이션 계산
	offset := (page - 1) * pageSize

	// 총 게시물 수 조회
	var countQuery *bun.SelectQuery
	tableName := board.TableName

	countQuery = s.db.NewSelect().
		Table(tableName).
		ColumnExpr("COUNT(*) AS count")

	var count int
	err = countQuery.Scan(ctx, &count)
	if err != nil {
		return nil, 0, err
	}

	// 게시물 목록 조회
	// 테이블명을 적절한 구분자로 감싸고 별칭은 AS 키워드로 구분
	var query *bun.SelectQuery
	var tableExpr string

	if utils.IsPostgres(s.db) {
		tableExpr = fmt.Sprintf("\"%s\" AS p", board.TableName)
	} else {
		tableExpr = fmt.Sprintf("`%s` AS p", board.TableName)
	}

	query = s.db.NewSelect().
		TableExpr(tableExpr).
		Column("p.*").
		ColumnExpr("u.username").
		Join("LEFT JOIN users AS u ON u.id = p.user_id").
		OrderExpr(fmt.Sprintf("p.%s %s", sortField, sortDir)).
		Limit(pageSize).
		Offset(offset)

	// 쿼리 실행
	var rows []map[string]any
	err = query.Scan(ctx, &rows)
	if err != nil {
		return nil, 0, err
	}

	// 결과 변환
	validPosts := make([]*models.DynamicPost, 0, len(rows))
	for _, row := range rows {
		// ID 값 확인
		if row["id"] == nil {
			continue
		}

		// 유틸리티 함수를 사용한 타입 변환
		postID := utils.InterfaceToInt64(row["id"])
		if postID == 0 {
			continue
		}

		post := &models.DynamicPost{
			ID:        postID,
			Title:     utils.InterfaceToString(row["title"]),
			Content:   utils.InterfaceToString(row["content"]),
			UserID:    utils.InterfaceToInt64(row["user_id"]),
			Username:  utils.InterfaceToString(row["username"]),
			ViewCount: utils.InterfaceToInt(row["view_count"]),
			CreatedAt: utils.InterfaceToTime(row["created_at"], time.Now()),
			UpdatedAt: utils.InterfaceToTime(row["updated_at"], time.Now()),
			Fields:    make(map[string]models.DynamicField),
			RawData:   row,
		}

		// 동적 필드 처리
		for _, field := range fields {
			if val, ok := row[field.ColumnName]; ok {
				post.Fields[field.Name] = models.DynamicField{
					Name:       field.Name,
					ColumnName: field.ColumnName,
					Value:      val,
					FieldType:  field.FieldType,
					Required:   field.Required,
				}
			}
		}

		validPosts = append(validPosts, post)
	}

	return validPosts, count, nil
}

func (s *boardService) SearchPosts(ctx context.Context, boardID int64, query string, page, pageSize int) ([]*models.DynamicPost, int, error) {
	// 게시판 정보 조회
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, 0, ErrBoardNotFound
	}

	// 검색어 처리

	// 검색 가능한 필드 조회
	var searchableFields []*models.BoardField
	err = s.db.NewSelect().
		Model(&searchableFields).
		Where("board_id = ? AND searchable = ?", boardID, true).
		Order("sort_order ASC").
		Scan(ctx)

	if err != nil {
		return nil, 0, err
	}

	// 카운트 쿼리용 조건 (별칭 없음)
	countConditions := []string{
		"title LIKE ?",
		"content LIKE ?",
	}

	// 선택 쿼리용 조건 (p. 별칭 포함)
	selectConditions := []string{
		"p.title LIKE ?",
		"p.content LIKE ?",
	}

	// 검색 패턴 생성
	searchPattern := "%" + query + "%"

	// 파라미터 준비 (각 쿼리에 대해 복제)
	countParams := []any{
		searchPattern,
		searchPattern,
	}

	selectParams := []any{
		searchPattern,
		searchPattern,
	}

	// 동적 필드에 대한 조건 추가
	for _, field := range searchableFields {
		// 카운트 쿼리용 (별칭 없음)
		countConditions = append(countConditions, fmt.Sprintf("%s LIKE ?", field.ColumnName))
		countParams = append(countParams, searchPattern)

		// 선택 쿼리용 (p. 별칭 포함)
		selectConditions = append(selectConditions, fmt.Sprintf("p.%s LIKE ?", field.ColumnName))
		selectParams = append(selectParams, searchPattern)
	}

	// 각각의 WHERE 절 생성
	countWhereClause := strings.Join(countConditions, " OR ")
	selectWhereClause := strings.Join(selectConditions, " OR ")

	// 페이지네이션 계산
	offset := (page - 1) * pageSize

	// 총 게시물 수 조회 (별칭 없음)
	var countQuery *bun.SelectQuery
	tableName := board.TableName

	countQuery = s.db.NewSelect().
		Table(tableName).
		ColumnExpr("COUNT(*) AS count").
		Where(countWhereClause, countParams...)

	var count int
	err = countQuery.Scan(ctx, &count)
	if err != nil {
		return nil, 0, err
	}

	// 게시물 목록 조회 (p. 별칭 사용)
	var selectQuery *bun.SelectQuery
	var tableExpr string

	if utils.IsPostgres(s.db) {
		tableExpr = fmt.Sprintf("\"%s\" AS p", board.TableName)
	} else {
		tableExpr = fmt.Sprintf("`%s` AS p", board.TableName)
	}

	selectQuery = s.db.NewSelect().
		TableExpr(tableExpr).
		Column("p.id", "p.title", "p.content", "p.user_id", "p.view_count", "p.created_at", "p.updated_at").
		ColumnExpr("u.username").
		Join("LEFT JOIN users AS u ON u.id = p.user_id").
		Where(selectWhereClause, selectParams...).
		OrderExpr("p.created_at DESC").
		Limit(pageSize).
		Offset(offset)

	// 쿼리 실행 및 결과 처리
	var rows []map[string]any
	err = selectQuery.Scan(ctx, &rows)
	if err != nil {
		return nil, 0, err
	}

	// 게시판 필드 정보 조회
	fields, err := s.boardRepo.GetFieldsByBoardID(ctx, boardID)
	if err != nil {
		return nil, 0, err
	}

	// 결과 변환
	validPosts := make([]*models.DynamicPost, 0, len(rows))
	for _, row := range rows {
		// ID 값 확인
		if row["id"] == nil {
			continue
		}

		// 타입 변환에 유틸리티 함수 사용
		postID := utils.InterfaceToInt64(row["id"])
		if postID == 0 {
			continue
		}

		post := &models.DynamicPost{
			ID:        postID,
			Title:     utils.InterfaceToString(row["title"]),
			Content:   utils.InterfaceToString(row["content"]),
			UserID:    utils.InterfaceToInt64(row["user_id"]),
			Username:  utils.InterfaceToString(row["username"]),
			ViewCount: utils.InterfaceToInt(row["view_count"]),
			CreatedAt: utils.InterfaceToTime(row["created_at"], time.Now()),
			UpdatedAt: utils.InterfaceToTime(row["updated_at"], time.Now()),
			Fields:    make(map[string]models.DynamicField),
			RawData:   row,
		}

		// 동적 필드 처리
		for _, field := range fields {
			if val, ok := row[field.ColumnName]; ok {
				post.Fields[field.Name] = models.DynamicField{
					Name:       field.Name,
					ColumnName: field.ColumnName,
					Value:      val,
					FieldType:  field.FieldType,
					Required:   field.Required,
				}
			}
		}

		validPosts = append(validPosts, post)
	}

	return validPosts, count, nil
}

// 썸네일 관련
func (s *boardService) GetPostThumbnails(ctx context.Context, boardID int64, postIDs []int64) (map[int64]string, error) {
	thumbnails := make(map[int64]string)

	// 첨부 파일 조회를 위한 쿼리
	query := s.db.NewSelect().
		Model((*models.Attachment)(nil)).
		Where("board_id = ?", boardID).
		Where("post_id IN (?)", bun.In(postIDs)).
		Where("is_image = ?", true).
		Order("created_at ASC")

	var attachments []*models.Attachment
	err := query.Scan(ctx, &attachments)
	if err != nil {
		return nil, err
	}

	// 각 게시물마다 첫 번째 이미지를 썸네일로 선택
	for _, attachment := range attachments {
		// 해당 게시물의 첫 번째 이미지만 저장 (이미 썸네일이 있는 경우 건너뜀)
		if _, exists := thumbnails[attachment.PostID]; !exists {
			// 모든 경로를 URL 형식으로 변환 (슬래시 사용)
			thumbnails[attachment.PostID] = filepath.ToSlash(attachment.DownloadURL)

			// URL이 /attachments로 시작하지 않으면 첨부파일 다운로드 URL 사용
			if !strings.HasPrefix(thumbnails[attachment.PostID], "/attachments") {
				thumbnails[attachment.PostID] = fmt.Sprintf("/attachments/%d/download", attachment.ID)
			}
		}
	}

	return thumbnails, nil
}

// 메서드 구현
func (s *boardService) SearchPostsWithStatus(ctx context.Context, boardID int64, query, status string, page, pageSize int) ([]*models.DynamicPost, int, error) {
	// 게시판 정보 조회
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, 0, ErrBoardNotFound
	}

	// 검색 가능한 필드 조회
	var searchableFields []*models.BoardField
	err = s.db.NewSelect().
		Model(&searchableFields).
		Where("board_id = ? AND searchable = ?", boardID, true).
		Order("sort_order ASC").
		Scan(ctx)

	if err != nil {
		return nil, 0, err
	}

	// 카운트 쿼리용 조건 (별칭 없음)
	var countConditions []string
	var countParams []any

	// 선택 쿼리용 조건 (p. 별칭 포함)
	var selectConditions []string
	var selectParams []any

	// 검색어가 있는 경우 검색 조건 추가
	if query != "" {
		// 제목 및 내용 검색
		countConditions = append(countConditions, "title LIKE ?", "content LIKE ?")
		selectConditions = append(selectConditions, "p.title LIKE ?", "p.content LIKE ?")

		// 검색 패턴 생성
		searchPattern := "%" + query + "%"
		countParams = append(countParams, searchPattern, searchPattern)
		selectParams = append(selectParams, searchPattern, searchPattern)

		// 동적 필드에 대한 검색 조건 추가
		for _, field := range searchableFields {
			if field.Name != "status" { // 상태는 별도로 처리
				countConditions = append(countConditions, fmt.Sprintf("%s LIKE ?", field.ColumnName))
				selectConditions = append(selectConditions, fmt.Sprintf("p.%s LIKE ?", field.ColumnName))
				countParams = append(countParams, searchPattern)
				selectParams = append(selectParams, searchPattern)
			}
		}
	}

	// 상태 필터 추가
	if status != "" {
		countConditions = append(countConditions, "status = ?")
		selectConditions = append(selectConditions, "p.status = ?")
		countParams = append(countParams, status)
		selectParams = append(selectParams, status)
	}

	// 각각의 WHERE 절 생성
	var countWhereClause string
	var selectWhereClause string

	if len(countConditions) > 0 {
		if query != "" && status != "" {
			// 검색어와 상태 모두 있는 경우
			countWhereClause = "(" + strings.Join(countConditions[:len(countConditions)-1], " OR ") + ") AND " + countConditions[len(countConditions)-1]
			selectWhereClause = "(" + strings.Join(selectConditions[:len(selectConditions)-1], " OR ") + ") AND " + selectConditions[len(selectConditions)-1]
		} else {
			// 검색어나 상태 중 하나만 있는 경우
			countWhereClause = strings.Join(countConditions, " OR ")
			selectWhereClause = strings.Join(selectConditions, " OR ")
		}
	}

	// 페이지네이션 계산
	offset := (page - 1) * pageSize

	// 총 게시물 수 조회
	var countQuery *bun.SelectQuery
	tableName := board.TableName

	countQuery = s.db.NewSelect().
		Table(tableName).
		ColumnExpr("COUNT(*) AS count")

	if countWhereClause != "" {
		countQuery = countQuery.Where(countWhereClause, countParams...)
	}

	var count int
	err = countQuery.Scan(ctx, &count)
	if err != nil {
		return nil, 0, err
	}

	// 게시물 목록 조회
	var selectQuery *bun.SelectQuery
	var tableExpr string

	if utils.IsPostgres(s.db) {
		tableExpr = fmt.Sprintf("\"%s\" AS p", board.TableName)
	} else {
		tableExpr = fmt.Sprintf("`%s` AS p", board.TableName)
	}

	selectQuery = s.db.NewSelect().
		TableExpr(tableExpr).
		Column("p.*").
		ColumnExpr("u.username").
		Join("LEFT JOIN users AS u ON u.id = p.user_id")

	if selectWhereClause != "" {
		selectQuery = selectQuery.Where(selectWhereClause, selectParams...)
	}

	selectQuery = selectQuery.
		OrderExpr("p.created_at DESC").
		Limit(pageSize).
		Offset(offset)

	// 쿼리 실행 및 결과 처리
	var rows []map[string]any
	err = selectQuery.Scan(ctx, &rows)
	if err != nil {
		return nil, 0, err
	}

	// 게시판 필드 정보 조회
	fields, err := s.boardRepo.GetFieldsByBoardID(ctx, boardID)
	if err != nil {
		return nil, 0, err
	}

	// 결과 변환
	validPosts := make([]*models.DynamicPost, 0, len(rows))
	for _, row := range rows {
		// ID 값 확인
		if row["id"] == nil {
			continue
		}

		// 타입 변환에 유틸리티 함수 사용
		postID := utils.InterfaceToInt64(row["id"])
		if postID == 0 {
			continue
		}

		post := &models.DynamicPost{
			ID:        postID,
			Title:     utils.InterfaceToString(row["title"]),
			Content:   utils.InterfaceToString(row["content"]),
			UserID:    utils.InterfaceToInt64(row["user_id"]),
			Username:  utils.InterfaceToString(row["username"]),
			ViewCount: utils.InterfaceToInt(row["view_count"]),
			CreatedAt: utils.InterfaceToTime(row["created_at"], time.Now()),
			UpdatedAt: utils.InterfaceToTime(row["updated_at"], time.Now()),
			Fields:    make(map[string]models.DynamicField),
			RawData:   row,
		}

		// 동적 필드 처리
		for _, field := range fields {
			if val, ok := row[field.ColumnName]; ok {
				post.Fields[field.Name] = models.DynamicField{
					Name:       field.Name,
					ColumnName: field.ColumnName,
					Value:      val,
					FieldType:  field.FieldType,
					Required:   field.Required,
				}
			}
		}

		validPosts = append(validPosts, post)
	}

	return validPosts, count, nil
}
