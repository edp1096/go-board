// internal/service/referrer_service.go
package service

import (
	"context"
	"sync"
	"time"

	"github.com/edp1096/go-board/internal/models"
	"github.com/edp1096/go-board/internal/repository"
	"github.com/edp1096/go-board/internal/utils"
)

type ReferrerService interface {
	RecordVisit(ctx context.Context, referrerURL, targetURL, visitorIP, userAgent string, userID *int64) error
	GetTopReferrers(ctx context.Context, limit, days int) ([]*models.ReferrerSummary, error)
	GetTopReferrersByDomain(ctx context.Context, limit, days int) ([]*models.ReferrerSummary, error)
	GetReferrersByType(ctx context.Context, days int) ([]*models.ReferrerTypeStats, error)
	GetReferrersByDate(ctx context.Context, days int) ([]*models.ReferrerTimeStats, error)
	GetTotal(ctx context.Context, days int) (int, error)
	EnrichReferrerData(referrers []*models.ReferrerSummary) // DNS 정보 보강
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

func (s *referrerService) GetTopReferrers(ctx context.Context, limit, days int) ([]*models.ReferrerSummary, error) {
	return s.referrerRepo.GetTopReferrers(ctx, limit, days)
}

func (s *referrerService) GetTopReferrersByDomain(ctx context.Context, limit, days int) ([]*models.ReferrerSummary, error) {
	return s.referrerRepo.GetTopReferrersByDomain(ctx, limit, days)
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

// EnrichReferrerData는 레퍼러 데이터에 DNS 조회 정보를 추가합니다.
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

				// 1. 방문자 IP 목록에 대한 역DNS 조회
				if len(ref.VisitorIPs) > 0 {
					// 첫 번째 IP에 대해서만 역DNS 조회 (성능상 이유로)
					firstIP := ref.VisitorIPs[0]
					if firstIP != "" && firstIP != "unknown" {
						ptr, _ := utils.LookupPTR(firstIP)
						ref.ReverseDNS = ptr
					}
				}

				// 2. 레퍼러 도메인에 대한 정DNS 조회
				if ref.ReferrerDomain != "" && ref.ReferrerDomain != "direct" && ref.ReferrerDomain != "unknown" {
					ips := utils.GetDomainInfo(ref.ReferrerDomain)
					ref.ForwardDNS = ips
				}
			}
		}()
	}

	wg.Wait()
}
