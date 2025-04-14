// internal/repository/referrer_repository.go
package repository

import (
	"context"
	"time"

	"github.com/edp1096/go-board/internal/models"
	"github.com/edp1096/go-board/internal/utils"

	"github.com/uptrace/bun"
)

type ReferrerRepository interface {
	Create(ctx context.Context, stat *models.ReferrerStat) error
	GetTopReferrers(ctx context.Context, limit int, days int) ([]*models.ReferrerSummary, error)
	GetTopReferrersByDomain(ctx context.Context, limit int, days int) ([]*models.ReferrerSummary, error)
	GetReferrersByType(ctx context.Context, days int) ([]*models.ReferrerTypeStats, error)
	GetReferrersByDate(ctx context.Context, days int) ([]*models.ReferrerTimeStats, error)
	GetTotal(ctx context.Context, days int) (int, error)
	GetMostCommonIpForReferrer(ctx context.Context, referrerURL string, startDate time.Time) (string, string, error)
}

type referrerRepository struct {
	db *bun.DB
}

func NewReferrerRepository(db *bun.DB) ReferrerRepository {
	return &referrerRepository{db: db}
}

func (r *referrerRepository) Create(ctx context.Context, stat *models.ReferrerStat) error {
	_, err := r.db.NewInsert().Model(stat).Exec(ctx)
	return err
}

func (r *referrerRepository) GetTopReferrers(ctx context.Context, limit int, days int) ([]*models.ReferrerSummary, error) {
	if limit <= 0 {
		limit = 10
	}

	// 필터링 날짜 계산
	startDate := time.Now().AddDate(0, 0, -days)

	var results []struct {
		ReferrerURL    string `bun:"referrer_url"`
		ReferrerDomain string `bun:"referrer_domain"`
		ReferrerType   string `bun:"referrer_type"`
		Count          int    `bun:"count"`
		UniqueCount    int    `bun:"unique_count"`
	}

	query := r.db.NewSelect().
		TableExpr("referrer_stats").
		ColumnExpr("referrer_url").
		ColumnExpr("referrer_domain").
		ColumnExpr("referrer_type").
		ColumnExpr("COUNT(*) AS count")

	// 데이터베이스별 고유 방문자 카운트 처리
	if utils.IsPostgres(r.db) || utils.IsSQLite(r.db) || utils.IsMySQL(r.db) {
		query = query.ColumnExpr("COUNT(DISTINCT visitor_ip) AS unique_count")
	}

	err := query.Where("visit_time >= ?", startDate).
		GroupExpr("referrer_url, referrer_domain, referrer_type").
		OrderExpr("count DESC").
		Limit(limit).
		Scan(ctx, &results)

	if err != nil {
		return nil, err
	}

	// 백분율 계산을 위한 총 방문 수
	total, err := r.GetTotal(ctx, days)
	if err != nil {
		return nil, err
	}

	// 결과 변환
	summaries := make([]*models.ReferrerSummary, len(results))
	for i, res := range results {
		percent := 0.0
		if total > 0 {
			percent = (float64(res.Count) / float64(total)) * 100
		}

		// 각 레퍼러에 대한 가장 일반적인 IP와 UserAgent 가져오기
		visitorIP, userAgent, _ := r.GetMostCommonIpForReferrer(ctx, res.ReferrerURL, startDate)

		summaries[i] = &models.ReferrerSummary{
			ReferrerURL:    res.ReferrerURL,
			ReferrerDomain: res.ReferrerDomain,
			ReferrerType:   res.ReferrerType,
			Count:          res.Count,
			UniqueCount:    res.UniqueCount,
			PercentTotal:   percent,
			VisitorIP:      visitorIP,
			UserAgent:      userAgent,
		}
	}

	return summaries, nil
}

// 레퍼러 URL에 대한 가장 일반적인 방문자 IP와 User-Agent를 찾습니다
func (r *referrerRepository) GetMostCommonIpForReferrer(ctx context.Context, referrerURL string, startDate time.Time) (string, string, error) {
	type IpInfo struct {
		VisitorIP string `bun:"visitor_ip"`
		UserAgent string `bun:"user_agent"`
		Count     int    `bun:"count"`
	}

	var ipInfo IpInfo

	err := r.db.NewSelect().
		TableExpr("referrer_stats").
		ColumnExpr("visitor_ip").
		ColumnExpr("user_agent").
		ColumnExpr("COUNT(*) as count").
		Where("referrer_url = ? AND visit_time >= ?", referrerURL, startDate).
		GroupExpr("visitor_ip, user_agent").
		OrderExpr("count DESC").
		Limit(1).
		Scan(ctx, &ipInfo)

	if err != nil {
		return "", "", err
	}

	return ipInfo.VisitorIP, ipInfo.UserAgent, nil
}

