// internal/service/system_settings_service.go
package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/edp1096/toy-board/internal/models"
	"github.com/edp1096/toy-board/internal/repository"
)

type SystemSettingsService interface {
	GetSetting(ctx context.Context, key string) (string, error)
	GetAllSettings(ctx context.Context) ([]*models.SystemSetting, error)
	SetSetting(ctx context.Context, key, value, description string) error
	GetApprovalMode(ctx context.Context) (string, error)
	GetApprovalDays(ctx context.Context) (int, error)
	UpdateApprovalSettings(ctx context.Context, mode string, days int) error
}

type systemSettingsService struct {
	repo repository.SystemSettingsRepository
}

func NewSystemSettingsService(repo repository.SystemSettingsRepository) SystemSettingsService {
	return &systemSettingsService{
		repo: repo,
	}
}

func (s *systemSettingsService) GetSetting(ctx context.Context, key string) (string, error) {
	setting, err := s.repo.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return setting.SettingValue, nil
}

func (s *systemSettingsService) GetAllSettings(ctx context.Context) ([]*models.SystemSetting, error) {
	return s.repo.GetAll(ctx)
}

func (s *systemSettingsService) SetSetting(ctx context.Context, key, value, description string) error {
	// 최대 재시도 횟수
	maxRetries := 3
	var err error

	for i := 0; i < maxRetries; i++ {
		err = s.repo.Set(ctx, key, value, description)

		// 성공하면 종료
		if err == nil {
			return nil
		}

		// DB가 잠겨 있는 경우 재시도 (SQLite BUSY 에러)
		if strings.Contains(err.Error(), "database is locked") || strings.Contains(err.Error(), "SQLITE_BUSY") {
			// 약간의 딜레이 후 재시도
			time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
			continue
		}

		// 다른 종류의 오류는 바로 리턴
		return err
	}

	return err
}

func (s *systemSettingsService) GetApprovalMode(ctx context.Context) (string, error) {
	value, err := s.GetSetting(ctx, "approval_mode")
	if err != nil {
		// 기본값으로 즉시 승인 반환
		return models.ApprovalModeImmediate, nil
	}
	return value, nil
}

func (s *systemSettingsService) GetApprovalDays(ctx context.Context) (int, error) {
	value, err := s.GetSetting(ctx, "approval_days")
	if err != nil {
		// 기본값 3일 반환
		return 3, nil
	}

	days, err := strconv.Atoi(value)
	if err != nil {
		// 숫자로 변환할 수 없는 경우 기본값 3일 반환
		return 3, nil
	}

	return days, nil
}

func (s *systemSettingsService) UpdateApprovalSettings(ctx context.Context, mode string, days int) error {
	// 승인 모드 업데이트
	err := s.SetSetting(ctx, "approval_mode", mode, "회원가입 승인 모드 (immediate: 즉시, delayed: n일 후, manual: 관리자 승인)")
	if err != nil {
		return err
	}

	// 승인 일수 업데이트
	return s.SetSetting(ctx, "approval_days", strconv.Itoa(days), "회원가입 후 자동 승인까지의 대기 일수 (approval_mode가 delayed인 경우)")
}
