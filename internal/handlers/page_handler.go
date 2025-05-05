// internal/handlers/page_handler.go
package handlers

import (
	"strconv"
	"time"

	"github.com/edp1096/go-board/internal/models"
	"github.com/edp1096/go-board/internal/service"
	"github.com/edp1096/go-board/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gosimple/slug"
)

type PageHandler struct {
	pageService service.PageService
}

func NewPageHandler(pageService service.PageService) *PageHandler {
	return &PageHandler{
		pageService: pageService,
	}
}

// GetPage 페이지 조회 핸들러
func (h *PageHandler) GetPage(c *fiber.Ctx) error {
	// 슬러그로 페이지 조회
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).Render("error", fiber.Map{
			"title":   "잘못된 요청",
			"message": "페이지 식별자가 없습니다.",
		})
	}

	page, err := h.pageService.GetPageBySlug(c.Context(), slug)
	if err != nil {
		return c.Status(fiber.StatusNotFound).Render("error", fiber.Map{
			"title":   "페이지를 찾을 수 없음",
			"message": "요청하신 페이지를 찾을 수 없습니다.",
		})
	}

	// 페이지가 비활성화된 경우
	if !page.Active {
		return c.Status(fiber.StatusNotFound).Render("error", fiber.Map{
			"title":   "페이지를 찾을 수 없음",
			"message": "요청하신 페이지를 찾을 수 없습니다.",
		})
	}

	// 메타 데이터 생성
	metaDescription := utils.TruncateText(page.Content, 150)

	return utils.RenderWithUser(c, "page/view", fiber.Map{
		"title":           page.Title,
		"description":     metaDescription,
		"page":            page,
		"metaTitle":       page.Title,
		"metaDescription": metaDescription,
		"metaURL":         c.BaseURL() + c.Path(),
		"metaSiteName":    "게시판 시스템",
	})
}

// CreatePagePage 페이지 생성 폼 핸들러
func (h *PageHandler) CreatePagePage(c *fiber.Ctx) error {
	return utils.RenderWithUser(c, "page/create", fiber.Map{
		"title": "페이지 생성",
	})
}

// CreatePage 페이지 생성 처리 핸들러
func (h *PageHandler) CreatePage(c *fiber.Ctx) error {
	// 현재 로그인한 관리자 확인
	user := c.Locals("user").(*models.User)
	if user.Role != models.RoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "페이지 생성 권한이 없습니다.",
		})
	}

	// 폼 데이터 가져오기
	title := c.FormValue("title")
	content := c.FormValue("content")
	pageSlug := c.FormValue("slug")
	showInMenu := c.FormValue("show_in_menu") == "on"

	// 필수 필드 검증
	if title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "제목을 입력해주세요.",
		})
	}

	// 슬러그 생성
	if pageSlug == "" {
		pageSlug = slug.Make(title)
	}

	// 페이지 객체 생성
	page := &models.Page{
		Title:      title,
		Content:    content,
		Slug:       pageSlug,
		Active:     true,
		ShowInMenu: showInMenu,
		SortOrder:  0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 페이지 생성
	if err := h.pageService.CreatePage(c.Context(), page); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "페이지 생성 중 오류가 발생했습니다: " + err.Error(),
		})
	}

	// JSON 요청인 경우
	if c.Get("Accept") == "application/json" {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "페이지가 생성되었습니다.",
			"slug":    page.Slug,
		})
	}

	// 웹 요청인 경우 생성된 페이지로 리다이렉트
	return c.Redirect("/page/" + page.Slug)
}

// EditPagePage 페이지 수정 폼 핸들러
func (h *PageHandler) EditPagePage(c *fiber.Ctx) error {
	// 페이지 ID 가져오기
	pageID, err := strconv.ParseInt(c.Params("pageID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).Render("error", fiber.Map{
			"title":   "잘못된 요청",
			"message": "페이지 ID가 유효하지 않습니다.",
		})
	}

	// 페이지 조회
	page, err := h.pageService.GetPageByID(c.Context(), pageID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).Render("error", fiber.Map{
			"title":   "페이지를 찾을 수 없음",
			"message": "요청하신 페이지를 찾을 수 없습니다.",
		})
	}

	return utils.RenderWithUser(c, "page/edit", fiber.Map{
		"title": "페이지 수정",
		"page":  page,
	})
}

