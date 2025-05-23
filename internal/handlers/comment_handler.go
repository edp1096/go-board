// internal/handlers/comment_handler.go
package handlers

import (
	"strconv"

	"github.com/edp1096/toy-board/internal/models"
	"github.com/edp1096/toy-board/internal/service"
	"github.com/edp1096/toy-board/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// CommentHandler - 댓글 관련 핸들러
type CommentHandler struct {
	commentService service.CommentService
	boardService   service.BoardService
}

// NewCommentHandler - 새 CommentHandler 생성
func NewCommentHandler(commentService service.CommentService, boardService service.BoardService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
		boardService:   boardService,
	}
}

// GetComments - 게시물 댓글 목록 조회 API
func (h *CommentHandler) GetComments(c *fiber.Ctx) error {
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

	// 답글 포함 여부 (기본값: true)
	includeReplies := c.Query("includeReplies", "true") == "true"

	// 현재 로그인한 사용자 정보
	user := c.Locals("user")
	var isAdmin bool

	if user != nil {
		userObj := user.(*models.User)
		isAdmin = (userObj.Role == models.RoleAdmin)
	}

	// 댓글 목록 조회
	comments, err := h.commentService.GetCommentsByPostID(c.Context(), boardID, postID, includeReplies, isAdmin)
	if err != nil {
		if err == service.ErrCommentsDisabled {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "이 게시판에서는 댓글 기능이 비활성화되었습니다",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "댓글을 불러오는데 실패했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"comments": comments,
	})
}

// CreateComment - 새 댓글 작성 API
func (h *CommentHandler) CreateComment(c *fiber.Ctx) error {
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
		Content  string `json:"content"`
		ParentID *int64 `json:"parentId"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "요청 형식이 잘못되었습니다",
		})
	}

	// 댓글 내용 검증
	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "댓글 내용을 입력해주세요",
		})
	}

	// IP 주소 획득
	visitorIP := utils.GetClientIP(c)

	// 댓글 생성
	comment, err := h.commentService.CreateComment(c.Context(), boardID, postID, user.ID, req.Content, req.ParentID, visitorIP)
	if err != nil {
		if err == service.ErrCommentsDisabled {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "이 게시판에서는 댓글 기능이 비활성화되었습니다",
			})
		} else if err == service.ErrCommentNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "답글을 작성하려는 댓글을 찾을 수 없습니다",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "댓글 작성에 실패했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"comment": comment,
	})
}

// UpdateComment - 댓글 수정 API
func (h *CommentHandler) UpdateComment(c *fiber.Ctx) error {
	commentID, err := strconv.ParseInt(c.Params("commentID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 댓글 ID입니다",
		})
	}

	// 댓글 조회 (먼저 댓글을 가져와야 boardID를 알 수 있음)
	comment, err := h.commentService.GetCommentByID(c.Context(), commentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "댓글을 찾을 수 없습니다",
		})
	}

	// 현재 로그인한 사용자 정보
	user := c.Locals("user").(*models.User)
	isAdmin := user.Role == models.RoleAdmin

	// 게시판 매니저 또는 소모임 moderator 여부 확인
	isModerator := false
	if !isAdmin {
		// 매니저 여부 확인
		isManager, _ := h.boardService.IsBoardManager(c.Context(), comment.BoardID, user.ID)
		if isManager {
			isModerator = true
		} else {
			// 게시판 타입 확인
			board, err := h.boardService.GetBoardByID(c.Context(), comment.BoardID)
			if err == nil && board.BoardType == models.BoardTypeGroup {
				// 소모임 moderator 여부 확인
				isModerator, _ = h.boardService.IsParticipantModerator(c.Context(), comment.BoardID, user.ID)
			}
		}
	}

	// 요청 본문 파싱
	var req struct {
		Content string `json:"content"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "요청 형식이 잘못되었습니다",
		})
	}

	// 댓글 내용 검증
	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "댓글 내용을 입력해주세요",
		})
	}

	// 댓글 수정
	updatedComment, err := h.commentService.UpdateComment(c.Context(), commentID, user.ID, req.Content, isAdmin || isModerator)
	if err != nil {
		if err == service.ErrCommentNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "댓글을 찾을 수 없습니다",
			})
		} else if err == service.ErrNoPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "댓글을 수정할 권한이 없습니다",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "댓글 수정에 실패했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"comment": updatedComment,
	})
}

// DeleteComment - 댓글 삭제 API
func (h *CommentHandler) DeleteComment(c *fiber.Ctx) error {
	commentID, err := strconv.ParseInt(c.Params("commentID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 댓글 ID입니다",
		})
	}

	// 댓글 조회 (먼저 댓글을 가져와야 boardID를 알 수 있음)
	comment, err := h.commentService.GetCommentByID(c.Context(), commentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "댓글을 찾을 수 없습니다",
		})
	}

	// 현재 로그인한 사용자 정보
	user := c.Locals("user").(*models.User)
	isAdmin := user.Role == models.RoleAdmin

	// 게시판 매니저 또는 소모임 moderator 여부 확인
	isModerator := false
	if !isAdmin {
		// 매니저 여부 확인
		isManager, _ := h.boardService.IsBoardManager(c.Context(), comment.BoardID, user.ID)
		if isManager {
			isModerator = true
		} else {
			// 게시판 타입 확인
			board, err := h.boardService.GetBoardByID(c.Context(), comment.BoardID)
			if err == nil && board.BoardType == models.BoardTypeGroup {
				// 소모임 moderator 여부 확인
				isModerator, _ = h.boardService.IsParticipantModerator(c.Context(), comment.BoardID, user.ID)
			}
		}
	}

	// 댓글 삭제
	err = h.commentService.DeleteComment(c.Context(), commentID, user.ID, isAdmin || isModerator)
	if err != nil {
		if err == service.ErrCommentNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "댓글을 찾을 수 없습니다",
			})
		} else if err == service.ErrNoPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "댓글을 삭제할 권한이 없습니다",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "댓글 삭제에 실패했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "댓글이 삭제되었습니다",
	})
}
