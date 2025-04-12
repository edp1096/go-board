// internal/models/system_settings.go
package models

import (
	"time"

	"github.com/uptrace/bun"
)

// 승인 모드 상수
const (
	ApprovalModeImmediate = "immediate" // 즉시 승인
	ApprovalModeDelayed   = "delayed"   // n일 후 승인
	ApprovalModeManual    = "manual"    // 관리자 수동 승인
)

// SystemSetting 시스템 설정 모델
type SystemSetting struct {
	bun.BaseModel `bun:"table:system_settings,alias:ss"`

	ID           int64     `bun:"id,pk,autoincrement" json:"id"`
	SettingKey   string    `bun:"setting_key,unique,notnull" json:"settingKey"`
	SettingValue string    `bun:"setting_value,notnull" json:"settingValue"`
	Description  string    `bun:"description" json:"description"`
	CreatedAt    time.Time `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`
	UpdatedAt    time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updatedAt"`
}
