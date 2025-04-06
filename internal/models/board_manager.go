// internal/models/board_manager.go
package models

import (
	"time"

	"github.com/uptrace/bun"
)

// BoardManager 모델 - 게시판과 매니저의 관계 정보 저장
type BoardManager struct {
	bun.BaseModel `bun:"table:board_managers,alias:bm"`

	BoardID   int64     `bun:"board_id,pk" json:"boardId"`
	UserID    int64     `bun:"user_id,pk" json:"userId"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`

	// 관계
	Board *Board `bun:"rel:belongs-to,join:board_id=id" json:"board,omitempty"`
	User  *User  `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
}
