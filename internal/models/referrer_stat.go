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
