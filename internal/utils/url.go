// internal/utils/url.go
package utils

import (
	"net/url"
	"strings"
)

// ExtractDomain은 URL에서 도메인을 추출합니다
func ExtractDomain(urlStr string) string {
	if urlStr == "direct" {
		return "direct"
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

	// 도메인 추출 (예: www.google.com -> google.com)
	domainParts := strings.Split(host, ".")
	if len(domainParts) >= 2 {
		return strings.Join(domainParts[len(domainParts)-2:], ".")
	}
	return host
}

// ClassifyReferrerType은 도메인을 기반으로 레퍼러 타입을 결정합니다
func ClassifyReferrerType(domain string) string {
	if domain == "direct" || domain == "" {
		return "direct"
	}

	domain = strings.ToLower(domain)

	// 검색 엔진
	searchEngines := []string{"google", "naver", "daum", "bing", "yahoo", "baidu", "yandex"}
	for _, engine := range searchEngines {
		if strings.Contains(domain, engine) {
			return "search"
		}
	}

	// 소셜 미디어
	socialMedia := []string{"facebook", "twitter", "instagram", "linkedin", "pinterest", "tiktok", "youtube"}
	for _, social := range socialMedia {
		if strings.Contains(domain, social) {
			return "social"
		}
	}

	return "other"
}
