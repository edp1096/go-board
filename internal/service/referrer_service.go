// internal/service/referrer_service.go
package service

import (
	"context"
	"sync"
	"time"

	"github.com/edp1096/toy-board/internal/models"
	"github.com/edp1096/toy-board/internal/repository"
	"github.com/edp1096/toy-board/internal/utils"
)

type ReferrerService interface {
	RecordVisit(ctx context.Context, referrerURL, targetURL, visitorIP, userAgent string, userID *int64) error
	GetTopReferrers(ctx context.Context, limit, days int) ([]*models.ReferrerSummary, error)
	GetTopReferrersByDomain(ctx context.Context, limit, days int) ([]*models.ReferrerSummary, error)
	GetReferrersByType(ctx context.Context, days int) ([]*models.ReferrerTypeStats, error)
	GetReferrersByDate(ctx context.Context, days int) ([]*models.ReferrerTimeStats, error)
	GetTotal(ctx context.Context, days int) (int, error)
	EnrichReferrerData(referrers []*models.ReferrerSummary) // DNS 정보 보강
	GetIPDetails(ctx context.Context, ipAddress string, startDate time.Time) (*models.IPDetail, error)
}

type referrerService struct {
	referrerRepo repository.ReferrerRepository
}

func NewReferrerService(referrerRepo repository.ReferrerRepository) ReferrerService {
	return &referrerService{
		referrerRepo: referrerRepo,
	}
}

func (s *referrerService) RecordVisit(ctx context.Context, referrerURL, targetURL, visitorIP, userAgent string, userID *int64) error {
	// 도메인 추출 및 타입 분류
	domain := utils.ExtractDomain(referrerURL)
	refType := utils.ClassifyReferrerType(domain)

	stat := &models.ReferrerStat{
		ReferrerURL:    referrerURL,
		ReferrerDomain: domain,
		ReferrerType:   refType,
		TargetURL:      targetURL,
		VisitorIP:      visitorIP,
		UserID:         userID,
		UserAgent:      userAgent,
		VisitTime:      time.Now(),
	}

	return s.referrerRepo.Create(ctx, stat)
}

// GetTopReferrers 함수에 User-Agent 분석 추가
func (s *referrerService) GetTopReferrers(ctx context.Context, limit, days int) ([]*models.ReferrerSummary, error) {
	referrers, err := s.referrerRepo.GetTopReferrers(ctx, limit, days)
	if err != nil {
		return nil, err
	}

	// 타겟 URL과 User-Agent 분석 정보 추가
	startDate := time.Now().AddDate(0, 0, -days)
	for _, ref := range referrers {
		// 타겟 URL 정보 추가
		if ref.ReferrerURL != "direct" && ref.ReferrerURL != "" {
			targetURLs, err := s.referrerRepo.GetTargetURLsForReferrer(ctx, ref.ReferrerURL, startDate)
			if err == nil {
				ref.TargetURLs = targetURLs
			}
		}

		// IP별 User-Agent 정보 조회 (새로운 방식)
		ipDetails, err := s.referrerRepo.GetUniqueIPsWithUA(ctx, ref.ReferrerURL, startDate)
		if err == nil {
			ref.IPDetails = ipDetails

			// 기존 호환성을 위한 필드도 유지
			ips := make([]string, 0, len(ipDetails))
			uas := make([]string, 0, len(ipDetails))

			for _, detail := range ipDetails {
				ips = append(ips, detail.IP)
				uas = append(uas, detail.UserAgent)
			}

			ref.VisitorIPs = ips
			ref.UserAgents = uas
		} else {
			// 호환성을 위해 기존 방식으로 대체
			ips, uas, _ := s.referrerRepo.GetUniqueIPsForReferrer(ctx, ref.ReferrerURL, startDate)
			ref.VisitorIPs = ips
			ref.UserAgents = uas
		}

		// User-Agent 통계 초기화
		ref.UAStats.Browsers = make(map[string]int)
		ref.UAStats.OSes = make(map[string]int)

		// User-Agent 분석
		botCount := 0
		humanCount := 0
		mobileCount := 0
		desktopCount := 0

		// IP별 User-Agent 정보 기반 통계 계산
		for _, detail := range ref.IPDetails {
			uaInfo := utils.AnalyzeUserAgent(detail.UserAgent)

			// 봇/사람 카운트
			if uaInfo.IsBot {
				botCount++
				ref.UAStats.Browsers["Bot"]++
			} else {
				humanCount++
				ref.UAStats.Browsers[uaInfo.Browser]++
			}

			// 모바일/데스크톱 카운트
			if uaInfo.IsMobile {
				mobileCount++
			} else {
				desktopCount++
			}

			// OS 카운트
			ref.UAStats.OSes[uaInfo.OS]++
		}

		ref.UAStats.BotCount = botCount
		ref.UAStats.HumanCount = humanCount
		ref.UAStats.MobileCount = mobileCount
		ref.UAStats.DesktopCount = desktopCount
	}

	return referrers, nil
}

