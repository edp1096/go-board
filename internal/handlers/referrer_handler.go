// internal/handlers/referrer_handler.go
package handlers

import (
	"strconv"

	"github.com/edp1096/go-board/internal/service"
	"github.com/edp1096/go-board/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type ReferrerHandler struct {
	referrerService service.ReferrerService
}

func NewReferrerHandler(referrerService service.ReferrerService) *ReferrerHandler {
	return &ReferrerHandler{
		referrerService: referrerService,
	}
}

// ReferrerStatsPage renders the referrer statistics page
func (h *ReferrerHandler) ReferrerStatsPage(c *fiber.Ctx) error {
	// Get query parameters
	days, _ := strconv.Atoi(c.Query("days", "30"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	// Ensure sane defaults
	if days <= 0 {
		days = 30
	}
	if limit <= 0 {
		limit = 10
	}

	// Get top referrers
	topReferrers, err := h.referrerService.GetTopReferrers(c.Context(), limit, days)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "레퍼러 통계를 불러오는데 실패했습니다",
			"error":   err.Error(),
		})
	}

	// Get time-based statistics
	timeStats, err := h.referrerService.GetReferrersByDate(c.Context(), days)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "시간별 통계를 불러오는데 실패했습니다",
			"error":   err.Error(),
		})
	}

	// Get total count
	total, err := h.referrerService.GetTotal(c.Context(), days)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "총 방문 수를 불러오는데 실패했습니다",
			"error":   err.Error(),
		})
	}

	return utils.RenderWithUser(c, "admin/referrer_stats", fiber.Map{
		"title":          "레퍼러 통계",
		"days":           days,
		"limit":          limit,
		"topReferrers":   topReferrers,
		"timeStats":      timeStats,
		"total":          total,
		"pageScriptPath": "/static/js/pages/admin-referrer-stats.js",
	})
}

// GetReferrerData returns JSON data for AJAX requests
func (h *ReferrerHandler) GetReferrerData(c *fiber.Ctx) error {
	// Get query parameters
	days, _ := strconv.Atoi(c.Query("days", "30"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	// Ensure sane defaults
	if days <= 0 {
		days = 30
	}
	if limit <= 0 {
		limit = 10
	}

	// Get top referrers
	topReferrers, err := h.referrerService.GetTopReferrers(c.Context(), limit, days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "레퍼러 통계를 불러오는데 실패했습니다",
		})
	}

	// Get time-based statistics
	timeStats, err := h.referrerService.GetReferrersByDate(c.Context(), days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "시간별 통계를 불러오는데 실패했습니다",
		})
	}

	// Get total count
	total, err := h.referrerService.GetTotal(c.Context(), days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "총 방문 수를 불러오는데 실패했습니다",
		})
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"topReferrers": topReferrers,
		"timeStats":    timeStats,
		"total":        total,
	})
}
