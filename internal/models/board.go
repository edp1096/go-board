// internal/models/board.go (updated version with comments_enabled field)
package models

import (
	"time"

	"github.com/uptrace/bun"
)

// 게시판 유형
type BoardType string

const (
	BoardTypeNormal  BoardType = "normal"  // 일반 게시판
	BoardTypeGallery BoardType = "gallery" // 갤러리 게시판
	BoardTypeQnA     BoardType = "qna"     // 질문/답변 게시판
)

// 게시판 필드 유형
type FieldType string

const (
	FieldTypeText     FieldType = "text"     // 텍스트 필드
	FieldTypeTextarea FieldType = "textarea" // 텍스트 영역
	FieldTypeNumber   FieldType = "number"   // 숫자
	FieldTypeDate     FieldType = "date"     // 날짜
	FieldTypeSelect   FieldType = "select"   // 선택 옵션
	FieldTypeCheckbox FieldType = "checkbox" // 체크박스
	FieldTypeFile     FieldType = "file"     // 파일 업로드
)

// Board 모델 - 게시판 정보 저장
type Board struct {
	bun.BaseModel `bun:"table:boards,alias:b"`

	ID              int64     `bun:"id,pk,autoincrement" json:"id"`
	Name            string    `bun:"name,notnull" json:"name"`                   // 게시판 이름
	Slug            string    `bun:"slug,unique,notnull" json:"slug"`            // URL용 슬러그
	Description     string    `bun:"description" json:"description"`             // 게시판 설명
	BoardType       BoardType `bun:"board_type,notnull" json:"boardType"`        // 게시판 유형
	TableName       string    `bun:"table_name,notnull,unique" json:"tableName"` // 실제 DB 테이블 이름
	Active          bool      `bun:"active,notnull,default:true" json:"active"`
	CommentsEnabled bool      `bun:"comments_enabled,notnull,default:true" json:"commentsEnabled"` // 댓글 기능 활성화 여부
	AllowAnonymous  bool      `bun:"allow_anonymous,notnull,default:false" json:"allowAnonymous"`  // 익명 사용자 접근 허용 여부
	CreatedAt       time.Time `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`
	UpdatedAt       time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updatedAt"`

	// 관계
	Fields []*BoardField `bun:"rel:has-many,join:id=board_id" json:"fields"`
}

// BoardField 모델 - 게시판 각 필드 정의
type BoardField struct {
	bun.BaseModel `bun:"table:board_fields,alias:bf"`

	ID          int64     `bun:"id,pk,autoincrement" json:"id"`
	BoardID     int64     `bun:"board_id,notnull" json:"boardId"`
	Name        string    `bun:"name,notnull" json:"name"`                // 필드 이름
	ColumnName  string    `bun:"column_name,notnull" json:"columnName"`   // DB 칼럼 이름
	DisplayName string    `bun:"display_name,notnull" json:"displayName"` // 표시 이름
	FieldType   FieldType `bun:"field_type,notnull" json:"fieldType"`     // 필드 유형
	Required    bool      `bun:"required,notnull" json:"required"`        // 필수 여부
	Sortable    bool      `bun:"sortable,notnull" json:"sortable"`        // 정렬 가능 여부
	Searchable  bool      `bun:"searchable,notnull" json:"searchable"`    // 검색 가능 여부
	Options     string    `bun:"options" json:"options"`                  // 필드 옵션 (JSON 형식)
	SortOrder   int       `bun:"sort_order,notnull" json:"sortOrder"`     // 표시 순서
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`
	UpdatedAt   time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updatedAt"`

	// 관계
	Board *Board `bun:"rel:belongs-to,join:board_id=id" json:"board,omitempty"`
}

// 게시물 공통 속성 (각 동적 게시판에 공통으로 적용됨)
type PostCommon struct {
	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	Title     string    `bun:"title,notnull" json:"title"`
	Content   string    `bun:"content,notnull" json:"content"`
	UserID    int64     `bun:"user_id,notnull" json:"userId"`
	ViewCount int       `bun:"view_count,notnull,default:0" json:"viewCount"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updatedAt"`

	// 관계 (추가하지 않음 - 동적 테이블이므로)
}
