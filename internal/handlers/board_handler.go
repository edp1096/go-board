// internal/handlers/board_handler.go

package handlers

import (
	"fmt"
	"go-board/internal/models"
	"go-board/internal/service"
	"go-board/internal/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

type BoardHandler struct {
	boardService service.BoardService
}

func NewBoardHandler(boardService service.BoardService) *BoardHandler {
	return &BoardHandler{
		boardService: boardService,
	}
}

// ListBoards 게시판 목록 조회
func (h *BoardHandler) ListBoards(c *fiber.Ctx) error {
	// 활성 게시판만 조회
	boards, err := h.boardService.ListBoards(c.Context(), true)
	if err != nil {
		fmt.Printf("게시판 목록 조회 오류: %v\n", err)
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "게시판 목록을 불러오는데 실패했습니다",
			"error":   err.Error(),
		})
	}

	// fmt.Printf("게시판 목록 조회 성공: %d개 게시판\n", len(boards))

	// 템플릿 파일이 존재하는지 직접 확인
	templatesFS := c.App().Config().Views.(*html.Engine)
	if templatesFS == nil {
		fmt.Println("템플릿 엔진이 nil입니다!")
	}

	// // 디버그용 템플릿 정보 출력
	// fmt.Printf("렌더링할 템플릿: %s\n", "board/list")

	return utils.RenderWithUser(c, "board/list", fiber.Map{
		"title":  "게시판 목록",
		"boards": boards,
	})
}

// GetBoard 게시판 상세 조회
func (h *BoardHandler) GetBoard(c *fiber.Ctx) error {
	boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "잘못된 게시판 ID입니다",
			"error":   err.Error(),
		})
	}

	// board, err := h.boardService.GetBoardByID(c.Context(), boardID)
	_, err = h.boardService.GetBoardByID(c.Context(), boardID)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "게시판을 찾을 수 없습니다",
			"error":   err.Error(),
		})
	}

	return c.Redirect("/boards/" + strconv.FormatInt(boardID, 10) + "/posts")
}

// ListPosts 게시물 목록 조회
func (h *BoardHandler) ListPosts(c *fiber.Ctx) error {
	boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "잘못된 게시판 ID입니다",
			"error":   err.Error(),
		})
	}

	// 게시판 정보 조회
	board, err := h.boardService.GetBoardByID(c.Context(), boardID)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "게시판을 찾을 수 없습니다",
			"error":   err.Error(),
		})
	}

	// 페이지네이션 파라미터
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	pageSize := 20 // 페이지당 게시물 수

	// 정렬 파라미터
	sortField := c.Query("sort", "created_at")
	sortDir := c.Query("dir", "desc")

	// 검색 파라미터
	query := c.Query("q")

	var posts []*models.DynamicPost
	var total int

	// 검색 또는 일반 목록 조회
	if query != "" {
		posts, total, err = h.boardService.SearchPosts(c.Context(), boardID, query, page, pageSize)
	} else {
		posts, total, err = h.boardService.ListPosts(c.Context(), boardID, page, pageSize, sortField, sortDir)
	}

	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "게시물 목록을 불러오는데 실패했습니다",
			"error":   err.Error(),
		})
	}

	// 페이지네이션 계산
	totalPages := (total + pageSize - 1) / pageSize

	return utils.RenderWithUser(c, "board/posts", fiber.Map{
		"title":      board.Name,
		"board":      board,
		"posts":      posts,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
		"total":      total,
		"sortField":  sortField,
		"sortDir":    sortDir,
		"query":      query,
	})
}

// GetPost 게시물 상세 조회
func (h *BoardHandler) GetPost(c *fiber.Ctx) error {
	boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "잘못된 게시판 ID입니다",
			"error":   err.Error(),
		})
	}

	postID, err := strconv.ParseInt(c.Params("postID"), 10, 64)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "잘못된 게시물 ID입니다",
			"error":   err.Error(),
		})
	}

	// 게시판 정보 조회
	board, err := h.boardService.GetBoardByID(c.Context(), boardID)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "게시판을 찾을 수 없습니다",
			"error":   err.Error(),
		})
	}

	// 게시물 정보 조회
	post, err := h.boardService.GetPost(c.Context(), boardID, postID)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "게시물을 찾을 수 없습니다",
			"error":   err.Error(),
		})
	}

	return utils.RenderWithUser(c, "board/view", fiber.Map{
		"title": post.Title,
		"board": board,
		"post":  post,
	})
}

