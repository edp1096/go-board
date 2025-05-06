// internal/handlers/page_handler.go
package handlers

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/edp1096/go-board/config"
	"github.com/edp1096/go-board/internal/models"
	"github.com/edp1096/go-board/internal/service"
	"github.com/edp1096/go-board/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gosimple/slug"
)

type PageHandler struct {
	pageService service.PageService
	config      *config.Config
}

func NewPageHandler(pageService service.PageService, cfg *config.Config) *PageHandler {
	return &PageHandler{
		pageService: pageService,
		config:      cfg,
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

// CreatePageScreen 페이지 생성 폼 핸들러
func (h *PageHandler) CreatePageScreen(c *fiber.Ctx) error {
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
	sessionId := c.FormValue("editorSessionId", "") // 에디터 세션 ID

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

	// 세션 ID가 있으면 임시 이미지 처리
	if sessionId != "" {
		if err := h.MoveSessionPageImages(c.Context(), page.ID, sessionId, &content); err != nil {
			// 이미지 이동 오류는 로깅만 하고 계속 진행
			fmt.Printf("임시 이미지 이동 중 오류: %v\n", err)
		} else if content != page.Content {
			// 내용이 변경되었으면 페이지 내용 업데이트
			page.Content = content
			if err := h.pageService.UpdatePage(c.Context(), page); err != nil {
				fmt.Printf("페이지 내용 업데이트 중 오류: %v\n", err)
			}
		}
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

// MoveSessionPageImages는 세션별 임시 폴더의 이미지를 페이지 ID 폴더로 이동시키고 HTML 내용의 경로를 업데이트합니다
func (h *PageHandler) MoveSessionPageImages(ctx context.Context, pageID int64, sessionId string, content *string) error {
	if content == nil || *content == "" || sessionId == "" {
		return nil
	}

	// 임시 경로 패턴 (예: "/uploads/pages/temp/SESSION_ID/images/")
	tempPattern := fmt.Sprintf("/uploads/pages/temp/%s/images/", sessionId)

	// 새 경로 패턴 (예: "/uploads/pages/123/images/")
	newPattern := fmt.Sprintf("/uploads/pages/%d/images/", pageID)

	// 새 디렉토리 생성
	newDir := filepath.Join(h.config.UploadDir, "pages", strconv.FormatInt(pageID, 10), "images")
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return fmt.Errorf("페이지 이미지 디렉토리 생성 실패: %w", err)
	}

	// 임시 디렉토리 경로
	tempDir := filepath.Join(h.config.UploadDir, "pages", "temp", sessionId, "images")

	// 임시 디렉토리가 존재하는지 확인
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		// 임시 디렉토리가 없으면 이동할 이미지도 없음
		return nil
	}

	// 경로 매핑 생성 (URL 업데이트용)
	pathMapping := make(map[string]string)

	// 임시 디렉토리를 나중에 정리하기 위해 기록
	tempDirsToCleanup := []string{}

	// 1. 메인 이미지 디렉토리 처리
	files, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("임시 디렉토리 읽기 실패: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue // 디렉토리는 건너뛰기
		}

		fileName := file.Name()
		oldPath := filepath.Join(tempDir, fileName)
		newPath := filepath.Join(newDir, fileName)

		// 파일 복사 (이동 대신) - 오류가 나도 진행
		err := CopyFile(oldPath, newPath)
		if err != nil {
			fmt.Printf("파일 복사 중 오류: %v\n", err)
			continue
		}

		// URL 경로 매핑 추가
		oldURL := tempPattern + fileName
		newURL := newPattern + fileName
		pathMapping[oldURL] = newURL

		// 나중에 정리할 디렉토리 기록
		tempDirsToCleanup = append(tempDirsToCleanup, tempDir)
	}

	// 2. 썸네일 디렉토리 처리
	thumbsDir := filepath.Join(tempDir, "thumbs")
	if _, err := os.Stat(thumbsDir); err == nil {
		newThumbsDir := filepath.Join(newDir, "thumbs")
		if err := os.MkdirAll(newThumbsDir, 0755); err != nil {
			fmt.Printf("썸네일 디렉토리 생성 실패: %v\n", err)
		} else {
			// 썸네일 파일들 복사
			thumbFiles, err := os.ReadDir(thumbsDir)
			if err != nil {
				fmt.Printf("썸네일 디렉토리 읽기 실패: %v\n", err)
			} else {
				for _, thumbFile := range thumbFiles {
					if thumbFile.IsDir() {
						continue
					}

					thumbName := thumbFile.Name()
					oldThumbPath := filepath.Join(thumbsDir, thumbName)
					newThumbPath := filepath.Join(newThumbsDir, thumbName)

					// 썸네일 파일 복사 (이동 대신)
					err := CopyFile(oldThumbPath, newThumbPath)
					if err != nil {
						fmt.Printf("썸네일 복사 중 오류: %v\n", err)
						continue
					}

					// 썸네일 URL 경로 매핑 추가
					oldThumbURL := tempPattern + "thumbs/" + thumbName
					newThumbURL := newPattern + "thumbs/" + thumbName
					pathMapping[oldThumbURL] = newThumbURL
				}

				// 나중에 정리할 디렉토리 기록
				tempDirsToCleanup = append(tempDirsToCleanup, thumbsDir)
			}
		}
	}

	// 3. HTML 내용의 이미지 경로 업데이트
	newContent := *content
	for oldURL, newURL := range pathMapping {
		newContent = strings.Replace(newContent, oldURL, newURL, -1)
	}

	// 변경된 내용 저장
	*content = newContent

	// 4. 임시 디렉토리 정리를 위한 고루틴 시작 (비동기로 처리)
	go func(sessionId string, dirs []string) {
		// 10초 후에 정리 시도 (브라우저가 파일 참조를 해제할 시간을 줌)
		time.Sleep(10 * time.Second)

		// 1. 먼저 개별 파일 삭제 시도
		for _, dir := range dirs {
			files, err := os.ReadDir(dir)
			if err != nil {
				continue // 디렉토리 접근 오류 무시
			}

			for _, file := range files {
				if !file.IsDir() {
					// 파일만 삭제
					filePath := filepath.Join(dir, file.Name())
					os.Remove(filePath) // 오류 무시
				}
			}
		}

		// 2. 디렉토리들을 깊이 기준으로 정렬 (더 깊은 경로가 먼저 삭제되도록)
		sort.Slice(dirs, func(i, j int) bool {
			return len(dirs[i]) > len(dirs[j])
		})

		// 3. 첫 번째 시도: 정렬된 순서대로 디렉토리 삭제
		for _, dir := range dirs {
			os.Remove(dir) // 오류 무시
		}

		// 4. 세션 ID 폴더와 그 부모 디렉토리까지 삭제 시도
		tempSessionDir := filepath.Join(h.config.UploadDir, "pages", "temp", sessionId)
		os.Remove(filepath.Join(tempSessionDir, "images", "thumbs")) // 썸네일 디렉토리
		os.Remove(filepath.Join(tempSessionDir, "images"))           // 이미지 디렉토리
		os.Remove(tempSessionDir)                                    // 세션 디렉토리

		// 5. 두 번째 시도 (3초 후): 더 강력한 방법으로 삭제
		time.Sleep(3 * time.Second)

		// 세션 폴더를 강제로 재귀적 삭제 시도
		err := filepath.Walk(tempSessionDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // 오류 무시하고 계속 진행
			}

			// 디렉토리는 건너뛰고 파일만 처리 (첫 번째 패스)
			if !info.IsDir() {
				os.Remove(path) // 오류 무시
			}
			return nil
		})

		if err == nil {
			// 파일이 삭제된 후, 디렉토리 삭제 시도 (깊이 우선으로)
			var dirsToRemove []string

			filepath.Walk(tempSessionDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil // 오류 무시
				}

				if info.IsDir() && path != tempSessionDir {
					dirsToRemove = append(dirsToRemove, path)
				}
				return nil
			})

			// 디렉토리를 깊이 기준으로 정렬
			sort.Slice(dirsToRemove, func(i, j int) bool {
				return len(dirsToRemove[i]) > len(dirsToRemove[j])
			})

			// 정렬된 순서로 삭제
			for _, dir := range dirsToRemove {
				os.Remove(dir) // 오류 무시
			}

			// 루트 세션 디렉토리 삭제
			os.Remove(tempSessionDir)
		}

		// 6. 마지막 시도: 전체 temp 디렉토리가 비어있는지 확인하고 삭제 시도
		tempDir := filepath.Join(h.config.UploadDir, "pages", "temp")

		entries, err := os.ReadDir(tempDir)
		if err == nil && len(entries) == 0 {
			// temp 디렉토리가 비어있으면 삭제
			os.Remove(tempDir)
		}
	}(sessionId, tempDirsToCleanup)

	return nil
}

