// internal/service/qna_service.go
package service

import (
	"context"
	"errors"
	"go-board/internal/models"
	"go-board/internal/repository"
	"time"

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
	VoteQuestion(ctx context.Context, boardID, questionID, userID int64, value int) (int, error)
	VoteAnswer(ctx context.Context, answerID, userID int64, value int) (int, error)

	// 상태 관련 메서드
	UpdateQuestionStatus(ctx context.Context, boardID, questionID, userID int64, status string) error
	SetBestAnswer(ctx context.Context, boardID, questionID, answerID, userID int64) error
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
	// 게시판 정보 조회
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	// 질문 가져오기
	post, err := s.boardSvc.GetPost(ctx, boardID, questionID)
	if err != nil {
		return nil, err
	}

	// answer_count 필드 값 업데이트
	answerCount := 1
	if field, ok := post.Fields["answer_count"]; ok && field.Value != nil {
		// 기존 값에 1 추가
		var currentCount int
		switch val := post.Fields["answer_count"].Value.(type) {
		case int:
			currentCount = val
		case float64:
			currentCount = int(val)
		case string:
			// 문자열을 숫자로 변환 시도 (필요한 경우)
		}
		answerCount = currentCount + 1
	}

	// 게시물 업데이트 쿼리
	setClause := "answer_count = ?"
	params := []interface{}{answerCount, questionID}

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

// GetAnswersByQuestionID는 질문의 모든 답변을 조회합니다.
func (s *qnaService) GetAnswersByQuestionID(ctx context.Context, boardID, questionID int64) ([]*models.Answer, error) {
	// 게시물 가져오기 - 베스트 답변 ID 확인을 위해
	post, err := s.boardSvc.GetPost(ctx, boardID, questionID)
	if err != nil {
		return nil, err
	}

	// 베스트 답변 ID 가져오기
	var bestAnswerID int64 = 0
	if field, ok := post.Fields["best_answer_id"]; ok && field.Value != nil {
		switch val := post.Fields["best_answer_id"].Value.(type) {
		case int:
			bestAnswerID = int64(val)
		case int64:
			bestAnswerID = val
		case float64:
			bestAnswerID = int64(val)
		}
	}

	// 답변 목록 조회
	var answers []*models.Answer
	err = s.db.NewSelect().
		Model(&answers).
		Relation("User").
		Where("board_id = ? AND question_id = ?", boardID, questionID).
		OrderExpr("vote_count DESC, created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	// 베스트 답변 표시
	for _, answer := range answers {
		if answer.ID == bestAnswerID {
			answer.IsBestAnswer = true
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
		Where("id = ?", answerID).
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
		switch val := post.Fields["best_answer_id"].Value.(type) {
		case int:
			bestAnswerID = int64(val)
		case int64:
			bestAnswerID = val
		case float64:
			bestAnswerID = int64(val)
		}
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
		Model((*models.Vote)(nil)).
		Where("target_id = ? AND target_type = ?", answerID, "answer").
		Exec(ctx)
	if err != nil {
		return err
	}

	// 질문의 답변 수 업데이트
	// answer_count 필드 값 계산
	answerCount := 0
	if field, ok := post.Fields["answer_count"]; ok && field.Value != nil {
		// 기존 값에서 1 감소
		var currentCount int
		switch val := post.Fields["answer_count"].Value.(type) {
		case int:
			currentCount = val
		case float64:
			currentCount = int(val)
		case string:
			// 문자열을 숫자로 변환 시도 (필요한 경우)
		}
		answerCount = currentCount - 1
		if answerCount < 0 {
			answerCount = 0
		}
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

// VoteQuestion은 질문에 투표합니다.
func (s *qnaService) VoteQuestion(ctx context.Context, boardID, questionID, userID int64, value int) (int, error) {
	return s.handleVote(ctx, boardID, questionID, userID, "question", value)
}

// VoteAnswer는 답변에 투표합니다.
func (s *qnaService) VoteAnswer(ctx context.Context, answerID, userID int64, value int) (int, error) {
	// 답변 정보 조회
	answer, err := s.GetAnswerByID(ctx, answerID)
	if err != nil {
		return 0, err
	}

	return s.handleVote(ctx, answer.BoardID, answerID, userID, "answer", value)
}

// handleVote는 투표 처리를 담당하는 공통 함수입니다.
func (s *qnaService) handleVote(ctx context.Context, boardID, targetID, userID int64, targetType string, value int) (int, error) {
	// 트랜잭션 시작
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 이전 투표 기록 확인
	var existingVote models.Vote
	err = tx.NewSelect().
		Model(&existingVote).
		Where("board_id = ? AND target_id = ? AND target_type = ? AND user_id = ?",
			boardID, targetID, targetType, userID).
		Scan(ctx)

	var voteChange int
	if err == nil {
		// 이전 투표가 있는 경우
		if existingVote.Value == value {
			// 동일한 투표 취소
			_, err = tx.NewDelete().
				Model((*models.Vote)(nil)).
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
		vote := &models.Vote{
			UserID:     userID,
			BoardID:    boardID,
			TargetID:   targetID,
			TargetType: targetType,
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
		voteChange = value
	}

	// 대상 투표 수 업데이트
	var currentCount int
	var newCount int
	var updateQuery *bun.UpdateQuery

	if targetType == "question" {
		// 게시판 정보 조회
		board, err := s.boardRepo.GetByID(ctx, boardID)
		if err != nil {
			return 0, err
		}

		// 질문 조회
		post, err := s.boardSvc.GetPost(ctx, boardID, targetID)
		if err != nil {
			return 0, err
		}

		// 현재 투표 수 가져오기
		if field, ok := post.Fields["vote_count"]; ok && field.Value != nil {
			switch val := post.Fields["vote_count"].Value.(type) {
			case int:
				currentCount = val
			case int64:
				currentCount = int(val)
			case float64:
				currentCount = int(val)
			}
		}

		// 새 투표 수 계산
		newCount = currentCount + voteChange

		// 업데이트 쿼리 생성
		updateQuery = tx.NewUpdate().
			Table(board.TableName).
			Set("vote_count = ?", newCount).
			Where("id = ?", targetID)
	} else {
		// 답변에 대한 투표 수 업데이트
		var answer models.Answer
		err = tx.NewSelect().
			Model(&answer).
			Column("vote_count").
			Where("id = ?", targetID).
			Scan(ctx)
		if err != nil {
			return 0, err
		}

		// 현재 투표 수 가져오기
		currentCount = answer.VoteCount

		// 새 투표 수 계산
		newCount = currentCount + voteChange

		// 업데이트 쿼리 생성
		updateQuery = tx.NewUpdate().
			Model((*models.Answer)(nil)).
			Set("vote_count = ?", newCount).
			Where("id = ?", targetID)
	}

	// 업데이트 쿼리 실행
	_, err = updateQuery.Exec(ctx)
	if err != nil {
		return 0, err
	}

	// 트랜잭션 커밋
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return newCount, nil
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
