// internal/models/answer.go
package models

import (
	"time"

	"github.com/uptrace/bun"
)

// Answer 모델 - Q&A 게시판의 답변 정보
type Answer struct {
	bun.BaseModel `bun:"table:qna_answers,alias:a"`

	ID           int64     `bun:"id,pk,autoincrement" json:"id"`
	BoardID      int64     `bun:"board_id,notnull" json:"boardId"`
	QuestionID   int64     `bun:"question_id,notnull" json:"questionId"`
	UserID       int64     `bun:"user_id,notnull" json:"userId"`
	Content      string    `bun:"content,notnull" json:"content"`
	VoteCount    int       `bun:"vote_count,notnull,default:0" json:"voteCount"`
	ParentID     *int64    `bun:"parent_id" json:"parentId,omitempty"`
	IsBestAnswer bool      `bun:"-" json:"isBestAnswer"`
	CreatedAt    time.Time `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`
	UpdatedAt    time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updatedAt"`

	// 관계
	User     *User     `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	Children []*Answer `bun:"-" json:"children,omitempty"`
}

// QuestionVote 모델 - 질문 투표 정보 저장
type QuestionVote struct {
	bun.BaseModel `bun:"table:qna_question_votes,alias:qv"`

	ID         int64     `bun:"id,pk,autoincrement" json:"id"`
	UserID     int64     `bun:"user_id,notnull" json:"userId"`
	BoardID    int64     `bun:"board_id,notnull" json:"boardId"`
	QuestionID int64     `bun:"question_id,notnull" json:"questionId"`
	Value      int       `bun:"value,notnull" json:"value"` // 1 (up) 또는 -1 (down)
	CreatedAt  time.Time `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`
	UpdatedAt  time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updatedAt"`
}

// AnswerVote 모델 - 답변 투표 정보 저장
type AnswerVote struct {
	bun.BaseModel `bun:"table:qna_answer_votes,alias:av"`

	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	UserID    int64     `bun:"user_id,notnull" json:"userId"`
	BoardID   int64     `bun:"board_id,notnull" json:"boardId"`
	AnswerID  int64     `bun:"answer_id,notnull" json:"answerId"`
	Value     int       `bun:"value,notnull" json:"value"` // 1 (up) 또는 -1 (down)
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updatedAt"`
}
