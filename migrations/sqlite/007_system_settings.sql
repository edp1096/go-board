-- migrations/sqlite/007_system_settings.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE system_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    setting_key TEXT NOT NULL UNIQUE,
    setting_value TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- 기본 시스템 설정 삽입
-- +goose StatementBegin
INSERT INTO system_settings (setting_key, setting_value, description) VALUES
('approval_mode', 'immediate', '회원가입 승인 모드 (immediate: 즉시, delayed: n일 후, manual: 관리자 승인)'),
('approval_days', '3', '회원가입 후 자동 승인까지의 대기 일수 (approval_mode가 delayed인 경우)');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS system_settings;
-- +goose StatementEnd