// CreatePostPage 게시물 작성 페이지 렌더링
func (h *BoardHandler) CreatePostPage(c *fiber.Ctx) error {
	boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "잘못된 게시판 ID입니다",
			"error":   err.Error(),
		})
	}

	// 게시판 정보 조회
	board, err := h.boardService.GetBoardByID(c.Context(), boardID)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "게시판을 찾을 수 없습니다",
			"error":   err.Error(),
		})
	}

	return utils.RenderWithUser(c, "board/create", fiber.Map{
		"title": "게시물 작성",
		"board": board,
	})
}

// CreatePost 게시물 작성 처리
func (h *BoardHandler) CreatePost(c *fiber.Ctx) error {
	boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 게시판 ID입니다",
		})
	}

	// 게시판 정보 조회
	board, err := h.boardService.GetBoardByID(c.Context(), boardID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "게시판을 찾을 수 없습니다",
		})
	}

	// 현재 로그인한 사용자 가져오기
	user := c.Locals("user").(*models.User)

	// 기본 필드 가져오기
	title := c.FormValue("title")
	content := c.FormValue("content")

	if title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "제목을 입력해주세요",
		})
	}

	// 동적 게시물 객체 생성
	post := &models.DynamicPost{
		Title:   title,
		Content: content,
		UserID:  user.ID,
		Fields:  make(map[string]models.DynamicField),
	}

	// 동적 필드 처리
	for _, field := range board.Fields {
		// 필드값 가져오기
		value := c.FormValue(field.Name)

		// 필수 필드 검증
		if field.Required && value == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": field.DisplayName + "을(를) 입력해주세요",
				"field":   field.Name,
			})
		}

		// 필드 값 변환 및 검증
		var fieldValue any = value

		switch field.FieldType {
		case models.FieldTypeNumber:
			if value != "" {
				num, err := strconv.Atoi(value)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"success": false,
						"message": field.DisplayName + "은(는) 숫자여야 합니다",
						"field":   field.Name,
					})
				}
				fieldValue = num
			} else {
				fieldValue = nil
			}
		case models.FieldTypeCheckbox:
			fieldValue = value == "on" || value == "true" || value == "1"
		}

		// 동적 필드 추가
		post.Fields[field.Name] = models.DynamicField{
			Name:       field.Name,
			ColumnName: field.ColumnName,
			Value:      fieldValue,
			FieldType:  field.FieldType,
			Required:   field.Required,
		}
	}

	// 게시물 생성
	err = h.boardService.CreatePost(c.Context(), boardID, post)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "게시물 작성에 실패했습니다: " + err.Error(),
		})
	}

	// JSON 요청인 경우
	if c.Get("Accept") == "application/json" {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "게시물이 작성되었습니다",
			"id":      post.ID,
		})
	}

	// 웹 요청인 경우 게시물 상세 페이지로 리다이렉트
	return c.Redirect("/boards/" + strconv.FormatInt(boardID, 10) + "/posts/" + strconv.FormatInt(post.ID, 10))
}

// EditPostPage 게시물 수정 페이지 렌더링
func (h *BoardHandler) EditPostPage(c *fiber.Ctx) error {
	boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "잘못된 게시판 ID입니다",
			"error":   err.Error(),
		})
	}

	postID, err := strconv.ParseInt(c.Params("postID"), 10, 64)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "잘못된 게시물 ID입니다",
			"error":   err.Error(),
		})
	}

	// 게시판 정보 조회
	board, err := h.boardService.GetBoardByID(c.Context(), boardID)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "게시판을 찾을 수 없습니다",
			"error":   err.Error(),
		})
	}

	// 게시물 정보 조회
	post, err := h.boardService.GetPost(c.Context(), boardID, postID)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "게시물을 찾을 수 없습니다",
			"error":   err.Error(),
		})
	}

	// 현재 로그인한 사용자 가져오기
	user := c.Locals("user").(*models.User)

	// 본인 게시물 또는 관리자만 수정 가능
	if user.ID != post.UserID && user.Role != models.RoleAdmin {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "게시물을 수정할 권한이 없습니다",
		})
	}

	return utils.RenderWithUser(c, "board/edit", fiber.Map{
		"title": "게시물 수정",
		"board": board,
		"post":  post,
	})
}

