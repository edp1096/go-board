// internal/handlers/category_handler.go
package handlers

import (
	"log"
	"strconv"
	"time"

	"github.com/edp1096/go-board/internal/models"
	"github.com/edp1096/go-board/internal/service"
	"github.com/edp1096/go-board/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gosimple/slug"
)

type CategoryHandler struct {
	categoryService service.CategoryService
	boardService    service.BoardService
	pageService     service.PageService
}

func NewCategoryHandler(
	categoryService service.CategoryService,
	boardService service.BoardService,
	pageService service.PageService,
) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
		boardService:    boardService,
		pageService:     pageService,
	}
}

// ListCategories 관리자용 카테고리 목록 핸들러
func (h *CategoryHandler) ListCategories(c *fiber.Ctx) error {
	var allCategories []*models.Category
	var err error

	// 	allCategories, err = h.categoryService.ListCategories(c.Context(), false, nil)
	// 계층 구조
	allCategories, err = h.categoryService.ListCategoriesWithRelations(c.Context(), false)

	for i, category := range allCategories {
		// 부모 카테고리 정보 설정
		if category.ParentID != nil {
			parentCategory, err := h.categoryService.GetCategoryByID(c.Context(), *category.ParentID)
			if err == nil {
				allCategories[i].Parent = parentCategory
			} else {
				log.Printf("부모 카테고리 조회 실패: %v", err)
			}
		}
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).Render("error", fiber.Map{
			"title":   "오류",
			"message": "카테고리 목록을 불러오는데 실패했습니다: " + err.Error(),
		})
	}

	return utils.RenderWithUser(c, "admin/categories/list", fiber.Map{
		"title":      "카테고리 관리",
		"categories": allCategories,
	})
}

// CreateCategoryPage 카테고리 생성 폼 핸들러
func (h *CategoryHandler) CreateCategoryPage(c *fiber.Ctx) error {
	// 부모 카테고리 목록 조회 (최상위 카테고리만)
	parentCategories, err := h.categoryService.ListCategories(c.Context(), true, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).Render("error", fiber.Map{
			"title":   "오류",
			"message": "카테고리 목록을 불러오는데 실패했습니다: " + err.Error(),
		})
	}

	return utils.RenderWithUser(c, "admin/categories/create", fiber.Map{
		"title":            "카테고리 생성",
		"parentCategories": parentCategories,
	})
}

// CreateCategory 카테고리 생성 처리 핸들러
func (h *CategoryHandler) CreateCategory(c *fiber.Ctx) error {
	// 현재 로그인한 관리자 확인
	user := c.Locals("user").(*models.User)
	if user.Role != models.RoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "카테고리 생성 권한이 없습니다.",
		})
	}

	// 폼 데이터 가져오기
	name := c.FormValue("name")
	description := c.FormValue("description")
	categorySlug := c.FormValue("slug")
	parentIDStr := c.FormValue("parent_id")
	sortOrderStr := c.FormValue("sort_order")
	active := c.FormValue("active") == "on"

	// 필수 필드 검증
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "이름을 입력해주세요.",
		})
	}

	// 슬러그 생성
	if categorySlug == "" {
		categorySlug = slug.Make(name)
	}

	// 부모 카테고리 ID 변환
	var parentID *int64
	if parentIDStr != "" && parentIDStr != "0" {
		id, err := strconv.ParseInt(parentIDStr, 10, 64)
		if err == nil && id > 0 {
			parentID = &id
		}
	}

	// 정렬 순서 변환
	sortOrder := 0
	if sortOrderStr != "" {
		if order, err := strconv.Atoi(sortOrderStr); err == nil {
			sortOrder = order
		}
	}

	// 카테고리 객체 생성
	category := &models.Category{
		Name:        name,
		Description: description,
		Slug:        categorySlug,
		ParentID:    parentID,
		SortOrder:   sortOrder,
		Active:      active,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 카테고리 생성
	if err := h.categoryService.CreateCategory(c.Context(), category); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "카테고리 생성 중 오류가 발생했습니다: " + err.Error(),
		})
	}

	// JSON 요청인 경우
	if c.Get("Accept") == "application/json" {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "카테고리가 생성되었습니다.",
			"id":      category.ID,
		})
	}

	// 웹 요청인 경우 카테고리 목록으로 리다이렉트
	return c.Redirect("/admin/categories")
}

