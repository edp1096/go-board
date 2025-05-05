// internal/service/category_service.go
package service

import (
	"context"
	"errors"
	"time"

	"github.com/edp1096/go-board/internal/models"
	"github.com/edp1096/go-board/internal/repository"

	"github.com/gosimple/slug"
)

var (
	ErrCategoryNotFound = errors.New("카테고리를 찾을 수 없음")
	ErrInvalidCategory  = errors.New("유효하지 않은 카테고리")
)

type CategoryService interface {
	CreateCategory(ctx context.Context, category *models.Category) error
	GetCategoryByID(ctx context.Context, id int64) (*models.Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error)
	UpdateCategory(ctx context.Context, category *models.Category) error
	DeleteCategory(ctx context.Context, id int64) error
	ListCategories(ctx context.Context, onlyActive bool, parentID *int64) ([]*models.Category, error)

	// 계층 구조 관련
	GetCategoryTree(ctx context.Context, onlyActive bool) ([]*models.Category, error)

	// 카테고리-아이템 관계 관리
	AddItemToCategory(ctx context.Context, categoryID, itemID int64, itemType string) error
	RemoveItemFromCategory(ctx context.Context, categoryID, itemID int64, itemType string) error
	GetItemsByCategory(ctx context.Context, categoryID int64) ([]*models.CategoryItem, error)
	GetCategoriesByItem(ctx context.Context, itemID int64, itemType string) ([]*models.Category, error)

	// 메뉴 구성
	GetMenuStructure(ctx context.Context, onlyRoot bool) ([]map[string]any, error)
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
	boardRepo    repository.BoardRepository
	pageRepo     repository.PageRepository
}

func NewCategoryService(
	categoryRepo repository.CategoryRepository,
	boardRepo repository.BoardRepository,
	pageRepo repository.PageRepository,
) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
		boardRepo:    boardRepo,
		pageRepo:     pageRepo,
	}
}

func (s *categoryService) CreateCategory(ctx context.Context, category *models.Category) error {
	// 슬러그가 없으면 생성
	if category.Slug == "" {
		category.Slug = slug.Make(category.Name)
	}

	// 생성 시간 설정
	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now

	return s.categoryRepo.Create(ctx, category)
}

func (s *categoryService) GetCategoryByID(ctx context.Context, id int64) (*models.Category, error) {
	return s.categoryRepo.GetByID(ctx, id)
}

func (s *categoryService) GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error) {
	return s.categoryRepo.GetBySlug(ctx, slug)
}

func (s *categoryService) UpdateCategory(ctx context.Context, category *models.Category) error {
	// 업데이트 시간 설정
	category.UpdatedAt = time.Now()
	return s.categoryRepo.Update(ctx, category)
}

func (s *categoryService) DeleteCategory(ctx context.Context, id int64) error {
	return s.categoryRepo.Delete(ctx, id)
}

func (s *categoryService) ListCategories(ctx context.Context, onlyActive bool, parentID *int64) ([]*models.Category, error) {
	return s.categoryRepo.List(ctx, onlyActive, parentID)
}

// GetCategoryTree 카테고리 트리 구조 가져오기
func (s *categoryService) GetCategoryTree(ctx context.Context, onlyActive bool) ([]*models.Category, error) {
	// 모든 카테고리 조회 (플랫 구조)
	allCategories, err := s.categoryRepo.List(ctx, onlyActive, nil)
	if err != nil {
		return nil, err
	}

	// 카테고리 맵 생성 (ID -> 카테고리)
	categoryMap := make(map[int64]*models.Category)
	for _, category := range allCategories {
		categoryMap[category.ID] = category
		// 빈 자식 배열 초기화
		category.Children = make([]*models.Category, 0)
	}

	// 계층 구조 구성
	rootCategories := make([]*models.Category, 0)
	for _, category := range allCategories {
		if category.ParentID == nil {
			// 최상위 카테고리
			rootCategories = append(rootCategories, category)
		} else {
			// 하위 카테고리
			if parent, ok := categoryMap[*category.ParentID]; ok {
				parent.Children = append(parent.Children, category)
			}
		}
	}

	return rootCategories, nil
}

