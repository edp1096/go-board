// internal/repository/comment_repository.go
package repository

import (
	"context"

	"github.com/edp1096/go-board/internal/models"

	"github.com/uptrace/bun"
)

// CommentRepository 인터페이스
type CommentRepository interface {
	Create(ctx context.Context, comment *models.Comment) error
	GetByID(ctx context.Context, id int64) (*models.Comment, error)
	GetByPostID(ctx context.Context, boardID, postID int64, includeReplies bool) ([]*models.Comment, error)
	Update(ctx context.Context, comment *models.Comment) error
	Delete(ctx context.Context, id int64) error
	DeleteByPostID(ctx context.Context, boardID, postID int64) error
}

// commentRepository 구현체
type commentRepository struct {
	db *bun.DB
}

// 새 CommentRepository 생성
func NewCommentRepository(db *bun.DB) CommentRepository {
	return &commentRepository{db: db}
}

// Create - 새 댓글 생성
func (r *commentRepository) Create(ctx context.Context, comment *models.Comment) error {
	_, err := r.db.NewInsert().Model(comment).Exec(ctx)
	return err
}

// GetByID - ID로 댓글 조회
func (r *commentRepository) GetByID(ctx context.Context, id int64) (*models.Comment, error) {
	comment := new(models.Comment)
	err := r.db.NewSelect().
		Model(comment).
		Relation("User").
		Where("c.id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

// GetByPostID - 게시물 ID로 댓글 목록 조회
func (r *commentRepository) GetByPostID(ctx context.Context, boardID, postID int64, includeReplies bool) ([]*models.Comment, error) {
	var comments []*models.Comment

	query := r.db.NewSelect().
		Model(&comments).
		Relation("User").
		Where("c.board_id = ?", boardID).
		Where("c.post_id = ?", postID)

	if !includeReplies {
		// 최상위 댓글만 조회
		query = query.Where("c.parent_id IS NULL")
	}

	err := query.Order("c.created_at ASC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	if includeReplies {
		// 최상위 댓글과 답글을 구분하여 계층 구조 생성
		parentComments := make([]*models.Comment, 0)
		childrenMap := make(map[int64][]*models.Comment)

		for _, comment := range comments {
			if comment.ParentID == nil {
				// 최상위 댓글
				parentComments = append(parentComments, comment)
			} else {
				// 답글
				parentID := *comment.ParentID
				childrenMap[parentID] = append(childrenMap[parentID], comment)
			}
		}

		// 각 최상위 댓글에 답글 연결
		for _, parent := range parentComments {
			if children, exists := childrenMap[parent.ID]; exists {
				parent.Children = children
			}
		}

		return parentComments, nil
	}

	return comments, nil
}

// Update - 댓글 수정
func (r *commentRepository) Update(ctx context.Context, comment *models.Comment) error {
	_, err := r.db.NewUpdate().
		Model(comment).
		Column("content", "updated_at").
		WherePK().
		Exec(ctx)
	return err
}

// Delete - 댓글 삭제
func (r *commentRepository) Delete(ctx context.Context, id int64) error {
	// 먼저 이 댓글에 달린 답글들을 삭제
	_, err := r.db.NewDelete().
		Model((*models.Comment)(nil)).
		Where("parent_id = ?", id).
		Exec(ctx)
	if err != nil {
		return err
	}

	// 그 후 해당 댓글 삭제
	_, err = r.db.NewDelete().
		Model((*models.Comment)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// DeleteByPostID - 게시물에 속한 모든 댓글 삭제
func (r *commentRepository) DeleteByPostID(ctx context.Context, boardID, postID int64) error {
	_, err := r.db.NewDelete().
		Model((*models.Comment)(nil)).
		Where("board_id = ?", boardID).
		Where("post_id = ?", postID).
		Exec(ctx)
	return err
}
