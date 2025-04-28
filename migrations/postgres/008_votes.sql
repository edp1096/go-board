-- migrations/postgres/008_votes.sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE boards ADD COLUMN votes_enabled BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE post_votes (
    id SERIAL PRIMARY KEY,
    post_id INTEGER NOT NULL,
    board_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    value INTEGER NOT NULL, -- 1: 좋아요, -1: 싫어요
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (post_id, user_id)
);

CREATE TABLE comment_votes (
    id SERIAL PRIMARY KEY,
    comment_id INTEGER NOT NULL,
    board_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    value INTEGER NOT NULL, -- 1: 좋아요, -1: 싫어요
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (comment_id, user_id)
);

ALTER TABLE comments ADD COLUMN like_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE comments ADD COLUMN dislike_count INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE comments DROP COLUMN dislike_count;
ALTER TABLE comments DROP COLUMN like_count;
DROP TABLE IF EXISTS comment_votes;
DROP TABLE IF EXISTS post_votes;
ALTER TABLE boards DROP COLUMN votes_enabled;
-- +goose StatementEnd