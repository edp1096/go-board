package utils

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// IP 주소 획득 함수
func GetClientIP(c *fiber.Ctx) string {
	// X-Forwarded-For 헤더 확인
	forwardedFor := c.Get("X-Forwarded-For")
	if forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// X-Real-IP 헤더 확인
	realIP := c.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// fasthttp 컨텍스트에서 직접 원격 IP 주소 획득
	remoteAddr := c.Context().RemoteAddr().String()
	if remoteAddr != "" {
		// IP:포트 형식에서 IP 부분만 추출
		if strings.Contains(remoteAddr, ":") {
			ip := strings.Split(remoteAddr, ":")[0]
			return ip
		}

		return remoteAddr
	}

	// 기본 IP 사용
	return c.IP()
}
