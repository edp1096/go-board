// internal/repository/page_repository.go
package repository

import (
	"context"

	"github.com/edp1096/go-board/internal/models"

	"github.com/uptrace/bun"
)

type PageRepository interface {
	Create(ctx context.Context, page *models.Page) error
	GetByID(ctx context.Context, id int64) (*models.Page, error)
	GetBySlug(ctx context.Context, slug string) (*models.Page, error)
	Update(ctx context.Context, page *models.Page) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, onlyActive bool) ([]*models.Page, error)
}

type pageRepository struct {
	db *bun.DB
}

func NewPageRepository(db *bun.DB) PageRepository {
	return &pageRepository{db: db}
}

func (r *pageRepository) Create(ctx context.Context, page *models.Page) error {
	_, err := r.db.NewInsert().Model(page).Exec(ctx)
	return err
}

func (r *pageRepository) GetByID(ctx context.Context, id int64) (*models.Page, error) {
	page := new(models.Page)
	err := r.db.NewSelect().
		Model(page).
		Where("p.id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return page, nil
}

func (r *pageRepository) GetBySlug(ctx context.Context, slug string) (*models.Page, error) {
	page := new(models.Page)
	err := r.db.NewSelect().
		Model(page).
		Where("p.slug = ?", slug).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return page, nil
}

func (r *pageRepository) Update(ctx context.Context, page *models.Page) error {
	_, err := r.db.NewUpdate().Model(page).WherePK().Exec(ctx)
	return err
}

func (r *pageRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model((*models.Page)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *pageRepository) List(ctx context.Context, onlyActive bool) ([]*models.Page, error) {
	var pages []*models.Page
	query := r.db.NewSelect().Model(&pages)

	if onlyActive {
		query = query.Where("active = ?", true)
	}

	err := query.Order("sort_order ASC").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return pages, nil
}
