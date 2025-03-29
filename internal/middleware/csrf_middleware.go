// internal/middleware/csrf_middleware.go
package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"math/rand"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	csrfTokenLength = 32
	csrfHeaderName  = "X-CSRF-Token"
	csrfCookieName  = "csrf_token"
	csrfFormField   = "_csrf"
)

// CSRFConfig CSRF 미들웨어 설정
type CSRFConfig struct {
	// TokenLookup defines where to look for the CSRF token, default: "header:X-CSRF-Token,form:_csrf,cookie:csrf_token"
	TokenLookup string
	// Cookie defines cookie options
	Cookie fiber.Cookie
	// KeyLookup is a string in the form of "<source>:<n>" that is used
	// to create a key from. Optional. Default: "header:X-CSRF-Token"
	KeyLookup string
	// CookieName is the name of the cookie to store the CSRF token
	CookieName string
	// FieldName is the name of the form field to use for the CSRF token
	FieldName string
	// HeaderName is the name of the header to use for the CSRF token
	HeaderName string
	// Expiration is the duration before the CSRF token expires (default: 1 hour)
	Expiration time.Duration
	// Cookie path
	CookiePath string
}

// CSRF 미들웨어 생성 함수
func CSRF(config ...CSRFConfig) fiber.Handler {
	// 기본 설정
	cfg := CSRFConfig{
		TokenLookup: "header:" + csrfHeaderName + ",form:" + csrfFormField + ",cookie:" + csrfCookieName,
		CookieName:  csrfCookieName,
		FieldName:   csrfFormField,
		HeaderName:  csrfHeaderName,
		Expiration:  1 * time.Hour,
		CookiePath:  "/",
	}

	// 사용자 설정 적용
	if len(config) > 0 {
		if config[0].TokenLookup != "" {
			cfg.TokenLookup = config[0].TokenLookup
		}
		if config[0].CookieName != "" {
			cfg.CookieName = config[0].CookieName
		}
		if config[0].FieldName != "" {
			cfg.FieldName = config[0].FieldName
		}
		if config[0].HeaderName != "" {
			cfg.HeaderName = config[0].HeaderName
		}
		if config[0].Expiration > 0 {
			cfg.Expiration = config[0].Expiration
		}
		if config[0].CookiePath != "" {
			cfg.CookiePath = config[0].CookiePath
		}
	}

	// 난수 생성기 초기화
	rand.Seed(time.Now().UnixNano())

	// Fiber 핸들러 반환
	return func(c *fiber.Ctx) error {
		// GET, HEAD, OPTIONS 요청은 CSRF 검증 제외
		method := c.Method()
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			// CSRF 토큰 생성 및 쿠키 설정
			token := generateToken()
			setCookie(c, cfg.CookieName, token, cfg.Expiration, cfg.CookiePath)
			c.Locals("csrf", token)
			return c.Next()
		}

		// POST, PUT, DELETE 등 상태 변경 요청은 CSRF 검증
		// 토큰 확인
		var clientToken string
		lookups := strings.Split(cfg.TokenLookup, ",")

		// 여러 소스에서 토큰 검색
		for _, lookup := range lookups {
			parts := strings.Split(lookup, ":")
			if len(parts) != 2 {
				continue
			}

			source, key := parts[0], parts[1]
			switch source {
			case "header":
				clientToken = c.Get(key)
			case "form":
				clientToken = c.FormValue(key)
			case "cookie":
				clientToken = c.Cookies(key)
			}

			if clientToken != "" {
				break
			}
		}

		// 서버 토큰 가져오기
		serverToken := c.Cookies(cfg.CookieName)

		// 토큰 비교
		if serverToken == "" || clientToken == "" || !validateToken(clientToken, serverToken) {
			// XHR 요청 또는 JSON 응답을 원하는 경우 JSON으로 응답
			if c.Get("X-Requested-With") == "XMLHttpRequest" ||
				c.Get("Accept") == "application/json" ||
				strings.Contains(c.Get("Accept"), "application/json") {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"success": false,
					"error":   true,
					"message": "CSRF 토큰이 유효하지 않습니다",
				})
			}

			// 일반 요청은 기본 오류 응답
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid CSRF token",
			})
		}

		// 새 토큰 생성 (매 요청마다 갱신)
		newToken := generateToken()
		setCookie(c, cfg.CookieName, newToken, cfg.Expiration, cfg.CookiePath)
		c.Locals("csrf", newToken)

		return c.Next()
	}
}

// 랜덤 토큰 생성
func generateToken() string {
	b := make([]byte, csrfTokenLength)
	for i := range b {
		b[i] = byte(rand.Intn(256))
	}
	return base64.StdEncoding.EncodeToString(b)
}

// 쿠키 설정 헬퍼 함수
func setCookie(c *fiber.Ctx, name, value string, expiration time.Duration, path string) {
	cookie := fiber.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Expires:  time.Now().Add(expiration),
		HTTPOnly: true,
		SameSite: "Lax",
	}
	c.Cookie(&cookie)
}

// 토큰 검증 함수
func validateToken(clientToken, serverToken string) bool {
	return subtle.ConstantTimeCompare([]byte(clientToken), []byte(serverToken)) == 1
}
