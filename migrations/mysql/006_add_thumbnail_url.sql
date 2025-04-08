-- migrations/mysql/006_add_thumbnail_url.sql
-- +goose Up
ALTER TABLE attachments ADD COLUMN thumbnail_url VARCHAR(255) DEFAULT NULL AFTER download_url;

-- +goose Down
ALTER TABLE attachments DROP COLUMN thumbnail_url;