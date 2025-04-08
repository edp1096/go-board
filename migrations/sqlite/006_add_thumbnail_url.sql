-- migrations/sqlite/006_add_thumbnail_url.sql
-- +goose Up
ALTER TABLE attachments ADD COLUMN thumbnail_url TEXT;

-- +goose Down
-- SQLite는 컬럼 삭제를 직접 지원하지 않으므로 테이블 복사 및 재생성이 필요합니다.
-- 마이그레이션 롤백시 수동으로 처리하는 것이 좋습니다.