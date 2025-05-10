// internal/handlers/referrer_handler.go
package handlers

import (
	"strconv"
	"time"

	"github.com/edp1096/toy-board/internal/models"
	"github.com/edp1096/toy-board/internal/service"
	"github.com/edp1096/toy-board/internal/utils"

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

// ReferrerStatsPage 레퍼러 통계 페이지를 렌더링합니다
func (h *ReferrerHandler) ReferrerStatsPage(c *fiber.Ctx) error {
	// 쿼리 파라미터 가져오기
	days, _ := strconv.Atoi(c.Query("days", "30"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	viewMode := c.Query("view", "url")           // url, domain, type
	showDNS := c.Query("dns", "false") == "true" // DNS 조회 표시 여부

	// 기본값 설정
	if days <= 0 {
		days = 30
	}
	if limit <= 0 {
		limit = 10
	}

	var topReferrers []*models.ReferrerSummary
	var err error

	if viewMode == "domain" {
		// 도메인별 상위 레퍼러 조회
		topReferrers, err = h.referrerService.GetTopReferrersByDomain(c.Context(), limit, days)
	} else {
		// URL별 상위 레퍼러 조회 (기본)
		topReferrers, err = h.referrerService.GetTopReferrers(c.Context(), limit, days)
	}

	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "레퍼러 통계를 불러오는데 실패했습니다",
			"error":   err.Error(),
		})
	}

	// DNS 조회 요청이 있는 경우에만 DNS 정보 채우기
	if showDNS {
		h.referrerService.EnrichReferrerData(topReferrers)
	}

	// 레퍼러 타입 통계 조회
	typeStats, err := h.referrerService.GetReferrersByType(c.Context(), days)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "레퍼러 타입 통계를 불러오는데 실패했습니다",
			"error":   err.Error(),
		})
	}

	// 시간별 통계 조회
	timeStats, err := h.referrerService.GetReferrersByDate(c.Context(), days)
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "시간별 통계를 불러오는데 실패했습니다",
			"error":   err.Error(),
		})
	}

	// 총 방문 수 조회
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
		"viewMode":       viewMode,
		"showDNS":        showDNS,
		"topReferrers":   topReferrers,
		"typeStats":      typeStats,
		"timeStats":      timeStats,
		"total":          total,
		"pageScriptPath": "/static/js/pages/admin-referrer-stats.js",
	})
}

// GetReferrerData API 요청용 JSON 데이터 반환
func (h *ReferrerHandler) GetReferrerData(c *fiber.Ctx) error {
	// 쿼리 파라미터 가져오기
	days, _ := strconv.Atoi(c.Query("days", "30"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	viewMode := c.Query("view", "url")           // url, domain, type, target
	mode := c.Query("mode", "all")               // all, top, types, time
	showDNS := c.Query("dns", "false") == "true" // DNS 조회 표시 여부

	// IP 상세 정보 요청인 경우
	ipDetail := c.Query("ip")
	if ipDetail != "" {
		// 필터링 날짜 계산
		startDate := time.Now().AddDate(0, 0, -days)

		// IP 상세 정보 조회
		detail, err := h.referrerService.GetIPDetails(c.Context(), ipDetail, startDate)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "IP 상세 정보를 조회하는데 실패했습니다",
				"error":   err.Error(),
			})
		}

		// 성공 응답
		return c.JSON(fiber.Map{
			"success": true,
			"data":    detail,
		})
	}

	// 기본값 설정
	if days <= 0 {
		days = 30
	}
	if limit <= 0 {
		limit = 10
	}

	// 응답 데이터 준비
	data := fiber.Map{
		"success": true,
	}

	// 요청된 모드에 따라 데이터 조회
	if mode == "all" || mode == "top" {
		var topReferrers []*models.ReferrerSummary
		var err error

		if viewMode == "domain" {
			topReferrers, err = h.referrerService.GetTopReferrersByDomain(c.Context(), limit, days)
		} else {
			topReferrers, err = h.referrerService.GetTopReferrers(c.Context(), limit, days)
		}

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "레퍼러 통계를 불러오는데 실패했습니다",
			})
		}

		// DNS 정보 추가 요청이 있는 경우
		if showDNS {
			h.referrerService.EnrichReferrerData(topReferrers)
		}

		data["topReferrers"] = topReferrers
	}

	if mode == "all" || mode == "types" {
		typeStats, err := h.referrerService.GetReferrersByType(c.Context(), days)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "레퍼러 타입 통계를 불러오는데 실패했습니다",
			})
		}

		data["typeStats"] = typeStats
	}

	if mode == "all" || mode == "time" {
		timeStats, err := h.referrerService.GetReferrersByDate(c.Context(), days)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "시간별 통계를 불러오는데 실패했습니다",
			})
		}

		data["timeStats"] = timeStats
	}

	if mode == "all" {
		total, err := h.referrerService.GetTotal(c.Context(), days)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "총 방문 수를 불러오는데 실패했습니다",
			})
		}

		data["total"] = total
	}

	return c.JSON(data)
}
