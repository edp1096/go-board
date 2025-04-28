// internal/handlers/vote_handler.go
package handlers

import (
	"strconv"

	"github.com/edp1096/go-board/internal/models"
	"github.com/edp1096/go-board/internal/service"
	"github.com/gofiber/fiber/v2"
)

type VoteHandler struct {
	postVoteService    service.PostVoteService
	commentVoteService service.CommentVoteService
}

func NewVoteHandler(postVoteService service.PostVoteService, commentVoteService service.CommentVoteService) *VoteHandler {
	return &VoteHandler{
		postVoteService:    postVoteService,
		commentVoteService: commentVoteService,
	}
}

// 게시물 투표 처리
func (h *VoteHandler) VotePost(c *fiber.Ctx) error {
	boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 게시판 ID입니다",
		})
	}

	postID, err := strconv.ParseInt(c.Params("postID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 게시물 ID입니다",
		})
	}

	// 현재 로그인한 사용자 정보
	user := c.Locals("user").(*models.User)

	// 요청 본문 파싱
	var req struct {
		Value int `json:"value"` // 1: 좋아요, -1: 싫어요, 0: 취소
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "요청 형식이 잘못되었습니다",
		})
	}

	// 투표 처리
	likes, dislikes, userVote, err := h.postVoteService.VotePost(c.Context(), boardID, postID, user.ID, req.Value)
	if err != nil {
		if err == service.ErrVotesDisabled {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "투표 처리 중 오류가 발생했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"likes":    likes,
		"dislikes": dislikes,
		"userVote": userVote,
	})
}

// 댓글 투표 처리
func (h *VoteHandler) VoteComment(c *fiber.Ctx) error {
	commentID, err := strconv.ParseInt(c.Params("commentID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 댓글 ID입니다",
		})
	}

	// 현재 로그인한 사용자 정보
	user := c.Locals("user").(*models.User)

	// 요청 본문 파싱
	var req struct {
		BoardID string `json:"boardId"`
		Value   int    `json:"value"` // 1: 좋아요, -1: 싫어요, 0: 취소
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "요청 형식이 잘못되었습니다",
		})
	}

	boardID, _ := strconv.ParseInt(req.BoardID, 10, 64)

	// 투표 처리
	likes, dislikes, userVote, err := h.commentVoteService.VoteComment(c.Context(), boardID, commentID, user.ID, req.Value)
	if err != nil {
		if err == service.ErrVotesDisabled {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "투표 처리 중 오류가 발생했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"likes":    likes,
		"dislikes": dislikes,
		"userVote": userVote,
	})
}

// 게시물 투표 상태 조회
func (h *VoteHandler) GetPostVoteStatus(c *fiber.Ctx) error {
	// boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	_, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 게시판 ID입니다",
		})
	}

	postID, err := strconv.ParseInt(c.Params("postID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 게시물 ID입니다",
		})
	}

	// 사용자 정보
	var userID int64
	user := c.Locals("user")
	if user != nil {
		userID = user.(*models.User).ID
	}

	// 투표 상태 조회
	voteValue, err := h.postVoteService.GetPostVoteStatus(c.Context(), postID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "투표 상태 조회 중 오류가 발생했습니다",
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"voteValue": voteValue,
	})
}

// 댓글 투표 상태 조회
func (h *VoteHandler) GetCommentVoteStatus(c *fiber.Ctx) error {
	commentID, err := strconv.ParseInt(c.Params("commentID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 댓글 ID입니다",
		})
	}

	// 사용자 정보
	var userID int64
	user := c.Locals("user")
	if user != nil {
		userID = user.(*models.User).ID
	}

	// 투표 상태 조회
	voteValue, err := h.commentVoteService.GetCommentVoteStatus(c.Context(), commentID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "투표 상태 조회 중 오류가 발생했습니다",
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"voteValue": voteValue,
	})
}

// 여러 댓글 투표 상태 조회
func (h *VoteHandler) GetMultipleCommentVoteStatuses(c *fiber.Ctx) error {
	// 요청 본문 파싱
	var req struct {
		BoardID    int64   `json:"boardId"`
		CommentIDs []int64 `json:"commentIds"`
	}

	// if err := c.BodyParser(&req); err != nil {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"success": false,
	// 		"message": "요청 형식이 잘못되었습니다",
	// 	})
	// }

	// 사용자 정보
	var userID int64
	user := c.Locals("user")
	if user != nil {
		userID = user.(*models.User).ID
	}

	// 여러 댓글 투표 상태 조회
	voteStatuses, err := h.commentVoteService.GetMultipleCommentVoteStatuses(c.Context(), req.CommentIDs, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "투표 상태 조회 중 오류가 발생했습니다",
		})
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"voteStatuses": voteStatuses,
	})
}
