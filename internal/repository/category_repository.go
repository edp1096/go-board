// internal/repository/category_repository.go
package repository

import (
	"context"

	"github.com/edp1096/go-board/internal/models"

	"github.com/uptrace/bun"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *models.Category) error
	GetByID(ctx context.Context, id int64) (*models.Category, error)
	GetBySlug(ctx context.Context, slug string) (*models.Category, error)
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, onlyActive bool, parentID *int64) ([]*models.Category, error)

	// 카테고리-아이템 관계 관리
	AddItem(ctx context.Context, categoryItem *models.CategoryItem) error
	RemoveItem(ctx context.Context, categoryID, itemID int64, itemType string) error
	GetItemsByCategory(ctx context.Context, categoryID int64) ([]*models.CategoryItem, error)
	GetCategoriesByItem(ctx context.Context, itemID int64, itemType string) ([]*models.Category, error)
}

type categoryRepository struct {
	db *bun.DB
}

func NewCategoryRepository(db *bun.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *models.Category) error {
	_, err := r.db.NewInsert().Model(category).Exec(ctx)
	return err
}

func (r *categoryRepository) GetByID(ctx context.Context, id int64) (*models.Category, error) {
	category := new(models.Category)
	err := r.db.NewSelect().
		Model(category).
		Relation("Parent").
		Relation("Children", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("active = ?", true).Order("sort_order ASC")
		}).
		Where("c.id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (r *categoryRepository) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	category := new(models.Category)
	err := r.db.NewSelect().
		Model(category).
		Relation("Parent").
		Relation("Children", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("active = ?", true).Order("sort_order ASC")
		}).
		Where("c.slug = ?", slug).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *models.Category) error {
	_, err := r.db.NewUpdate().Model(category).WherePK().Exec(ctx)
	return err
}

func (r *categoryRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model((*models.Category)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

// 카테고리 목록
func (r *categoryRepository) List(ctx context.Context, onlyActive bool, parentID *int64) ([]*models.Category, error) {
	var categories []*models.Category
	query := r.db.NewSelect().Model(&categories)

	if onlyActive {
		query = query.Where("active = ?", true)
	}

	// parentID 매개변수 처리 수정
	if parentID != nil {
		// 특정 부모 카테고리의 하위 카테고리 조회
		query = query.Where("parent_id = ?", *parentID)
	}
	// parentID == nil인 경우 조건을 추가하지 않음 (모든 카테고리 조회)

	err := query.Order("sort_order ASC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

// 카테고리-아이템 관계 관리 메서드

func (r *categoryRepository) AddItem(ctx context.Context, categoryItem *models.CategoryItem) error {
	_, err := r.db.NewInsert().Model(categoryItem).Exec(ctx)
	return err
}

func (r *categoryRepository) RemoveItem(ctx context.Context, categoryID, itemID int64, itemType string) error {
	_, err := r.db.NewDelete().
		Model((*models.CategoryItem)(nil)).
		Where("category_id = ? AND item_id = ? AND item_type = ?", categoryID, itemID, itemType).
		Exec(ctx)
	return err
}

func (r *categoryRepository) GetItemsByCategory(ctx context.Context, categoryID int64) ([]*models.CategoryItem, error) {
	var items []*models.CategoryItem
	err := r.db.NewSelect().
		Model(&items).
		Where("category_id = ?", categoryID).
		Order("sort_order ASC").
		Scan(ctx)
	return items, err
}

func (r *categoryRepository) GetCategoriesByItem(ctx context.Context, itemID int64, itemType string) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.NewSelect().
		Model(&categories).
		Join("JOIN category_items AS ci ON ci.category_id = c.id").
		Where("ci.item_id = ? AND ci.item_type = ?", itemID, itemType).
		Order("c.sort_order ASC").
		Scan(ctx)
	return categories, err
}
