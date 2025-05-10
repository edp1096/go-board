// internal/service/post_vote_service.go
package service

import (
	"context"
	"errors"
	"time"

	"github.com/edp1096/toy-board/internal/models"
	"github.com/edp1096/toy-board/internal/repository"
)

var (
	ErrVotesDisabled    = errors.New("이 게시판에서는 좋아요/싫어요 기능이 비활성화되었습니다")
	ErrInvalidVoteValue = errors.New("유효하지 않은 투표 값입니다")
)

type PostVoteService interface {
	VotePost(ctx context.Context, boardID, postID, userID int64, value int) (likes int, dislikes int, userVote int, err error)
	GetPostVoteStatus(ctx context.Context, postID, userID int64) (int, error)
	GetPostVoteCounts(ctx context.Context, postID int64) (likes int, dislikes int, err error)
}

type postVoteService struct {
	postVoteRepo repository.PostVoteRepository
	boardService BoardService
}

func NewPostVoteService(postVoteRepo repository.PostVoteRepository, boardService BoardService) PostVoteService {
	return &postVoteService{
		postVoteRepo: postVoteRepo,
		boardService: boardService,
	}
}

func (s *postVoteService) VotePost(ctx context.Context, boardID, postID, userID int64, value int) (int, int, int, error) {
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

	// 기존 투표 확인
	existingVote, err := s.postVoteRepo.GetByPostAndUser(ctx, postID, userID)
	now := time.Now()

	if err == nil { // 기존 투표가 있는 경우
		if value == 0 { // 투표 취소
			err = s.postVoteRepo.Delete(ctx, existingVote.ID)
		} else if existingVote.Value != value { // 투표 변경
			existingVote.Value = value
			existingVote.UpdatedAt = now
			err = s.postVoteRepo.Update(ctx, existingVote)
		} else { // 동일한 투표 (취소)
			err = s.postVoteRepo.Delete(ctx, existingVote.ID)
			value = 0 // 투표 취소로 간주
		}
	} else { // 새 투표
		if value == 0 { // 취소 요청은 무시
			return 0, 0, 0, nil
		}

		newVote := &models.PostVote{
			PostID:    postID,
			BoardID:   boardID,
			UserID:    userID,
			Value:     value,
			CreatedAt: now,
			UpdatedAt: now,
		}
		err = s.postVoteRepo.Create(ctx, newVote)
	}

	if err != nil {
		return 0, 0, 0, err
	}

	// 투표 수 업데이트
	err = s.postVoteRepo.UpdatePostVoteCount(ctx, boardID, postID)
	if err != nil {
		return 0, 0, 0, err
	}

	// 현재 투표 수 조회
	likes, dislikes, err := s.GetPostVoteCounts(ctx, postID)
	if err != nil {
		return 0, 0, 0, err
	}

	return likes, dislikes, value, nil
}

func (s *postVoteService) GetPostVoteStatus(ctx context.Context, postID, userID int64) (int, error) {
	if userID == 0 { // 로그인하지 않은 경우
		return 0, nil
	}

	vote, err := s.postVoteRepo.GetByPostAndUser(ctx, postID, userID)
	if err != nil {
		return 0, nil // 투표 없음
	}

	return vote.Value, nil
}

func (s *postVoteService) GetPostVoteCounts(ctx context.Context, postID int64) (int, int, error) {
	likes, err := s.postVoteRepo.CountByPost(ctx, postID, 1)
	if err != nil {
		return 0, 0, err
	}

	dislikes, err := s.postVoteRepo.CountByPost(ctx, postID, -1)
	if err != nil {
		return 0, 0, err
	}

	return likes, dislikes, nil
}
