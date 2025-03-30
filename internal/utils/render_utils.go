// internal/utils/render_utils.go

package utils

import (
	"github.com/gofiber/fiber/v2"
)

// RenderWithUser는 사용자 정보가 포함된 템플릿 렌더링을 제공합니다.
func RenderWithUser(c *fiber.Ctx, template string, data fiber.Map) error {
	// fmt.Printf("템플릿 렌더링 시도: %s\n", template)

	if data == nil {
		data = fiber.Map{}
	}

	// // .html 확장자가 없으면 추가
	// if !strings.HasSuffix(template, ".html") {
	// 	template = template + ".html"
	// 	fmt.Printf("템플릿 확장자 추가: %s\n", template)
	// }

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

	// 템플릿 렌더링 시도
	err := c.Render(template, data)
	if err != nil {
		// fmt.Printf("템플릿 렌더링 오류: %v\n", err)
		return err
	}

	return nil
}
