// internal/middleware/referrer_middleware.go
package middleware

import (
	"context"
	"strings"

	"github.com/edp1096/go-board/internal/models"
	"github.com/edp1096/go-board/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ReferrerMiddleware interface {
	CaptureReferrer(c *fiber.Ctx) error
}

type referrerMiddleware struct {
	referrerService service.ReferrerService
}

func NewReferrerMiddleware(referrerService service.ReferrerService) ReferrerMiddleware {
	return &referrerMiddleware{
		referrerService: referrerService,
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

	// Get visitor IP
	visitorIP := c.IP()
	if visitorIP == "" {
		visitorIP = "unknown"
	}

	// Get user agent
	userAgent := c.Get("User-Agent")

	// Get user ID if logged in
	var userID *int64
	if user := c.Locals("user"); user != nil {
		u := user.(*models.User)
		userID = &u.ID
	}

	// Record the visit asynchronously to not block the request
	go func(ctx context.Context, rs service.ReferrerService) {
		_ = rs.RecordVisit(ctx, referrerURL, targetURL, visitorIP, userAgent, userID)
	}(context.Background(), m.referrerService)

	return c.Next()
}
