// internal/service/qna_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/edp1096/go-board/internal/models"
	"github.com/edp1096/go-board/internal/repository"
	"github.com/edp1096/go-board/internal/utils"

	"github.com/uptrace/bun"
)

var (
	ErrAnswerNotFound = errors.New("답변을 찾을 수 없음")
)

// QnAService는 Q&A 게시판 관련 서비스입니다.
type QnAService interface {
	// 답변 관련 메서드
	CreateAnswer(ctx context.Context, boardID, questionID, userID int64, content string) (*models.Answer, error)
	GetAnswersByQuestionID(ctx context.Context, boardID, questionID int64) ([]*models.Answer, error)
	GetAnswerByID(ctx context.Context, answerID int64) (*models.Answer, error)
	UpdateAnswer(ctx context.Context, answerID, userID int64, content string, isAdmin bool) (*models.Answer, error)
	DeleteAnswer(ctx context.Context, answerID, userID int64, isAdmin bool) error

	// 투표 관련 메서드
	GetQuestionVoteCount(ctx context.Context, boardID, questionID int64) (int, error)
	VoteQuestion(ctx context.Context, boardID, questionID, userID int64, value int) (int, error)
	VoteAnswer(ctx context.Context, answerID, userID int64, value int) (int, error)

	// 상태 관련 메서드
	UpdateQuestionStatus(ctx context.Context, boardID, questionID, userID int64, status string) error
	SetBestAnswer(ctx context.Context, boardID, questionID, answerID, userID int64) error

	// 답글 관련 메서드
	CreateAnswerReply(ctx context.Context, answerID, userID int64, content string) (*models.Answer, error)
}

type qnaService struct {
	db        *bun.DB
	boardRepo repository.BoardRepository
	boardSvc  BoardService
}

// NewQnAService는 QnAService의 새 인스턴스를 생성합니다.
func NewQnAService(db *bun.DB, boardRepo repository.BoardRepository, boardSvc BoardService) QnAService {
	return &qnaService{
		db:        db,
		boardRepo: boardRepo,
		boardSvc:  boardSvc,
	}
}

