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
	GetUniqueIPsForReferrer(ctx context.Context, referrerURL string, startDate time.Time) ([]string, []string, error)
	GetUniqueIPsForDomain(ctx context.Context, domain string, startDate time.Time) ([]string, []string, error)
	GetUniqueIPsWithUA(ctx context.Context, referrerURL string, startDate time.Time) ([]models.IPUserAgentInfo, error)
	GetUniqueIPsWithUAForDomain(ctx context.Context, domain string, startDate time.Time) ([]models.IPUserAgentInfo, error)
	GetTargetURLsForReferrer(ctx context.Context, referrerURL string, startDate time.Time) ([]string, error)
	GetIPDetails(ctx context.Context, ipAddress string, startDate time.Time) (*models.IPDetail, error)
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

		visitorIPs, userAgents, _ := r.GetUniqueIPsForReferrer(ctx, res.ReferrerURL, startDate)

		summaries[i] = &models.ReferrerSummary{
			ReferrerURL:    res.ReferrerURL,
			ReferrerDomain: res.ReferrerDomain,
			ReferrerType:   res.ReferrerType,
			Count:          res.Count,
			UniqueCount:    res.UniqueCount,
			PercentTotal:   percent,
			VisitorIPs:     visitorIPs,
			UserAgents:     userAgents,
		}
	}

	return summaries, nil
}

