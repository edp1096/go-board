package models

import (
	"time"

	"github.com/uptrace/bun"
)

// PostVote 모델 - 게시물 좋아요/싫어요 정보 저장
type PostVote struct {
	bun.BaseModel `bun:"table:post_votes,alias:pv"`

	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	PostID    int64     `bun:"post_id,notnull" json:"postId"`
	BoardID   int64     `bun:"board_id,notnull" json:"boardId"`
	UserID    int64     `bun:"user_id,notnull" json:"userId"`
	Value     int       `bun:"value,notnull" json:"value"` // 1: 좋아요, -1: 싫어요
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updatedAt"`

	// 관계
	User *User `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
}
