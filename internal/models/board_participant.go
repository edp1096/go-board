package models

import (
	"time"

	"github.com/uptrace/bun"
)

// 참여자 역할 유형
type ParticipantRole string

const (
	ParticipantRoleMember    ParticipantRole = "member"    // 일반 회원
	ParticipantRoleModerator ParticipantRole = "moderator" // 중재자
)

// BoardParticipant 모델 - 소모임 게시판 참여자 정보
type BoardParticipant struct {
	bun.BaseModel `bun:"table:board_participants,alias:bp"`

	BoardID   int64           `bun:"board_id,pk" json:"boardId"`
	UserID    int64           `bun:"user_id,pk" json:"userId"`
	Role      ParticipantRole `bun:"role,notnull" json:"role"`
	CreatedAt time.Time       `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`

	// 관계
	Board *Board `bun:"rel:belongs-to,join:board_id=id" json:"board,omitempty"`
	User  *User  `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
}
