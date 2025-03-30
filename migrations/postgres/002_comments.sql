-- migrations/postgres/002_comments.sql
-- +goose Up
-- +goose StatementBegin
-- Add comments_enabled column to boards table
ALTER TABLE boards ADD COLUMN comments_enabled BOOLEAN NOT NULL DEFAULT true;
-- +goose StatementEnd

-- +goose StatementBegin
-- Create comments table
CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    post_id INT NOT NULL,
    board_id INT NOT NULL,
    user_id INT NOT NULL,
    content TEXT NOT NULL,
    parent_id INT DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
-- Add index for post_id
CREATE INDEX idx_comments_post_id ON comments (post_id);
-- +goose StatementEnd

-- +goose StatementBegin
-- Add index for board_id
CREATE INDEX idx_comments_board_id ON comments (board_id);
-- +goose StatementEnd

-- +goose StatementBegin
-- Add index for parent_id
CREATE INDEX idx_comments_parent_id ON comments (parent_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop the comments table
DROP TABLE IF EXISTS comments;
-- +goose StatementEnd

-- +goose StatementBegin
-- Remove the comments_enabled column from boards table
ALTER TABLE boards DROP COLUMN IF EXISTS comments_enabled;
-- +goose StatementEnd