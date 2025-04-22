// internal/utils/ua_analyzer.go
package utils

import (
	"regexp"
	"strings"
)

// UserAgentInfo는 User-Agent 분석 결과를 담는 구조체입니다
type UserAgentInfo struct {
	OriginalUA string `json:"originalUA"` // 원본 User-Agent
	IsBot      bool   `json:"isBot"`      // 봇 여부
	BotName    string `json:"botName"`    // 봇 이름 (봇인 경우)
	Browser    string `json:"browser"`    // 브라우저 이름
	BrowserVer string `json:"browserVer"` // 브라우저 버전
	OS         string `json:"os"`         // 운영체제
	OSVer      string `json:"osVer"`      // 운영체제 버전
	Device     string `json:"device"`     // 기기 정보
	IsMobile   bool   `json:"isMobile"`   // 모바일 여부
}

// 주요 봇 패턴 정의
var botPatterns = map[string]*regexp.Regexp{
	"Googlebot":      regexp.MustCompile(`(?i)googlebot`),
	"Bingbot":        regexp.MustCompile(`(?i)bingbot`),
	"Yandex":         regexp.MustCompile(`(?i)yandex`),
	"Baidu":          regexp.MustCompile(`(?i)baiduspider`),
	"Facebook":       regexp.MustCompile(`(?i)facebookexternalhit`),
	"Twitter":        regexp.MustCompile(`(?i)twitterbot`),
	"DuckDuckGo":     regexp.MustCompile(`(?i)duckduckbot`),
	"Python Crawler": regexp.MustCompile(`(?i)python-requests|python-urllib`),
	"Generic Bot":    regexp.MustCompile(`(?i)bot|crawler|spider|slurp|search`),
}

// 브라우저 패턴 정의
var browserPatterns = map[string]*regexp.Regexp{
	"Edge":      regexp.MustCompile(`(?i)Edg\/([0-9.]+)`),
	"Samsung":   regexp.MustCompile(`(?i)SamsungBrowser\/([0-9.]+)`),
	"Opera":     regexp.MustCompile(`(?i)Opera\/([0-9.]+)`),
	"UCBrowser": regexp.MustCompile(`(?i)UCBrowser\/([0-9.]+)`),
	"Chrome":    regexp.MustCompile(`(?i)Chrome\/([0-9.]+)`),
	"Firefox":   regexp.MustCompile(`(?i)Firefox\/([0-9.]+)`),
	"Safari":    regexp.MustCompile(`(?i)Safari\/([0-9.]+)`),
	"IE":        regexp.MustCompile(`(?i)MSIE ([0-9.]+)`),
}

// OS 패턴 정의
var osPatterns = map[string]*regexp.Regexp{
	"Windows": regexp.MustCompile(`(?i)Windows NT ([0-9.]+)`),
	"macOS":   regexp.MustCompile(`(?i)Mac OS X ([0-9_\.]+)`),
	"iOS":     regexp.MustCompile(`(?i)iPhone OS ([0-9_]+)`),
	"Android": regexp.MustCompile(`(?i)Android ([0-9.]+)`),
	"Linux":   regexp.MustCompile(`(?i)Linux`),
	"Ubuntu":  regexp.MustCompile(`(?i)Ubuntu`),
	"FreeBSD": regexp.MustCompile(`(?i)FreeBSD`),
}

// 모바일 패턴 정의
var mobilePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)Mobile`),
	regexp.MustCompile(`(?i)Android`),
	regexp.MustCompile(`(?i)iPhone`),
	regexp.MustCompile(`(?i)iPad`),
	regexp.MustCompile(`(?i)Windows Phone`),
}

// AnalyzeUserAgent는 User-Agent 문자열을 분석하여 정보를 반환합니다
func AnalyzeUserAgent(ua string) UserAgentInfo {
	info := UserAgentInfo{
		OriginalUA: ua,
	}

	// 빈 UA 처리
	if ua == "" {
		info.Browser = "Unknown"
		info.OS = "Unknown"
		return info
	}

	// 봇 감지
	for botName, pattern := range botPatterns {
		if pattern.MatchString(ua) {
			info.IsBot = true
			info.BotName = botName
			info.Browser = "Bot"
			return info
		}
	}

	// 브라우저 감지
	info.Browser = "Unknown"
	for browser, pattern := range browserPatterns {
		matches := pattern.FindStringSubmatch(ua)
		if len(matches) > 1 {
			info.Browser = browser
			info.BrowserVer = matches[1]
			break
		}
	}

	// 특별 케이스: Safari와 Chrome을 구분 (많은 브라우저가 Safari 문자열을 포함)
	if info.Browser == "Safari" && strings.Contains(ua, "Chrome") {
		chromeMatches := browserPatterns["Chrome"].FindStringSubmatch(ua)
		if len(chromeMatches) > 1 {
			info.Browser = "Chrome"
			info.BrowserVer = chromeMatches[1]
		}
	}

	// OS 감지
	info.OS = "Unknown"
	for os, pattern := range osPatterns {
		matches := pattern.FindStringSubmatch(ua)
		if len(matches) > 1 {
			info.OS = os
			info.OSVer = matches[1]
			// macOS 버전 표시 개선
			if os == "macOS" {
				info.OSVer = strings.Replace(info.OSVer, "_", ".", -1)
			}
			break
		} else if pattern.MatchString(ua) {
			info.OS = os
			break
		}
	}

	// 모바일 감지
	for _, pattern := range mobilePatterns {
		if pattern.MatchString(ua) {
			info.IsMobile = true
			break
		}
	}

	// 디바이스 정보 추론
	if info.IsMobile {
		if strings.Contains(ua, "iPhone") {
			info.Device = "iPhone"
		} else if strings.Contains(ua, "iPad") {
			info.Device = "iPad"
		} else if strings.Contains(ua, "Android") {
			// 안드로이드 기기명 추출 시도
			deviceMatch := regexp.MustCompile(`(?i);\s*([^;]+)\s+Build/`).FindStringSubmatch(ua)
			if len(deviceMatch) > 1 {
				info.Device = deviceMatch[1]
			} else {
				info.Device = "Android Device"
			}
		} else {
			info.Device = "Mobile Device"
		}
	} else {
		info.Device = "Desktop/Laptop"
	}

	return info
}

// GetCategoryIcon은 User-Agent 카테고리에 따른 아이콘 클래스를 반환합니다
func GetUAIcon(uaInfo UserAgentInfo) string {
	if uaInfo.IsBot {
		return "fa-robot"
	}

	if uaInfo.IsMobile {
		return "fa-mobile-alt"
	}

	switch uaInfo.Browser {
	case "Chrome":
		return "fa-chrome"
	case "Firefox":
		return "fa-firefox-browser"
	case "Safari":
		return "fa-safari"
	case "Edge":
		return "fa-edge"
	case "IE":
		return "fa-internet-explorer"
	case "Opera":
		return "fa-opera"
	default:
		return "fa-globe"
	}
}

// GetBotName은 봇 User-Agent에서 봇 이름을 추출합니다
func GetBotName(ua string) string {
	for botName, pattern := range botPatterns {
		if pattern.MatchString(ua) {
			return botName
		}
	}
	return "Unknown Bot"
}
