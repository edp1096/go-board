-- migrations/postgres/004_anonymous_access.sql
-- +goose Up
-- +goose StatementBegin
-- Add allow_anonymous column to boards table
ALTER TABLE boards ADD COLUMN allow_anonymous BOOLEAN NOT NULL DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove the allow_anonymous column from boards table
ALTER TABLE boards DROP COLUMN IF EXISTS allow_anonymous;
-- +goose StatementEnd