// CreateAnswer는 새 답변을 생성합니다.
func (s *qnaService) CreateAnswer(ctx context.Context, boardID, questionID, userID int64, content string) (*models.Answer, error) {
	// 답변 객체 생성
	now := time.Now()
	answer := &models.Answer{
		BoardID:    boardID,
		QuestionID: questionID,
		UserID:     userID,
		Content:    content,
		VoteCount:  0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// 트랜잭션 시작
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 답변 저장
	_, err = tx.NewInsert().Model(answer).Exec(ctx)
	if err != nil {
		return nil, err
	}

	// 질문의 답변 수 업데이트
	// 질문 가져오기
	post, err := s.boardSvc.GetPost(ctx, boardID, questionID)
	if err != nil {
		return nil, err
	}

	// answer_count 필드 값 업데이트
	answerCount := 1
	if field, ok := post.Fields["answer_count"]; ok && field.Value != nil {
		// 기존 값에 1 추가
		currentCount := utils.InterfaceToInt(field.Value)
		answerCount = currentCount + 1
	}

	// 게시물 업데이트 쿼리
	setClause := "answer_count = ?"
	params := []any{answerCount, questionID}

	// 게시판 정보 조회
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	// 질문 답변 수 업데이트 쿼리 실행
	_, err = tx.NewUpdate().
		Table(board.TableName).
		Set(setClause, params[0]).
		Where("id = ?", params[1]).
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	// 사용자 정보 조회하여 반환 결과에 포함
	var user models.User
	err = tx.NewSelect().
		Model(&user).
		Where("id = ?", userID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	// 트랜잭션 커밋
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 반환할 결과에 사용자 정보 포함
	answer.User = &user

	return answer, nil
}

// GetAnswersByQuestionID는 질문의 모든 답변과 답글을 조회합니다.
func (s *qnaService) GetAnswersByQuestionID(ctx context.Context, boardID, questionID int64) ([]*models.Answer, error) {
	// 게시물 가져오기 - 베스트 답변 ID 확인을 위해
	post, err := s.boardSvc.GetPost(ctx, boardID, questionID)
	if err != nil {
		return nil, err
	}

	// 베스트 답변 ID 가져오기
	var bestAnswerID int64 = 0
	if field, ok := post.Fields["best_answer_id"]; ok && field.Value != nil {
		bestAnswerID = utils.InterfaceToInt64(post.Fields["best_answer_id"].Value)
	}

	// 모든 답변 및 답글 조회
	var allAnswers []*models.Answer
	err = s.db.NewSelect().
		Model(&allAnswers).
		Relation("User").
		Where("board_id = ? AND question_id = ?", boardID, questionID).
		OrderExpr("CASE WHEN parent_id IS NULL THEN 0 ELSE 1 END, vote_count DESC, a.created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	// 답변과 답글 분리
	answers := make([]*models.Answer, 0)
	repliesMap := make(map[int64][]*models.Answer)

	for _, answer := range allAnswers {
		if answer.ParentID == nil {
			// 답변인 경우
			if answer.ID == bestAnswerID {
				answer.IsBestAnswer = true
			}
			answers = append(answers, answer)
		} else {
			// 답글인 경우
			parentID := *answer.ParentID
			repliesMap[parentID] = append(repliesMap[parentID], answer)
		}
	}

	// 답변에 답글 연결
	for _, answer := range answers {
		if replies, exists := repliesMap[answer.ID]; exists {
			answer.Children = replies
		}
	}

	return answers, nil
}

// GetAnswerByID는 ID로 답변을 조회합니다.
func (s *qnaService) GetAnswerByID(ctx context.Context, answerID int64) (*models.Answer, error) {
	answer := new(models.Answer)

	err := s.db.NewSelect().
		Model(answer).
		Relation("User").
		Where("a.id = ?", answerID).
		Scan(ctx)

	if err != nil {
		return nil, ErrAnswerNotFound
	}

	return answer, nil
}

// UpdateAnswer는 답변을 수정합니다.
func (s *qnaService) UpdateAnswer(ctx context.Context, answerID, userID int64, content string, isAdmin bool) (*models.Answer, error) {
	// 답변 조회
	answer, err := s.GetAnswerByID(ctx, answerID)
	if err != nil {
		return nil, err
	}

	// 권한 확인 (답변 작성자 또는 관리자만 수정 가능)
	if answer.UserID != userID && !isAdmin {
		return nil, ErrNoPermission
	}

	// 답변 수정
	answer.Content = content
	answer.UpdatedAt = time.Now()

	_, err = s.db.NewUpdate().
		Model(answer).
		Column("content", "updated_at").
		Where("id = ?", answerID).
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	return answer, nil
}

// DeleteAnswer는 답변을 삭제합니다.
func (s *qnaService) DeleteAnswer(ctx context.Context, answerID, userID int64, isAdmin bool) error {
	// 답변 조회
	answer, err := s.GetAnswerByID(ctx, answerID)
	if err != nil {
		return err
	}

	// 권한 확인 (답변 작성자 또는 관리자만 삭제 가능)
	if answer.UserID != userID && !isAdmin {
		return ErrNoPermission
	}

	// 트랜잭션 시작
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 답변이 베스트 답변인지 확인
	board, err := s.boardRepo.GetByID(ctx, answer.BoardID)
	if err != nil {
		return err
	}

	// 질문 조회
	post, err := s.boardSvc.GetPost(ctx, answer.BoardID, answer.QuestionID)
	if err != nil {
		return err
	}

	// 베스트 답변인 경우 표시 삭제
	var bestAnswerID int64 = 0
	if field, ok := post.Fields["best_answer_id"]; ok && field.Value != nil {
		bestAnswerID = utils.InterfaceToInt64(post.Fields["best_answer_id"].Value)
	}

	if bestAnswerID == answerID {
		// 베스트 답변 표시 제거
		_, err = tx.NewUpdate().
			Table(board.TableName).
			Set("best_answer_id = NULL").
			Where("id = ?", answer.QuestionID).
			Exec(ctx)
		if err != nil {
			return err
		}
	}

	// 답변 삭제
	_, err = tx.NewDelete().
		Model((*models.Answer)(nil)).
		Where("id = ?", answerID).
		Exec(ctx)
	if err != nil {
		return err
	}

	// 관련 투표 삭제
	_, err = tx.NewDelete().
		Model((*models.AnswerVote)(nil)).
		Where("answer_id = ?", answerID).
		Exec(ctx)
	if err != nil {
		return err
	}

	// 질문의 답변 수 업데이트 - answer_count 필드 값 계산
	answerCount := 0
	if field, ok := post.Fields["answer_count"]; ok && field.Value != nil {
		// 기존 값에서 1 감소
		currentCount := utils.InterfaceToInt(post.Fields["answer_count"].Value)
		answerCount = max(currentCount-1, 0)
	}

	// 질문 답변 수 업데이트 쿼리 실행
	_, err = tx.NewUpdate().
		Table(board.TableName).
		Set("answer_count = ?", answerCount).
		Where("id = ?", answer.QuestionID).
		Exec(ctx)

	if err != nil {
		return err
	}

	// 트랜잭션 커밋
	return tx.Commit()
}

// GetQuestionVoteCount는 질문의 현재 투표 수를 조회합니다.
func (s *qnaService) GetQuestionVoteCount(ctx context.Context, boardID, questionID int64) (int, error) {
	// 투표 수 계산
	var voteSum int
	err := s.db.NewSelect().
		Model((*models.QuestionVote)(nil)).
		ColumnExpr("COALESCE(SUM(value), 0) AS vote_sum").
		Where("question_id = ?", questionID).
		Scan(ctx, &voteSum)

	if err != nil {
		return 0, err
	}

	return voteSum, nil
}

// VoteQuestion은 질문에 투표합니다.
func (s *qnaService) VoteQuestion(ctx context.Context, boardID, questionID, userID int64, value int) (int, error) {
	// 질문 존재 여부 확인
	// post, err := s.boardSvc.GetPost(ctx, boardID, questionID)
	_, err := s.boardSvc.GetPost(ctx, boardID, questionID)
	if err != nil {
		return 0, fmt.Errorf("질문 조회 실패: %w", err)
	}

	// 트랜잭션 시작
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 이전 투표 기록 확인
	var existingVote models.QuestionVote
	err = tx.NewSelect().
		Model(&existingVote).
		Where("board_id = ? AND question_id = ? AND user_id = ?", boardID, questionID, userID).
		Scan(ctx)

	// var voteChange int
	if err == nil {
		// 이전 투표가 있는 경우
		if existingVote.Value == value {
			// 동일한 투표 취소
			_, err = tx.NewDelete().
				Model((*models.QuestionVote)(nil)).
				Where("id = ?", existingVote.ID).
				Exec(ctx)
			if err != nil {
				return 0, err
			}
			// voteChange = -value
		} else {
			// 다른 방향으로 투표 변경
			existingVote.Value = value
			existingVote.UpdatedAt = time.Now()

			_, err = tx.NewUpdate().
				Model(&existingVote).
				Column("value", "updated_at").
				Where("id = ?", existingVote.ID).
				Exec(ctx)
			if err != nil {
				return 0, err
			}
			// voteChange = value * 2 // 기존 값의 반대로 변경 (-1 → 1 또는 1 → -1)
		}
	} else {
		// 새 투표 생성
		vote := &models.QuestionVote{
			UserID:     userID,
			BoardID:    boardID,
			QuestionID: questionID,
			Value:      value,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		_, err = tx.NewInsert().
			Model(vote).
			Exec(ctx)
		if err != nil {
			return 0, err
		}
		// voteChange = value
	}

	// 질문 투표 수 계산
	var voteSum int
	err = tx.NewSelect().
		Model((*models.QuestionVote)(nil)).
		ColumnExpr("COALESCE(SUM(value), 0) AS vote_sum").
		Where("question_id = ?", questionID).
		Scan(ctx, &voteSum)
	if err != nil {
		return 0, err
	}

	// 게시판 정보 조회
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return 0, ErrBoardNotFound
	}

	// 질문의 vote_count 필드 업데이트
	_, err = s.db.NewUpdate().
		Table(board.TableName).
		Set("vote_count = ?", voteSum).
		Where("id = ?", questionID).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("투표 수 업데이트 실패: %w", err)
	}

	// 트랜잭션 커밋
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return voteSum, nil
}

// VoteAnswer는 답변에 투표합니다.
func (s *qnaService) VoteAnswer(ctx context.Context, answerID, userID int64, value int) (int, error) {
	// 답변 정보 조회
	answer, err := s.GetAnswerByID(ctx, answerID)
	if err != nil {
		return 0, err
	}

	// 트랜잭션 시작
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 이전 투표 기록 확인
	var existingVote models.AnswerVote
	err = tx.NewSelect().
		Model(&existingVote).
		Where("answer_id = ? AND user_id = ?",
			answerID, userID).
		Scan(ctx)

	var voteChange int
	if err == nil {
		// 이전 투표가 있는 경우
		if existingVote.Value == value {
			// 동일한 투표 취소
			_, err = tx.NewDelete().
				Model((*models.AnswerVote)(nil)).
				Where("id = ?", existingVote.ID).
				Exec(ctx)
			if err != nil {
				return 0, err
			}
			voteChange = -value
		} else {
			// 다른 방향으로 투표 변경
			existingVote.Value = value
			existingVote.UpdatedAt = time.Now()

			_, err = tx.NewUpdate().
				Model(&existingVote).
				Column("value", "updated_at").
				Where("id = ?", existingVote.ID).
				Exec(ctx)
			if err != nil {
				return 0, err
			}
			voteChange = value * 2 // 기존 값의 반대로 변경 (-1 → 1 또는 1 → -1)
		}
	} else {
		// 새 투표 생성
		vote := &models.AnswerVote{
			UserID:    userID,
			BoardID:   answer.BoardID,
			AnswerID:  answerID,
			Value:     value,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = tx.NewInsert().
			Model(vote).
			Exec(ctx)
		if err != nil {
			return 0, err
		}
		voteChange = value
	}

	// 답변 투표 수 업데이트
	newVoteCount := answer.VoteCount + voteChange
	_, err = tx.NewUpdate().
		Model((*models.Answer)(nil)).
		Set("vote_count = ?", newVoteCount).
		Where("id = ?", answerID).
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	// 트랜잭션 커밋
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return newVoteCount, nil
}

// UpdateQuestionStatus는 질문의 상태를 업데이트합니다.
func (s *qnaService) UpdateQuestionStatus(ctx context.Context, boardID, questionID, userID int64, status string) error {
	// 게시판 정보 조회
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return err
	}

	// 질문 조회
	post, err := s.boardSvc.GetPost(ctx, boardID, questionID)
	if err != nil {
		return err
	}

	// 권한 확인 (질문 작성자만 상태 변경 가능)
	if post.UserID != userID {
		return ErrNoPermission
	}

	// 상태 값 검증
	if status != "solved" && status != "unsolved" {
		return errors.New("유효하지 않은 상태입니다")
	}

	// 상태 업데이트
	_, err = s.db.NewUpdate().
		Table(board.TableName).
		Set("status = ?", status).
		Where("id = ?", questionID).
		Exec(ctx)

	return err
}

// SetBestAnswer는 베스트 답변을 설정합니다.
func (s *qnaService) SetBestAnswer(ctx context.Context, boardID, questionID, answerID, userID int64) error {
	// 게시판 정보 조회
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return err
	}

	// 질문 조회
	post, err := s.boardSvc.GetPost(ctx, boardID, questionID)
	if err != nil {
		return err
	}

	// 권한 확인 (질문 작성자만 베스트 답변 설정 가능)
	if post.UserID != userID {
		return ErrNoPermission
	}

	// 답변 존재 확인
	answer, err := s.GetAnswerByID(ctx, answerID)
	if err != nil {
		return err
	}

	// 답변이 실제로 이 질문에 속하는지 확인
	if answer.QuestionID != questionID {
		return errors.New("해당 답변은 이 질문에 속하지 않습니다")
	}

	// 트랜잭션 시작
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 베스트 답변 설정
	_, err = tx.NewUpdate().
		Table(board.TableName).
		Set("best_answer_id = ?", answerID).
		Set("status = ?", "solved"). // 자동으로 해결됨 상태로 변경
		Where("id = ?", questionID).
		Exec(ctx)
	if err != nil {
		return err
	}

	// 트랜잭션 커밋
	return tx.Commit()
}

// CreateAnswerReply는 답변에 대한 답글을 생성합니다.
func (s *qnaService) CreateAnswerReply(ctx context.Context, answerID, userID int64, content string) (*models.Answer, error) {
	// 부모 답변 조회
	parentAnswer, err := s.GetAnswerByID(ctx, answerID)
	if err != nil {
		return nil, ErrAnswerNotFound
	}

	// 이미 답글인 경우 거부 (중첩 답글 방지)
	if parentAnswer.ParentID != nil {
		return nil, errors.New("답글에 대한 답글은 작성할 수 없습니다")
	}

	// 답글 객체 생성
	now := time.Now()
	reply := &models.Answer{
		BoardID:    parentAnswer.BoardID,
		QuestionID: parentAnswer.QuestionID,
		UserID:     userID,
		Content:    content,
		ParentID:   &answerID, // 부모 답변 ID 설정
		VoteCount:  0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// 트랜잭션 시작
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 답글 저장
	_, err = tx.NewInsert().Model(reply).Exec(ctx)
	if err != nil {
		return nil, err
	}

	// 사용자 정보 조회하여 반환 결과에 포함
	var user models.User
	err = tx.NewSelect().
		Model(&user).
		Where("id = ?", userID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	// 트랜잭션 커밋
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 반환할 결과에 사용자 정보 포함
	reply.User = &user

	return reply, nil
}
