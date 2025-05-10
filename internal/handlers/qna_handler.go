// internal/handlers/qna_handler.go
package handlers

import (
	"strconv"

	"github.com/edp1096/toy-board/internal/models"
	"github.com/edp1096/toy-board/internal/service"
	"github.com/edp1096/toy-board/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// QnAHandler - Q&A 기능 관련 핸들러
type QnAHandler struct {
	boardService service.BoardService
	qnaService   service.QnAService
}

// NewQnAHandler - 새 QnAHandler 생성
func NewQnAHandler(boardService service.BoardService, qnaService service.QnAService) *QnAHandler {
	return &QnAHandler{
		boardService: boardService,
		qnaService:   qnaService,
	}
}

// GetAnswers - 질문에 대한 답변 목록 조회 API
func (h *QnAHandler) GetAnswers(c *fiber.Ctx) error {
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
	user := c.Locals("user")
	var isAdmin bool

	if user != nil {
		userObj := user.(*models.User)
		isAdmin = (userObj.Role == models.RoleAdmin)
	}

	// 답변 목록 조회
	answers, err := h.qnaService.GetAnswersByQuestionID(c.Context(), boardID, postID, isAdmin)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "답변을 불러오는데 실패했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"answers": answers,
	})
}

// CreateAnswer - 답변 작성 API
func (h *QnAHandler) CreateAnswer(c *fiber.Ctx) error {
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
		Content string `json:"content"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "요청 형식이 잘못되었습니다",
		})
	}

	// 답변 내용 검증
	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "답변 내용을 입력해주세요",
		})
	}

	// IP 주소 획득
	visitorIP := utils.GetClientIP(c)

	// 답변 생성
	answer, err := h.qnaService.CreateAnswer(c.Context(), boardID, postID, user.ID, req.Content, visitorIP)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "답변 작성에 실패했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{"success": true, "answer": answer})
}

// UpdateAnswer - 답변 수정 API
func (h *QnAHandler) UpdateAnswer(c *fiber.Ctx) error {
	answerID, err := strconv.ParseInt(c.Params("answerID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 답변 ID입니다",
		})
	}

	// 현재 로그인한 사용자 정보
	user := c.Locals("user").(*models.User)
	isAdmin := user.Role == models.RoleAdmin

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

	// 답변 내용 검증
	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "답변 내용을 입력해주세요",
		})
	}

	// 답변 수정
	answer, err := h.qnaService.UpdateAnswer(c.Context(), answerID, user.ID, req.Content, isAdmin)
	if err != nil {
		if err == service.ErrAnswerNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "답변을 찾을 수 없습니다",
			})
		} else if err == service.ErrNoPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "답변을 수정할 권한이 없습니다",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "답변 수정에 실패했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"answer":  answer,
	})
}

// DeleteAnswer - 답변 삭제 API
func (h *QnAHandler) DeleteAnswer(c *fiber.Ctx) error {
	answerID, err := strconv.ParseInt(c.Params("answerID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 답변 ID입니다",
		})
	}

	// 현재 로그인한 사용자 정보
	user := c.Locals("user").(*models.User)
	isAdmin := user.Role == models.RoleAdmin

	// 답변 삭제
	err = h.qnaService.DeleteAnswer(c.Context(), answerID, user.ID, isAdmin)
	if err != nil {
		if err == service.ErrAnswerNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "답변을 찾을 수 없습니다",
				"error":   err.Error(),
			})
		} else if err == service.ErrNoPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "답변을 삭제할 권한이 없습니다",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "답변 삭제에 실패했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "답변이 삭제되었습니다",
	})
}

// GetQuestionVoteCount는 질문의 투표 수를 조회하는 API입니다.
func (h *QnAHandler) GetQuestionVoteCount(c *fiber.Ctx) error {
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

	// 투표 수 조회
	voteCount, err := h.qnaService.GetQuestionVoteCount(c.Context(), boardID, postID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "투표 수 조회에 실패했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"voteCount": voteCount,
	})
}

// VoteQuestion - 질문 투표 API
func (h *QnAHandler) VoteQuestion(c *fiber.Ctx) error {
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
		Direction string `json:"direction"` // "up" 또는 "down"
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "요청 형식이 잘못되었습니다",
		})
	}

	// 방향 검증
	if req.Direction != "up" && req.Direction != "down" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "유효하지 않은 투표 방향입니다 (up 또는 down)",
		})
	}

	// 투표 처리
	voteValue := 1
	if req.Direction == "down" {
		voteValue = -1
	}

	newCount, err := h.qnaService.VoteQuestion(c.Context(), boardID, postID, user.ID, voteValue)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "투표 처리에 실패했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"voteCount": newCount,
	})
}

// VoteAnswer - 답변 투표 API
func (h *QnAHandler) VoteAnswer(c *fiber.Ctx) error {
	answerID, err := strconv.ParseInt(c.Params("answerID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 답변 ID입니다",
		})
	}

	// 현재 로그인한 사용자 정보
	user := c.Locals("user").(*models.User)

	// 요청 본문 파싱
	var req struct {
		Direction string `json:"direction"` // "up" 또는 "down"
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "요청 형식이 잘못되었습니다",
		})
	}

	// 방향 검증
	if req.Direction != "up" && req.Direction != "down" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "유효하지 않은 투표 방향입니다 (up 또는 down)",
		})
	}

	// 투표 처리
	voteValue := 1
	if req.Direction == "down" {
		voteValue = -1
	}

	newCount, err := h.qnaService.VoteAnswer(c.Context(), answerID, user.ID, voteValue)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "투표 처리에 실패했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"voteCount": newCount,
	})
}

// UpdateQuestionStatus - 질문 상태 업데이트 API
func (h *QnAHandler) UpdateQuestionStatus(c *fiber.Ctx) error {
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
		Status string `json:"status"` // "solved" 또는 "unsolved"
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "요청 형식이 잘못되었습니다",
		})
	}

	// 상태 검증
	if req.Status != "solved" && req.Status != "unsolved" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "유효하지 않은 상태입니다 (solved 또는 unsolved)",
		})
	}

	// 질문 상태 업데이트
	err = h.qnaService.UpdateQuestionStatus(c.Context(), boardID, postID, user.ID, req.Status)
	if err != nil {
		if err == service.ErrNoPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "질문 상태를 변경할 권한이 없습니다",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "질문 상태 업데이트에 실패했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "질문 상태가 업데이트되었습니다",
	})
}

// SetBestAnswer - 베스트 답변 설정 API
func (h *QnAHandler) SetBestAnswer(c *fiber.Ctx) error {
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
		AnswerID int64 `json:"answerId"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "요청 형식이 잘못되었습니다",
		})
	}

	// 베스트 답변 설정
	err = h.qnaService.SetBestAnswer(c.Context(), boardID, postID, req.AnswerID, user.ID)
	if err != nil {
		if err == service.ErrNoPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "베스트 답변을 선택할 권한이 없습니다",
			})
		} else if err == service.ErrAnswerNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "답변을 찾을 수 없습니다",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "베스트 답변 설정에 실패했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "베스트 답변이 설정되었습니다",
	})
}

// internal/handlers/qna_handler.go 파일에 추가할 메소드

// CreateAnswerReply - 답변에 대한 답글 작성 API
func (h *QnAHandler) CreateAnswerReply(c *fiber.Ctx) error {
	answerID, err := strconv.ParseInt(c.Params("answerID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 답변 ID입니다",
		})
	}

	// 현재 로그인한 사용자 정보
	user := c.Locals("user").(*models.User)

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

	// 답글 내용 검증
	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "답글 내용을 입력해주세요",
		})
	}

	// IP 주소 획득
	visitorIP := utils.GetClientIP(c)

	// 답글 생성
	reply, err := h.qnaService.CreateAnswerReply(c.Context(), answerID, user.ID, req.Content, visitorIP)
	if err != nil {
		if err == service.ErrAnswerNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "답변을 찾을 수 없습니다",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "답글 작성에 실패했습니다: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"reply":   reply,
	})
}
