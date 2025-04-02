package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Attachment struct {
	bun.BaseModel `bun:"table:attachments,alias:a"`
	ID            int64     `bun:"id,pk,autoincrement" json:"id"`
	BoardID       int64     `bun:"board_id,notnull" json:"boardId"`
	PostID        int64     `bun:"post_id,notnull" json:"postId"`
	UserID        int64     `bun:"user_id,notnull" json:"userId"`
	FileName      string    `bun:"file_name,notnull" json:"fileName"`
	StorageName   string    `bun:"storage_name,notnull" json:"storageName"`
	FilePath      string    `bun:"file_path,notnull" json:"filePath"`
	FileSize      int64     `bun:"file_size,notnull" json:"fileSize"`
	MimeType      string    `bun:"mime_type,notnull" json:"mimeType"`
	IsImage       bool      `bun:"is_image,notnull" json:"isImage"`
	DownloadURL   string    `bun:"download_url,notnull" json:"downloadUrl"` // Not use
	DownloadCount int       `bun:"download_count,notnull,default:0" json:"downloadCount"`
	CreatedAt     time.Time `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`
}