func (s *referrerService) GetTopReferrersByDomain(ctx context.Context, limit, days int) ([]*models.ReferrerSummary, error) {
	referrers, err := s.referrerRepo.GetTopReferrersByDomain(ctx, limit, days)
	if err != nil {
		return nil, err
	}

	// 타겟 URL과 User-Agent 분석 정보 추가
	startDate := time.Now().AddDate(0, 0, -days)
	for _, ref := range referrers {
		// IP별 User-Agent 정보 조회 (새로운 방식)
		ipDetails, err := s.referrerRepo.GetUniqueIPsWithUAForDomain(ctx, ref.ReferrerDomain, startDate)
		if err == nil {
			ref.IPDetails = ipDetails

			// 기존 호환성을 위한 필드도 유지
			ips := make([]string, 0, len(ipDetails))
			uas := make([]string, 0, len(ipDetails))

			for _, detail := range ipDetails {
				ips = append(ips, detail.IP)
				uas = append(uas, detail.UserAgent)
			}

			ref.VisitorIPs = ips
			ref.UserAgents = uas
		} else {
			// 호환성을 위해 기존 방식으로 대체
			ips, uas, _ := s.referrerRepo.GetUniqueIPsForDomain(ctx, ref.ReferrerDomain, startDate)
			ref.VisitorIPs = ips
			ref.UserAgents = uas
		}

		// User-Agent 통계 초기화
		ref.UAStats.Browsers = make(map[string]int)
		ref.UAStats.OSes = make(map[string]int)

		// User-Agent 분석
		botCount := 0
		humanCount := 0
		mobileCount := 0
		desktopCount := 0

		// IP별 User-Agent 정보 기반 통계 계산
		for _, detail := range ref.IPDetails {
			uaInfo := utils.AnalyzeUserAgent(detail.UserAgent)

			// 봇/사람 카운트
			if uaInfo.IsBot {
				botCount++
				ref.UAStats.Browsers["Bot"]++
			} else {
				humanCount++
				ref.UAStats.Browsers[uaInfo.Browser]++
			}

			// 모바일/데스크톱 카운트
			if uaInfo.IsMobile {
				mobileCount++
			} else {
				desktopCount++
			}

			// OS 카운트
			ref.UAStats.OSes[uaInfo.OS]++
		}

		ref.UAStats.BotCount = botCount
		ref.UAStats.HumanCount = humanCount
		ref.UAStats.MobileCount = mobileCount
		ref.UAStats.DesktopCount = desktopCount
	}

	return referrers, nil
}

func (s *referrerService) GetReferrersByType(ctx context.Context, days int) ([]*models.ReferrerTypeStats, error) {
	return s.referrerRepo.GetReferrersByType(ctx, days)
}

func (s *referrerService) GetReferrersByDate(ctx context.Context, days int) ([]*models.ReferrerTimeStats, error) {
	return s.referrerRepo.GetReferrersByDate(ctx, days)
}

func (s *referrerService) GetTotal(ctx context.Context, days int) (int, error) {
	return s.referrerRepo.GetTotal(ctx, days)
}

// EnrichReferrerData는 레퍼러 데이터에 DNS 조회 정보와 타겟 URL 정보를 추가합니다.
// 페이지 로딩 시에는 역DNS 정보를 조회하지 않고 필수 정보만 로드합니다.
func (s *referrerService) EnrichReferrerData(referrers []*models.ReferrerSummary) {
	// 병렬 처리를 위한 워커 풀 구현
	type dnsTask struct {
		index int
		ref   *models.ReferrerSummary
	}

	const workerCount = 5
	taskCh := make(chan dnsTask, len(referrers))

	// 태스크 생성
	for i, ref := range referrers {
		taskCh <- dnsTask{
			index: i,
			ref:   ref,
		}
	}
	close(taskCh)

	// 워커 실행
	var wg sync.WaitGroup
	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for task := range taskCh {
				ref := task.ref

				// 레퍼러 도메인에 대한 정DNS 조회
				if ref.ReferrerDomain != "" && ref.ReferrerDomain != "direct" && ref.ReferrerDomain != "unknown" {
					ips := utils.GetDomainInfo(ref.ReferrerDomain)
					ref.ForwardDNS = ips
				}

				// 타겟 URL 정보 추가
				if ref.ReferrerURL != "" && ref.ReferrerURL != "direct" && len(ref.TargetURLs) == 0 {
					// 기본적으로 30일 데이터 사용
					startDate := time.Now().AddDate(0, 0, -30)
					targetURLs, err := s.referrerRepo.GetTargetURLsForReferrer(context.Background(), ref.ReferrerURL, startDate)
					if err == nil {
						ref.TargetURLs = targetURLs
					}
				}
			}
		}()
	}

	wg.Wait()
}

// GetIPDetails 함수는 모달 클릭 시 사용되며, 역DNS 조회를 포함합니다
func (s *referrerService) GetIPDetails(ctx context.Context, ipAddress string, startDate time.Time) (*models.IPDetail, error) {
	// 저장소에서 기본 IP 정보 가져오기
	detail, err := s.referrerRepo.GetIPDetails(ctx, ipAddress, startDate)
	if err != nil {
		return nil, err
	}

	// User-Agent 분석 추가
	for i, ua := range detail.UserAgents {
		uaInfo := utils.AnalyzeUserAgent(ua)

		// 분석 결과가 있으면 저장
		if i < len(detail.UAInfo) {
			detail.UAInfo[i] = models.UAInfo{
				UserAgent: ua,
				IsBot:     uaInfo.IsBot,
				BotName:   uaInfo.BotName,
				Browser:   uaInfo.Browser,
				OS:        uaInfo.OS,
				IsMobile:  uaInfo.IsMobile,
			}
		} else {
			detail.UAInfo = append(detail.UAInfo, models.UAInfo{
				UserAgent: ua,
				IsBot:     uaInfo.IsBot,
				BotName:   uaInfo.BotName,
				Browser:   uaInfo.Browser,
				OS:        uaInfo.OS,
				IsMobile:  uaInfo.IsMobile,
			})
		}
	}

	// 모달을 통해 요청될 때만 역DNS 정보 추가 (API 호출 시)
	ptr, _ := utils.LookupPTR(ipAddress)
	detail.ReverseDNS = ptr

	return detail, nil
}
