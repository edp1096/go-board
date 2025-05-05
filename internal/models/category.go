// internal/models/category.go
package models

import (
	"time"

	"github.com/uptrace/bun"
)

// Category 모델 - 카테고리 정보 저장
type Category struct {
	bun.BaseModel `bun:"table:categories,alias:c"`

	ID          int64     `bun:"id,pk,autoincrement" json:"id"`
	Name        string    `bun:"name,notnull" json:"name"`            // 카테고리 이름
	Slug        string    `bun:"slug,unique,notnull" json:"slug"`     // URL용 슬러그
	Description string    `bun:"description" json:"description"`      // 카테고리 설명
	ParentID    *int64    `bun:"parent_id" json:"parentId"`           // 부모 카테고리 ID (계층 구조)
	SortOrder   int       `bun:"sort_order,notnull" json:"sortOrder"` // 정렬 순서
	Active      bool      `bun:"active,notnull" json:"active"`        // 활성화 여부
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`
	UpdatedAt   time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updatedAt"`

	// 관계
	Parent   *Category   `bun:"rel:belongs-to,join:parent_id=id" json:"parent,omitempty"`
	Children []*Category `bun:"rel:has-many,join:id=parent_id" json:"children,omitempty"`
}

// CategoryItem 모델 - 카테고리와 아이템(게시판, 페이지)의 관계 정보
type CategoryItem struct {
	bun.BaseModel `bun:"table:category_items,alias:ci"`

	ID         int64     `bun:"id,pk,autoincrement" json:"id"`
	CategoryID int64     `bun:"category_id,notnull" json:"categoryId"` // 카테고리 ID
	ItemID     int64     `bun:"item_id,notnull" json:"itemId"`         // 아이템 ID (게시판 또는 페이지 ID)
	ItemType   string    `bun:"item_type,notnull" json:"itemType"`     // 아이템 타입 ("board" 또는 "page")
	SortOrder  int       `bun:"sort_order,notnull" json:"sortOrder"`   // 정렬 순서
	CreatedAt  time.Time `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`

	// 관계
	Category *Category `bun:"rel:belongs-to,join:category_id=id" json:"category,omitempty"`
}
