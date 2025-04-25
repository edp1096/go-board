// internal/handlers/whois_handler.go
package handlers

import (
	"net"
	"strings"

	"github.com/edp1096/go-board/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type WhoisHandler struct {
}

func NewWhoisHandler() *WhoisHandler {
	return &WhoisHandler{}
}

// GetWhoisInfo는 도메인 또는 IP 주소에 대한 WHOIS 정보를 제공합니다
func (h *WhoisHandler) GetWhoisInfo(c *fiber.Ctx) error {
	// 도메인 파라미터 확인
	domain := c.Query("domain")
	ip := c.Query("ip")

	// 도메인 또는 IP 주소 중 하나는 필요
	if domain == "" && ip == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "도메인 또는 IP 주소 파라미터가 필요합니다",
		})
	}

	// IP 주소 처리
	if ip != "" {
		// IP 주소 유효성 검사
		if net.ParseIP(ip) == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "유효하지 않은 IP 주소입니다",
			})
		}

		// WHOIS 정보 조회
		whoisInfo, err := utils.GetIPWhois(ip)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "IP WHOIS 정보 조회 실패",
				"error":   err.Error(),
			})
		}

		// 결과 반환
		return c.JSON(fiber.Map{
			"success": true,
			"data":    whoisInfo,
		})
	}

	// 도메인 처리
	// 도메인 유효성 검사
	if !isDomainValid(domain) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "유효하지 않은 도메인입니다",
		})
	}

	// WHOIS 정보 조회
	whoisInfo, err := utils.GetWhoisInfo(domain)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "도메인 WHOIS 정보 조회 실패",
			"error":   err.Error(),
		})
	}

	// 결과 반환
	return c.JSON(fiber.Map{
		"success": true,
		"data":    whoisInfo,
	})
}

// 간단한 도메인 유효성 검사
func isDomainValid(domain string) bool {
	// 기본 도메인 유효성 검사
	if domain == "" || len(domain) > 253 {
		return false
	}

	// 도메인에는 최소 하나의 점이 있어야 함
	if !strings.Contains(domain, ".") {
		return false
	}

	// 간단한 형식 검사 (더 복잡한 검사는 필요시 추가)
	parts := strings.Split(domain, ".")
	return len(parts) >= 2
}
