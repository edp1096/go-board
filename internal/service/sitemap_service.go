// internal/service/sitemap_service.go
package service

import (
	"context"
	"fmt"
	"go-board/internal/models"
	"go-board/internal/repository"
	"go-board/internal/utils"
	"time"
)

// 사이트맵 생성 서비스 인터페이스
type SitemapService interface {
	GenerateSitemap(ctx context.Context, baseURL string) (*utils.Sitemap, error)
	GenerateSitemapIndex(ctx context.Context, baseURL string, maxURLsPerFile int) (*utils.SitemapIndex, map[string]*utils.Sitemap, error)
}

// 사이트맵 서비스 구현체
type sitemapService struct {
	boardRepo    repository.BoardRepository
	boardService BoardService
}

// 새 사이트맵 서비스 생성
func NewSitemapService(boardRepo repository.BoardRepository, boardService BoardService) SitemapService {
	return &sitemapService{
		boardRepo:    boardRepo,
		boardService: boardService,
	}
}

// 사이트맵 생성 메서드
func (s *sitemapService) GenerateSitemap(ctx context.Context, baseURL string) (*utils.Sitemap, error) {
	sitemap := utils.NewSitemap()

	// 메인 페이지 추가
	sitemap.AddURL(baseURL, time.Now(), "daily", "1.0")

	// 정적 페이지 추가
	sitemap.AddURL(baseURL+"/boards", time.Now(), "daily", "0.9")

	// 게시판 목록 가져오기
	boards, err := s.boardRepo.List(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("게시판 목록 조회 실패: %w", err)
	}

	// 각 게시판 페이지 추가
	for _, board := range boards {
		boardURL := fmt.Sprintf("%s/boards/%d/posts", baseURL, board.ID)
		sitemap.AddURL(boardURL, board.UpdatedAt, "daily", "0.8")

		// 게시물 목록 가져오기 (최근 100개만)
		posts, _, err := s.boardService.ListPosts(ctx, board.ID, 1, 100, "created_at", "desc")
		if err != nil {
			// 오류가 있더라도 계속 진행
			continue
		}

		// 각 게시물 페이지 추가
		for _, post := range posts {
			postURL := fmt.Sprintf("%s/boards/%d/posts/%d", baseURL, board.ID, post.ID)
			postDate := post.UpdatedAt
			if postDate.IsZero() {
				postDate = post.CreatedAt
			}

			// 게시판 유형에 따라 중요도 조정
			priority := "0.6" // 기본값
			changeFreq := "monthly"

			if board.BoardType == models.BoardTypeGallery {
				changeFreq = "weekly"
				priority = "0.7"
			} else if board.BoardType == models.BoardTypeQnA {
				changeFreq = "weekly"
				priority = "0.7"
			}

			sitemap.AddURL(postURL, postDate, changeFreq, priority)
		}
	}

	return sitemap, nil
}

// 사이트맵 인덱스 생성 메서드
func (s *sitemapService) GenerateSitemapIndex(ctx context.Context, baseURL string, maxURLsPerFile int) (*utils.SitemapIndex, map[string]*utils.Sitemap, error) {
	// 전체 사이트맵 생성
	sitemap, err := s.GenerateSitemap(ctx, baseURL)
	if err != nil {
		return nil, nil, err
	}

	// 파일 분할
	return sitemap.Split(baseURL, maxURLsPerFile)
}
