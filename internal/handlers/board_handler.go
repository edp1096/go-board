// internal/handlers/board_handler.go

package handlers

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/edp1096/go-board/config"
	"github.com/edp1096/go-board/internal/models"
	"github.com/edp1096/go-board/internal/service"
	"github.com/edp1096/go-board/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type BoardHandler struct {
	boardService   service.BoardService
	commentService service.CommentService
	uploadService  service.UploadService
	config         *config.Config
}

func NewBoardHandler(
	boardService service.BoardService,
	commentService service.CommentService,
	uploadService service.UploadService,
	cfg *config.Config,
) *BoardHandler {
	return &BoardHandler{
		boardService:   boardService,
		commentService: commentService,
		uploadService:  uploadService,
		config:         cfg,
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

	// 현재 로그인한 사용자 정보
	user := c.Locals("user")
	var userID int64
	var isAdmin bool

	if user != nil {
		userObj := user.(*models.User)
		userID = userObj.ID
		isAdmin = (userObj.Role == models.RoleAdmin)
	}

	// 각 게시판에 대해 참여자 여부 확인
	type BoardWithAccess struct {
		Board         *models.Board
		IsParticipant bool
		IsManager     bool
	}

	boardsWithAccess := make([]BoardWithAccess, 0, len(boards))

	for _, board := range boards {
		boardAccess := BoardWithAccess{
			Board: board,
		}

		if user != nil {
			// 관리자는 모든 게시판에 접근 가능
			if isAdmin {
				boardAccess.IsParticipant = true
				boardAccess.IsManager = true
			} else {
				// 소모임 게시판인 경우 참여자 여부 확인
				if board.BoardType == models.BoardTypeGroup {
					isParticipant, _ := h.boardService.IsParticipant(c.Context(), board.ID, userID)
					boardAccess.IsParticipant = isParticipant

					isManager, _ := h.boardService.IsBoardManager(c.Context(), board.ID, userID)
					boardAccess.IsManager = isManager
				}
			}
		}

		boardsWithAccess = append(boardsWithAccess, boardAccess)
	}

	return utils.RenderWithUser(c, "board/list", fiber.Map{
		"title":            "게시판 목록",
		"boardsWithAccess": boardsWithAccess,
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
	if board.BoardType == models.BoardTypeGallery {
		pageSize = 16 // 갤러리는 한 페이지에 16개 항목으로 조정
	}

	// 정렬 파라미터
	sortField := c.Query("sort", "created_at")
	sortDir := c.Query("dir", "desc")

	// 검색 파라미터
	query := c.Query("q")

	// Q&A 게시판 필터 파라미터
	status := c.Query("status")

	var posts []*models.DynamicPost
	var total int

	// 검색 또는 일반 목록 조회
	if query != "" || (board.BoardType == models.BoardTypeQnA && status != "") {
		// Q&A 게시판인 경우 status 필터를 추가
		if board.BoardType == models.BoardTypeQnA && status != "" {
			posts, total, err = h.boardService.SearchPostsWithStatus(c.Context(), boardID, query, status, page, pageSize)
		} else {
			posts, total, err = h.boardService.SearchPosts(c.Context(), boardID, query, page, pageSize)
		}
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

	// 현재 로그인한 사용자 정보
	user := c.Locals("user")
	var userID int64
	var isAdmin bool
	var isManager bool
	var isModerator bool

	if user != nil {
		userObj := user.(*models.User)
		userID = userObj.ID
		isAdmin = (userObj.Role == models.RoleAdmin)
		isManager, _ = h.boardService.IsBoardManager(c.Context(), boardID, userID)

		// 소모임 게시판인 경우 moderator 여부도 확인
		if board.BoardType == models.BoardTypeGroup {
			isModerator, _ = h.boardService.IsParticipantModerator(c.Context(), boardID, userID)
		}
	}

	// 비밀글 필터링
	filteredPosts := make([]*models.DynamicPost, 0, len(posts))
	for _, post := range posts {
		// 비밀글이 아니면 바로 추가
		if !post.IsPrivate {
			filteredPosts = append(filteredPosts, post)
			continue
		}

		// 비밀글인 경우 접근 권한 확인
		if user != nil && (post.UserID == userID || isAdmin || isManager || isModerator) {
			filteredPosts = append(filteredPosts, post)
		}
	}

	// 비밀글 필터링 후 게시물 목록 갱신
	posts = filteredPosts

	// 페이지네이션 계산
	totalPages := (total + pageSize - 1) / pageSize

	// 게시판 타입이 갤러리인 경우 썸네일 정보 추가
	if board.BoardType == models.BoardTypeGallery {
		// 게시물 ID 목록 수집
		postIDs := make([]int64, 0, len(posts))
		for _, post := range posts {
			postIDs = append(postIDs, post.ID)
		}

		// 썸네일 조회
		thumbnails, err := h.boardService.GetPostThumbnails(c.Context(), boardID, postIDs)
		if err != nil {
			// 썸네일 조회 실패 시 로그만 남기고 계속 진행
			fmt.Printf("썸네일 조회 실패: %v\n", err)
		} else {
			// 각 게시물에 썸네일 URL 추가
			for _, post := range posts {
				if url, ok := thumbnails[post.ID]; ok {
					// 백슬래시를 슬래시로 변환 (Windows 경로 문제 해결)
					post.RawData["ThumbnailURL"] = strings.ReplaceAll(url, "\\", "/")
				}
			}
		}
	}

	// 템플릿 선택 (게시판 타입에 따라)
	templateName := "board/posts"
	if board.BoardType == models.BoardTypeGallery {
		templateName = "board/gallery_posts"
	} else if board.BoardType == models.BoardTypeQnA {
		templateName = "board/qna_posts"
	}

	return utils.RenderWithUser(c, templateName, fiber.Map{
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
		"status":     status,
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

	// 현재 로그인한 사용자 정보
	user := c.Locals("user")
	var userID int64
	isAdmin := false
	isManager := false
	isModerator := false

	// 비밀글 접근 권한 확인
	if post.IsPrivate {
		// 권한이 없는 경우
		hasAccess := false

		if user != nil {
			userObj := user.(*models.User)
			userID = userObj.ID

			// 1. 작성자인 경우
			if userObj.ID == post.UserID {
				hasAccess = true
			} else if userObj.Role == models.RoleAdmin {
				// 2. 관리자인 경우
				hasAccess = true
			} else {
				// 3. 게시판 매니저인 경우
				isManager, _ = h.boardService.IsBoardManager(c.Context(), boardID, userObj.ID)
				if isManager {
					hasAccess = true
				} else if board.BoardType == models.BoardTypeGroup {
					isModerator, _ = h.boardService.IsParticipantModerator(c.Context(), boardID, userObj.ID)
					if isModerator {
						hasAccess = true
					}
				}
			}
		}

		// 접근 권한이 없으면 오류 페이지 표시
		if !hasAccess {
			return utils.RenderWithUser(c, "error", fiber.Map{
				"title":   "접근 제한",
				"message": "비밀글은 작성자와 관리자만 볼 수 있습니다.",
			})
		}
	} else {
		// 사용자가 로그인한 경우 매니저 여부 확인
		if user != nil {
			userObj := user.(*models.User)
			isManager, _ = h.boardService.IsBoardManager(c.Context(), boardID, userObj.ID)

			// 소모임 게시판인 경우 moderator 여부도 확인
			if board.BoardType == models.BoardTypeGroup {
				isModerator, _ = h.boardService.IsParticipantModerator(c.Context(), boardID, userObj.ID)
			}
		}
	}

	// // 사용자가 로그인한 경우 매니저 여부 확인
	// if user != nil {
	// 	userObj := user.(*models.User)
	// 	isManager, _ = h.boardService.IsBoardManager(c.Context(), boardID, userObj.ID)
	// }

	pageSize := 20 // 페이지당 게시물 수
	if board.BoardType == models.BoardTypeGallery {
		pageSize = 16 // 갤러리는 한 페이지에 16개 항목으로 조정
	}
	page := 1                 // 첫 페이지만 표시
	sortField := "created_at" // 생성일 기준
	sortDir := "desc"         // 내림차순 (최신순)

	posts, total, err := h.boardService.ListPosts(c.Context(), boardID, page, pageSize, sortField, sortDir)
	if err != nil {
		// 게시물 목록 조회 실패 시에도 게시물 상세 정보는 표시
		posts = []*models.DynamicPost{}
		total = 0
	}

	// 비밀글 필터링
	filteredPosts := make([]*models.DynamicPost, 0, len(posts))
	for _, post := range posts {
		// 비밀글이 아니면 바로 추가
		if !post.IsPrivate {
			filteredPosts = append(filteredPosts, post)
			continue
		}

		// 비밀글인 경우 접근 권한 확인
		if user != nil && (post.UserID == userID || isAdmin || isManager || isModerator) {
			filteredPosts = append(filteredPosts, post)
		}
	}

	// 비밀글 필터링 후 게시물 목록 갱신
	posts = filteredPosts

	// 게시판 타입이 갤러리인 경우 썸네일 정보 추가
	if board.BoardType == models.BoardTypeGallery && len(posts) > 0 {
		// 게시물 ID 목록 수집
		postIDs := make([]int64, 0, len(posts))
		for _, post := range posts {
			postIDs = append(postIDs, post.ID)
		}

		// 썸네일 조회
		thumbnails, err := h.boardService.GetPostThumbnails(c.Context(), boardID, postIDs)
		if err != nil {
			// 썸네일 조회 실패 시 로그만 남기고 계속 진행
			fmt.Printf("썸네일 조회 실패: %v\n", err)
		} else {
			// 각 게시물에 썸네일 URL 추가
			for _, post := range posts {
				if url, ok := thumbnails[post.ID]; ok {
					// 백슬래시를 슬래시로 변환 (Windows 경로 문제 해결)
					post.RawData["ThumbnailURL"] = strings.ReplaceAll(url, "\\", "/")
				}
			}
		}
	}

	// 페이지네이션 계산
	totalPages := (total + pageSize - 1) / pageSize

	// 템플릿 선택 - gallery_view는 없음
	templateName := "board/view"
	if board.BoardType == models.BoardTypeQnA {
		templateName = "board/qna_view"
	}

	// 메타 데이터 추출
	// metaDescription := ""
	// if len(post.Content) > 150 {
	// 	metaDescription = utils.PlainText(post.Content[:150]) + "..."
	// } else {
	// 	metaDescription = utils.PlainText(post.Content)
	// }
	metaDescription := utils.TruncateText(post.Content, 150)

	// 썸네일 URL 처리
	var thumbnailURL string
	if board.BoardType == models.BoardTypeGallery {
		thumbnails, _ := h.boardService.GetPostThumbnails(c.Context(), boardID, []int64{postID})
		if url, ok := thumbnails[postID]; ok {
			thumbnailURL = url
		}
	}

	// // 서버 주소 가져오기
	// serverURL := "https://" + h.config.ServerAddress
	// if h.config.Environment == "development" {
	// 	serverURL = "http://" + h.config.ServerAddress
	// }

	return utils.RenderWithUser(c, templateName, fiber.Map{
		"title":       post.Title,
		"description": metaDescription,
		"board":       board,
		"post":        post,
		"isManager":   isManager,
		"isModerator": isModerator,
		// SEO 메타 태그 데이터
		"metaAuthor":      post.Username,
		"metaTitle":       post.Title,
		"metaDescription": metaDescription,
		// "metaURL":         serverURL + c.Path(),
		"metaURL":      c.BaseURL() + c.Path(),
		"metaSiteName": "게시판 시스템",
		"metaImage":    thumbnailURL,
		// 게시물 목록 데이터 (페이지 하단에 표시할 목록)
		"posts":      posts,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
		"total":      total,
		"sortField":  sortField,
		"sortDir":    sortDir,
		"query":      "", // 검색어는 없음
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
		"title":                "게시물 작성",
		"board":                board,
		"maxUploadSizeMB":      h.config.MaxUploadSize / config.BytesPerMB,
		"maxImageUploadSizeMB": h.config.MaxImageUploadSize / config.BytesPerMB,
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

	// 폼 데이터 출력
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "폼 데이터를 찾을 수 없습니다",
		})
	}

	// 기본 필드 가져오기
	title := c.FormValue("title")
	content := c.FormValue("content")
	isPrivate := c.FormValue("is_private") == "on"

	if title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "제목을 입력해주세요",
		})
	}

	// 동적 게시물 객체 생성
	post := &models.DynamicPost{
		Title:     title,
		Content:   content,
		UserID:    user.ID,
		IsPrivate: isPrivate,
		Fields:    make(map[string]models.DynamicField),
	}

	// 동적 필드 처리
	for _, field := range board.Fields {
		// 필드값 가져오기 (파일 필드는 별도 처리)
		if field.FieldType == models.FieldTypeFile {
			// 파일 필드는 멀티파트 폼에서 "files" 이름으로 통일해서 처리
			if form != nil && len(form.File["files"]) > 0 {
				// 파일이 있으면 필드에 "파일 있음" 표시만 하고, 실제 파일은 나중에 처리
				post.Fields[field.Name] = models.DynamicField{
					Name:       field.Name,
					ColumnName: field.ColumnName,
					Value:      "파일 있음", // 실제 파일 처리는 게시물 생성 후 수행
					FieldType:  field.FieldType,
					Required:   field.Required,
				}
				continue
			} else if field.Required {
				// 필수 파일 필드인데 파일이 없는 경우
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": field.DisplayName + "을(를) 입력해주세요",
					"field":   field.Name,
				})
			}
		}

		// 다른 타입의 필드 처리
		value := c.FormValue(field.Name)

		// 필수 필드 검증
		if field.Required && value == "" && field.FieldType != models.FieldTypeFile {
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

		// 동적 필드 추가 (파일 필드가 아닌 경우)
		if field.FieldType != models.FieldTypeFile || value != "" {
			post.Fields[field.Name] = models.DynamicField{
				Name:       field.Name,
				ColumnName: field.ColumnName,
				Value:      fieldValue,
				FieldType:  field.FieldType,
				Required:   field.Required,
			}
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

	// 파일 첨부가 있는 경우 처리
	if form != nil && len(form.File["files"]) > 0 {
		files := form.File["files"]

		// 업로드 경로 생성
		uploadPath := filepath.Join("uploads", "boards", strconv.FormatInt(boardID, 10), "posts", strconv.FormatInt(post.ID, 10), "attachments")

		var uploadedFiles []*utils.UploadedFile
		var err error

		// 게시판 타입에 따라 다른 업로드 함수 사용
		if board.BoardType == models.BoardTypeGallery {
			// 갤러리 게시판은 이미지 타입도 허용
			uploadedFiles, err = utils.UploadGalleryFiles(files, uploadPath, h.config.MaxImageUploadSize)
		} else {
			// 일반 게시판은 기존대로 처리
			uploadedFiles, err = utils.UploadAttachments(files, uploadPath, h.config.MaxUploadSize)
		}

		if err != nil {
			// // 실패해도 게시물은 생성되므로 계속 진행
			// fmt.Println("파일 업로드 실패:", err)
		} else if h.uploadService != nil {
			// 데이터베이스에 첨부 파일 정보 저장
			_, err := h.uploadService.SaveAttachments(c.Context(), boardID, post.ID, user.ID, uploadedFiles)
			if err != nil {
				// fmt.Println("첨부 파일 저장 실패:", err)
			}
		}
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

	// 권한 확인 (작성자, 관리자 또는 게시판 매니저만 수정 가능)
	hasPermission := false

	// 1. 작성자인 경우
	if user.ID == post.UserID {
		hasPermission = true
	} else if user.Role == models.RoleAdmin {
		// 2. 전체 관리자인 경우
		hasPermission = true
	} else {
		// 3. 게시판 매니저인지 확인
		isManager, err := h.boardService.IsBoardManager(c.Context(), boardID, user.ID)
		if err == nil && isManager {
			hasPermission = true
		} else if board.BoardType == models.BoardTypeGroup {
			// 4. 소모임 게시판인 경우 moderator 여부 확인
			isModerator, err := h.boardService.IsParticipantModerator(c.Context(), boardID, user.ID)
			if err == nil && isModerator {
				hasPermission = true
			}
		}
	}

	if !hasPermission {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "게시물을 수정할 권한이 없습니다",
		})
	}

	return utils.RenderWithUser(c, "board/edit", fiber.Map{
		"title":                "게시물 수정",
		"board":                board,
		"post":                 post,
		"maxUploadSizeMB":      h.config.MaxUploadSize / config.BytesPerMB,
		"maxImageUploadSizeMB": h.config.MaxImageUploadSize / config.BytesPerMB,
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

	// 권한 확인 (작성자, 관리자 또는 게시판 매니저만 수정 가능)
	hasPermission := false

	// 1. 작성자인 경우
	if user.ID == post.UserID {
		hasPermission = true
	} else if user.Role == models.RoleAdmin {
		// 2. 전체 관리자인 경우
		hasPermission = true
	} else {
		// 3. 게시판 매니저 또는 소모임 moderator인지 확인
		isManager, err := h.boardService.IsBoardManager(c.Context(), boardID, user.ID)
		if err == nil && isManager {
			hasPermission = true
		} else if board.BoardType == models.BoardTypeGroup {
			// 소모임 게시판이면 moderator 여부 확인
			isModerator, err := h.boardService.IsParticipantModerator(c.Context(), boardID, user.ID)
			if err == nil && isModerator {
				hasPermission = true
			}
		}
	}

	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "게시물을 수정할 권한이 없습니다",
		})
	}

	// 기본 필드 가져오기
	title := c.FormValue("title")
	content := c.FormValue("content")
	isPrivate := c.FormValue("is_private") == "on"

	if title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "제목을 입력해주세요",
		})
	}

	// 기본 필드 업데이트
	post.Title = title
	post.Content = content
	post.IsPrivate = isPrivate

	// 삭제할 첨부 파일 목록 가져오기
	deleteAttachments := c.FormValue("delete_attachments[]")
	if deleteAttachments == "" {
		// Fiber가 배열을 다른 방식으로 처리하는지 확인
		deleteAttachments = c.FormValue("delete_attachments")
	}

	var deleteAttachmentIDs []int64
	if deleteAttachments != "" {
		// 쉼표로 구분된 값으로 가정
		attachmentIDStrings := strings.SplitSeq(deleteAttachments, ",")
		for idStr := range attachmentIDStrings {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}

			id, err := strconv.ParseInt(idStr, 10, 64)
			if err == nil {
				deleteAttachmentIDs = append(deleteAttachmentIDs, id)
			}
		}
	}

	// 파일 업로드 확인
	form, err := c.MultipartForm()
	var hasNewFiles bool
	if err == nil && form != nil && len(form.File["files"]) > 0 {
		hasNewFiles = true
	}

	// 갤러리 게시판인 경우 기존 첨부 파일 확인
	var hasExistingAttachments bool
	var existingAttachments []*models.Attachment

	if board.BoardType == models.BoardTypeGallery && h.uploadService != nil {
		existingAttachments, err = h.uploadService.GetAttachmentsByPostID(c.Context(), boardID, postID)
		if err == nil && len(existingAttachments) > 0 {
			// 삭제되지 않을 첨부 파일이 있는지 확인
			for _, attachment := range existingAttachments {
				isMarkedForDeletion := false
				for _, deleteID := range deleteAttachmentIDs {
					if attachment.ID == deleteID {
						isMarkedForDeletion = true
						break
					}
				}

				if !isMarkedForDeletion {
					hasExistingAttachments = true
					break
				}
			}
		}
	}

	// 동적 필드 처리
	for _, field := range board.Fields {
		// 필드값 가져오기
		value := c.FormValue(field.Name)

		// 갤러리 이미지 필드인 경우 특별 처리
		if board.BoardType == models.BoardTypeGallery && field.FieldType == models.FieldTypeFile && field.Required {
			// 새 파일이 업로드되었거나, 기존 이미지가 있고 모두 삭제되지 않는 경우 - 필수 검증 건너뛰기
			if hasNewFiles || hasExistingAttachments {
				// 값만 설정하고 필수 검증 건너뛰기
				var fieldValue any = value
				post.Fields[field.Name] = models.DynamicField{
					Name:       field.Name,
					ColumnName: field.ColumnName,
					Value:      fieldValue,
					FieldType:  field.FieldType,
					Required:   field.Required,
				}
				continue
			}
		}

		// 일반적인 필수 필드 검증
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

	// 삭제할 첨부 파일 처리
	if deleteAttachments != "" {
		for _, id := range deleteAttachmentIDs {
			err = h.uploadService.DeleteAttachment(c.Context(), id)
			if err != nil {
				// 오류 로깅만 하고 계속 진행
				fmt.Printf("첨부 파일 삭제 실패 (ID: %d): %v\n", id, err)
			}
		}
	}

	// 파일 첨부 처리
	if hasNewFiles && form != nil {
		files := form.File["files"]

		// 업로드 경로 생성
		uploadPath := filepath.Join("uploads", "boards", strconv.FormatInt(boardID, 10), "posts", strconv.FormatInt(postID, 10), "attachments")

		var uploadedFiles []*utils.UploadedFile
		var err error

		// 게시판 타입에 따라 다른 업로드 함수 사용
		if board.BoardType == models.BoardTypeGallery {
			// 갤러리 게시판은 이미지 타입도 허용
			uploadedFiles, err = utils.UploadGalleryFiles(files, uploadPath, h.config.MaxImageUploadSize)
		} else {
			// 일반 게시판은 기존대로 처리
			uploadedFiles, err = utils.UploadAttachments(files, uploadPath, h.config.MaxUploadSize)
		}

		if err != nil {
			// // 오류 로깅만 하고 계속 진행
			// fmt.Printf("파일 업로드 실패: %v\n", err)
		} else if h.uploadService != nil {
			// 데이터베이스에 첨부 파일 정보 저장
			_, err := h.uploadService.SaveAttachments(c.Context(), boardID, postID, user.ID, uploadedFiles)
			if err != nil {
				// fmt.Printf("첨부 파일 정보 저장 실패: %v\n", err)
			}
		}
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

	// 권한 확인 (작성자, 관리자 또는 게시판 매니저만 삭제 가능)
	hasPermission := false

	// 1. 작성자인 경우
	if user.ID == post.UserID {
		hasPermission = true
	} else if user.Role == models.RoleAdmin {
		// 2. 전체 관리자인 경우
		hasPermission = true
	} else {
		// 3. 게시판 매니저 또는 소모임 moderator인지 확인
		isManager, err := h.boardService.IsBoardManager(c.Context(), boardID, user.ID)
		if err == nil && isManager {
			hasPermission = true
		} else if board.BoardType == models.BoardTypeGroup {
			// 소모임 게시판이면 moderator 여부 확인
			isModerator, err := h.boardService.IsParticipantModerator(c.Context(), boardID, user.ID)
			if err == nil && isModerator {
				hasPermission = true
			}
		}
	}

	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "게시물을 삭제할 권한이 없습니다",
		})
	}

	// 댓글 삭제 (댓글 서비스가 존재하는 경우)
	if h.commentService != nil {
		err = h.commentService.DeleteCommentsByPostID(c.Context(), boardID, postID)
		if err != nil {
			// 댓글 삭제 오류는 로깅만 하고 진행 (게시물 삭제가 우선)
			log.Printf("게시물 댓글 삭제 실패 (boardID: %d, postID: %d): %v", boardID, postID, err)
		}
	}

	// 첨부 파일 삭제 (uploadService가 있는 경우)
	if h.uploadService != nil {
		err = h.uploadService.DeleteAttachmentsByPostID(c.Context(), boardID, postID)
		if err != nil {
			// 첨부 파일 삭제 오류는 로깅만 하고 진행 (게시물 삭제가 우선)
			log.Printf("게시물 첨부 파일 삭제 실패 (boardID: %d, postID: %d): %v", boardID, postID, err)
		}
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
