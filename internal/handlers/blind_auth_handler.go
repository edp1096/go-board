// internal/handlers/blind_auth_handler.go
package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/edp1096/go-board/internal/service"
	"github.com/edp1096/go-board/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type BlindAuthHandler struct {
	authService service.AuthService
	secretKey   string
}

func NewBlindAuthHandler(authService service.AuthService, secretKey string) *BlindAuthHandler {
	return &BlindAuthHandler{
		authService: authService,
		secretKey:   secretKey,
	}
}

// BlindAuth 핸들러 - 외부 시스템에서 온 사용자 자동 인증
func (h *BlindAuthHandler) BlindAuth(c *fiber.Ctx) error {
	// URL에서 토큰 확인
	token := c.Query("token")
	if token == "" {
		return c.Redirect("/") // 토큰이 없으면 홈으로
	}

	// 토큰 검증 및 페이로드 추출
	claims, err := h.validateToken(token)
	if err != nil {
		// 오류 로깅만 하고 실패해도 일반 페이지로 리다이렉트
		fmt.Printf("토큰 검증 실패: %v\n", err)
		return c.Redirect("/")
	}

	// 필수 필드 확인
	externalID, ok1 := claims["external_id"].(string)
	username, ok2 := claims["username"].(string)
	email, ok3 := claims["email"].(string)

	if !ok1 || !ok2 || !ok3 || externalID == "" || username == "" {
		return c.Redirect("/")
	}

	// 사용자 조회 또는 생성
	externalSystem := claims["system"].(string)
	if externalSystem == "" {
		externalSystem = "main_system" // 기본값
	}

	user, err := h.authService.GetUserByExternalID(c.Context(), externalID, externalSystem)

	// 사용자가 없으면 자동 생성
	if err != nil || user == nil {
		fullName, _ := claims["full_name"].(string)
		if fullName == "" {
			fullName = username // 기본값
		}

		// 사용자명 충돌 방지 (선택적)
		existingUser, _ := h.authService.GetUserByUsername(c.Context(), username)
		if existingUser != nil {
			// 이미 같은 사용자명이 있으면 suffix 추가
			username = username + "_" + externalID[:6]
		}

		// 외부 사용자 등록
		user, err = h.authService.RegisterExternal(
			c.Context(),
			username,
			email,
			fullName,
			externalID,
			externalSystem,
		)

		if err != nil {
			fmt.Printf("사용자 생성 실패: %v\n", err)
			return c.Redirect("/")
		}
	}

	// go-board 내부 토큰 발급
	internalToken, err := h.authService.GenerateTokenForUser(c.Context(), user.ID)
	if err != nil {
		fmt.Printf("토큰 생성 실패: %v\n", err)
		return c.Redirect("/")
	}

	// 토큰을 쿠키에 저장
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    internalToken,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		SameSite: "none", // 크로스 도메인 쿠키
		Secure:   true,   // HTTPS 필수
	})

	// 원래 요청하려던 페이지로 리다이렉트
	targetPath := c.Query("redirect", "/")
	return c.Redirect(targetPath)
}

// 외부 시스템 로그아웃 API
func (h *BlindAuthHandler) ExternalLogout(c *fiber.Ctx) error {
	// 요청 파라미터 검증
	externalID := c.FormValue("external_id")
	externalSystem := c.FormValue("system", "main_system")
	apiKey := c.FormValue("api_key")

	// API 키 검증
	expectedKey := utils.GenerateHMAC(externalID+externalSystem, h.secretKey)
	if apiKey != expectedKey {
		return c.Status(401).JSON(fiber.Map{
			"success": false,
			"message": "인증 실패",
		})
	}

	// 외부 ID로 사용자 검색
	user, err := h.authService.GetUserByExternalID(c.Context(), externalID, externalSystem)
	if err == nil && user != nil {
		// 해당 사용자의 모든 토큰 무효화
		err = h.authService.InvalidateUserTokens(c.Context(), user.ID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "토큰 무효화 실패",
			})
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

// 토큰 검증 함수
func (h *BlindAuthHandler) validateToken(tokenString string) (map[string]interface{}, error) {
	// 토큰 구조: header.payload.signature
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("잘못된 토큰 형식")
	}

	// 서명 검증
	message := parts[0] + "." + parts[1]
	signatureBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, err
	}

	// HMAC SHA-256 서명 계산
	mac := hmac.New(sha256.New, []byte(h.secretKey))
	mac.Write([]byte(message))
	expectedSignature := mac.Sum(nil)

	// 서명 비교
	if !hmac.Equal(signatureBytes, expectedSignature) {
		return nil, fmt.Errorf("유효하지 않은 토큰 서명")
	}

	// 페이로드 디코딩
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	// JSON 페이로드 파싱
	var claims map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, err
	}

	// 만료 시간 확인
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("토큰이 만료됨")
		}
	}

	return claims, nil
}
