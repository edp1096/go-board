// internal/middleware/global_auth.go

package middleware

import (
	"log"
	"strings"

	"dynamic-board/internal/service"

	"github.com/gofiber/fiber/v2"
)

// GlobalAuth는 모든 요청에서 인증을 시도하지만 실패해도 차단하지 않는 미들웨어입니다.
func GlobalAuth(authService service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 토큰 가져오기 (쿠키 또는 Authorization 헤더에서)
		var token string

		// 쿠키에서 먼저 확인
		cookie := c.Cookies("auth_token")
		if cookie != "" {
			token = cookie
		} else {
			// Authorization 헤더에서 확인
			authHeader := c.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// 토큰이 있으면 검증 시도
		if token != "" {
			user, err := authService.ValidateToken(c.Context(), token)
			if err == nil && user != nil {
				// 성공하면 컨텍스트에 사용자 정보 저장
				c.Locals("user", user)
				log.Printf("전역 인증 성공: 사용자 %s (ID: %d)", user.Username, user.ID)
			} else if err != nil {
				log.Printf("전역 인증 실패: %v", err)
			}
		}

		// 항상 다음 핸들러로 진행 (실패해도 차단하지 않음)
		return c.Next()
	}
}
