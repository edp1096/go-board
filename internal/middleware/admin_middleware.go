// internal/middleware/admin_middleware.go
package middleware

import (
	"strings"

	"github.com/edp1096/toy-board/internal/models"
	"github.com/edp1096/toy-board/internal/service"

	"github.com/gofiber/fiber/v2"
)

type AdminMiddleware interface {
	RequireAdmin(c *fiber.Ctx) error
}

type adminMiddleware struct {
	authService service.AuthService
}

func NewAdminMiddleware(authService service.AuthService) AdminMiddleware {
	return &adminMiddleware{
		authService: authService,
	}
}

func (m *adminMiddleware) RequireAdmin(c *fiber.Ctx) error {
	// 사용자 정보 가져오기
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		// API 요청인 경우 401 응답
		if strings.HasPrefix(c.Path(), "/api/") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "인증이 필요합니다",
			})
		}

		// 웹 페이지 요청인 경우 로그인 페이지로 리다이렉트
		return c.Redirect("/auth/login?redirect=" + c.Path())
	}

	// 관리자 권한 확인
	if user.Role != models.RoleAdmin {
		// API 요청인 경우 403 응답
		if strings.HasPrefix(c.Path(), "/api/") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "관리자 권한이 필요합니다",
			})
		}

		// 웹 페이지 요청인 경우 오류 페이지 렌더링
		return c.Status(fiber.StatusForbidden).Render("error", fiber.Map{
			"message": "관리자 권한이 필요합니다",
			"error":   "접근 권한이 없습니다",
		})
	}

	return c.Next()
}
