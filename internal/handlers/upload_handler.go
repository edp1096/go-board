// internal/handlers/upload_handler.go
package handlers

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/edp1096/toy-board/config"
	"github.com/edp1096/toy-board/internal/models"
	"github.com/edp1096/toy-board/internal/service"
	"github.com/edp1096/toy-board/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type UploadHandler struct {
	uploadService service.UploadService
	boardService  service.BoardService
	config        *config.Config
}

func NewUploadHandler(
	uploadService service.UploadService,
	boardService service.BoardService,
	cfg *config.Config,
) *UploadHandler {
	return &UploadHandler{
		uploadService: uploadService,
		boardService:  boardService,
		config:        cfg,
	}
}

// UploadAttachments는 게시물 첨부 파일을 업로드합니다
func (h *UploadHandler) UploadAttachments(c *fiber.Ctx) error {
	// 게시판 ID 확인
	boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 게시판 ID입니다",
		})
	}

	// 게시물 ID 확인
	postID, err := strconv.ParseInt(c.Params("postID", "0"), 10, 64)
	if err != nil {
		postID = 0 // 임시 저장
	}

	// 현재 사용자 확인
	user := c.Locals("user").(*models.User)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "로그인이 필요합니다",
		})
	}

	// 파일 확인
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "파일 업로드 데이터가 올바르지 않습니다",
		})
	}

	files := form.File["files"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "업로드할 파일이 없습니다",
		})
	}

	// 업로드 경로 생성
	uploadPath := filepath.Join("boards", strconv.FormatInt(boardID, 10), "posts", strconv.FormatInt(postID, 10), "attachments")

	// 파일 업로드
	uploadedFiles, err := utils.UploadAttachments(files, uploadPath, h.config.MaxUploadSize, h.config.UploadDir)
	if err != nil {
		if strings.Contains(err.Error(), "파일 크기가 허용 한도를 초과") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"code":    "file_too_large",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "파일 업로드 실패: " + err.Error(),
		})
	}

	// 데이터베이스에 첨부 파일 정보 저장
	attachments, err := h.uploadService.SaveAttachments(c.Context(), boardID, postID, user.ID, uploadedFiles)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "첨부 파일 정보 저장 실패: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":     true,
		"attachments": attachments,
	})
}

// UploadImages는 에디터용 이미지를 업로드합니다
func (h *UploadHandler) UploadImages(c *fiber.Ctx) error {
	// 게시판 ID 확인
	boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 게시판 ID입니다",
		})
	}

	// 사용자 확인
	user := c.Locals("user").(*models.User)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "로그인이 필요합니다",
		})
	}

	// 파일 확인
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "파일 업로드 데이터가 올바르지 않습니다",
		})
	}

	// 에디터 요구사항에 맞게 필드 이름 검색
	var files []*multipart.FileHeader
	for key, fileHeaders := range form.File {
		// 필드 이름이 upload-files[]인 경우 또는 인덱스가 있는 경우
		if key == "upload-files[]" || strings.HasPrefix(key, "upload-files[") {
			files = append(files, fileHeaders...)
		}
	}

	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "업로드할 이미지가 없습니다",
		})
	}

	// 업로드 경로 생성
	uploadPath := filepath.Join("boards", strconv.FormatInt(boardID, 10), "medias")

	// 이미지 업로드 - 수정된 utils.UploadImages 함수 호출
	uploadedFiles, err := utils.UploadImages(files, uploadPath, h.config.MaxMediaUploadSize, h.config.UploadDir)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "이미지 업로드 실패: " + err.Error(),
		})
	}

	// 에디터 요구사항에 맞는 응답 포맷
	response := make([]map[string]any, 0, len(uploadedFiles))
	for _, file := range uploadedFiles {
		fileResponse := map[string]any{
			"storagename": file.StorageName,
			"thumbnail":   file.ThumbnailURL,
			"url":         file.URL,
		}

		// // WebP 파일인 경우 애니메이션 여부 확인
		// if strings.HasSuffix(strings.ToLower(file.OriginalName), ".webp") {
		// 	isAnimated, _ := utils.IsAnimatedWebP(file.Path)
		// 	fileResponse["animation"] = isAnimated
		// } else {
		// 	fileResponse["animation"] = false
		// }

		response = append(response, fileResponse)
	}

	return c.JSON(fiber.Map{
		"files": response,
	})
}