func (r *referrerRepository) GetTopReferrersByDomain(ctx context.Context, limit int, days int) ([]*models.ReferrerSummary, error) {
	if limit <= 0 {
		limit = 10
	}

	// 필터링 날짜 계산
	startDate := time.Now().AddDate(0, 0, -days)

	var results []struct {
		ReferrerDomain string `bun:"referrer_domain"`
		ReferrerType   string `bun:"referrer_type"`
		Count          int    `bun:"count"`
		UniqueCount    int    `bun:"unique_count"`
	}

	query := r.db.NewSelect().
		TableExpr("referrer_stats").
		ColumnExpr("referrer_domain").
		ColumnExpr("referrer_type").
		ColumnExpr("COUNT(*) AS count")

	// 데이터베이스별 고유 방문자 카운트 처리
	if utils.IsPostgres(r.db) || utils.IsSQLite(r.db) || utils.IsMySQL(r.db) {
		query = query.ColumnExpr("COUNT(DISTINCT visitor_ip) AS unique_count")
	}

	err := query.Where("visit_time >= ?", startDate).
		GroupExpr("referrer_domain, referrer_type").
		OrderExpr("count DESC").
		Limit(limit).
		Scan(ctx, &results)

	if err != nil {
		return nil, err
	}

	// 백분율 계산을 위한 총 방문 수
	total, err := r.GetTotal(ctx, days)
	if err != nil {
		return nil, err
	}

	// 결과 변환
	summaries := make([]*models.ReferrerSummary, len(results))
	for i, res := range results {
		percent := 0.0
		if total > 0 {
			percent = (float64(res.Count) / float64(total)) * 100
		}

		// 도메인에 대한 가장 일반적인 IP 가져오기 쿼리
		var commonIPInfo struct {
			VisitorIP string `bun:"visitor_ip"`
			UserAgent string `bun:"user_agent"`
		}

		// 해당 도메인에 대한 가장 많이 사용된 IP와 User-Agent 찾기
		err := r.db.NewSelect().
			TableExpr("referrer_stats").
			ColumnExpr("visitor_ip").
			ColumnExpr("user_agent").
			Where("referrer_domain = ? AND visit_time >= ?", res.ReferrerDomain, startDate).
			GroupExpr("visitor_ip, user_agent").
			OrderExpr("COUNT(*) DESC").
			Limit(1).
			Scan(ctx, &commonIPInfo)

		visitorIP := ""
		userAgent := ""
		if err == nil {
			visitorIP = commonIPInfo.VisitorIP
			userAgent = commonIPInfo.UserAgent
		}

		summaries[i] = &models.ReferrerSummary{
			ReferrerDomain: res.ReferrerDomain,
			ReferrerType:   res.ReferrerType,
			Count:          res.Count,
			UniqueCount:    res.UniqueCount,
			PercentTotal:   percent,
			VisitorIP:      visitorIP,
			UserAgent:      userAgent,
		}
	}

	return summaries, nil
}

func (r *referrerRepository) GetReferrersByType(ctx context.Context, days int) ([]*models.ReferrerTypeStats, error) {
	// 필터링 날짜 계산
	startDate := time.Now().AddDate(0, 0, -days)

	var results []struct {
		Type  string `bun:"referrer_type"`
		Count int    `bun:"count"`
	}

	err := r.db.NewSelect().
		TableExpr("referrer_stats").
		ColumnExpr("referrer_type").
		ColumnExpr("COUNT(*) AS count").
		Where("visit_time >= ?", startDate).
		GroupExpr("referrer_type").
		OrderExpr("count DESC").
		Scan(ctx, &results)

	if err != nil {
		return nil, err
	}

	// 백분율 계산을 위한 총 방문 수
	total, err := r.GetTotal(ctx, days)
	if err != nil {
		return nil, err
	}

	// 결과 변환
	stats := make([]*models.ReferrerTypeStats, len(results))
	for i, r := range results {
		percent := 0.0
		if total > 0 {
			percent = (float64(r.Count) / float64(total)) * 100
		}

		stats[i] = &models.ReferrerTypeStats{
			Type:         r.Type,
			Count:        r.Count,
			PercentTotal: percent,
		}
	}

	return stats, nil
}

func (r *referrerRepository) GetReferrersByDate(ctx context.Context, days int) ([]*models.ReferrerTimeStats, error) {
	// 필터링 날짜 계산
	startDate := time.Now().AddDate(0, 0, -days)

	var results []*models.ReferrerTimeStats
	var query *bun.SelectQuery

	// 데이터베이스별 날짜 포맷팅
	if utils.IsPostgres(r.db) {
		query = r.db.NewSelect().
			TableExpr("referrer_stats").
			ColumnExpr("TO_CHAR(visit_time, 'YYYY-MM-DD') AS date").
			ColumnExpr("COUNT(*) AS count")
	} else if utils.IsSQLite(r.db) {
		query = r.db.NewSelect().
			TableExpr("referrer_stats").
			ColumnExpr("strftime('%Y-%m-%d', visit_time) AS date").
			ColumnExpr("COUNT(*) AS count")
	} else {
		// MySQL
		query = r.db.NewSelect().
			TableExpr("referrer_stats").
			ColumnExpr("DATE_FORMAT(visit_time, '%Y-%m-%d') AS date").
			ColumnExpr("COUNT(*) AS count")
	}

	err := query.Where("visit_time >= ?", startDate).
		GroupExpr("date").
		OrderExpr("date ASC").
		Scan(ctx, &results)

	return results, err
}

func (r *referrerRepository) GetTotal(ctx context.Context, days int) (int, error) {
	// 필터링 날짜 계산
	startDate := time.Now().AddDate(0, 0, -days)

	var count int
	err := r.db.NewSelect().
		TableExpr("referrer_stats").
		ColumnExpr("COUNT(*) AS count").
		Where("visit_time >= ?", startDate).
		Scan(ctx, &count)

	return count, err
}
