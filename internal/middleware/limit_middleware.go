// internal/middleware/body_limit_middleware.go
package middleware

import (
	"fmt"
	"strings"

	"github.com/edp1096/go-board/config"
	"github.com/edp1096/go-board/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// 업로드 경로를 확인하는 헬퍼 함수
func isUploadPath(path string) bool {
	// API 업로드 경로
	if strings.Contains(path, "/api/boards") && strings.Contains(path, "/upload") {
		return true
	}
	if strings.Contains(path, "/api/boards") && strings.Contains(path, "/attachments") {
		return true
	}

	// 게시물 작성/수정 경로
	if strings.Contains(path, "/boards") && strings.Contains(path, "/posts") {
		if strings.HasSuffix(path, "/create") || strings.Contains(path, "/edit") {
			return true
		}
		// POST/PUT 메서드 체크는 미들웨어에서 수행
	}

	return false
}

// BodyLimitMiddleware는 경로에 따라 다른 본문 크기 제한을 적용하는 미들웨어입니다.
// StreamRequestBody가 활성화된 상태에서 사용됩니다.
func BodyLimitMiddleware(cfg *config.Config) fiber.Handler {
	// 일반 요청과 업로드 요청의 최대 크기 계산
	regularLimit := int64(cfg.MaxBodyLimit)

	// 업로드 제한은 일반 파일과 이미지 중 더 큰 값
	uploadLimit := max(cfg.MaxImageUploadSize, cfg.MaxUploadSize)

	return func(c *fiber.Ctx) error {
		path := c.Path()
		method := c.Method()
		contentLength := int64(c.Request().Header.ContentLength())
		contentType := c.Get("Content-Type")

		// 업로드 경로 확인
		isUpload := isUploadPath(path)

		// StreamRequestBody를 활용한 크기 제한 적용
		// 1. Content-Length 헤더가 있는 경우 즉시 확인
		if contentLength > 0 {
			var limit int64
			// POST/PUT 메서드에서 multipart 요청인 경우만 업로드로 간주
			if (method == "POST" || method == "PUT") && strings.Contains(contentType, "multipart/form-data") {
				// strings.Contains(contentType, "multipart/form-data") && isUpload {
				limit = uploadLimit
			} else {
				limit = regularLimit
			}

			if contentLength > limit {
				return handleSizeExceeded(c, limit)
			}
		}

		// 2. Content-Length가 없거나 스트리밍 처리 필요시 별도 로직
		// (StreamRequestBody가 true인 경우 Fiber 내부에서 처리)

		// 파일 크기 제한 정보를 컨텍스트에 저장 (핸들러에서 사용)
		if isUpload {
			c.Locals("maxUploadSize", cfg.MaxUploadSize)
			c.Locals("maxImageUploadSize", cfg.MaxImageUploadSize)
		}

		return c.Next()
	}
}

// 크기 초과 시 오류 처리
func handleSizeExceeded(c *fiber.Ctx, limit int64) error {
	// API 요청인 경우 JSON 응답
	if strings.HasPrefix(c.Path(), "/api/") || c.XHR() {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("요청 크기가 허용된 최대 크기(%dMB)를 초과했습니다.", limit/config.BytesPerMB),
			"code":    "request_too_large",
		})
	}

	// 일반 페이지 요청인 경우
	return utils.RenderWithUser(c, "error", fiber.Map{
		"title":   "요청 크기 초과",
		"message": fmt.Sprintf("요청된 크기가 허용된 최대 크기(%dMB)를 초과했습니다.", limit/config.BytesPerMB),
	})
}
