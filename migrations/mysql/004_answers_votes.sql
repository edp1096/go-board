-- migrations/mysql/004_answers_votes.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE answers (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    board_id INT NOT NULL,
    question_id INT NOT NULL,
    user_id INT NOT NULL,
    content TEXT NOT NULL,
    vote_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE votes (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    board_id INT NOT NULL,
    target_id INT NOT NULL,
    target_type VARCHAR(20) NOT NULL, -- 'question' or 'answer'
    value INT NOT NULL, -- 1 (up) or -1 (down)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
    UNIQUE KEY unique_vote (user_id, target_id, target_type)
);
-- +goose StatementEnd

-- Indexes for performance optimization
-- +goose StatementBegin
CREATE INDEX idx_votes_target ON votes(target_id, target_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS votes;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS answers;
-- +goose StatementEnd