// EditCategoryPage 카테고리 수정 폼 핸들러
func (h *CategoryHandler) EditCategoryPage(c *fiber.Ctx) error {
	// 카테고리 ID 가져오기
	categoryID, err := strconv.ParseInt(c.Params("categoryID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).Render("error", fiber.Map{
			"title":   "잘못된 요청",
			"message": "카테고리 ID가 유효하지 않습니다.",
		})
	}

	// 카테고리 조회
	category, err := h.categoryService.GetCategoryByID(c.Context(), categoryID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).Render("error", fiber.Map{
			"title":   "카테고리를 찾을 수 없음",
			"message": "요청하신 카테고리를 찾을 수 없습니다.",
		})
	}

	// 부모 카테고리 목록 조회 (최상위 카테고리만)
	parentCategories, err := h.categoryService.ListCategories(c.Context(), true, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).Render("error", fiber.Map{
			"title":   "오류",
			"message": "카테고리 목록을 불러오는데 실패했습니다: " + err.Error(),
		})
	}

	// 해당 카테고리에 속한 항목 조회
	categoryItems, err := h.categoryService.GetItemsByCategory(c.Context(), categoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).Render("error", fiber.Map{
			"title":   "오류",
			"message": "카테고리 항목을 불러오는데 실패했습니다: " + err.Error(),
		})
	}

	// 항목 상세 정보 조회
	boardItems := make([]map[string]interface{}, 0)
	pageItems := make([]map[string]interface{}, 0)

	for _, item := range categoryItems {
		if item.ItemType == "board" {
			board, err := h.boardService.GetBoardByID(c.Context(), item.ItemID)
			if err != nil {
				continue
			}
			boardItems = append(boardItems, map[string]interface{}{
				"id":        item.ID,
				"itemID":    board.ID,
				"name":      board.Name,
				"sortOrder": item.SortOrder,
			})
		} else if item.ItemType == "page" {
			page, err := h.pageService.GetPageByID(c.Context(), item.ItemID)
			if err != nil {
				continue
			}
			pageItems = append(pageItems, map[string]interface{}{
				"id":        item.ID,
				"itemID":    page.ID,
				"name":      page.Title,
				"sortOrder": item.SortOrder,
			})
		}
	}

	return utils.RenderWithUser(c, "admin/categories/edit", fiber.Map{
		"title":            "카테고리 수정",
		"category":         category,
		"parentCategories": parentCategories,
		"boardItems":       boardItems,
		"pageItems":        pageItems,
	})
}

// UpdateCategory 카테고리 수정 처리 핸들러
func (h *CategoryHandler) UpdateCategory(c *fiber.Ctx) error {
	// 현재 로그인한 관리자 확인
	user := c.Locals("user").(*models.User)
	if user.Role != models.RoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "카테고리 수정 권한이 없습니다.",
		})
	}

	// 카테고리 ID 가져오기
	categoryID, err := strconv.ParseInt(c.Params("categoryID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "카테고리 ID가 유효하지 않습니다.",
		})
	}

	// 카테고리 조회
	category, err := h.categoryService.GetCategoryByID(c.Context(), categoryID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "카테고리를 찾을 수 없습니다.",
		})
	}

	// 폼 데이터 가져오기
	name := c.FormValue("name")
	description := c.FormValue("description")
	categorySlug := c.FormValue("slug")
	parentIDStr := c.FormValue("parent_id")
	sortOrderStr := c.FormValue("sort_order")
	active := c.FormValue("active") == "on"

	// 필수 필드 검증
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "이름을 입력해주세요.",
		})
	}

	// 슬러그 업데이트
	if categorySlug != "" && categorySlug != category.Slug {
		category.Slug = categorySlug
	}

	// 부모 카테고리 ID 변환
	var parentID *int64
	if parentIDStr != "" && parentIDStr != "0" {
		id, err := strconv.ParseInt(parentIDStr, 10, 64)
		if err == nil && id > 0 && id != categoryID { // 자기 자신을 부모로 설정할 수 없음
			parentID = &id
		}
	}

	// 정렬 순서 변환
	sortOrder := category.SortOrder
	if sortOrderStr != "" {
		if order, err := strconv.Atoi(sortOrderStr); err == nil {
			sortOrder = order
		}
	}

	// 카테고리 객체 업데이트
	category.Name = name
	category.Description = description
	category.ParentID = parentID
	category.SortOrder = sortOrder
	category.Active = active
	category.UpdatedAt = time.Now()

	// 카테고리 업데이트
	if err := h.categoryService.UpdateCategory(c.Context(), category); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "카테고리 수정 중 오류가 발생했습니다: " + err.Error(),
		})
	}

	// JSON 요청인 경우
	if c.Get("Accept") == "application/json" {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "카테고리가 수정되었습니다.",
		})
	}

	// 웹 요청인 경우 카테고리 목록으로 리다이렉트 (수정: 이 부분에서 문제가 있을 수 있음)
	return c.Redirect("/admin/categories")
}

