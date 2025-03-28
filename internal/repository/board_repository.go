// internal/repository/board_repository.go
package repository

import (
	"context"
	"dynamic-board/internal/models"

	"github.com/uptrace/bun"
)

type BoardRepository interface {
	Create(ctx context.Context, board *models.Board) error
	GetByID(ctx context.Context, id int64) (*models.Board, error)
	GetBySlug(ctx context.Context, slug string) (*models.Board, error)
	Update(ctx context.Context, board *models.Board) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, onlyActive bool) ([]*models.Board, error)

	// 게시판 필드 관련
	CreateField(ctx context.Context, field *models.BoardField) error
	GetFieldsByBoardID(ctx context.Context, boardID int64) ([]*models.BoardField, error)
	UpdateField(ctx context.Context, field *models.BoardField) error
	DeleteField(ctx context.Context, id int64) error
}

type boardRepository struct {
	db *bun.DB
}

func NewBoardRepository(db *bun.DB) BoardRepository {
	return &boardRepository{db: db}
}

func (r *boardRepository) Create(ctx context.Context, board *models.Board) error {
	_, err := r.db.NewInsert().Model(board).Exec(ctx)
	return err
}

func (r *boardRepository) GetByID(ctx context.Context, id int64) (*models.Board, error) {
	board := new(models.Board)
	err := r.db.NewSelect().
		Model(board).
		Relation("Fields", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("sort_order ASC")
		}).
		Where("b.id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return board, nil
}

func (r *boardRepository) GetBySlug(ctx context.Context, slug string) (*models.Board, error) {
	board := new(models.Board)
	err := r.db.NewSelect().
		Model(board).
		Relation("Fields", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("sort_order ASC")
		}).
		Where("b.slug = ?", slug).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return board, nil
}

func (r *boardRepository) Update(ctx context.Context, board *models.Board) error {
	_, err := r.db.NewUpdate().Model(board).WherePK().Exec(ctx)
	return err
}

func (r *boardRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model((*models.Board)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *boardRepository) List(ctx context.Context, onlyActive bool) ([]*models.Board, error) {
	var boards []*models.Board
	query := r.db.NewSelect().Model(&boards)

	if onlyActive {
		query = query.Where("active = ?", true)
	}

	err := query.Order("name ASC").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return boards, nil
}

// 게시판 필드 관련 메서드

func (r *boardRepository) CreateField(ctx context.Context, field *models.BoardField) error {
	_, err := r.db.NewInsert().Model(field).Exec(ctx)
	return err
}

func (r *boardRepository) GetFieldsByBoardID(ctx context.Context, boardID int64) ([]*models.BoardField, error) {
	var fields []*models.BoardField
	err := r.db.NewSelect().
		Model(&fields).
		Where("board_id = ?", boardID).
		Order("sort_order ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return fields, nil
}

func (r *boardRepository) UpdateField(ctx context.Context, field *models.BoardField) error {
	_, err := r.db.NewUpdate().Model(field).WherePK().Exec(ctx)
	return err
}

func (r *boardRepository) DeleteField(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model((*models.BoardField)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}