// UploadImages는 에디터용 이미지와 비디오를 업로드합니다
func (h *UploadHandler) UploadMedias(c *fiber.Ctx) error {
	// 게시판 ID 확인
	boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 게시판 ID입니다",
		})
	}

	// 사용자 확인
	user := c.Locals("user").(*models.User)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "로그인이 필요합니다",
		})
	}

	// 파일 확인
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "파일 업로드 데이터가 올바르지 않습니다",
		})
	}

	// 에디터 요구사항에 맞게 필드 이름 검색
	var files []*multipart.FileHeader
	for key, fileHeaders := range form.File {
		// 필드 이름이 upload-files[]인 경우 또는 인덱스가 있는 경우
		if key == "upload-files[]" || strings.HasPrefix(key, "upload-files[") {
			files = append(files, fileHeaders...)
		}
	}

	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "업로드할 미디어 파일이 없습니다",
		})
	}

	// 업로드 경로 생성
	uploadPath := filepath.Join("boards", strconv.FormatInt(boardID, 10), "medias")

	// 이미지와 비디오 업로드 - UploadMedias 함수 호출
	uploadedFiles, err := utils.UploadMedias(files, uploadPath, h.config.MaxMediaUploadSize, h.config.UploadDir)
	if err != nil {
		if strings.Contains(err.Error(), "파일 크기가 허용 한도를 초과") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"code":    "file_too_large",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "미디어 파일 업로드 실패: " + err.Error(),
		})
	}

	// 에디터 요구사항에 맞는 응답 포맷
	response := make([]map[string]any, 0, len(uploadedFiles))
	for _, file := range uploadedFiles {
		fileResponse := map[string]any{
			"storagename": file.StorageName,
			"thumbnail":   file.ThumbnailURL,
			"url":         file.URL,
			"is_image":    file.IsImage,
			"is_video":    file.IsVideo,
		}

		response = append(response, fileResponse)
	}

	return c.JSON(fiber.Map{
		"files": response,
	})
}

// GetAttachments는 게시물의 첨부 파일 목록을 반환합니다
func (h *UploadHandler) GetAttachments(c *fiber.Ctx) error {
	// 게시판 ID 확인
	boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 게시판 ID입니다",
		})
	}

	// 게시물 ID 확인
	postID, err := strconv.ParseInt(c.Params("postID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 게시물 ID입니다",
		})
	}

	// 첨부 파일 목록 조회
	attachments, err := h.uploadService.GetAttachmentsByPostID(c.Context(), boardID, postID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "첨부 파일 목록 조회 실패: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":     true,
		"attachments": attachments,
	})
}

// DownloadAttachment는 첨부 파일을 다운로드합니다
func (h *UploadHandler) DownloadAttachment(c *fiber.Ctx) error {
	attachmentID, err := strconv.ParseInt(c.Params("attachmentID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 첨부 파일 ID입니다",
		})
	}

	attachment, err := h.uploadService.GetAttachmentByID(c.Context(), attachmentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "첨부 파일을 찾을 수 없습니다",
		})
	}

	h.uploadService.IncrementDownloadCount(c.Context(), attachmentID)

	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, attachment.FileName))
	c.Set("Content-Type", attachment.MimeType)

	return c.SendFile(attachment.FilePath)
}

// DeleteAttachment는 첨부 파일을 삭제합니다
func (h *UploadHandler) DeleteAttachment(c *fiber.Ctx) error {
	// 첨부 파일 ID 확인
	attachmentID, err := strconv.ParseInt(c.Params("attachmentID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 첨부 파일 ID입니다",
		})
	}

	// 현재 사용자 확인
	user := c.Locals("user").(*models.User)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "로그인이 필요합니다",
		})
	}

	// 첨부 파일 정보 조회
	attachment, err := h.uploadService.GetAttachmentByID(c.Context(), attachmentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "첨부 파일을 찾을 수 없습니다",
		})
	}

	// 게시물 정보 조회
	post, err := h.boardService.GetPost(c.Context(), attachment.BoardID, attachment.PostID, false)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "게시물을 찾을 수 없습니다",
		})
	}

	// 권한 확인 (작성자 또는 관리자만 삭제 가능)
	if user.ID != post.UserID && user.Role != models.RoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "첨부 파일을 삭제할 권한이 없습니다",
		})
	}

	// 첨부 파일 삭제
	err = h.uploadService.DeleteAttachment(c.Context(), attachmentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "첨부 파일 삭제 실패: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "첨부 파일이 삭제되었습니다",
	})
}

