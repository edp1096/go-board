// internal/handlers/sitemap_handler.go
package handlers

import (
	"fmt"

	"github.com/edp1096/go-board/internal/service"
	"github.com/edp1096/go-board/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// 사이트맵 핸들러
type SitemapHandler struct {
	sitemapService service.SitemapService
}

// 새 사이트맵 핸들러 생성
func NewSitemapHandler(sitemapService service.SitemapService) *SitemapHandler {
	return &SitemapHandler{
		sitemapService: sitemapService,
	}
}

// GetSitemapIndex는 사이트맵 인덱스를 제공하는 핸들러
func (h *SitemapHandler) GetSitemapIndex(c *fiber.Ctx) error {
	// 기본 URL 가져오기
	baseURL := getBaseURL(c)

	// 사이트맵 인덱스 생성
	index, sitemaps, err := h.sitemapService.GenerateSitemapIndex(c.Context(), baseURL, 1000)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("사이트맵 생성 오류: " + err.Error())
	}

	// 인덱스 파일과 개별 사이트맵 파일 저장 (메모리에 저장)
	c.Locals("sitemaps", sitemaps)

	// 인덱스 XML 생성
	xmlContent, err := index.ToXML()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("XML 생성 오류: " + err.Error())
	}

	// XML 응답 전송
	c.Set("Content-Type", "application/xml; charset=utf-8")
	return c.SendString(xmlContent)
}

// GetSitemapFile은 개별 사이트맵 파일을 제공하는 핸들러
func (h *SitemapHandler) GetSitemapFile(c *fiber.Ctx) error {
	// 인덱스 값 가져오기
	index := c.Params("index")
	// 파일명 구성
	filename := fmt.Sprintf("sitemap_%s.xml", index)

	// 사이트맵 파일 가져오기
	sitemaps, ok := c.Locals("sitemaps").(map[string]*utils.Sitemap)
	if !ok {
		// 사이트맵이 없으면 다시 생성
		baseURL := getBaseURL(c)
		_, sitemapsNew, err := h.sitemapService.GenerateSitemapIndex(c.Context(), baseURL, 1000)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("사이트맵 생성 오류: " + err.Error())
		}

		// 요청된 파일 찾기
		sitemap, exists := sitemapsNew[filename]
		if !exists {
			return c.Status(fiber.StatusNotFound).SendString("요청한 사이트맵 파일을 찾을 수 없습니다")
		}

		// XML 생성
		xmlContent, err := sitemap.ToXML()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("XML 생성 오류: " + err.Error())
		}

		// XML 응답 전송
		c.Set("Content-Type", "application/xml; charset=utf-8")
		return c.SendString(xmlContent)
	}

	// 캐시된 사이트맵에서 요청된 파일 찾기
	sitemap, exists := sitemaps[filename]
	if !exists {
		return c.Status(fiber.StatusNotFound).SendString("요청한 사이트맵 파일을 찾을 수 없습니다")
	}

	// XML 생성
	xmlContent, err := sitemap.ToXML()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("XML 생성 오류: " + err.Error())
	}

	// XML 응답 전송
	c.Set("Content-Type", "application/xml; charset=utf-8")
	return c.SendString(xmlContent)
}

// 기본 URL 가져오기 헬퍼 함수
func getBaseURL(c *fiber.Ctx) string {
	baseURL := c.Locals("baseURL")
	if baseURL != nil {
		return baseURL.(string)
	}

	protocol := "https"
	if c.Protocol() == "http" {
		protocol = "http"
	}

	return protocol + "://" + c.Hostname()
}
