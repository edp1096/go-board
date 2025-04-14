// internal/utils/url.go
package utils

import (
	"net/url"
	"strings"
)

// ExtractDomain은 URL에서 도메인을 추출합니다
func ExtractDomain(urlStr string) string {
	if urlStr == "direct" || urlStr == "" {
		return "direct"
	}

	// URL 체계(scheme) 확인/추가
	if !strings.Contains(urlStr, "://") {
		urlStr = "http://" + urlStr
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return "unknown"
	}

	host := u.Hostname()
	if host == "" {
		// URL이 유효하지 않은 경우 원본 문자열에서 도메인 추출 시도
		parts := strings.Split(urlStr, "/")
		if len(parts) > 0 {
			host = parts[0]
		}
	}

	// 도메인 추출 (예: www.example.com -> example.com)
	domainParts := strings.Split(host, ".")
	if len(domainParts) >= 2 {
		return strings.Join(domainParts[len(domainParts)-2:], ".")
	}

	return host
}

// ClassifyReferrerType은 도메인을 기반으로 레퍼러 타입을 결정합니다
func ClassifyReferrerType(domain string) string {
	if domain == "direct" || domain == "" || domain == "unknown" {
		return "direct"
	}

	domain = strings.ToLower(domain)

	// 도메인 기반으로 타입 분류 (특정 회사명 사용 제거)
	if strings.Contains(domain, "search") ||
		strings.HasSuffix(domain, "google.com") ||
		strings.HasSuffix(domain, "naver.com") ||
		strings.HasSuffix(domain, "daum.net") ||
		strings.HasSuffix(domain, "bing.com") ||
		strings.HasSuffix(domain, "yahoo.com") ||
		strings.HasSuffix(domain, "baidu.com") ||
		strings.HasSuffix(domain, "yandex.ru") {
		return "search"
	}

	if strings.Contains(domain, "social") ||
		strings.HasSuffix(domain, "facebook.com") ||
		strings.HasSuffix(domain, "twitter.com") ||
		strings.HasSuffix(domain, "x.com") ||
		strings.HasSuffix(domain, "instagram.com") ||
		strings.HasSuffix(domain, "linkedin.com") ||
		strings.HasSuffix(domain, "pinterest.com") ||
		strings.HasSuffix(domain, "tiktok.com") ||
		strings.HasSuffix(domain, "youtube.com") {
		return "social"
	}

	return "other"
}

// IsBot은 User-Agent 문자열을 기반으로 봇인지 판단합니다
func IsBot(userAgent string) bool {
	if userAgent == "" {
		return false // 빈 User-Agent는 봇으로 간주하지 않음
	}

	userAgent = strings.ToLower(userAgent)
	botKeywords := []string{
		"bot", "crawler", "spider", "slurp", "googlebot",
		"bingbot", "yandex", "baidu", "sogou", "duckduckgo",
		"semrush", "ahrefs", "python-requests", "go-http-client",
	}

	for _, keyword := range botKeywords {
		if strings.Contains(userAgent, keyword) {
			return true
		}
	}

	return false
}

// GetReferrerDisplayText는 레퍼러 타입에 따라 표시 텍스트를 반환합니다
func GetReferrerDisplayText(refType string) string {
	switch refType {
	case "direct":
		return "직접 방문"
	case "search":
		return "검색 엔진"
	case "social":
		return "소셜 미디어"
	default:
		return "기타"
	}
}
