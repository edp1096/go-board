// internal/repository/referrer_repository.go
package repository

import (
	"context"
	"go-board/internal/models"
	"go-board/internal/utils"
	"time"

	"github.com/uptrace/bun"
)

type ReferrerRepository interface {
	Create(ctx context.Context, stat *models.ReferrerStat) error
	GetTopReferrers(ctx context.Context, limit int, days int) ([]*models.ReferrerSummary, error)
	GetReferrersByDate(ctx context.Context, days int) ([]*models.ReferrerTimeStats, error)
	GetTotal(ctx context.Context, days int) (int, error)
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

	// Calculate the date for filtering
	startDate := time.Now().AddDate(0, 0, -days)

	var results []struct {
		ReferrerURL string `bun:"referrer_url"`
		Count       int    `bun:"count"`
		UniqueCount int    `bun:"unique_count"`
	}

	query := r.db.NewSelect().
		TableExpr("referrer_stats").
		ColumnExpr("referrer_url").
		ColumnExpr("COUNT(*) AS count")

	// Handle different database dialects for the unique count
	if utils.IsPostgres(r.db) {
		query = query.ColumnExpr("COUNT(DISTINCT visitor_ip) AS unique_count")
	} else if utils.IsSQLite(r.db) {
		query = query.ColumnExpr("COUNT(DISTINCT visitor_ip) AS unique_count")
	} else {
		// MySQL
		query = query.ColumnExpr("COUNT(DISTINCT visitor_ip) AS unique_count")
	}

	err := query.Where("visit_time >= ?", startDate).
		GroupExpr("referrer_url").
		OrderExpr("count DESC").
		Limit(limit).
		Scan(ctx, &results)

	if err != nil {
		return nil, err
	}

	// Get total count for percentage calculation
	total, err := r.GetTotal(ctx, days)
	if err != nil {
		return nil, err
	}

	// Convert to summary objects with percentages
	summaries := make([]*models.ReferrerSummary, len(results))
	for i, r := range results {
		percent := 0.0
		if total > 0 {
			percent = (float64(r.Count) / float64(total)) * 100
		}

		summaries[i] = &models.ReferrerSummary{
			ReferrerURL:  r.ReferrerURL,
			Count:        r.Count,
			UniqueCount:  r.UniqueCount,
			PercentTotal: percent,
		}
	}

	return summaries, nil
}

func (r *referrerRepository) GetReferrersByDate(ctx context.Context, days int) ([]*models.ReferrerTimeStats, error) {
	// Calculate the date for filtering
	startDate := time.Now().AddDate(0, 0, -days)

	var results []*models.ReferrerTimeStats
	var query *bun.SelectQuery

	// The SQL will be different depending on the database type
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
	// Calculate the date for filtering
	startDate := time.Now().AddDate(0, 0, -days)

	var count int
	err := r.db.NewSelect().
		TableExpr("referrer_stats").
		ColumnExpr("COUNT(*) AS count").
		Where("visit_time >= ?", startDate).
		Scan(ctx, &count)

	return count, err
}
