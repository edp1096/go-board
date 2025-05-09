// utils/template.go
package utils

import (
	"html"
	"regexp"
	"strings"
)

// HTML을 일반 텍스트로 변환하는 함수
func PlainText(htmlContent string) string {
	// 1. 이미지 태그 완전히 제거
	imgRegex := regexp.MustCompile(`<img[^>]*>`)
	text := imgRegex.ReplaceAllString(htmlContent, "")

	// 2. 모든 HTML 태그 제거
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	text = tagRegex.ReplaceAllString(text, "")

	// 3. HTML 엔티티 디코딩
	text = html.UnescapeString(text)

	// 4. 연속된 공백 제거 및 트림
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	// 5. 빈 문자열 처리
	if text == "" {
		text = "게시판 시스템의 게시물입니다."
	}

	return text
}

// 텍스트 길이 제한 함수
func TruncateText(text string, maxLength int) string {
	plainText := PlainText(text)
	if len(plainText) <= maxLength {
		return plainText
	}
	return plainText[:maxLength] + "..."
}

// TrimSpace
func TrimSpace(textContent string) string {
	return strings.TrimSpace(textContent)
}