// GetUniqueIPsForReferrer는 레퍼러 URL에 대한 모든 고유 방문자 IP와 User-Agent를 찾습니다
func (r *referrerRepository) GetUniqueIPsForReferrer(ctx context.Context, referrerURL string, startDate time.Time) ([]string, []string, error) {
	type IpInfo struct {
		VisitorIP string `bun:"visitor_ip"`
		UserAgent string `bun:"user_agent"`
	}

	var ipInfos []IpInfo

	err := r.db.NewSelect().
		TableExpr("referrer_stats").
		ColumnExpr("DISTINCT visitor_ip").
		ColumnExpr("user_agent").
		Where("referrer_url = ? AND visit_time >= ?", referrerURL, startDate).
		GroupExpr("visitor_ip, user_agent").
		OrderExpr("COUNT(*) DESC").
		Limit(50). // 최대 50개까지만 표시 (UI 과부하 방지)
		Scan(ctx, &ipInfos)

	if err != nil {
		return nil, nil, err
	}

	// 중복 제거를 위한 맵 사용
	ipMap := make(map[string]bool)
	uaMap := make(map[string]bool)

	for _, info := range ipInfos {
		ipMap[info.VisitorIP] = true
		uaMap[info.UserAgent] = true
	}

	// 배열 변환
	ips := make([]string, 0, len(ipMap))
	for ip := range ipMap {
		ips = append(ips, ip)
	}

	userAgents := make([]string, 0, len(uaMap))
	for ua := range uaMap {
		userAgents = append(userAgents, ua)
	}

	return ips, userAgents, nil
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

		// 기존의 단일 IP 조회 코드를 제거하고 새 함수로 대체
		// 해당 도메인에 대한 가장 많이 사용된 IP와 User-Agent 찾기
		visitorIPs, userAgents, _ := r.GetUniqueIPsForDomain(ctx, res.ReferrerDomain, startDate)

		summaries[i] = &models.ReferrerSummary{
			ReferrerDomain: res.ReferrerDomain,
			ReferrerType:   res.ReferrerType,
			Count:          res.Count,
			UniqueCount:    res.UniqueCount,
			PercentTotal:   percent,
			VisitorIPs:     visitorIPs, // 배열로 변경
			UserAgents:     userAgents, // 배열로 변경
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

// GetUniqueIPsForDomain은 도메인에 대한 모든 고유 방문자 IP와 User-Agent를 찾습니다
func (r *referrerRepository) GetUniqueIPsForDomain(ctx context.Context, domain string, startDate time.Time) ([]string, []string, error) {
	type IpInfo struct {
		VisitorIP string `bun:"visitor_ip"`
		UserAgent string `bun:"user_agent"`
	}

	var ipInfos []IpInfo

	err := r.db.NewSelect().
		TableExpr("referrer_stats").
		ColumnExpr("DISTINCT visitor_ip").
		ColumnExpr("user_agent").
		Where("referrer_domain = ? AND visit_time >= ?", domain, startDate).
		GroupExpr("visitor_ip, user_agent").
		OrderExpr("COUNT(*) DESC").
		Limit(50). // 최대 50개까지만 표시
		Scan(ctx, &ipInfos)

	if err != nil {
		return nil, nil, err
	}

	// 중복 제거를 위한 맵 사용
	ipMap := make(map[string]bool)
	uaMap := make(map[string]bool)

	for _, info := range ipInfos {
		ipMap[info.VisitorIP] = true
		uaMap[info.UserAgent] = true
	}

	// 배열 변환
	ips := make([]string, 0, len(ipMap))
	for ip := range ipMap {
		ips = append(ips, ip)
	}

	userAgents := make([]string, 0, len(uaMap))
	for ua := range uaMap {
		userAgents = append(userAgents, ua)
	}

	return ips, userAgents, nil
}

// 특정 레퍼러의 대상 URL 목록을 가져옵니다
func (r *referrerRepository) GetTargetURLsForReferrer(ctx context.Context, referrerURL string, startDate time.Time) ([]string, error) {
	var targets []struct {
		TargetURL string `bun:"target_url"`
		Count     int    `bun:"count"`
	}

	err := r.db.NewSelect().
		TableExpr("referrer_stats").
		ColumnExpr("target_url").
		ColumnExpr("COUNT(*) as count").
		Where("referrer_url = ? AND visit_time >= ?", referrerURL, startDate).
		GroupExpr("target_url").
		OrderExpr("count DESC").
		Limit(5). // 상위 5개만
		Scan(ctx, &targets)

	if err != nil {
		return nil, err
	}

	// 결과 변환
	urls := make([]string, len(targets))
	for i, target := range targets {
		urls[i] = target.TargetURL
	}

	return urls, nil
}

func (r *referrerRepository) GetIPDetails(ctx context.Context, ipAddress string, startDate time.Time) (*models.IPDetail, error) {
	// IP 주소에 대한 방문 정보 조회
	var visits []struct {
		TargetURL   string    `bun:"target_url"`
		UserAgent   string    `bun:"user_agent"`
		VisitTime   time.Time `bun:"visit_time"`
		ReferrerURL string    `bun:"referrer_url"`
	}

	err := r.db.NewSelect().
		TableExpr("referrer_stats").
		ColumnExpr("target_url, user_agent, visit_time, referrer_url").
		Where("visitor_ip = ? AND visit_time >= ?", ipAddress, startDate).
		OrderExpr("visit_time DESC").
		Limit(50). // 최대 50개 방문 기록
		Scan(ctx, &visits)

	if err != nil {
		return nil, err
	}

	// 방문 기록이 없으면 빈 결과 반환
	if len(visits) == 0 {
		return &models.IPDetail{
			IP: ipAddress,
		}, nil
	}

	// 결과 구성
	detail := &models.IPDetail{
		IP:         ipAddress,
		VisitCount: len(visits),
		LastVisit:  visits[0].VisitTime.Format("2006-01-02 15:04:05"),
	}

	// 중복 제거용 맵
	targetMap := make(map[string]bool)
	uaMap := make(map[string]bool)
	referrerMap := make(map[string]bool)

	// 고유 타겟 URL, UserAgent, 레퍼러 추출
	for _, visit := range visits {
		if !targetMap[visit.TargetURL] {
			targetMap[visit.TargetURL] = true
			detail.TargetURLs = append(detail.TargetURLs, visit.TargetURL)
		}

		if !uaMap[visit.UserAgent] {
			uaMap[visit.UserAgent] = true
			detail.UserAgents = append(detail.UserAgents, visit.UserAgent)
		}

		if !referrerMap[visit.ReferrerURL] {
			referrerMap[visit.ReferrerURL] = true
			detail.ReferrerURLs = append(detail.ReferrerURLs, visit.ReferrerURL)
		}
	}

	return detail, nil
}

// GetUniqueIPsWithUA는 레퍼러 URL에 대한 모든 고유 방문자 IP와 User-Agent를 함께 반환합니다
func (r *referrerRepository) GetUniqueIPsWithUA(ctx context.Context, referrerURL string, startDate time.Time) ([]models.IPUserAgentInfo, error) {
	type IpInfo struct {
		VisitorIP string `bun:"visitor_ip"`
		UserAgent string `bun:"user_agent"`
	}

	var ipInfos []IpInfo

	err := r.db.NewSelect().
		TableExpr("referrer_stats").
		ColumnExpr("visitor_ip").
		ColumnExpr("user_agent").
		Where("referrer_url = ? AND visit_time >= ?", referrerURL, startDate).
		GroupExpr("visitor_ip, user_agent").
		OrderExpr("COUNT(*) DESC").
		Limit(50). // 최대 50개까지만 표시 (UI 과부하 방지)
		Scan(ctx, &ipInfos)

	if err != nil {
		return nil, err
	}

	// IP 정보를 새 구조체로 변환
	result := make([]models.IPUserAgentInfo, 0, len(ipInfos))
	for _, info := range ipInfos {
		// User-Agent 분석
		uaInfo := utils.AnalyzeUserAgent(info.UserAgent)

		// IP별 상세 정보 생성
		ipDetail := models.IPUserAgentInfo{
			IP:        info.VisitorIP,
			UserAgent: info.UserAgent,
			IsBot:     uaInfo.IsBot,
		}

		result = append(result, ipDetail)
	}

	return result, nil
}

// GetUniqueIPsWithUAForDomain은 도메인에 대한 모든 고유 방문자 IP와 User-Agent를 함께 반환합니다
func (r *referrerRepository) GetUniqueIPsWithUAForDomain(ctx context.Context, domain string, startDate time.Time) ([]models.IPUserAgentInfo, error) {
	type IpInfo struct {
		VisitorIP string `bun:"visitor_ip"`
		UserAgent string `bun:"user_agent"`
	}

	var ipInfos []IpInfo

	err := r.db.NewSelect().
		TableExpr("referrer_stats").
		ColumnExpr("visitor_ip").
		ColumnExpr("user_agent").
		Where("referrer_domain = ? AND visit_time >= ?", domain, startDate).
		GroupExpr("visitor_ip, user_agent").
		OrderExpr("COUNT(*) DESC").
		Limit(50). // 최대 50개까지만 표시
		Scan(ctx, &ipInfos)

	if err != nil {
		return nil, err
	}

	// IP 정보를 새 구조체로 변환
	result := make([]models.IPUserAgentInfo, 0, len(ipInfos))
	for _, info := range ipInfos {
		// User-Agent 분석
		uaInfo := utils.AnalyzeUserAgent(info.UserAgent)

		// IP별 상세 정보 생성
		ipDetail := models.IPUserAgentInfo{
			IP:        info.VisitorIP,
			UserAgent: info.UserAgent,
			IsBot:     uaInfo.IsBot,
		}

		result = append(result, ipDetail)
	}

	return result, nil
}
