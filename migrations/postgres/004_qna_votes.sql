-- migrations/postgres/004_qna_votes.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE qna_answers (
    id SERIAL PRIMARY KEY,
    board_id INT NOT NULL,
    question_id INT NOT NULL,
    user_id INT NOT NULL,
    content TEXT NOT NULL,
    vote_count INT NOT NULL DEFAULT 0,
    parent_id INTEGER DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES qna_answers(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_qna_answers_parent_id ON qna_answers(parent_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE qna_question_votes (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    board_id INT NOT NULL,
    question_id INT NOT NULL,
    value INT NOT NULL, -- 1 (up) or -1 (down)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
    UNIQUE (user_id, question_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE qna_answer_votes (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    board_id INT NOT NULL,
    answer_id INT NOT NULL,
    value INT NOT NULL, -- 1 (up) or -1 (down)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
    UNIQUE (user_id, answer_id)
);
-- +goose StatementEnd

-- Indexes for performance optimization
-- +goose StatementBegin
CREATE INDEX idx_question_votes_question ON qna_question_votes(question_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_answer_votes_answer ON qna_answer_votes(answer_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS qna_answer_votes;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS qna_question_votes;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS qna_answers;
-- +goose StatementEnd