// internal/utils/render_utils.go

package utils

import (
	"github.com/gofiber/fiber/v2"
)

// RenderWithUser는 사용자 정보가 포함된 템플릿 렌더링을 제공합니다.
func RenderWithUser(c *fiber.Ctx, template string, data fiber.Map) error {
	if data == nil {
		data = fiber.Map{}
	}

	// c.Locals에 저장된 "user" 값을 데이터에 추가합니다.
	user := c.Locals("user")
	if user != nil {
		data["user"] = user
	}
	
	// CSRF 토큰 추가
	if csrf := c.Locals("csrf"); csrf != nil {
		data["csrf"] = csrf
	}

	// UTF-8 인코딩 명시적 설정
	c.Set("Content-Type", "text/html; charset=utf-8")

	return c.Render(template, data)
}
