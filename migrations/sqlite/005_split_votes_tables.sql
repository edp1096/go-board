-- migrations/sqlite/005_split_votes_tables.sql
-- +goose Up
-- +goose StatementBegin
-- 1. 기존 테이블 백업 (안전을 위해)
CREATE TABLE qna_votes_backup AS SELECT * FROM qna_votes;
-- +goose StatementEnd

-- +goose StatementBegin
-- 2. 새 질문 투표 테이블 생성
CREATE TABLE qna_question_votes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    board_id INTEGER NOT NULL,
    question_id INTEGER NOT NULL,
    value INTEGER NOT NULL, -- 1 (up) or -1 (down)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
    UNIQUE (user_id, question_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
-- 3. 기존 질문 투표 데이터 이전
INSERT INTO qna_question_votes (user_id, board_id, question_id, value, created_at, updated_at)
SELECT user_id, board_id, target_id, value, created_at, updated_at 
FROM qna_votes 
WHERE target_type = 'question';
-- +goose StatementEnd

-- +goose StatementBegin
-- 4. 기존 테이블을 answer_votes 용으로 새로 생성 (SQLite는 ALTER TABLE 기능이 제한적)
CREATE TABLE qna_answer_votes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    board_id INTEGER NOT NULL,
    answer_id INTEGER NOT NULL,
    value INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
    UNIQUE (user_id, answer_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
-- 5. 답변 투표 데이터 이전
INSERT INTO qna_answer_votes (user_id, board_id, answer_id, value, created_at, updated_at)
SELECT user_id, board_id, target_id, value, created_at, updated_at 
FROM qna_votes 
WHERE target_type = 'answer';
-- +goose StatementEnd

-- +goose StatementBegin
-- 6. 기존 테이블 삭제
DROP TABLE IF EXISTS qna_votes;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- 1. qna_votes 테이블 다시 생성
CREATE TABLE qna_votes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    board_id INTEGER NOT NULL,
    target_id INTEGER NOT NULL,
    target_type TEXT NOT NULL,
    value INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
    UNIQUE (user_id, target_id, target_type)
);
-- +goose StatementEnd

-- +goose StatementBegin
-- 2. 질문 투표 데이터 복원
INSERT INTO qna_votes (user_id, board_id, target_id, target_type, value, created_at, updated_at)
SELECT user_id, board_id, question_id, 'question', value, created_at, updated_at
FROM qna_question_votes;
-- +goose StatementEnd

-- +goose StatementBegin
-- 3. 답변 투표 데이터 복원
INSERT INTO qna_votes (user_id, board_id, target_id, target_type, value, created_at, updated_at)
SELECT user_id, board_id, answer_id, 'answer', value, created_at, updated_at
FROM qna_answer_votes;
-- +goose StatementEnd

-- +goose StatementBegin
-- 4. 새 테이블 삭제
DROP TABLE IF EXISTS qna_question_votes;
-- +goose StatementEnd

-- +goose StatementBegin
-- 5. 답변 투표 테이블 삭제
DROP TABLE IF EXISTS qna_answer_votes;
-- +goose StatementEnd

-- +goose StatementBegin
-- 6. 백업 테이블 삭제
DROP TABLE IF EXISTS qna_votes_backup;
-- +goose StatementEnd