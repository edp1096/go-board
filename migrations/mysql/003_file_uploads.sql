-- migrations/mysql/003_file_uploads.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE attachments (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    board_id BIGINT NOT NULL,
    post_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    storage_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    is_image BOOLEAN NOT NULL DEFAULT false,
    download_url VARCHAR(255) NOT NULL,
    download_count INT DEFAULT 0,
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