// internal/models/page.go
package models

import (
	"time"

	"github.com/uptrace/bun"
)

// Page 모델 - 정적 페이지 정보 저장
type Page struct {
	bun.BaseModel `bun:"table:pages,alias:p"`

	ID         int64     `bun:"id,pk,autoincrement" json:"id"`
	Title      string    `bun:"title,notnull" json:"title"`             // 페이지 제목
	Slug       string    `bun:"slug,unique,notnull" json:"slug"`        // URL용 슬러그
	Content    string    `bun:"content" json:"content"`                 // 페이지 내용
	Active     bool      `bun:"active,notnull" json:"active"`           // 페이지 활성화 여부
	ShowInMenu bool      `bun:"show_in_menu,notnull" json:"showInMenu"` // 메뉴에 표시 여부
	SortOrder  int       `bun:"sort_order,notnull" json:"sortOrder"`    // 정렬 순서
	CreatedAt  time.Time `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`
	UpdatedAt  time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updatedAt"`
}
