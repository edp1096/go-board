// internal/handlers/auth_handler.go

package handlers

import (
	"os"
	"strings"
	"time"

	"github.com/edp1096/toy-board/internal/models"
	"github.com/edp1096/toy-board/internal/service"
	"github.com/edp1096/toy-board/internal/utils"

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
	// 이미 로그인된 사용자는 메인 페이지로 리다이렉트
	if c.Locals("user") != nil {
		return c.Redirect("/")
	}

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

	// 기본 유효성 검사
	trimmedUsername := strings.TrimSpace(username)
	if trimmedUsername == "" || len(trimmedUsername) < 3 {
		return utils.RenderWithUser(c, "auth/login", fiber.Map{
			"title":    "로그인",
			"error":    "아이디 또는 비밀번호가 올바르지 않습니다",
			"username": username,
			"redirect": redirect,
		})
	}

	// 비밀번호 빈 값 체크
	if strings.TrimSpace(password) == "" {
		return utils.RenderWithUser(c, "auth/login", fiber.Map{
			"title":    "로그인",
			"error":    "아이디 또는 비밀번호가 올바르지 않습니다",
			"username": username,
			"redirect": redirect,
		})
	}

	// 금지된 문자 체크 (옵션)
	validator := utils.NewInputValidator()
	if validator.ContainsForbiddenChars(trimmedUsername) {
		return utils.RenderWithUser(c, "auth/login", fiber.Map{
			"title":    "로그인",
			"error":    "아이디 또는 비밀번호가 올바르지 않습니다",
			"username": username,
			"redirect": redirect,
		})
	}

	// 로그인 처리 - 검증된 사용자명 전달
	_, token, err := h.authService.Login(c.Context(), trimmedUsername, password)
	if err != nil {
		return utils.RenderWithUser(c, "auth/login", fiber.Map{
			"title":    "로그인",
			"error":    "아이디 또는 비밀번호가 올바르지 않습니다",
			"username": username,
			"redirect": redirect,
		})
	}

	// 토큰을 쿠키에 저장 - 보안 설정 적용
	secure := os.Getenv("APP_ENV") == "production" // 운영환경에서만 Secure 활성화

	cookie := fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour), // 24시간 유효
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "lax",
	}
	c.Cookie(&cookie)

	// 리다이렉트
	return c.Redirect(redirect)
}

// RegisterPage 회원가입 페이지 렌더링
func (h *AuthHandler) RegisterPage(c *fiber.Ctx) error {
	// 이미 로그인된 사용자는 메인 페이지로 리다이렉트
	if c.Locals("user") != nil {
		return c.Redirect("/")
	}

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

	// 유효성 검사기 생성
	validator := utils.NewInputValidator()

	// 사용자명 검증
	if valid, errorMsg := validator.ValidateUsername(username); !valid {
		return utils.RenderWithUser(c, "auth/register", fiber.Map{
			"title":    "회원가입",
			"error":    errorMsg,
			"username": username,
			"email":    email,
			"fullName": fullName,
		})
	}

	// 이메일 검증
	if valid, errorMsg := validator.ValidateEmail(email); !valid {
		return utils.RenderWithUser(c, "auth/register", fiber.Map{
			"title":    "회원가입",
			"error":    errorMsg,
			"username": username,
			"email":    email,
			"fullName": fullName,
		})
	}

	// 비밀번호 검증
	if valid, errorMsg := validator.ValidatePassword(password); !valid {
		return utils.RenderWithUser(c, "auth/register", fiber.Map{
			"title":    "회원가입",
			"error":    errorMsg,
			"username": username,
			"email":    email,
			"fullName": fullName,
		})
	}

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

	// 이름 검증
	if valid, errorMsg := validator.ValidateFullName(fullName); !valid {
		return utils.RenderWithUser(c, "auth/register", fiber.Map{
			"title":    "회원가입",
			"error":    errorMsg,
			"username": username,
			"email":    email,
			"fullName": fullName,
		})
	}

	// 검증이 모두 통과된 데이터로 회원가입 처리
	_, err := h.authService.Register(
		c.Context(),
		strings.TrimSpace(username),
		strings.TrimSpace(email),
		password,
		strings.TrimSpace(fullName),
	)

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
	// 쿠키 명시적으로 만료 설정하여 삭제
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour), // 과거 시간으로 설정하여 즉시 만료
		HTTPOnly: true,
		SameSite: "lax",
	})

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

	// 유효성 검사기 생성
	validator := utils.NewInputValidator()

	// 이메일 검증
	if valid, errorMsg := validator.ValidateEmail(email); !valid {
		return utils.RenderWithUser(c, "auth/profile", fiber.Map{
			"title": "내 프로필",
			"error": errorMsg,
		})
	}

	// 이름 검증
	if valid, errorMsg := validator.ValidateFullName(fullName); !valid {
		return utils.RenderWithUser(c, "auth/profile", fiber.Map{
			"title": "내 프로필",
			"error": errorMsg,
		})
	}

	// 검증된 값으로 사용자 정보 업데이트
	user.Email = strings.TrimSpace(email)
	user.FullName = strings.TrimSpace(fullName)

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
		// 새 비밀번호 검증
		if valid, errorMsg := validator.ValidatePassword(newPassword); !valid {
			return utils.RenderWithUser(c, "auth/profile", fiber.Map{
				"title":   "내 프로필",
				"error":   "새 비밀번호: " + errorMsg,
				"success": "이메일과 이름이 업데이트되었습니다.",
			})
		}

		err := h.authService.ChangePassword(c.Context(), user.ID, currentPassword, newPassword)
		if err != nil {
			return utils.RenderWithUser(c, "auth/profile", fiber.Map{
				"title":   "내 프로필",
				"error":   "비밀번호 변경에 실패했습니다: " + err.Error(),
				"success": "이메일과 이름이 업데이트되었습니다.",
			})
		}
	}

	// 성공 메시지와 함께 프로필 페이지 다시 렌더링
	return utils.RenderWithUser(c, "auth/profile", fiber.Map{
		"title":   "내 프로필",
		"success": "프로필이 업데이트되었습니다.",
	})
}