// DeleteCategory 카테고리 삭제 처리 핸들러
func (h *CategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	// 현재 로그인한 관리자 확인
	user := c.Locals("user").(*models.User)
	if user.Role != models.RoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "카테고리 삭제 권한이 없습니다.",
		})
	}

	// 카테고리 ID 가져오기
	categoryID, err := strconv.ParseInt(c.Params("categoryID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "카테고리 ID가 유효하지 않습니다.",
		})
	}

	// 카테고리 삭제
	if err := h.categoryService.DeleteCategory(c.Context(), categoryID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "카테고리 삭제 중 오류가 발생했습니다: " + err.Error(),
		})
	}

	// JSON 요청인 경우
	if c.Get("Accept") == "application/json" {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "카테고리가 삭제되었습니다.",
		})
	}

	// 웹 요청인 경우 카테고리 목록으로 리다이렉트
	return c.Redirect("/admin/categories")
}

// API 핸들러 - 카테고리 항목 관리

// AddItemToCategory 카테고리에 항목 추가 처리 핸들러
func (h *CategoryHandler) AddItemToCategory(c *fiber.Ctx) error {
	// 현재 로그인한 관리자 확인
	user := c.Locals("user").(*models.User)
	if user.Role != models.RoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "카테고리 항목 추가 권한이 없습니다.",
		})
	}

	// 카테고리 ID 가져오기
	categoryID, err := strconv.ParseInt(c.Params("categoryID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "카테고리 ID가 유효하지 않습니다.",
		})
	}

	// JSON 데이터 파싱
	var data struct {
		ItemID   int64  `json:"itemId"`
		ItemType string `json:"itemType"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "요청 데이터가 유효하지 않습니다.",
		})
	}

	// 항목 추가
	if err := h.categoryService.AddItemToCategory(c.Context(), categoryID, data.ItemID, data.ItemType); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "항목 추가 중 오류가 발생했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "항목이 추가되었습니다.",
	})
}

// RemoveItemFromCategory 카테고리에서 항목 제거 처리 핸들러
func (h *CategoryHandler) RemoveItemFromCategory(c *fiber.Ctx) error {
	// 현재 로그인한 관리자 확인
	user := c.Locals("user").(*models.User)
	if user.Role != models.RoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "카테고리 항목 제거 권한이 없습니다.",
		})
	}

	// 카테고리 ID 가져오기
	categoryID, err := strconv.ParseInt(c.Params("categoryID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "카테고리 ID가 유효하지 않습니다.",
		})
	}

	// 항목 ID와 타입 가져오기
	itemID, err := strconv.ParseInt(c.Params("itemID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "항목 ID가 유효하지 않습니다.",
		})
	}

	itemType := c.Params("itemType")
	if itemType != "board" && itemType != "page" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "항목 타입이 유효하지 않습니다.",
		})
	}

	// 항목 제거
	if err := h.categoryService.RemoveItemFromCategory(c.Context(), categoryID, itemID, itemType); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "항목 제거 중 오류가 발생했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "항목이 제거되었습니다.",
	})
}

// GetCategoryMenuStructure 메뉴 구조 조회 API 핸들러
func (h *CategoryHandler) GetCategoryMenuStructure(c *fiber.Ctx) error {
	// 메뉴 구조 조회 - 최상위 카테고리만 표시하도록 수정
	menuStructure, err := h.categoryService.GetMenuStructure(c.Context(), true)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "메뉴 구조를 불러오는데 실패했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    menuStructure,
	})
}

// ListCategoriesAPI 모든 카테고리 조회 API
func (h *CategoryHandler) ListCategoriesAPI(c *fiber.Ctx) error {
	// 활성 카테고리만 조회할지 여부
	onlyActive := c.Query("active") == "true"

	// 특정 부모 카테고리의 하위 카테고리만 조회할지 여부
	var parentID *int64
	if parentIDStr := c.Query("parent_id"); parentIDStr != "" {
		id, err := strconv.ParseInt(parentIDStr, 10, 64)
		if err == nil {
			parentID = &id
		}
	}

	// 카테고리 목록 조회
	categories, err := h.categoryService.ListCategories(c.Context(), onlyActive, parentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "카테고리 목록을 불러오는데 실패했습니다: " + err.Error(),
		})
	}

	// 간소화된 카테고리 데이터 구성
	categoryData := make([]map[string]interface{}, 0, len(categories))
	for _, category := range categories {
		categoryData = append(categoryData, map[string]interface{}{
			"id":          category.ID,
			"name":        category.Name,
			"slug":        category.Slug,
			"description": category.Description,
			"parentId":    category.ParentID,
			"active":      category.Active,
			"sortOrder":   category.SortOrder,
		})
	}

	return c.JSON(fiber.Map{
		"success":    true,
		"categories": categoryData,
	})
}
