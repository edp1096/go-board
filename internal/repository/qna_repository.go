// internal/repository/qna_repository.go
package repository

import (
	"context"

	"github.com/edp1096/toy-board/internal/models"
	"github.com/uptrace/bun"
)

// QnARepository는 Q&A 게시판 관련 저장소 인터페이스입니다.
type QnARepository interface {
	// 답변 관련 메서드
	CreateAnswer(ctx context.Context, answer *models.Answer) error
	GetAnswersByQuestionID(ctx context.Context, boardID, questionID int64, showIP bool) ([]*models.Answer, error)
	GetAnswerByID(ctx context.Context, answerID int64) (*models.Answer, error)
	UpdateAnswer(ctx context.Context, answer *models.Answer) error
	DeleteAnswer(ctx context.Context, answerID int64) error

	// 투표 관련 메서드
	GetQuestionVoteCount(ctx context.Context, questionID int64) (int, error)
	CreateQuestionVote(ctx context.Context, vote *models.QuestionVote) error
	GetQuestionVote(ctx context.Context, boardID, questionID, userID int64) (*models.QuestionVote, error)
	UpdateQuestionVote(ctx context.Context, vote *models.QuestionVote) error
	DeleteQuestionVote(ctx context.Context, voteID int64) error

	// 답변 투표 관련 메서드
	GetAnswerVote(ctx context.Context, answerID, userID int64) (*models.AnswerVote, error)
	CreateAnswerVote(ctx context.Context, vote *models.AnswerVote) error
	UpdateAnswerVote(ctx context.Context, vote *models.AnswerVote) error
	DeleteAnswerVote(ctx context.Context, voteID int64) error

	// 기타 메서드
	UpdateQuestionStatus(ctx context.Context, boardID, questionID int64, status string) error
	SetBestAnswer(ctx context.Context, boardID, questionID, answerID int64) error
	UpdateQuestionVoteCount(ctx context.Context, boardID, questionID int64, voteSum int) error
	UpdateAnswerVoteCount(ctx context.Context, answerID int64, newVoteCount int) error

	// 트랜잭션 관련
	BeginTx(ctx context.Context) (bun.Tx, error)
}

type qnaRepository struct {
	db *bun.DB
}

// NewQnARepository는 QnARepository의 새 인스턴스를 생성합니다.
func NewQnARepository(db *bun.DB) QnARepository {
	return &qnaRepository{
		db: db,
	}
}

func (r *qnaRepository) CreateAnswer(ctx context.Context, answer *models.Answer) error {
	_, err := r.db.NewInsert().Model(answer).Exec(ctx)
	return err
}

func (r *qnaRepository) GetAnswersByQuestionID(ctx context.Context, boardID, questionID int64, showIP bool) ([]*models.Answer, error) {
	var allAnswers []*models.Answer
	query := r.db.NewSelect().
		Model(&allAnswers).
		Relation("User").
		Where("board_id = ? AND question_id = ?", boardID, questionID).
		OrderExpr("CASE WHEN parent_id IS NULL THEN 0 ELSE 1 END, vote_count DESC, a.created_at ASC")

	if !showIP {
		query = query.ExcludeColumn("ip_address")
	}

	err := query.Scan(ctx)
	return allAnswers, err
}

func (r *qnaRepository) GetAnswerByID(ctx context.Context, answerID int64) (*models.Answer, error) {
	answer := new(models.Answer)
	err := r.db.NewSelect().
		Model(answer).
		Relation("User").
		Where("a.id = ?", answerID).
		Scan(ctx)
	return answer, err
}

func (r *qnaRepository) UpdateAnswer(ctx context.Context, answer *models.Answer) error {
	_, err := r.db.NewUpdate().
		Model(answer).
		Column("content", "updated_at").
		Where("id = ?", answer.ID).
		Exec(ctx)
	return err
}

func (r *qnaRepository) DeleteAnswer(ctx context.Context, answerID int64) error {
	_, err := r.db.NewDelete().
		Model((*models.Answer)(nil)).
		Where("id = ?", answerID).
		Exec(ctx)
	return err
}

func (r *qnaRepository) GetQuestionVoteCount(ctx context.Context, questionID int64) (int, error) {
	var voteSum int
	err := r.db.NewSelect().
		Model((*models.QuestionVote)(nil)).
		ColumnExpr("COALESCE(SUM(value), 0) AS vote_sum").
		Where("question_id = ?", questionID).
		Scan(ctx, &voteSum)
	return voteSum, err
}

