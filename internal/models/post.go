// internal/models/post.go (동적 게시물 생성용 인터페이스)
package models

import (
	"time"
)

// 동적 게시물 생성을 위한 필드 정보
type DynamicField struct {
	Name         string    // 필드 이름
	ColumnName   string    // DB 칼럼 이름
	Value        any       // 필드 값
	FieldType    FieldType // 필드 유형
	Required     bool      // 필수 여부
	ErrorMessage string    // 유효성 검사 오류 메시지
}

// 게시물 생성/조회에 사용할 동적 구조체
type DynamicPost struct {
	// 기본 필드
	ID        int64     // 게시물 ID
	Title     string    // 제목
	Content   string    // 내용
	UserID    int64     // 작성자 ID
	Username  string    // 작성자 이름 (조인 데이터)
	ViewCount int       // 조회수
	CreatedAt time.Time // 생성일
	UpdatedAt time.Time // 수정일

	// 동적 필드 (키-값 쌍으로 저장)
	Fields map[string]DynamicField // 필드 이름 -> 필드 값

	// 원본 데이터 (내부 사용)
	RawData map[string]any
}