// 하위 카테고리 재귀적 로드
func (s *categoryService) loadChildCategories(ctx context.Context, parent *models.Category, onlyActive bool) error {
	childCategories, err := s.categoryRepo.List(ctx, onlyActive, &parent.ID)
	if err != nil {
		return err
	}

	parent.Children = childCategories

	// 각 하위 카테고리에 대해 재귀 호출
	for _, child := range childCategories {
		if err := s.loadChildCategories(ctx, child, onlyActive); err != nil {
			return err
		}
	}

	return nil
}

// 카테고리-아이템 관계 관리 메서드

func (s *categoryService) AddItemToCategory(ctx context.Context, categoryID, itemID int64, itemType string) error {
	// 카테고리 존재 여부 확인
	_, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return ErrCategoryNotFound
	}

	// 아이템 존재 여부 확인
	switch itemType {
	case "board":
		_, err = s.boardRepo.GetByID(ctx, itemID)
		if err != nil {
			return ErrBoardNotFound
		}
	case "page":
		_, err = s.pageRepo.GetByID(ctx, itemID)
		if err != nil {
			return ErrPageNotFound
		}
	default:
		return errors.New("지원하지 않는 아이템 타입")
	}

	// 기존 정렬 순서 중 가장 큰 값 조회
	items, err := s.categoryRepo.GetItemsByCategory(ctx, categoryID)
	if err != nil {
		return err
	}

	// 새 아이템의 정렬 순서 계산
	sortOrder := 0
	if len(items) > 0 {
		for _, item := range items {
			if item.SortOrder > sortOrder {
				sortOrder = item.SortOrder
			}
		}
		sortOrder++
	}

	// 새 관계 추가
	categoryItem := &models.CategoryItem{
		CategoryID: categoryID,
		ItemID:     itemID,
		ItemType:   itemType,
		SortOrder:  sortOrder,
		CreatedAt:  time.Now(),
	}

	return s.categoryRepo.AddItem(ctx, categoryItem)
}

func (s *categoryService) RemoveItemFromCategory(ctx context.Context, categoryID, itemID int64, itemType string) error {
	return s.categoryRepo.RemoveItem(ctx, categoryID, itemID, itemType)
}

func (s *categoryService) GetItemsByCategory(ctx context.Context, categoryID int64) ([]*models.CategoryItem, error) {
	return s.categoryRepo.GetItemsByCategory(ctx, categoryID)
}

func (s *categoryService) GetCategoriesByItem(ctx context.Context, itemID int64, itemType string) ([]*models.Category, error) {
	return s.categoryRepo.GetCategoriesByItem(ctx, itemID, itemType)
}

