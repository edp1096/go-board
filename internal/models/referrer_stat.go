// internal/models/referrer_stat.go
package models

import (
	"time"

	"github.com/uptrace/bun"
)

// ReferrerStat represents a single referrer statistic entry
type ReferrerStat struct {
	bun.BaseModel `bun:"table:referrer_stats,alias:rs"`

	ID             int64     `bun:"id,pk,autoincrement" json:"id"`
	ReferrerURL    string    `bun:"referrer_url,notnull" json:"referrerUrl"`
	ReferrerDomain string    `bun:"referrer_domain" json:"referrerDomain"`
	ReferrerType   string    `bun:"referrer_type" json:"referrerType"`
	TargetURL      string    `bun:"target_url,notnull" json:"targetUrl"`
	VisitorIP      string    `bun:"visitor_ip,notnull" json:"visitorIp"`
	UserID         *int64    `bun:"user_id" json:"userId,omitempty"`
	UserAgent      string    `bun:"user_agent" json:"userAgent"`
	VisitTime      time.Time `bun:"visit_time,notnull" json:"visitTime"`

	// Relations
	User *User `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
}

// ReferrerSummary represents aggregated referrer statistics
type ReferrerSummary struct {
	ReferrerURL    string  `json:"referrerUrl"`
	ReferrerDomain string  `json:"referrerDomain"`
	ReferrerType   string  `json:"referrerType"`
	Count          int     `json:"count"`
	UniqueCount    int     `json:"uniqueCount"`
	PercentTotal   float64 `json:"percentTotal"`

	// IP 목록
	VisitorIPs []string `json:"visitorIps"` // IP 목록
	UserAgents []string `json:"userAgents"` // User-Agent 목록
	ReverseDNS string   `json:"reverseDns"` // 역DNS 조회 결과
	ForwardDNS []string `json:"forwardDns"` // 정DNS 조회 결과

	// 방문 페이지 정보 (새로 추가)
	TargetURLs []string `json:"targetUrls,omitempty"` // 방문한 타겟 URL 목록

	// User-Agent 분석 통계 (추가)
	UAStats struct {
		BotCount     int            `json:"botCount"`     // 봇 수
		HumanCount   int            `json:"humanCount"`   // 사람 수
		MobileCount  int            `json:"mobileCount"`  // 모바일 수
		DesktopCount int            `json:"desktopCount"` // 데스크톱 수
		Browsers     map[string]int `json:"browsers"`     // 브라우저별 카운트
		OSes         map[string]int `json:"oses"`         // OS별 카운트
	} `json:"uaStats,omitempty"`
}

// IPDetail은 IP 주소별 상세 정보를 담는 구조체입니다
type IPDetail struct {
	IP           string   `json:"ip"`
	VisitCount   int      `json:"visitCount"`
	LastVisit    string   `json:"lastVisit"`
	TargetURLs   []string `json:"targetUrls"`
	UserAgents   []string `json:"userAgents"`
	ReferrerURLs []string `json:"referrerUrls"`
	ReverseDNS   string   `json:"reverseDns"`
	UAInfo       []UAInfo `json:"uaInfo"`
}

// UAInfo는 User-Agent 분석 정보를 담는 구조체입니다
type UAInfo struct {
	UserAgent string `json:"userAgent"` // 원본 User-Agent
	IsBot     bool   `json:"isBot"`     // 봇 여부
	BotName   string `json:"botName"`   // 봇 이름
	Browser   string `json:"browser"`   // 브라우저
	OS        string `json:"os"`        // 운영체제
	IsMobile  bool   `json:"isMobile"`  // 모바일 여부
}

// ReferrerTimeStats represents time-based referrer statistics
type ReferrerTimeStats struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// ReferrerTypeStats represents referrer type-based statistics
type ReferrerTypeStats struct {
	Type         string  `json:"type"`
	Count        int     `json:"count"`
	PercentTotal float64 `json:"percentTotal"`
}
