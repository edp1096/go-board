// internal/service/board_service.go
package service

import (
	"context"
	"dynamic-board/internal/models"
	"dynamic-board/internal/repository"
	"errors"
	"fmt"
	"strconv"
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
		board.TableName = fmt.Sprintf("board_%s", slug.Make(board.Name))
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
	values := map[string]interface{}{
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

	// 방법 1: Model 메서드 사용하기 (권장)
	res, err := s.db.NewInsert().
		Model(&values).
		Table(board.TableName).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("게시물 생성 실패: %w", err)
	}

	// 생성된 ID 반환 (데이터베이스에 따라 다를 수 있음)
	id, err := res.LastInsertId()
	if err == nil {
		post.ID = id
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
	query := s.db.NewSelect().
		Table(board.TableName).
		Column("p.*").
		ColumnExpr("u.username").
		Join("LEFT JOIN users AS u ON u.id = p.user_id").
		Where("p.id = ?", postID)

	// 쿼리 실행
	var row map[string]interface{}
	err = query.Scan(ctx, &row)
	if err != nil {
		return nil, ErrPostNotFound
	}

	// 조회수 증가
	// _, err = s.db.NewUpdate().
	s.db.NewUpdate().
		Table(board.TableName).
		Set("view_count = view_count + 1").
		Where("id = ?", postID).
		Exec(ctx)

	// DynamicPost 객체 생성
	post := &models.DynamicPost{
		ID:        postID,
		Title:     row["title"].(string),
		Content:   row["content"].(string),
		UserID:    row["user_id"].(int64),
		Username:  row["username"].(string),
		ViewCount: row["view_count"].(int) + 1, // 방금 증가한 조회수 반영
		CreatedAt: row["created_at"].(time.Time),
		UpdatedAt: row["updated_at"].(time.Time),
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
	values := map[string]interface{}{
		"title":      post.Title,
		"content":    post.Content,
		"updated_at": time.Now(),
	}

	// 동적 필드 추가
	for _, field := range post.Fields {
		values[field.ColumnName] = field.Value
	}

	// Model 메서드를 사용하여 업데이트
	_, err = s.db.NewUpdate().
		Model(&values).
		Table(board.TableName).
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
	_, err = s.db.NewDelete().
		Table(board.TableName).
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
	countQuery := s.db.NewSelect().
		Table(board.TableName).
		ColumnExpr("COUNT(*) AS count")

	var count int
	err = countQuery.Scan(ctx, &count)
	if err != nil {
		return nil, 0, err
	}

	// 게시물 목록 조회
	query := s.db.NewSelect().
		Table(board.TableName + " AS p").
		Column("p.*").
		ColumnExpr("u.username").
		Join("LEFT JOIN users AS u ON u.id = p.user_id").
		OrderExpr(fmt.Sprintf("p.%s %s", sortField, sortDir)).
		Limit(pageSize).
		Offset(offset)

	// 쿼리 실행
	var rows []map[string]interface{}
	err = query.Scan(ctx, &rows)
	if err != nil {
		return nil, 0, err
	}

	// 결과 변환
	posts := make([]*models.DynamicPost, len(rows))
	for i, row := range rows {
		// ID 가져오기 (데이터베이스와 드라이버에 따라 타입이 다를 수 있음)
		var postID int64
		switch id := row["id"].(type) {
		case int64:
			postID = id
		case int:
			postID = int64(id)
		case float64:
			postID = int64(id)
		case string:
			postID, _ = strconv.ParseInt(id, 10, 64)
		}

		// 사용자 ID 가져오기
		var userID int64
		switch uid := row["user_id"].(type) {
		case int64:
			userID = uid
		case int:
			userID = int64(uid)
		case float64:
			userID = int64(uid)
		}

		// 조회수 가져오기
		var viewCount int
		switch vc := row["view_count"].(type) {
		case int:
			viewCount = vc
		case int64:
			viewCount = int(vc)
		case float64:
			viewCount = int(vc)
		}

		post := &models.DynamicPost{
			ID:        postID,
			Title:     row["title"].(string),
			Content:   row["content"].(string),
			UserID:    userID,
			Username:  row["username"].(string),
			ViewCount: viewCount,
			CreatedAt: row["created_at"].(time.Time),
			UpdatedAt: row["updated_at"].(time.Time),
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

		posts[i] = post
	}

	return posts, count, nil
}

func (s *boardService) SearchPosts(ctx context.Context, boardID int64, query string, page, pageSize int) ([]*models.DynamicPost, int, error) {
	// 게시판 정보 조회
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, 0, ErrBoardNotFound
	}

	// 검색 가능한 필드 조회 - 수정된 부분
	var searchableFields []*models.BoardField
	err = s.db.NewSelect().
		Model(&searchableFields).
		Where("board_id = ? AND searchable = ?", boardID, true).
		Order("sort_order ASC").
		Scan(ctx)

	if err != nil {
		return nil, 0, err
	}

	// 검색 조건 생성
	conditions := []string{
		fmt.Sprintf("title LIKE '%%%s%%'", query),
		fmt.Sprintf("content LIKE '%%%s%%'", query),
	}

	for _, field := range searchableFields {
		conditions = append(conditions, fmt.Sprintf("%s LIKE '%%%s%%'", field.ColumnName, query))
	}

	whereClause := strings.Join(conditions, " OR ")

	// 페이지네이션 계산
	offset := (page - 1) * pageSize

	// 총 게시물 수 조회
	countQuery := s.db.NewSelect().
		Table(board.TableName).
		ColumnExpr("COUNT(*) AS count").
		Where(whereClause)

	var count int
	err = countQuery.Scan(ctx, &count)
	if err != nil {
		return nil, 0, err
	}

	// 게시물 목록 조회
	selectQuery := s.db.NewSelect().
		Table(board.TableName + " AS p").
		Column("p.*").
		ColumnExpr("u.username").
		Join("LEFT JOIN users AS u ON u.id = p.user_id").
		Where(whereClause).
		OrderExpr("p.created_at DESC").
		Limit(pageSize).
		Offset(offset)

	// 쿼리 실행 및 결과 처리
	var rows []map[string]interface{}
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
	posts := make([]*models.DynamicPost, len(rows))
	for i, row := range rows {
		// ID 가져오기 (데이터베이스와 드라이버에 따라 타입이 다를 수 있음)
		var postID int64
		switch id := row["id"].(type) {
		case int64:
			postID = id
		case int:
			postID = int64(id)
		case float64:
			postID = int64(id)
		case string:
			postID, _ = strconv.ParseInt(id, 10, 64)
		}

		// 사용자 ID 가져오기
		var userID int64
		switch uid := row["user_id"].(type) {
		case int64:
			userID = uid
		case int:
			userID = int64(uid)
		case float64:
			userID = int64(uid)
		}

		// 조회수 가져오기
		var viewCount int
		switch vc := row["view_count"].(type) {
		case int:
			viewCount = vc
		case int64:
			viewCount = int(vc)
		case float64:
			viewCount = int(vc)
		}

		post := &models.DynamicPost{
			ID:        postID,
			Title:     row["title"].(string),
			Content:   row["content"].(string),
			UserID:    userID,
			Username:  row["username"].(string),
			ViewCount: viewCount,
			CreatedAt: row["created_at"].(time.Time),
			UpdatedAt: row["updated_at"].(time.Time),
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

		posts[i] = post
	}

	return posts, count, nil
}
