-- migrations/mysql/005_split_votes_tables.sql
-- +goose Up
-- +goose StatementBegin
-- 1. 기존 테이블 백업 (안전을 위해)
CREATE TABLE qna_votes_backup AS SELECT * FROM qna_votes;
-- +goose StatementEnd

-- +goose StatementBegin
-- 2. 새 질문 투표 테이블 생성
CREATE TABLE qna_question_votes (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    board_id INT NOT NULL,
    question_id INT NOT NULL,
    value INT NOT NULL, -- 1 (up) or -1 (down)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
    UNIQUE KEY unique_question_vote (user_id, question_id)
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
-- 5. 답변 투표 테이블 구조 수정
ALTER TABLE qna_answer_votes
    DROP COLUMN target_type,
    CHANGE COLUMN target_id answer_id INT NOT NULL,
    ADD UNIQUE KEY unique_answer_vote (user_id, answer_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- 1. 고유 키 제약 조건 제거
ALTER TABLE qna_answer_votes DROP INDEX unique_answer_vote;
-- +goose StatementEnd

-- +goose StatementBegin
-- 2. 원래 테이블 구조로 복원
ALTER TABLE qna_answer_votes
    CHANGE COLUMN answer_id target_id INT NOT NULL,
    ADD COLUMN target_type VARCHAR(20) NOT NULL DEFAULT 'answer';
-- +goose StatementEnd

-- +goose StatementBegin
-- 3. 테이블 이름 원복
ALTER TABLE qna_answer_votes RENAME TO qna_votes;
-- +goose StatementEnd

-- +goose StatementBegin
-- 4. 질문 투표 데이터 복원
UPDATE qna_votes SET target_type = 'question' 
WHERE target_id IN (
    SELECT question_id FROM qna_question_votes
);
-- +goose StatementEnd

-- +goose StatementBegin
-- 5. 새 테이블 삭제
DROP TABLE qna_question_votes;
-- +goose StatementEnd

-- +goose StatementBegin
-- 6. 백업 테이블 삭제 (필요시)
DROP TABLE qna_votes_backup;
-- +goose StatementEnd