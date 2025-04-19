-- migrations/sqlite/003_file_uploads.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE attachments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    board_id INTEGER NOT NULL,
    post_id INTEGER NOT NULL, 
    user_id INTEGER NOT NULL,
    file_name TEXT NOT NULL,
    storage_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    mime_type TEXT NOT NULL,
    is_image TINYINT NOT NULL DEFAULT 0,
    download_url TEXT NOT NULL,
    thumbnail_url TEXT DEFAULT NULL,
    download_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_attachments_post ON attachments(board_id, post_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS attachments;
-- +goose StatementEnd