// CopyFile은 파일을 소스에서 타겟으로 복사합니다
func CopyFile(source, target string) error {
	// 소스 파일 열기
	s, err := os.Open(source)
	if err != nil {
		return err
	}
	defer s.Close()

	// 타겟 파일 생성
	t, err := os.Create(target)
	if err != nil {
		return err
	}
	defer t.Close()

	// 파일 복사
	_, err = io.Copy(t, s)
	if err != nil {
		return err
	}

	// 성공적으로 복사되면 파일 권한 설정
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	return os.Chmod(target, info.Mode())
}

// MovePageTempImages는 임시 폴더의 이미지를 페이지 ID 폴더로 이동시키고 HTML 내용의 경로를 업데이트합니다
func (h *PageHandler) MovePageTempImages(ctx context.Context, pageID int64, content *string) error {
	if content == nil || *content == "" {
		return nil
	}

	// 임시 경로 패턴 (예: "/uploads/pages/temp/images/")
	tempPattern := "/uploads/pages/temp/images/"

	// 새 경로 패턴 (예: "/uploads/pages/123/images/")
	newPattern := fmt.Sprintf("/uploads/pages/%d/images/", pageID)

	// 새 디렉토리 생성
	newDir := filepath.Join(h.config.UploadDir, "pages", strconv.FormatInt(pageID, 10), "images")
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return fmt.Errorf("페이지 이미지 디렉토리 생성 실패: %w", err)
	}

	// 임시 디렉토리 경로
	tempDir := filepath.Join(h.config.UploadDir, "pages", "temp", "images")

	// 임시 디렉토리가 존재하는지 확인
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		// 임시 디렉토리가 없으면 이동할 이미지도 없음
		return nil
	}

	// 임시 디렉토리의 모든 파일 가져오기
	files, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("임시 디렉토리 읽기 실패: %w", err)
	}

	// 이미지 파일 이동 및 경로 매핑 생성
	pathMapping := make(map[string]string)

	for _, file := range files {
		if file.IsDir() {
			continue // 디렉토리는 건너뛰기
		}

		fileName := file.Name()
		oldPath := filepath.Join(tempDir, fileName)
		newPath := filepath.Join(newDir, fileName)

		// 파일 이동
		if err := os.Rename(oldPath, newPath); err != nil {
			return fmt.Errorf("파일 이동 실패 (%s -> %s): %w", oldPath, newPath, err)
		}

		// URL 경로 매핑 추가
		oldURL := tempPattern + fileName
		newURL := newPattern + fileName
		pathMapping[oldURL] = newURL
	}

	// 서브디렉토리(예: thumbs) 이동
	thumbsDir := filepath.Join(tempDir, "thumbs")
	if _, err := os.Stat(thumbsDir); err == nil {
		newThumbsDir := filepath.Join(newDir, "thumbs")
		if err := os.MkdirAll(newThumbsDir, 0755); err != nil {
			return fmt.Errorf("썸네일 디렉토리 생성 실패: %w", err)
		}

		// 썸네일 파일들 이동
		thumbFiles, err := os.ReadDir(thumbsDir)
		if err != nil {
			return fmt.Errorf("썸네일 디렉토리 읽기 실패: %w", err)
		}

		for _, thumbFile := range thumbFiles {
			if thumbFile.IsDir() {
				continue
			}

			thumbName := thumbFile.Name()
			oldThumbPath := filepath.Join(thumbsDir, thumbName)
			newThumbPath := filepath.Join(newThumbsDir, thumbName)

			// 썸네일 파일 이동
			if err := os.Rename(oldThumbPath, newThumbPath); err != nil {
				return fmt.Errorf("썸네일 이동 실패 (%s -> %s): %w", oldThumbPath, newThumbPath, err)
			}

			// 썸네일 URL 경로 매핑 추가
			oldThumbURL := tempPattern + "thumbs/" + thumbName
			newThumbURL := newPattern + "thumbs/" + thumbName
			pathMapping[oldThumbURL] = newThumbURL
		}
	}

	// HTML 내용의 이미지 경로 업데이트
	newContent := *content
	for oldURL, newURL := range pathMapping {
		newContent = strings.Replace(newContent, oldURL, newURL, -1)
	}

	// 변경된 내용 저장
	*content = newContent

	return nil
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

	// 페이지 삭제 (이 호출에서 페이지 이미지도 함께 삭제됨)
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