// 페이지 이미지 업로드 처리
func (h *UploadHandler) UploadPageImages(c *fiber.Ctx) error {
	// 페이지 ID 확인 (없으면 0으로 설정 - 새 페이지 생성 시)
	pageID, err := strconv.ParseInt(c.Params("pageID", "0"), 10, 64)
	if err != nil {
		pageID = 0 // 임시 저장
	}

	// 현재 사용자 확인
	user := c.Locals("user").(*models.User)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "로그인이 필요합니다",
		})
	}

	// 관리자 권한 확인
	if user.Role != models.RoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "권한이 없습니다",
		})
	}

	// 파일 확인
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "파일 업로드 데이터가 올바르지 않습니다",
		})
	}

	// 에디터 요구사항에 맞게 필드 이름 검색
	var files []*multipart.FileHeader
	for key, fileHeaders := range form.File {
		// 필드 이름이 upload-files[]인 경우 또는 인덱스가 있는 경우
		if key == "upload-files[]" || strings.HasPrefix(key, "upload-files[") {
			files = append(files, fileHeaders...)
		}
	}

	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "업로드할 이미지가 없습니다",
		})
	}

	// 세션 ID 가져오기 (없으면 빈 문자열로 설정)
	sessionId := c.Query("sessionId", "")
	if sessionId == "" && pageID == 0 {
		// 세션 ID가 없고 페이지 ID도 없으면 사용자 ID를 대신 사용
		sessionId = strconv.FormatInt(user.ID, 10)
	}

	// 업로드 경로 생성
	var uploadPath string
	if pageID > 0 {
		// 기존 페이지 수정인 경우
		uploadPath = filepath.Join("pages", strconv.FormatInt(pageID, 10), "medias")
	} else {
		// 새 페이지 생성인 경우 - 세션별 임시 디렉토리 사용
		uploadPath = filepath.Join("pages", "temp", sessionId, "medias")
	}

	// 이미지 업로드
	uploadedFiles, err := utils.UploadImages(files, uploadPath, h.config.MaxMediaUploadSize, h.config.UploadDir)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "이미지 업로드 실패: " + err.Error(),
		})
	}

	// 응답 포맷
	response := make([]map[string]any, 0, len(uploadedFiles))
	for _, file := range uploadedFiles {
		fileResponse := map[string]any{
			"storagename": file.StorageName,
			"thumbnail":   file.ThumbnailURL,
			"url":         file.URL,
		}
		response = append(response, fileResponse)
	}

	return c.JSON(fiber.Map{
		"files": response,
	})
}

// 페이지 미디어 파일 업로드 처리
func (h *UploadHandler) UploadPageMedias(c *fiber.Ctx) error {
	// 페이지 ID 확인 (없으면 0으로 설정 - 새 페이지 생성 시)
	pageID, err := strconv.ParseInt(c.Params("pageID", "0"), 10, 64)
	if err != nil {
		pageID = 0 // 임시 저장
	}

	// 현재 사용자 확인
	user := c.Locals("user").(*models.User)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "로그인이 필요합니다",
		})
	}

	// 관리자 권한 확인
	if user.Role != models.RoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "권한이 없습니다",
		})
	}

	// 파일 확인
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "파일 업로드 데이터가 올바르지 않습니다",
		})
	}

	// 에디터 요구사항에 맞게 필드 이름 검색
	var files []*multipart.FileHeader
	for key, fileHeaders := range form.File {
		// 필드 이름이 upload-files[]인 경우 또는 인덱스가 있는 경우
		if key == "upload-files[]" || strings.HasPrefix(key, "upload-files[") {
			files = append(files, fileHeaders...)
		}
	}

	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "업로드할 미디어 파일이 없습니다",
		})
	}

	// 세션 ID 가져오기 (없으면 빈 문자열로 설정)
	sessionId := c.Query("sessionId", "")
	if sessionId == "" && pageID == 0 {
		// 세션 ID가 없고 페이지 ID도 없으면 사용자 ID를 대신 사용
		sessionId = strconv.FormatInt(user.ID, 10)
	}

	// 업로드 경로 생성
	var uploadPath string
	if pageID > 0 {
		// 기존 페이지 수정인 경우
		uploadPath = filepath.Join("pages", strconv.FormatInt(pageID, 10), "medias")
	} else {
		// 새 페이지 생성인 경우 - 세션별 임시 디렉토리 사용
		uploadPath = filepath.Join("pages", "temp", sessionId, "medias")
	}

	// 미디어 파일 업로드
	uploadedFiles, err := utils.UploadMedias(files, uploadPath, h.config.MaxMediaUploadSize, h.config.UploadDir)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "미디어 파일 업로드 실패: " + err.Error(),
		})
	}

	// 응답 포맷
	response := make([]map[string]any, 0, len(uploadedFiles))
	for _, file := range uploadedFiles {
		fileResponse := map[string]any{
			"storagename": file.StorageName,
			"thumbnail":   file.ThumbnailURL,
			"url":         file.URL,
			"is_image":    file.IsImage,
			"is_video":    file.IsVideo,
		}
		response = append(response, fileResponse)
	}

	return c.JSON(fiber.Map{
		"files": response,
	})
}
