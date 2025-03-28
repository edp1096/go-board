// internal/middleware/debug_middleware.go

package middleware

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

// RequestDebugger는 모든 요청의 세부 정보를 로그로 출력하는 미들웨어입니다.
func RequestDebugger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		log.Printf("[디버깅] 경로: %s, 메서드: %s", path, c.Method())

		// 쿠키 검사
		authCookie := c.Cookies("auth_token")
		if authCookie != "" {
			log.Printf("[디버깅] auth_token 쿠키 발견: %s...", authCookie[:10])
		} else {
			log.Printf("[디버깅] auth_token 쿠키 없음")
		}

		// 모든 쿠키 로깅
		cookies := c.GetReqHeaders()["Cookie"]
		if len(cookies) > 0 {
			log.Printf("[디버깅] 모든 쿠키 헤더: %s", cookies)
		}

		// 다음 핸들러로 진행
		err := c.Next()

		// 응답 후 사용자 정보 확인
		user := c.Locals("user")
		if user != nil {
			log.Printf("[디버깅] 응답 후 사용자 정보 있음")
		} else {
			log.Printf("[디버깅] 응답 후 사용자 정보 없음")
		}

		return err
	}
}