func (r *qnaRepository) GetQuestionVote(ctx context.Context, boardID, questionID, userID int64) (*models.QuestionVote, error) {
	var vote models.QuestionVote
	err := r.db.NewSelect().
		Model(&vote).
		Where("board_id = ? AND question_id = ? AND user_id = ?", boardID, questionID, userID).
		Scan(ctx)
	return &vote, err
}

func (r *qnaRepository) CreateQuestionVote(ctx context.Context, vote *models.QuestionVote) error {
	_, err := r.db.NewInsert().
		Model(vote).
		Exec(ctx)
	return err
}

func (r *qnaRepository) UpdateQuestionVote(ctx context.Context, vote *models.QuestionVote) error {
	_, err := r.db.NewUpdate().
		Model(vote).
		Column("value", "updated_at").
		Where("id = ?", vote.ID).
		Exec(ctx)
	return err
}

func (r *qnaRepository) DeleteQuestionVote(ctx context.Context, voteID int64) error {
	_, err := r.db.NewDelete().
		Model((*models.QuestionVote)(nil)).
		Where("id = ?", voteID).
		Exec(ctx)
	return err
}

func (r *qnaRepository) GetAnswerVote(ctx context.Context, answerID, userID int64) (*models.AnswerVote, error) {
	var vote models.AnswerVote
	err := r.db.NewSelect().
		Model(&vote).
		Where("answer_id = ? AND user_id = ?", answerID, userID).
		Scan(ctx)
	return &vote, err
}

func (r *qnaRepository) CreateAnswerVote(ctx context.Context, vote *models.AnswerVote) error {
	_, err := r.db.NewInsert().
		Model(vote).
		Exec(ctx)
	return err
}

func (r *qnaRepository) UpdateAnswerVote(ctx context.Context, vote *models.AnswerVote) error {
	_, err := r.db.NewUpdate().
		Model(vote).
		Column("value", "updated_at").
		Where("id = ?", vote.ID).
		Exec(ctx)
	return err
}

func (r *qnaRepository) DeleteAnswerVote(ctx context.Context, voteID int64) error {
	_, err := r.db.NewDelete().
		Model((*models.AnswerVote)(nil)).
		Where("id = ?", voteID).
		Exec(ctx)
	return err
}

func (r *qnaRepository) UpdateQuestionStatus(ctx context.Context, boardID, questionID int64, status string) error {
	_, err := r.db.NewUpdate().
		Table(getBoardTableName(ctx, r.db, boardID)).
		Set("status = ?", status).
		Where("id = ?", questionID).
		Exec(ctx)
	return err
}

func (r *qnaRepository) SetBestAnswer(ctx context.Context, boardID, questionID, answerID int64) error {
	_, err := r.db.NewUpdate().
		Table(getBoardTableName(ctx, r.db, boardID)).
		Set("best_answer_id = ?", answerID).
		Set("status = ?", "solved").
		Where("id = ?", questionID).
		Exec(ctx)
	return err
}

func (r *qnaRepository) UpdateQuestionVoteCount(ctx context.Context, boardID, questionID int64, voteSum int) error {
	_, err := r.db.NewUpdate().
		Table(getBoardTableName(ctx, r.db, boardID)).
		Set("vote_count = ?", voteSum).
		Where("id = ?", questionID).
		Exec(ctx)
	return err
}

func (r *qnaRepository) UpdateAnswerVoteCount(ctx context.Context, answerID int64, newVoteCount int) error {
	_, err := r.db.NewUpdate().
		Model((*models.Answer)(nil)).
		Set("vote_count = ?", newVoteCount).
		Where("id = ?", answerID).
		Exec(ctx)
	return err
}

func (r *qnaRepository) BeginTx(ctx context.Context) (bun.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

// 유틸리티 함수: 게시판 테이블 이름 가져오기
func getBoardTableName(ctx context.Context, db *bun.DB, boardID int64) string {
	var board struct {
		TableName string `bun:"table_name"`
	}

	_ = db.NewSelect().
		Table("boards").
		Column("table_name").
		Where("id = ?", boardID).
		Scan(ctx, &board)

	return board.TableName
}
