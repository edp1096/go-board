// internal/middleware/setup_middleware.go
package middleware

import (
	"strings"

	"github.com/edp1096/toy-board/internal/service"

	"github.com/gofiber/fiber/v2"
)

// InitializeMiddleware는 시스템 초기 설정이 필요한 경우 설정 페이지로 리다이렉트합니다
func InitializeMiddleware(setupService service.SetupService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 이미 설정 페이지인 경우 처리 진행
		if strings.HasPrefix(c.Path(), "/admin/setup") {
			return c.Next()
		}

		// 정적 리소스 요청은 허용
		if strings.HasPrefix(c.Path(), "/static") ||
			strings.HasPrefix(c.Path(), "/uploads") {
			return c.Next()
		}

		// 관리자 계정이 있는지 확인
		adminExists, err := setupService.IsAdminExists(c.Context())
		if err != nil {
			// 오류 발생 시 에러 페이지로 리다이렉트
			return c.Status(fiber.StatusInternalServerError).Render("error", fiber.Map{
				"title":   "오류",
				"message": "시스템 설정 확인 중 오류가 발생했습니다",
				"error":   err.Error(),
			})
		}

		// 관리자 계정이 없으면 설정 페이지로 리다이렉트
		if !adminExists {
			return c.Redirect("/admin/setup")
		}

		// 관리자 계정이 있으면 정상 처리 진행
		return c.Next()
	}
}
