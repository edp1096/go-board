// internal/repository/comment_vote_repository.go
package repository

import (
	"context"

	"github.com/edp1096/go-board/internal/models"
	"github.com/uptrace/bun"
)

type CommentVoteRepository interface {
	Create(ctx context.Context, vote *models.CommentVote) error
	GetByCommentAndUser(ctx context.Context, commentID, userID int64) (*models.CommentVote, error)
	Update(ctx context.Context, vote *models.CommentVote) error
	Delete(ctx context.Context, id int64) error
	CountByComment(ctx context.Context, commentID int64, value int) (int, error)
	UpdateCommentVoteCount(ctx context.Context, commentID int64) error
	GetVoteStatuses(ctx context.Context, commentIDs []int64, userID int64) (map[int64]int, error)
}

type commentVoteRepository struct {
	db *bun.DB
}

func NewCommentVoteRepository(db *bun.DB) CommentVoteRepository {
	return &commentVoteRepository{db: db}
}

func (r *commentVoteRepository) Create(ctx context.Context, vote *models.CommentVote) error {
	_, err := r.db.NewInsert().Model(vote).Exec(ctx)
	return err
}

func (r *commentVoteRepository) GetByCommentAndUser(ctx context.Context, commentID, userID int64) (*models.CommentVote, error) {
	vote := new(models.CommentVote)
	err := r.db.NewSelect().
		Model(vote).
		Where("comment_id = ? AND user_id = ?", commentID, userID).
		Scan(ctx)

	if err != nil {
		return nil, err
	}
	return vote, nil
}

func (r *commentVoteRepository) Update(ctx context.Context, vote *models.CommentVote) error {
	_, err := r.db.NewUpdate().
		Model(vote).
		Set("value = ?", vote.Value).
		Set("updated_at = ?", vote.UpdatedAt).
		Where("id = ?", vote.ID).
		Exec(ctx)

	return err
}

func (r *commentVoteRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().
		Model((*models.CommentVote)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

func (r *commentVoteRepository) CountByComment(ctx context.Context, commentID int64, value int) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.CommentVote)(nil)).
		Where("comment_id = ? AND value = ?", commentID, value).
		Count(ctx)

	return count, err
}

func (r *commentVoteRepository) UpdateCommentVoteCount(ctx context.Context, commentID int64) error {
	// 좋아요 수 계산
	likeCount, err := r.CountByComment(ctx, commentID, 1)
	if err != nil {
		return err
	}

	// 싫어요 수 계산
	dislikeCount, err := r.CountByComment(ctx, commentID, -1)
	if err != nil {
		return err
	}

	// 댓글 테이블 업데이트
	_, err = r.db.NewUpdate().
		Model((*models.Comment)(nil)).
		Set("like_count = ?", likeCount).
		Set("dislike_count = ?", dislikeCount).
		Where("id = ?", commentID).
		Exec(ctx)

	return err
}

func (r *commentVoteRepository) GetVoteStatuses(ctx context.Context, commentIDs []int64, userID int64) (map[int64]int, error) {
	var votes []*models.CommentVote
	err := r.db.NewSelect().
		Model(&votes).
		Where("comment_id IN (?) AND user_id = ?", bun.In(commentIDs), userID).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	result := make(map[int64]int)
	for _, vote := range votes {
		result[vote.CommentID] = vote.Value
	}

	return result, nil
}