// UpdatePage 페이지 수정 처리 핸들러
func (h *PageHandler) UpdatePage(c *fiber.Ctx) error {
	// 현재 로그인한 관리자 확인
	user := c.Locals("user").(*models.User)
	if user.Role != models.RoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "페이지 수정 권한이 없습니다.",
		})
	}

	// 페이지 ID 가져오기
	pageID, err := strconv.ParseInt(c.Params("pageID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "페이지 ID가 유효하지 않습니다.",
		})
	}

	// 페이지 조회
	page, err := h.pageService.GetPageByID(c.Context(), pageID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "페이지를 찾을 수 없습니다.",
		})
	}

	// 폼 데이터 가져오기
	title := c.FormValue("title")
	content := c.FormValue("content")
	pageSlug := c.FormValue("slug")
	showInMenu := c.FormValue("show_in_menu") == "on"
	active := c.FormValue("active") == "on"

	// 필수 필드 검증
	if title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "제목을 입력해주세요.",
		})
	}

	// 슬러그 업데이트
	if pageSlug != "" && pageSlug != page.Slug {
		page.Slug = pageSlug
	}

	// 페이지 객체 업데이트
	page.Title = title
	page.Content = content
	page.ShowInMenu = showInMenu
	page.Active = active
	page.UpdatedAt = time.Now()

	// 페이지 업데이트
	if err := h.pageService.UpdatePage(c.Context(), page); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "페이지 수정 중 오류가 발생했습니다: " + err.Error(),
		})
	}

	// JSON 요청인 경우
	if c.Get("Accept") == "application/json" {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "페이지가 수정되었습니다.",
		})
	}

	// 웹 요청인 경우 수정된 페이지로 리다이렉트 (수정: 이 부분을 페이지 목록으로 리다이렉트로 변경)
	return c.Redirect("/admin/pages")
}

// DeletePage 페이지 삭제 처리 핸들러
func (h *PageHandler) DeletePage(c *fiber.Ctx) error {
	// 현재 로그인한 관리자 확인
	user := c.Locals("user").(*models.User)
	if user.Role != models.RoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "페이지 삭제 권한이 없습니다.",
		})
	}

	// 페이지 ID 가져오기
	pageID, err := strconv.ParseInt(c.Params("pageID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "페이지 ID가 유효하지 않습니다.",
		})
	}

	// 페이지 삭제
	if err := h.pageService.DeletePage(c.Context(), pageID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "페이지 삭제 중 오류가 발생했습니다: " + err.Error(),
		})
	}

	// JSON 요청인 경우
	if c.Get("Accept") == "application/json" {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "페이지가 삭제되었습니다.",
		})
	}

	// 웹 요청인 경우 페이지 목록으로 리다이렉트
	return c.Redirect("/admin/pages")
}

// ListPages 관리자용 페이지 목록 핸들러
func (h *PageHandler) ListPages(c *fiber.Ctx) error {
	// 모든 페이지 조회
	pages, err := h.pageService.ListPages(c.Context(), false)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).Render("error", fiber.Map{
			"title":   "오류",
			"message": "페이지 목록을 불러오는데 실패했습니다: " + err.Error(),
		})
	}

	return utils.RenderWithUser(c, "admin/pages/list", fiber.Map{
		"title": "페이지 관리",
		"pages": pages,
	})
}

// ListPagesAPI 모든 페이지 조회 API
func (h *PageHandler) ListPagesAPI(c *fiber.Ctx) error {
	// 활성 페이지만 조회할지 여부
	onlyActive := c.Query("active") == "true"

	// 페이지 목록 조회
	pages, err := h.pageService.ListPages(c.Context(), onlyActive)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "페이지 목록을 불러오는데 실패했습니다: " + err.Error(),
		})
	}

	// 간소화된 페이지 데이터 구성
	pageData := make([]map[string]interface{}, 0, len(pages))
	for _, page := range pages {
		pageData = append(pageData, map[string]interface{}{
			"id":        page.ID,
			"title":     page.Title,
			"slug":      page.Slug,
			"active":    page.Active,
			"sortOrder": page.SortOrder,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"pages":   pageData,
	})
}
