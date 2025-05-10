// internal/handlers/system_settings_handler.go
package handlers

import (
	"strconv"

	"github.com/edp1096/toy-board/internal/models"
	"github.com/edp1096/toy-board/internal/service"
	"github.com/edp1096/toy-board/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type SystemSettingsHandler struct {
	settingsService service.SystemSettingsService
}

func NewSystemSettingsHandler(settingsService service.SystemSettingsService) *SystemSettingsHandler {
	return &SystemSettingsHandler{
		settingsService: settingsService,
	}
}

// SettingsPage 시스템 설정 페이지 렌더링
func (h *SystemSettingsHandler) SettingsPage(c *fiber.Ctx) error {
	// 모든 설정 가져오기
	settings, err := h.settingsService.GetAllSettings(c.Context())
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "설정을 불러오는데 실패했습니다",
			"error":   err.Error(),
		})
	}

	// 설정을 맵으로 변환 (키-값 쌍)
	settingsMap := make(map[string]string)
	for _, setting := range settings {
		settingsMap[setting.SettingKey] = setting.SettingValue
	}

	// 승인 모드 및 일수 기본값 설정
	approvalMode := settingsMap["approval_mode"]
	if approvalMode == "" {
		approvalMode = models.ApprovalModeImmediate
	}

	approvalDays := 3
	if daysStr, ok := settingsMap["approval_days"]; ok {
		if days, err := strconv.Atoi(daysStr); err == nil {
			approvalDays = days
		}
	}

	return utils.RenderWithUser(c, "admin/system_settings", fiber.Map{
		"title":          "시스템 설정",
		"settings":       settings,
		"settingsMap":    settingsMap,
		"approvalMode":   approvalMode,
		"approvalDays":   approvalDays,
		"pageScriptPath": "/static/js/pages/admin-system-settings.js",
	})
}

// UpdateSettings 시스템 설정 업데이트 처리
func (h *SystemSettingsHandler) UpdateSettings(c *fiber.Ctx) error {
	// 승인 모드 및 일수 가져오기
	approvalMode := c.FormValue("approval_mode")
	approvalDaysStr := c.FormValue("approval_days", "3")

	// 승인 일수 정수 변환
	approvalDays, err := strconv.Atoi(approvalDaysStr)
	if err != nil || approvalDays < 1 {
		approvalDays = 3 // 기본값
	}

	// 승인 모드 유효성 검사
	validModes := map[string]bool{
		models.ApprovalModeImmediate: true,
		models.ApprovalModeDelayed:   true,
		models.ApprovalModeManual:    true,
	}

	if !validModes[approvalMode] {
		approvalMode = models.ApprovalModeImmediate // 기본값
	}

	// 설정 업데이트
	err = h.settingsService.UpdateApprovalSettings(c.Context(), approvalMode, approvalDays)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "설정 업데이트에 실패했습니다: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "시스템 설정이 성공적으로 업데이트되었습니다",
	})
}