// UpdatePost 게시물 수정 처리
func (h *BoardHandler) UpdatePost(c *fiber.Ctx) error {
	boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 게시판 ID입니다",
		})
	}

	postID, err := strconv.ParseInt(c.Params("postID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 게시물 ID입니다",
		})
	}

	// 게시판 정보 조회
	board, err := h.boardService.GetBoardByID(c.Context(), boardID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "게시판을 찾을 수 없습니다",
		})
	}

	// 게시물 정보 조회
	post, err := h.boardService.GetPost(c.Context(), boardID, postID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "게시물을 찾을 수 없습니다",
		})
	}

	// 현재 로그인한 사용자 가져오기
	user := c.Locals("user").(*models.User)

	// 본인 게시물 또는 관리자만 수정 가능
	if user.ID != post.UserID && user.Role != models.RoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "게시물을 수정할 권한이 없습니다",
		})
	}

	// 기본 필드 가져오기
	title := c.FormValue("title")
	content := c.FormValue("content")

	if title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "제목을 입력해주세요",
		})
	}

	// 기본 필드 업데이트
	post.Title = title
	post.Content = content

	// 동적 필드 처리
	for _, field := range board.Fields {
		// 필드값 가져오기
		value := c.FormValue(field.Name)

		// 필수 필드 검증
		if field.Required && value == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": field.DisplayName + "을(를) 입력해주세요",
				"field":   field.Name,
			})
		}

		// 필드 값 변환 및 검증
		var fieldValue any = value

		switch field.FieldType {
		case models.FieldTypeNumber:
			if value != "" {
				num, err := strconv.Atoi(value)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"success": false,
						"message": field.DisplayName + "은(는) 숫자여야 합니다",
						"field":   field.Name,
					})
				}
				fieldValue = num
			} else {
				fieldValue = nil
			}
		case models.FieldTypeCheckbox:
			fieldValue = value == "on" || value == "true" || value == "1"
		}

		// 동적 필드 업데이트
		post.Fields[field.Name] = models.DynamicField{
			Name:       field.Name,
			ColumnName: field.ColumnName,
			Value:      fieldValue,
			FieldType:  field.FieldType,
			Required:   field.Required,
		}
	}

	// 게시물 업데이트
	err = h.boardService.UpdatePost(c.Context(), boardID, post)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "게시물 수정에 실패했습니다: " + err.Error(),
		})
	}

	// JSON 요청인 경우
	if c.Get("Accept") == "application/json" {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "게시물이 수정되었습니다",
		})
	}

	// 웹 요청인 경우 게시물 상세 페이지로 리다이렉트
	return c.Redirect("/boards/" + strconv.FormatInt(boardID, 10) + "/posts/" + strconv.FormatInt(postID, 10))
}

// DeletePost 게시물 삭제 처리
func (h *BoardHandler) DeletePost(c *fiber.Ctx) error {
	boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 게시판 ID입니다",
		})
	}

	postID, err := strconv.ParseInt(c.Params("postID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 게시물 ID입니다",
		})
	}

	// 게시물 정보 조회
	post, err := h.boardService.GetPost(c.Context(), boardID, postID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "게시물을 찾을 수 없습니다",
		})
	}

	// 현재 로그인한 사용자 가져오기
	user := c.Locals("user").(*models.User)

	// 본인 게시물 또는 관리자만 삭제 가능
	if user.ID != post.UserID && user.Role != models.RoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "게시물을 삭제할 권한이 없습니다",
		})
	}

	// 게시물 삭제
	err = h.boardService.DeletePost(c.Context(), boardID, postID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "게시물 삭제에 실패했습니다: " + err.Error(),
		})
	}

	// JSON 요청인 경우
	if c.Get("Accept") == "application/json" {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "게시물이 삭제되었습니다",
		})
	}

	// 웹 요청인 경우 게시판 목록 페이지로 리다이렉트
	return c.Redirect("/boards/" + strconv.FormatInt(boardID, 10) + "/posts")
}
