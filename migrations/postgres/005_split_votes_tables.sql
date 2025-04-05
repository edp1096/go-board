-- migrations/postgres/005_split_votes_tables.sql
-- +goose Up
-- +goose StatementBegin
-- 1. 기존 테이블 백업 (안전을 위해)
CREATE TABLE qna_votes_backup AS SELECT * FROM qna_votes;
-- +goose StatementEnd

-- +goose StatementBegin
-- 2. 새 질문 투표 테이블 생성
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
-- 3. 기존 질문 투표 데이터 이전
INSERT INTO qna_question_votes (user_id, board_id, question_id, value, created_at, updated_at)
SELECT user_id, board_id, target_id, value, created_at, updated_at 
FROM qna_votes 
WHERE target_type = 'question';
-- +goose StatementEnd

-- +goose StatementBegin
-- 4. 기존 테이블을 답변 투표용으로 변경
ALTER TABLE qna_votes RENAME TO qna_answer_votes;
-- +goose StatementEnd

-- +goose StatementBegin
-- 5-1. 칼럼 이름 변경
ALTER TABLE qna_answer_votes RENAME COLUMN target_id TO answer_id;
-- +goose StatementEnd

-- +goose StatementBegin
-- 5-2. 불필요한 칼럼 삭제
ALTER TABLE qna_answer_votes DROP COLUMN target_type;
-- +goose StatementEnd

-- +goose StatementBegin
-- 5-3. 고유 제약 조건 추가
ALTER TABLE qna_answer_votes ADD CONSTRAINT unique_answer_vote UNIQUE (user_id, answer_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- 1. 고유 제약 조건 제거
ALTER TABLE qna_answer_votes DROP CONSTRAINT unique_answer_vote;
-- +goose StatementEnd

-- +goose StatementBegin
-- 2-1. 칼럼 추가
ALTER TABLE qna_answer_votes ADD COLUMN target_type VARCHAR(20);
-- +goose StatementEnd

-- +goose StatementBegin
-- 2-2. 기본값 설정
UPDATE qna_answer_votes SET target_type = 'answer';
-- +goose StatementEnd

-- +goose StatementBegin
-- 2-3. NOT NULL 제약 조건 추가
ALTER TABLE qna_answer_votes ALTER COLUMN target_type SET NOT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
-- 2-4. 칼럼 이름 변경
ALTER TABLE qna_answer_votes RENAME COLUMN answer_id TO target_id;
-- +goose StatementEnd

-- +goose StatementBegin
-- 3. 테이블 이름 복원
ALTER TABLE qna_answer_votes RENAME TO qna_votes;
-- +goose StatementEnd

-- +goose StatementBegin
-- 4. 질문 투표 데이터 복원
INSERT INTO qna_votes (user_id, board_id, target_id, target_type, value, created_at, updated_at)
SELECT user_id, board_id, question_id, 'question', value, created_at, updated_at
FROM qna_question_votes;
-- +goose StatementEnd

-- +goose StatementBegin
-- 5. 새 테이블 삭제
DROP TABLE qna_question_votes;
-- +goose StatementEnd

-- +goose StatementBegin
-- 6. 백업 테이블 삭제 (필요시)
DROP TABLE qna_votes_backup;
-- +goose StatementEnd