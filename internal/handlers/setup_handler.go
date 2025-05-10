// internal/handlers/setup_handler.go
package handlers

import (
	"github.com/edp1096/toy-board/internal/service"
	"github.com/edp1096/toy-board/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type SetupHandler struct {
	authService  service.AuthService
	setupService service.SetupService
}

func NewSetupHandler(authService service.AuthService, setupService service.SetupService) *SetupHandler {
	return &SetupHandler{
		authService:  authService,
		setupService: setupService,
	}
}

// SetupPage 관리자 설정 페이지 렌더링
func (h *SetupHandler) SetupPage(c *fiber.Ctx) error {
	// 이미 관리자 계정이 있는지 확인
	adminExists, err := h.setupService.IsAdminExists(c.Context())
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "시스템 확인 중 오류가 발생했습니다",
			"error":   err.Error(),
		})
	}

	// 이미 관리자가 있으면 메인 페이지로 리다이렉트
	if adminExists {
		return c.Redirect("/")
	}

	return utils.RenderWithUser(c, "admin/setup", fiber.Map{
		"title": "관리자 계정 설정",
	})
}

// SetupAdmin 관리자 계정 생성 처리
func (h *SetupHandler) SetupAdmin(c *fiber.Ctx) error {
	// 이미 관리자 계정이 있는지 확인
	adminExists, err := h.setupService.IsAdminExists(c.Context())
	if err != nil {
		return utils.RenderWithUser(c, "error", fiber.Map{
			"title":   "오류",
			"message": "시스템 확인 중 오류가 발생했습니다",
			"error":   err.Error(),
		})
	}

	// 이미 관리자가 있으면 메인 페이지로 리다이렉트
	if adminExists {
		return c.Redirect("/")
	}

	// 폼 데이터 파싱
	username := c.FormValue("username")
	email := c.FormValue("email")
	password := c.FormValue("password")
	passwordConfirm := c.FormValue("password_confirm")
	fullName := c.FormValue("full_name")

	// 비밀번호 확인
	if password != passwordConfirm {
		return utils.RenderWithUser(c, "admin/setup", fiber.Map{
			"title":    "관리자 계정 설정",
			"error":    "비밀번호가 일치하지 않습니다",
			"username": username,
			"email":    email,
			"fullName": fullName,
		})
	}

	// 관리자 생성
	user, err := h.setupService.CreateAdminUser(c.Context(), username, email, password, fullName)
	if err != nil {
		return utils.RenderWithUser(c, "admin/setup", fiber.Map{
			"title":    "관리자 계정 설정",
			"error":    err.Error(),
			"username": username,
			"email":    email,
			"fullName": fullName,
		})
	}

	// 로그인 처리
	_, token, err := h.authService.Login(c.Context(), username, password)
	if err != nil {
		// 로그인 실패해도 계정은 생성되었으므로 로그인 페이지로 이동
		return c.Redirect("/auth/login?message=관리자 계정이 생성되었습니다. 로그인해 주세요.")
	}

	// 토큰을 쿠키에 저장
	cookie := fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		Expires:  user.CreatedAt.Add(24 * 3600), // 24시간 유효
		HTTPOnly: true,
	}
	c.Cookie(&cookie)

	// 관리자 대시보드로 리다이렉트
	return c.Redirect("/admin?message=관리자 계정이 성공적으로 생성되었습니다.")
}
