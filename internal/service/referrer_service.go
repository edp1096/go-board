// internal/service/referrer_service.go
package service

import (
	"context"
	"time"

	"github.com/edp1096/go-board/internal/models"
	"github.com/edp1096/go-board/internal/repository"
)

type ReferrerService interface {
	RecordVisit(ctx context.Context, referrerURL, targetURL, visitorIP, userAgent string, userID *int64) error
	GetTopReferrers(ctx context.Context, limit, days int) ([]*models.ReferrerSummary, error)
	GetReferrersByDate(ctx context.Context, days int) ([]*models.ReferrerTimeStats, error)
	GetTotal(ctx context.Context, days int) (int, error)
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
	stat := &models.ReferrerStat{
		ReferrerURL: referrerURL,
		TargetURL:   targetURL,
		VisitorIP:   visitorIP,
		UserID:      userID,
		UserAgent:   userAgent,
		VisitTime:   time.Now(),
	}

	return s.referrerRepo.Create(ctx, stat)
}

func (s *referrerService) GetTopReferrers(ctx context.Context, limit, days int) ([]*models.ReferrerSummary, error) {
	return s.referrerRepo.GetTopReferrers(ctx, limit, days)
}

func (s *referrerService) GetReferrersByDate(ctx context.Context, days int) ([]*models.ReferrerTimeStats, error) {
	return s.referrerRepo.GetReferrersByDate(ctx, days)
}

func (s *referrerService) GetTotal(ctx context.Context, days int) (int, error) {
	return s.referrerRepo.GetTotal(ctx, days)
}
