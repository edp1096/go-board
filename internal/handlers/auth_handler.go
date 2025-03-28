// internal/handlers/auth_handler.go

package handlers

import (
	"dynamic-board/internal/models"
	"dynamic-board/internal/service"
	"dynamic-board/internal/utils"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginPage 로그인 페이지 렌더링
func (h *AuthHandler) LoginPage(c *fiber.Ctx) error {
	return utils.RenderWithUser(c, "auth/login", fiber.Map{
		"title":    "로그인",
		"redirect": c.Query("redirect", "/"),
	})
}

// Login 로그인 처리
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	// 폼 데이터 파싱
	username := c.FormValue("username")
	password := c.FormValue("password")
	redirect := c.FormValue("redirect", "/")

	// 로그인 처리
	user, token, err := h.authService.Login(c.Context(), username, password)
	if err != nil {
		return utils.RenderWithUser(c, "auth/login", fiber.Map{
			"title":    "로그인",
			"error":    "아이디 또는 비밀번호가 올바르지 않습니다",
			"username": username,
			"redirect": redirect,
		})
	}

	// 로깅 추가 - 디버깅 용도
	log.Printf("로그인 성공: 사용자 %s (ID: %d)", user.Username, user.ID)

	// 토큰을 쿠키에 저장 - 설정 개선
	cookie := fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour), // 24시간 유효
		HTTPOnly: true,
		// 개발 환경에서는 Secure를 false로 설정
		Secure: false,
		// 개발 환경에서는 SameSite를 Lax로 설정
		SameSite: "lax",
	}
	c.Cookie(&cookie)

	log.Printf("쿠키 설정: %s=%s", cookie.Name, token[:10]+"...")

	// 리다이렉트
	return c.Redirect(redirect)
}

// RegisterPage 회원가입 페이지 렌더링
func (h *AuthHandler) RegisterPage(c *fiber.Ctx) error {
	return utils.RenderWithUser(c, "auth/register", fiber.Map{
		"title": "회원가입",
	})
}

// Register 회원가입 처리
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	// 폼 데이터 파싱
	username := c.FormValue("username")
	email := c.FormValue("email")
	password := c.FormValue("password")
	passwordConfirm := c.FormValue("password_confirm")
	fullName := c.FormValue("full_name")

	// 비밀번호 확인
	if password != passwordConfirm {
		return utils.RenderWithUser(c, "auth/register", fiber.Map{
			"title":    "회원가입",
			"error":    "비밀번호가 일치하지 않습니다",
			"username": username,
			"email":    email,
			"fullName": fullName,
		})
	}

	// 회원가입 처리
	_, err := h.authService.Register(c.Context(), username, email, password, fullName)
	if err != nil {
		return utils.RenderWithUser(c, "auth/register", fiber.Map{
			"title":    "회원가입",
			"error":    err.Error(),
			"username": username,
			"email":    email,
			"fullName": fullName,
		})
	}

	// 로그인 페이지로 리다이렉트 (성공 메시지 포함)
	return c.Redirect("/auth/login?message=회원가입이 완료되었습니다. 로그인해 주세요.")
}

// Logout 로그아웃 처리
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// 쿠키 삭제
	c.ClearCookie("auth_token")

	// 메인 페이지로 리다이렉트
	return c.Redirect("/")
}

// ProfilePage 프로필 페이지 렌더링
func (h *AuthHandler) ProfilePage(c *fiber.Ctx) error {
	// user := c.Locals("user").(*models.User)

	return utils.RenderWithUser(c, "auth/profile", fiber.Map{
		"title": "내 프로필",
	})
}

// UpdateProfile 프로필 업데이트 처리
func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	// 폼 데이터 파싱
	email := c.FormValue("email")
	fullName := c.FormValue("full_name")
	currentPassword := c.FormValue("current_password")
	newPassword := c.FormValue("new_password")

	// 사용자 정보 업데이트
	user.Email = email
	user.FullName = fullName

	// 프로필 업데이트
	err := h.authService.UpdateUser(c.Context(), user)
	if err != nil {
		return utils.RenderWithUser(c, "auth/profile", fiber.Map{
			"title": "내 프로필",
			"error": "프로필 업데이트에 실패했습니다: " + err.Error(),
		})
	}

	// 비밀번호 변경이 요청된 경우
	if currentPassword != "" && newPassword != "" {
		err := h.authService.ChangePassword(c.Context(), user.ID, currentPassword, newPassword)
		if err != nil {
			return utils.RenderWithUser(c, "auth/profile", fiber.Map{
				"title":   "내 프로필",
				"error":   "비밀번호 변경에 실패했습니다: " + err.Error(),
				"success": "프로필이 업데이트되었습니다.",
			})
		}
	}

	// 성공 메시지와 함께 프로필 페이지 다시 렌더링
	return utils.RenderWithUser(c, "auth/profile", fiber.Map{
		"title":   "내 프로필",
		"success": "프로필이 업데이트되었습니다.",
	})
}
