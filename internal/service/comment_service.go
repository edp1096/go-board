// internal/service/comment_service.go
package service

import (
	"context"
	"errors"
	"time"

	"github.com/edp1096/go-board/internal/models"
	"github.com/edp1096/go-board/internal/repository"
)

var (
	ErrCommentNotFound  = errors.New("댓글을 찾을 수 없음")
	ErrCommentsDisabled = errors.New("이 게시판에서는 댓글 기능이 비활성화되었습니다")
	ErrNoPermission     = errors.New("권한이 없습니다")
)

// CommentService 인터페이스
type CommentService interface {
	CreateComment(ctx context.Context, boardID, postID, userID int64, content string, parentID *int64, ipAddress string) (*models.Comment, error)
	GetCommentByID(ctx context.Context, id int64) (*models.Comment, error)
	GetCommentsByPostID(ctx context.Context, boardID, postID int64, includeReplies bool) ([]*models.Comment, error)
	UpdateComment(ctx context.Context, id, userID int64, content string, isAdmin bool) (*models.Comment, error)
	DeleteComment(ctx context.Context, id, userID int64, isAdmin bool) error
	DeleteCommentsByPostID(ctx context.Context, boardID, postID int64) error
}

// commentService 구현체
type commentService struct {
	commentRepo repository.CommentRepository
	boardRepo   repository.BoardRepository
}

// 새 CommentService 생성
func NewCommentService(commentRepo repository.CommentRepository, boardRepo repository.BoardRepository) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		boardRepo:   boardRepo,
	}
}

// CreateComment - 새 댓글 생성
func (s *commentService) CreateComment(ctx context.Context, boardID, postID, userID int64, content string, parentID *int64, ipAddress string) (*models.Comment, error) {
	// 게시판 정보 조회하여 댓글 기능 활성화 여부 확인
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	if !board.CommentsEnabled {
		return nil, ErrCommentsDisabled
	}

	// 부모 댓글 ID가 제공된 경우 해당 댓글 확인
	if parentID != nil {
		parentComment, err := s.commentRepo.GetByID(ctx, *parentID)
		if err != nil {
			return nil, ErrCommentNotFound
		}

		// 부모 댓글이 동일 게시물에 속하는지 확인
		if parentComment.PostID != postID || parentComment.BoardID != boardID {
			return nil, errors.New("부모 댓글이 다른 게시물에 속해 있습니다")
		}

		// 중첩 댓글 제한 (1단계만 허용)
		if parentComment.ParentID != nil {
			return nil, errors.New("댓글은 1단계까지만 중첩될 수 있습니다")
		}
	}

	// 댓글 객체 생성
	now := time.Now()
	comment := &models.Comment{
		PostID:    postID,
		BoardID:   boardID,
		UserID:    userID,
		Content:   content,
		ParentID:  parentID,
		IpAddress: ipAddress,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 댓글 저장
	err = s.commentRepo.Create(ctx, comment)
	if err != nil {
		return nil, err
	}

	// 댓글 저장 후 게시물의 댓글 수 업데이트
	s.updatePostCommentCount(ctx, boardID, postID)

	// 저장된 댓글 다시 조회 (사용자 정보 포함)
	return s.commentRepo.GetByID(ctx, comment.ID)
}

// GetCommentByID - 댓글 ID로 댓글 조회
func (s *commentService) GetCommentByID(ctx context.Context, id int64) (*models.Comment, error) {
	return s.commentRepo.GetByID(ctx, id)
}

// GetCommentsByPostID - 게시물 댓글 목록 조회
func (s *commentService) GetCommentsByPostID(ctx context.Context, boardID, postID int64, includeReplies bool) ([]*models.Comment, error) {
	// 게시판 정보 조회하여 댓글 기능 활성화 여부 확인
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	if !board.CommentsEnabled {
		return []*models.Comment{}, nil // 댓글 기능이 비활성화된 경우 빈 배열 반환
	}

	return s.commentRepo.GetByPostID(ctx, boardID, postID, includeReplies)
}

// UpdateComment - 댓글 수정
func (s *commentService) UpdateComment(ctx context.Context, id, userID int64, content string, isAdmin bool) (*models.Comment, error) {
	// 기존 댓글 조회
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCommentNotFound
	}

	// 권한 확인 (작성자 또는 관리자만 수정 가능)
	if comment.UserID != userID && !isAdmin {
		return nil, ErrNoPermission
	}

	// 댓글 업데이트
	comment.Content = content
	comment.UpdatedAt = time.Now()

	err = s.commentRepo.Update(ctx, comment)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// DeleteComment - 댓글 삭제
func (s *commentService) DeleteComment(ctx context.Context, id, userID int64, isAdmin bool) error {
	// 기존 댓글 조회
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return ErrCommentNotFound
	}

	// 권한 확인 (작성자 또는 관리자만 삭제 가능)
	if comment.UserID != userID && !isAdmin {
		return ErrNoPermission
	}

	err = s.commentRepo.Delete(ctx, id)

	// 게시물의 댓글 수 업데이트
	s.updatePostCommentCount(ctx, comment.BoardID, comment.PostID)

	return err
}

// DeleteCommentsByPostID 메서드 수정
func (s *commentService) DeleteCommentsByPostID(ctx context.Context, boardID, postID int64) error {
	err := s.commentRepo.DeleteByPostID(ctx, boardID, postID)
	if err != nil {
		return err
	}

	// 게시물의 댓글 수를 0으로 업데이트
	return s.commentRepo.UpdatePostCommentCount(ctx, boardID, postID, 0)
}

// 게시물의 댓글 수 업데이트
func (s *commentService) updatePostCommentCount(ctx context.Context, boardID, postID int64) error {
	// 댓글 수 조회
	count, err := s.commentRepo.CountByPostID(ctx, boardID, postID)
	if err != nil {
		return err
	}

	// 게시물 댓글 수 업데이트
	return s.commentRepo.UpdatePostCommentCount(ctx, boardID, postID, count)
}
