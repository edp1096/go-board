// internal/utils/auth_utils.go
package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateRandomString 랜덤 문자열 생성
func GenerateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "fallback_random_string"
	}
	return hex.EncodeToString(bytes)[:length]
}

// GenerateHMAC HMAC 생성
func GenerateHMAC(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// SecureCompare 상수 시간 비교 (타이밍 공격 방지)
func SecureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	result := 0
	for i := 0; i < len(a); i++ {
		result |= int(a[i] ^ b[i])
	}

	return result == 0
}
