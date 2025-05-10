// internal/repository/system_settings_repository.go
package repository

import (
	"context"
	"time"

	"github.com/edp1096/toy-board/internal/models"

	"github.com/uptrace/bun"
)

type SystemSettingsRepository interface {
	Get(ctx context.Context, key string) (*models.SystemSetting, error)
	GetAll(ctx context.Context) ([]*models.SystemSetting, error)
	Set(ctx context.Context, key string, value string, description string) error
	Update(ctx context.Context, setting *models.SystemSetting) error
}

type systemSettingsRepository struct {
	db *bun.DB
}

func NewSystemSettingsRepository(db *bun.DB) SystemSettingsRepository {
	return &systemSettingsRepository{db: db}
}

func (r *systemSettingsRepository) Get(ctx context.Context, key string) (*models.SystemSetting, error) {
	setting := new(models.SystemSetting)
	err := r.db.NewSelect().
		Model(setting).
		Where("setting_key = ?", key).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return setting, nil
}

func (r *systemSettingsRepository) GetAll(ctx context.Context) ([]*models.SystemSetting, error) {
	var settings []*models.SystemSetting
	err := r.db.NewSelect().
		Model(&settings).
		OrderExpr("id ASC").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (r *systemSettingsRepository) Set(ctx context.Context, key string, value string, description string) error {
	// 먼저 키가 존재하는지 확인
	setting := new(models.SystemSetting)
	err := r.db.NewSelect().
		Model(setting).
		Where("setting_key = ?", key).
		Scan(ctx)

	if err == nil {
		// 이미 존재하는 설정이라면 업데이트
		setting.SettingValue = value
		if description != "" {
			setting.Description = description
		}
		setting.UpdatedAt = time.Now()
		_, err = r.db.NewUpdate().
			Model(setting).
			WherePK().
			Exec(ctx)
		return err
	}

	// 존재하지 않는다면 새로 생성
	newSetting := &models.SystemSetting{
		SettingKey:   key,
		SettingValue: value,
		Description:  description,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = r.db.NewInsert().
		Model(newSetting).
		Exec(ctx)

	return err
}

func (r *systemSettingsRepository) Update(ctx context.Context, setting *models.SystemSetting) error {
	setting.UpdatedAt = time.Now()
	_, err := r.db.NewUpdate().
		Model(setting).
		WherePK().
		Exec(ctx)

	return err
}
