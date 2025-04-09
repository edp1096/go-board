-- migrations/postgres/007_qna_reply.sql
-- +goose Up
-- +goose StatementBegin
-- 답변 테이블에 parent_id 추가
ALTER TABLE qna_answers ADD COLUMN parent_id INTEGER DEFAULT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
-- parent_id에 인덱스 추가하여 조회 성능 개선
CREATE INDEX idx_qna_answers_parent_id ON qna_answers(parent_id);
-- +goose StatementEnd

-- +goose StatementBegin
-- self-referencing 외래 키 제약 추가
ALTER TABLE qna_answers ADD CONSTRAINT fk_qna_answers_parent_id
FOREIGN KEY (parent_id) REFERENCES qna_answers(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- 외래 키 제약 제거
ALTER TABLE qna_answers DROP CONSTRAINT fk_qna_answers_parent_id;
-- +goose StatementEnd

-- +goose StatementBegin
-- 인덱스 제거
DROP INDEX idx_qna_answers_parent_id;
-- +goose StatementEnd

-- +goose StatementBegin
-- parent_id 컬럼 제거
ALTER TABLE qna_answers DROP COLUMN parent_id;
-- +goose StatementEnd