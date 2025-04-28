-- migrations/sqlite/008_votes.sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE boards ADD COLUMN votes_enabled TINYINT NOT NULL DEFAULT 0;

CREATE TABLE post_votes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    board_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    value INTEGER NOT NULL, -- 1: 좋아요, -1: 싫어요
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (post_id, user_id)
);

CREATE TABLE comment_votes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    comment_id INTEGER NOT NULL,
    board_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    value INTEGER NOT NULL, -- 1: 좋아요, -1: 싫어요
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (comment_id, user_id)
);

-- SQLite에서는 한 번에 하나의 ALTER TABLE 문만 실행할 수 있습니다
ALTER TABLE comments ADD COLUMN like_count INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE comments ADD COLUMN dislike_count INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- SQLite는 칼럼 삭제를 지원하지 않으므로 테이블 재생성이 필요합니다
-- 이 예제에서는 단순화를 위해 DROP과 CREATE를 사용합니다
CREATE TABLE comments_backup (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    board_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    parent_id INTEGER DEFAULT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE
);

INSERT INTO comments_backup SELECT id, post_id, board_id, user_id, content, parent_id, created_at, updated_at FROM comments;
DROP TABLE comments;
ALTER TABLE comments_backup RENAME TO comments;

DROP TABLE IF EXISTS comment_votes;
DROP TABLE IF EXISTS post_votes;

CREATE TABLE boards_backup (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    description TEXT,
    board_type TEXT NOT NULL,
    table_name TEXT NOT NULL UNIQUE,
    active TINYINT NOT NULL DEFAULT 1,
    comments_enabled TINYINT NOT NULL DEFAULT 1,
    allow_anonymous TINYINT NOT NULL DEFAULT 0,
    allow_private TINYINT NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO boards_backup SELECT id, name, slug, description, board_type, table_name, active, comments_enabled, allow_anonymous, allow_private, sort_order, created_at, updated_at FROM boards;
DROP TABLE boards;
ALTER TABLE boards_backup RENAME TO boards;
-- +goose StatementEnd