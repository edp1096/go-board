package repository

import (
	"context"
	"log"

	"github.com/edp1096/go-board/internal/models"
	"github.com/uptrace/bun"
)

type BoardParticipantRepository interface {
	Create(ctx context.Context, participant *models.BoardParticipant) error
	GetByUserAndBoard(ctx context.Context, userID, boardID int64) (*models.BoardParticipant, error)
	GetParticipantsByBoardID(ctx context.Context, boardID int64) ([]*models.BoardParticipant, error)
	GetBoardsByUserID(ctx context.Context, userID int64) ([]*models.Board, error)
	Update(ctx context.Context, participant *models.BoardParticipant) error
	Delete(ctx context.Context, boardID, userID int64) error
	DeleteByBoardID(ctx context.Context, boardID int64) error
}

type boardParticipantRepository struct {
	db *bun.DB
}

func NewBoardParticipantRepository(db *bun.DB) BoardParticipantRepository {
	return &boardParticipantRepository{db: db}
}

func (r *boardParticipantRepository) Create(ctx context.Context, participant *models.BoardParticipant) error {
	_, err := r.db.NewInsert().Model(participant).Exec(ctx)
	return err
}

func (r *boardParticipantRepository) GetByUserAndBoard(ctx context.Context, userID, boardID int64) (*models.BoardParticipant, error) {
	participant := new(models.BoardParticipant)
	err := r.db.NewSelect().
		Model(participant).
		Where("user_id = ? AND board_id = ?", userID, boardID).
		Scan(ctx)

	return participant, err
}

func (r *boardParticipantRepository) GetParticipantsByBoardID(ctx context.Context, boardID int64) ([]*models.BoardParticipant, error) {
	var participants []*models.BoardParticipant
	query := r.db.NewSelect().
		Model(&participants).
		Relation("User").
		Where("board_id = ?", boardID).
		OrderExpr("bp.created_at ASC")
	err := query.Scan(ctx)

	if err != nil {
		log.Println(err)
	}

	return participants, err
}

func (r *boardParticipantRepository) GetBoardsByUserID(ctx context.Context, userID int64) ([]*models.Board, error) {
	var boards []*models.Board
	err := r.db.NewSelect().
		Model(&boards).
		Join("JOIN board_participants AS bp ON bp.board_id = b.id").
		Where("bp.user_id = ?", userID).
		Scan(ctx)

	return boards, err
}

func (r *boardParticipantRepository) Update(ctx context.Context, participant *models.BoardParticipant) error {
	query := r.db.NewUpdate().
		Model(participant).
		Column("role").
		Where("board_id = ? AND user_id = ?", participant.BoardID, participant.UserID)
	_, err := query.Exec(ctx)

	return err
}

func (r *boardParticipantRepository) Delete(ctx context.Context, boardID, userID int64) error {
	_, err := r.db.NewDelete().
		Model((*models.BoardParticipant)(nil)).
		Where("board_id = ? AND user_id = ?", boardID, userID).
		Exec(ctx)

	return err
}

func (r *boardParticipantRepository) DeleteByBoardID(ctx context.Context, boardID int64) error {
	_, err := r.db.NewDelete().
		Model((*models.BoardParticipant)(nil)).
		Where("board_id = ?", boardID).
		Exec(ctx)

	return err
}
