// internal/middleware/auth_middleware.go
package middleware

import (
	"dynamic-board/internal/service"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AuthMiddleware interface {
	RequireAuth(c *fiber.Ctx) error
}

type authMiddleware struct {
	authService service.AuthService
}

func NewAuthMiddleware(authService service.AuthService) AuthMiddleware {
	return &authMiddleware{
		authService: authService,
	}
}

func (m *authMiddleware) RequireAuth(c *fiber.Ctx) error {
	// 토큰 가져오기 (쿠키 또는 Authorization 헤더에서)
	var token string

	// 쿠키에서 먼저 확인
	cookie := c.Cookies("auth_token")
	if cookie != "" {
		token = cookie
		log.Printf("쿠키에서 토큰 발견: %s...", cookie[:10])
	} else {
		// Authorization 헤더에서 확인
		authHeader := c.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
			log.Printf("헤더에서 토큰 발견: %s...", token[:10])
		}
	}

	// 토큰이 없으면 로그인 페이지로 리다이렉트
	if token == "" {
		log.Printf("토큰이 없음: %s 경로로 접근 시도", c.Path())

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

	// 토큰 검증
	user, err := m.authService.ValidateToken(c.Context(), token)
	if err != nil {
		log.Printf("토큰 검증 실패: %v", err)

		// 토큰이 유효하지 않으면 쿠키 삭제
		c.ClearCookie("auth_token")

		// API 요청인 경우 401 응답
		if strings.HasPrefix(c.Path(), "/api/") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "유효하지 않은 인증 정보입니다",
			})
		}

		// 웹 페이지 요청인 경우 로그인 페이지로 리다이렉트
		return c.Redirect("/auth/login?redirect=" + c.Path())
	}

	log.Printf("인증 성공: 사용자 %s (ID: %d)", user.Username, user.ID)

	// 컨텍스트에 사용자 정보 저장
	c.Locals("user", user)

	return c.Next()
}
