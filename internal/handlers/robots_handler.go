// internal/handlers/robots_handler.go
package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// RobotsHandler 구조체
type RobotsHandler struct{}

var robotsTxt = `User-agent: *
Allow: /
Disallow: /admin/
Disallow: /private/
Disallow: /*.pdf$

Sitemap: %s/sitemap.xml
`

// 새 RobotsHandler 생성
func NewRobotsHandler() *RobotsHandler {
	return &RobotsHandler{}
}

// GetRobots는 robots.txt 파일을 제공하는 핸들러
func (h *RobotsHandler) GetRobots(c *fiber.Ctx) error {
	baseURL := getBaseURL(c)

	// robots.txt 내용 생성
	content := fmt.Sprintf(robotsTxt, baseURL)

	// 텍스트 응답 전송
	c.Set("Content-Type", "text/plain; charset=utf-8")
	return c.SendString(content)
}
