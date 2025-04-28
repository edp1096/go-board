// internal/service/comment_vote_service.go
package service

import (
	"context"
	"errors"
	"time"

	"github.com/edp1096/go-board/internal/models"
	"github.com/edp1096/go-board/internal/repository"
)

type CommentVoteService interface {
	VoteComment(ctx context.Context, boardID, commentID, userID int64, value int) (likes int, dislikes int, userVote int, err error)
	GetCommentVoteStatus(ctx context.Context, commentID, userID int64) (int, error)
	GetCommentVoteCounts(ctx context.Context, commentID int64) (likes int, dislikes int, err error)
	GetMultipleCommentVoteStatuses(ctx context.Context, commentIDs []int64, userID int64) (map[int64]int, error)
}

type commentVoteService struct {
	commentVoteRepo repository.CommentVoteRepository
	boardService    BoardService
	commentRepo     repository.CommentRepository
}

func NewCommentVoteService(commentVoteRepo repository.CommentVoteRepository, boardService BoardService, commentRepo repository.CommentRepository) CommentVoteService {
	return &commentVoteService{
		commentVoteRepo: commentVoteRepo,
		boardService:    boardService,
		commentRepo:     commentRepo,
	}
}

func (s *commentVoteService) VoteComment(ctx context.Context, boardID, commentID, userID int64, value int) (int, int, int, error) {
	// 투표 값 유효성 검사
	if value != 1 && value != -1 && value != 0 {
		return 0, 0, 0, ErrInvalidVoteValue
	}

	// 게시판 설정 확인
	board, err := s.boardService.GetBoardByID(ctx, boardID)
	if err != nil {
		return 0, 0, 0, err
	}

	if !board.VotesEnabled {
		return 0, 0, 0, ErrVotesDisabled
	}

	// 댓글 존재 확인
	// comment, err := s.commentRepo.GetByID(ctx, commentID)
	_, err = s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return 0, 0, 0, errors.New("댓글을 찾을 수 없습니다")
	}

	// 기존 투표 확인
	existingVote, err := s.commentVoteRepo.GetByCommentAndUser(ctx, commentID, userID)
	now := time.Now()

	if err == nil { // 기존 투표가 있는 경우
		if value == 0 { // 투표 취소
			err = s.commentVoteRepo.Delete(ctx, existingVote.ID)
		} else if existingVote.Value != value { // 투표 변경
			existingVote.Value = value
			existingVote.UpdatedAt = now
			err = s.commentVoteRepo.Update(ctx, existingVote)
		} else { // 동일한 투표 (취소)
			err = s.commentVoteRepo.Delete(ctx, existingVote.ID)
			value = 0 // 투표 취소로 간주
		}
	} else { // 새 투표
		if value == 0 { // 취소 요청은 무시
			return 0, 0, 0, nil
		}

		newVote := &models.CommentVote{
			CommentID: commentID,
			BoardID:   boardID,
			UserID:    userID,
			Value:     value,
			CreatedAt: now,
			UpdatedAt: now,
		}
		err = s.commentVoteRepo.Create(ctx, newVote)
	}

	if err != nil {
		return 0, 0, 0, err
	}

	// 투표 수 업데이트
	err = s.commentVoteRepo.UpdateCommentVoteCount(ctx, commentID)
	if err != nil {
		return 0, 0, 0, err
	}

	// 현재 투표 수 조회
	likes, dislikes, err := s.GetCommentVoteCounts(ctx, commentID)
	if err != nil {
		return 0, 0, 0, err
	}

	return likes, dislikes, value, nil
}

func (s *commentVoteService) GetCommentVoteStatus(ctx context.Context, commentID, userID int64) (int, error) {
	if userID == 0 { // 로그인하지 않은 경우
		return 0, nil
	}

	vote, err := s.commentVoteRepo.GetByCommentAndUser(ctx, commentID, userID)
	if err != nil {
		return 0, nil // 투표 없음
	}

	return vote.Value, nil
}

func (s *commentVoteService) GetCommentVoteCounts(ctx context.Context, commentID int64) (int, int, error) {
	likes, err := s.commentVoteRepo.CountByComment(ctx, commentID, 1)
	if err != nil {
		return 0, 0, err
	}

	dislikes, err := s.commentVoteRepo.CountByComment(ctx, commentID, -1)
	if err != nil {
		return 0, 0, err
	}

	return likes, dislikes, nil
}

func (s *commentVoteService) GetMultipleCommentVoteStatuses(ctx context.Context, commentIDs []int64, userID int64) (map[int64]int, error) {
	if userID == 0 { // 로그인하지 않은 경우
		return make(map[int64]int), nil
	}

	return s.commentVoteRepo.GetVoteStatuses(ctx, commentIDs, userID)
}
