// internal/repository/attachment_repository.go
package repository

import (
	"context"
	"go-board/internal/models"

	"github.com/uptrace/bun"
)

// AttachmentRepository는 첨부 파일 정보에 대한 데이터 액세스를 제공합니다
type AttachmentRepository interface {
	Create(ctx context.Context, attachment *models.Attachment) error
	GetByID(ctx context.Context, id int64) (*models.Attachment, error)
	GetByPostID(ctx context.Context, boardID, postID int64) ([]*models.Attachment, error)
	Delete(ctx context.Context, id int64) error
	DeleteByPostID(ctx context.Context, boardID, postID int64) error
	IncrementDownloadCount(ctx context.Context, id int64) error
}

type attachmentRepository struct {
	db *bun.DB
}

func NewAttachmentRepository(db *bun.DB) AttachmentRepository {
	return &attachmentRepository{db: db}
}

// Create는 첨부 파일 정보를 저장합니다
func (r *attachmentRepository) Create(ctx context.Context, attachment *models.Attachment) error {
	_, err := r.db.NewInsert().Model(attachment).Exec(ctx)
	return err
}

// GetByID는 첨부 파일 정보를 조회합니다
func (r *attachmentRepository) GetByID(ctx context.Context, id int64) (*models.Attachment, error) {
	attachment := new(models.Attachment)
	err := r.db.NewSelect().Model(attachment).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return attachment, nil
}

// GetByPostID는 게시물의 첨부 파일 목록을 조회합니다
func (r *attachmentRepository) GetByPostID(ctx context.Context, boardID, postID int64) ([]*models.Attachment, error) {
	var attachments []*models.Attachment
	err := r.db.NewSelect().
		Model(&attachments).
		Where("board_id = ?", boardID).
		Where("post_id = ?", postID).
		Order("created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return attachments, nil
}

// Delete는 첨부 파일 정보를 삭제합니다
func (r *attachmentRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model((*models.Attachment)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

// DeleteByPostID는 게시물의 모든 첨부 파일 정보를 삭제합니다
func (r *attachmentRepository) DeleteByPostID(ctx context.Context, boardID, postID int64) error {
	_, err := r.db.NewDelete().
		Model((*models.Attachment)(nil)).
		Where("board_id = ?", boardID).
		Where("post_id = ?", postID).
		Exec(ctx)
	return err
}

// IncrementDownloadCount는 다운로드 카운트를 증가시킵니다
func (r *attachmentRepository) IncrementDownloadCount(ctx context.Context, id int64) error {
	_, err := r.db.NewUpdate().
		Model((*models.Attachment)(nil)).
		Set("download_count = download_count + 1").
		Where("id = ?", id).
		Exec(ctx)
	return err
}
