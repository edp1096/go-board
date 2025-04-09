-- migrations/sqlite/007_qna_reply.sql
-- +goose Up
-- +goose StatementBegin
-- SQLite에서 ALTER TABLE로 외래 키 제약 조건 추가는 직접 지원하지 않음
-- 대신 테이블을 새로 만들고 데이터를 복사하는 방식으로 구현

-- 1. 임시 테이블 생성
CREATE TABLE qna_answers_new (
    id INTEGER NOT NULL PRIMARY KEY,
    board_id INTEGER NOT NULL,
    question_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    vote_count INTEGER NOT NULL DEFAULT 0,
    parent_id INTEGER DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES qna_answers_new(id) ON DELETE CASCADE
);

-- 2. 기존 데이터 복사
INSERT INTO qna_answers_new (id, board_id, question_id, user_id, content, vote_count, created_at, updated_at)
SELECT id, board_id, question_id, user_id, content, vote_count, created_at, updated_at
FROM qna_answers;

-- 3. 기존 테이블 삭제
DROP TABLE qna_answers;

-- 4. 새 테이블 이름 변경
ALTER TABLE qna_answers_new RENAME TO qna_answers;

-- 5. 인덱스 생성
CREATE INDEX idx_qna_answers_parent_id ON qna_answers(parent_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- SQLite에서는 컬럼 삭제를 직접 지원하지 않으므로 비슷한 접근 방식이 필요함

-- 1. 임시 테이블 생성 (parent_id 없이)
CREATE TABLE qna_answers_new (
    id INTEGER NOT NULL PRIMARY KEY,
    board_id INTEGER NOT NULL,
    question_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    vote_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 2. 데이터 복사 (parent_id 제외)
INSERT INTO qna_answers_new (id, board_id, question_id, user_id, content, vote_count, created_at, updated_at)
SELECT id, board_id, question_id, user_id, content, vote_count, created_at, updated_at
FROM qna_answers;

-- 3. 기존 테이블 삭제
DROP TABLE qna_answers;

-- 4. 새 테이블 이름 변경
ALTER TABLE qna_answers_new RENAME TO qna_answers;
-- +goose StatementEnd