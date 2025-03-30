// internal/models/comment.go
package models

import (
	"time"

	"github.com/uptrace/bun"
)

// Comment 모델 - 게시물 댓글 정보 저장
type Comment struct {
	bun.BaseModel `bun:"table:comments,alias:c"`

	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	PostID    int64     `bun:"post_id,notnull" json:"postId"`
	BoardID   int64     `bun:"board_id,notnull" json:"boardId"`
	UserID    int64     `bun:"user_id,notnull" json:"userId"`
	Content   string    `bun:"content,notnull" json:"content"`
	ParentID  *int64    `bun:"parent_id" json:"parentId,omitempty"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updatedAt"`

	// 관계
	User     *User      `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	Children []*Comment `bun:"rel:has-many,join:id=parent_id" json:"children,omitempty"`
}
