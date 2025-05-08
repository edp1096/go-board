// internal/utils/validation.go
package utils

import (
	"regexp"
	"strings"
)

// InputValidator 입력 유효성 검사 관련 함수 모음
type InputValidator struct {
	MinUsernameLength  int
	MinPasswordLength  int
	MinFullNameLength  int
	AllowedCharPattern string
	ForbiddenChars     []string
}

// NewInputValidator 기본 설정으로 InputValidator 생성
func NewInputValidator() *InputValidator {
	return &InputValidator{
		MinUsernameLength:  3,
		MinPasswordLength:  8,
		MinFullNameLength:  2,
		AllowedCharPattern: `^[a-zA-Z0-9가-힣_\-\.]+$`,
		ForbiddenChars:     []string{"<", ">", "\"", "'", ";", "--", "/*", "*/", "@@", "$", "|", "\\", "/"},
	}
}

// ValidateUsername 사용자명 유효성 검사
func (v *InputValidator) ValidateUsername(username string) (bool, string) {
	trimmed := strings.TrimSpace(username)

	if len(trimmed) < v.MinUsernameLength {
		return false, "사용자명은 최소 3자 이상이어야 합니다"
	}

	if v.ContainsForbiddenChars(trimmed) {
		return false, "사용자명에 허용되지 않는 특수문자가 포함되어 있습니다"
	}

	regex := regexp.MustCompile(v.AllowedCharPattern)
	if !regex.MatchString(trimmed) {
		return false, "사용자명은 알파벳, 숫자, 한글 및 일부 특수문자(_-.)만 사용 가능합니다"
	}

	return true, ""
}

// ValidateEmail 이메일 유효성 검사
func (v *InputValidator) ValidateEmail(email string) (bool, string) {
	trimmed := strings.TrimSpace(email)

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(trimmed) {
		return false, "유효하지 않은 이메일 형식입니다"
	}

	return true, ""
}

// ValidatePassword 비밀번호 유효성 검사
func (v *InputValidator) ValidatePassword(password string) (bool, string) {
	if len(password) < v.MinPasswordLength {
		return false, "비밀번호는 최소 8자 이상이어야 합니다"
	}

	// 최소 하나의 숫자, 하나의 소문자, 하나의 대문자 포함
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)

	if !hasNumber || !hasLower || !hasUpper {
		return false, "비밀번호는 최소 하나의 숫자, 소문자, 대문자를 포함해야 합니다"
	}

	return true, ""
}

// ValidateFullName 이름 유효성 검사
func (v *InputValidator) ValidateFullName(fullName string) (bool, string) {
	trimmed := strings.TrimSpace(fullName)

	if len(trimmed) < v.MinFullNameLength {
		return false, "이름은 최소 2자 이상이어야 합니다"
	}

	if v.ContainsForbiddenChars(trimmed) {
		return false, "이름에 허용되지 않는 특수문자가 포함되어 있습니다"
	}

	return true, ""
}

// ContainsForbiddenChars 금지된 문자 포함 여부 확인
func (v *InputValidator) ContainsForbiddenChars(input string) bool {
	for _, char := range v.ForbiddenChars {
		if strings.Contains(input, char) {
			return true
		}
	}
	return false
}
