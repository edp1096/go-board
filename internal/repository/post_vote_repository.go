// internal/repository/post_vote_repository.go
package repository

import (
	"context"
	"fmt"

	"github.com/edp1096/go-board/internal/models"
	"github.com/edp1096/go-board/internal/utils"
	"github.com/uptrace/bun"
)

type PostVoteRepository interface {
	Create(ctx context.Context, vote *models.PostVote) error
	GetByPostAndUser(ctx context.Context, postID, userID int64) (*models.PostVote, error)
	Update(ctx context.Context, vote *models.PostVote) error
	Delete(ctx context.Context, id int64) error
	CountByPost(ctx context.Context, postID int64, value int) (int, error)
	UpdatePostVoteCount(ctx context.Context, boardID, postID int64) error
}

type postVoteRepository struct {
	db *bun.DB
}

func NewPostVoteRepository(db *bun.DB) PostVoteRepository {
	return &postVoteRepository{db: db}
}

func (r *postVoteRepository) Create(ctx context.Context, vote *models.PostVote) error {
	_, err := r.db.NewInsert().Model(vote).Exec(ctx)
	return err
}

func (r *postVoteRepository) GetByPostAndUser(ctx context.Context, postID, userID int64) (*models.PostVote, error) {
	vote := new(models.PostVote)
	err := r.db.NewSelect().
		Model(vote).
		Where("post_id = ? AND user_id = ?", postID, userID).
		Scan(ctx)

	if err != nil {
		return nil, err
	}
	return vote, nil
}

func (r *postVoteRepository) Update(ctx context.Context, vote *models.PostVote) error {
	_, err := r.db.NewUpdate().
		Model(vote).
		Set("value = ?", vote.Value).
		Set("updated_at = ?", vote.UpdatedAt).
		Where("id = ?", vote.ID).
		Exec(ctx)

	return err
}

func (r *postVoteRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().
		Model((*models.PostVote)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

func (r *postVoteRepository) CountByPost(ctx context.Context, postID int64, value int) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.PostVote)(nil)).
		Where("post_id = ? AND value = ?", postID, value).
		Count(ctx)

	return count, err
}

func (r *postVoteRepository) UpdatePostVoteCount(ctx context.Context, boardID, postID int64) error {
	// 게시판 정보 조회
	var board models.Board
	err := r.db.NewSelect().
		Model(&board).
		Where("id = ?", boardID).
		Scan(ctx)

	if err != nil {
		return err
	}

	// 좋아요 수 계산
	likeCount, err := r.CountByPost(ctx, postID, 1)
	if err != nil {
		return err
	}

	// 싫어요 수 계산
	dislikeCount, err := r.CountByPost(ctx, postID, -1)
	if err != nil {
		return err
	}

	// 게시물 테이블 업데이트
	tableName := board.TableName
	var query string

	if utils.IsPostgres(r.db) || utils.IsSQLite(r.db) {
		// PostgreSQL, SQLite
		query = fmt.Sprintf(`UPDATE "%s" SET like_count = ?, dislike_count = ? WHERE id = ?`, tableName)
	} else {
		// MySQL, MariaDB
		query = fmt.Sprintf(`UPDATE `+"`%s`"+` SET like_count = ?, dislike_count = ? WHERE id = ?`, tableName)
	}

	_, err = r.db.Exec(query, likeCount, dislikeCount, postID)
	return err
}
