// internal/utils/render_utils.go

package utils

import (
	"log"

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
		log.Printf("템플릿 '%s'에 사용자 정보 추가", template)
		data["user"] = user
	} else {
		log.Printf("템플릿 '%s'에 사용자 정보 없음", template)
	}

	return c.Render(template, data)
}
