// internal/middleware/referrer_middleware.go
package middleware

import (
	"context"
	"strings"

	"github.com/edp1096/go-board/config"
	"github.com/edp1096/go-board/internal/models"
	"github.com/edp1096/go-board/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ReferrerMiddleware interface {
	CaptureReferrer(c *fiber.Ctx) error
}

type referrerMiddleware struct {
	referrerService service.ReferrerService
	config          *config.Config
}

func NewReferrerMiddleware(referrerService service.ReferrerService, cfg *config.Config) ReferrerMiddleware {
	return &referrerMiddleware{
		referrerService: referrerService,
		config:          cfg,
	}
}

func (m *referrerMiddleware) CaptureReferrer(c *fiber.Ctx) error {
	// Skip for static resources, API calls, or other irrelevant paths
	path := c.Path()
	if strings.HasPrefix(path, "/static") ||
		strings.HasPrefix(path, "/uploads") ||
		strings.HasPrefix(path, "/api") ||
		strings.HasPrefix(path, "/favicon.ico") {
		return c.Next()
	}

	// Get referrer from request header
	referrerURL := c.Get("Referer")
	if referrerURL == "" {
		referrerURL = "direct"
	}

	// Get target URL
	targetURL := c.Path()

	// Get visitor IP - 프록시 환경에 맞게 수정
	var visitorIP string
	// X-Forwarded-For 헤더 확인
	forwardedFor := c.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// 여러 IP가 있을 경우 맨 앞의 IP를 사용 (실제 클라이언트 IP)
		ips := strings.Split(forwardedFor, ",")
		visitorIP = strings.TrimSpace(ips[0])
	} else {
		// X-Real-IP 헤더 확인
		realIP := c.Get("X-Real-IP")
		if realIP != "" {
			visitorIP = realIP
		} else {
			// 둘 다 없으면 기본 IP 사용
			visitorIP = c.IP()
		}
	}

	if visitorIP == "" {
		remoteAddr := c.Context().RemoteAddr().String()
		if strings.Contains(remoteAddr, ":") {
			remoteAddr = strings.Split(remoteAddr, ":")[0]
		}
		visitorIP = remoteAddr
	}

	// Get user agent
	userAgent := c.Get("User-Agent")
	if userAgent == "" {
		userAgent = "unknown"
	}

	// Get user ID if logged in
	var userID *int64
	if user := c.Locals("user"); user != nil {
		u := user.(*models.User)
		userID = &u.ID
	}

	// DBDriver로 SQLite 여부 확인
	if m.config.DBDriver == "sqlite" {
		// SQLite인 경우 동기적으로 처리
		_ = m.referrerService.RecordVisit(c.Context(), referrerURL, targetURL, visitorIP, userAgent, userID)
	} else {
		// 다른 DB는 비동기 처리
		go func(ctx context.Context, rs service.ReferrerService) {
			_ = rs.RecordVisit(ctx, referrerURL, targetURL, visitorIP, userAgent, userID)
		}(context.Background(), m.referrerService)
	}

	return c.Next()
}