// 메뉴 구조 생성 메서드
func (s *categoryService) GetMenuStructure(ctx context.Context, onlyRoot bool) ([]map[string]any, error) {
	var categories []*models.Category
	var err error

	if onlyRoot {
		// 최상위 카테고리만 조회 (parent_id가 NULL인 카테고리)
		categories, err = s.categoryRepo.List(ctx, true, nil)
	} else {
		// 모든 활성 카테고리 조회
		allCategories, err := s.categoryRepo.List(ctx, true, nil)
		if err != nil {
			return nil, err
		}

		// 최상위 및 중간 카테고리만 필터링 (필요한 경우)
		categories = allCategories
	}

	if err != nil {
		return nil, err
	}

	menuItems := make([]map[string]interface{}, 0)

	// 카테고리별 메뉴 구성
	for _, category := range categories {
		// parent_id가 있는 카테고리는 최상위 메뉴에서 제외
		if onlyRoot && category.ParentID != nil {
			continue
		}

		// 카테고리 정보
		categoryItem := map[string]interface{}{
			"id":        category.ID,
			"name":      category.Name,
			"slug":      category.Slug,
			"type":      "category",
			"sortOrder": category.SortOrder,
			"children":  []map[string]interface{}{},
		}

		// 해당 카테고리에 속한 아이템 조회
		items, err := s.categoryRepo.GetItemsByCategory(ctx, category.ID)
		if err != nil {
			return nil, err
		}

		children := make([]map[string]interface{}, 0)

		// 아이템 정보 구성
		for _, item := range items {
			var itemDetail map[string]interface{}

			if item.ItemType == "board" {
				// 게시판 정보 조회
				board, err := s.boardRepo.GetByID(ctx, item.ItemID)
				if err != nil {
					continue
				}

				if !board.Active {
					continue
				}

				itemDetail = map[string]interface{}{
					"id":        board.ID,
					"name":      board.Name,
					"slug":      board.Slug,
					"type":      "board",
					"sortOrder": item.SortOrder,
				}
			} else if item.ItemType == "page" {
				// 페이지 정보 조회
				page, err := s.pageRepo.GetByID(ctx, item.ItemID)
				if err != nil {
					continue
				}

				if !page.Active || !page.ShowInMenu {
					continue
				}

				itemDetail = map[string]interface{}{
					"id":        page.ID,
					"name":      page.Title,
					"slug":      page.Slug,
					"type":      "page",
					"sortOrder": item.SortOrder,
				}
			}

			if itemDetail != nil {
				children = append(children, itemDetail)
			}
		}

		// 하위 카테고리 처리
		subCategories, err := s.categoryRepo.List(ctx, true, &category.ID)
		if err != nil {
			return nil, err
		}

		for _, subCat := range subCategories {
			subCategoryItem := map[string]interface{}{
				"id":        subCat.ID,
				"name":      subCat.Name,
				"slug":      subCat.Slug,
				"type":      "category",
				"sortOrder": subCat.SortOrder,
				"children":  []map[string]interface{}{},
			}

			// 하위 카테고리의 아이템 및 더 깊은 하위 카테고리 로드
			subItems, subSubCategories, err := s.loadCategoryContents(ctx, subCat.ID)
			if err != nil {
				return nil, err
			}

			// 모든 하위 항목을 children에 추가
			subChildren := append(subItems, subSubCategories...)
			subCategoryItem["children"] = subChildren

			children = append(children, subCategoryItem)
		}

		categoryItem["children"] = children
		menuItems = append(menuItems, categoryItem)
	}

	return menuItems, nil
}

// 카테고리 내용 로드 헬퍼 함수
func (s *categoryService) loadCategoryContents(ctx context.Context, categoryID int64) ([]map[string]interface{}, []map[string]interface{}, error) {
	// 카테고리 아이템 로드
	items, err := s.categoryRepo.GetItemsByCategory(ctx, categoryID)
	if err != nil {
		return nil, nil, err
	}

	itemDetails := make([]map[string]interface{}, 0)

	for _, item := range items {
		var itemDetail map[string]interface{}

		if item.ItemType == "board" {
			board, err := s.boardRepo.GetByID(ctx, item.ItemID)
			if err != nil {
				continue
			}

			if !board.Active {
				continue
			}

			itemDetail = map[string]interface{}{
				"id":        board.ID,
				"name":      board.Name,
				"slug":      board.Slug,
				"type":      "board",
				"sortOrder": item.SortOrder,
			}
		} else if item.ItemType == "page" {
			page, err := s.pageRepo.GetByID(ctx, item.ItemID)
			if err != nil {
				continue
			}

			if !page.Active || !page.ShowInMenu {
				continue
			}

			itemDetail = map[string]interface{}{
				"id":        page.ID,
				"name":      page.Title,
				"slug":      page.Slug,
				"type":      "page",
				"sortOrder": item.SortOrder,
			}
		}

		if itemDetail != nil {
			itemDetails = append(itemDetails, itemDetail)
		}
	}

	// 하위 카테고리 로드
	subCategories, err := s.categoryRepo.List(ctx, true, &categoryID)
	if err != nil {
		return nil, nil, err
	}

	subCategoryDetails := make([]map[string]interface{}, 0)

	for _, subCat := range subCategories {
		subCategoryItem := map[string]interface{}{
			"id":        subCat.ID,
			"name":      subCat.Name,
			"slug":      subCat.Slug,
			"type":      "category",
			"sortOrder": subCat.SortOrder,
			"children":  []map[string]interface{}{},
		}

		// 재귀적으로 하위 아이템 로드
		subItems, subSubCategories, err := s.loadCategoryContents(ctx, subCat.ID)
		if err != nil {
			return nil, nil, err
		}

		subChildren := append(subItems, subSubCategories...)
		subCategoryItem["children"] = subChildren

		subCategoryDetails = append(subCategoryDetails, subCategoryItem)
	}

	return itemDetails, subCategoryDetails, nil
}
