// internal/middleware/board_access.go
package middleware

import (
	"strconv"

	"github.com/edp1096/go-board/internal/service"

	"github.com/gofiber/fiber/v2"
)

// BoardAccessMiddleware 인터페이스
type BoardAccessMiddleware interface {
	CheckBoardAccess(c *fiber.Ctx) error
}

type boardAccessMiddleware struct {
	boardService service.BoardService
}

// NewBoardAccessMiddleware 생성자
func NewBoardAccessMiddleware(boardService service.BoardService) BoardAccessMiddleware {
	return &boardAccessMiddleware{
		boardService: boardService,
	}
}

// CheckBoardAccess 미들웨어 - 게시판 접근 권한 검사
func (m *boardAccessMiddleware) CheckBoardAccess(c *fiber.Ctx) error {
	// 게시판 ID 가져오기
	boardID, err := strconv.ParseInt(c.Params("boardID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "잘못된 게시판 ID입니다",
		})
	}

	// 게시판 정보 조회
	board, err := m.boardService.GetBoardByID(c.Context(), boardID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "게시판을 찾을 수 없습니다",
		})
	}

	// 게시판이 활성화되어 있는지 확인
	if !board.Active {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "비활성화된 게시판입니다",
		})
	}

	// 익명 접근이 허용되지 않는 경우 인증 확인
	if !board.AllowAnonymous {
		// 로그인한 사용자인지 확인
		user := c.Locals("user")
		if user == nil {
			// API 요청인 경우 401 응답
			if c.Path()[:4] == "/api" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"message": "이 게시판에 접근하려면 로그인이 필요합니다",
				})
			}

			// 웹 페이지 요청인 경우 로그인 페이지로 리다이렉트
			return c.Redirect("/auth/login?redirect=" + c.Path())
		}
	}

	// 게시판 정보를 Locals에 저장 (핸들러에서 다시 조회하지 않도록)
	c.Locals("board", board)

	return c.Next()
}
