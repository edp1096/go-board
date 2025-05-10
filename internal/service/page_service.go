// internal/service/page_service.go
package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/edp1096/toy-board/internal/models"
	"github.com/edp1096/toy-board/internal/repository"

	"github.com/gosimple/slug"
)

var (
	ErrPageNotFound = errors.New("페이지를 찾을 수 없음")
	ErrInvalidPage  = errors.New("유효하지 않은 페이지")
)

type PageService interface {
	CreatePage(ctx context.Context, page *models.Page) error
	GetPageByID(ctx context.Context, id int64) (*models.Page, error)
	GetPageBySlug(ctx context.Context, slug string) (*models.Page, error)
	UpdatePage(ctx context.Context, page *models.Page) error
	DeletePage(ctx context.Context, id int64) error
	ListPages(ctx context.Context, onlyActive bool) ([]*models.Page, error)
}

type pageService struct {
	pageRepo      repository.PageRepository
	uploadService UploadService
}

func NewPageService(pageRepo repository.PageRepository, uploadService UploadService) PageService {
	return &pageService{
		pageRepo:      pageRepo,
		uploadService: uploadService,
	}
}

func (s *pageService) CreatePage(ctx context.Context, page *models.Page) error {
	// 슬러그가 없으면 생성
	if page.Slug == "" {
		page.Slug = slug.Make(page.Title)
	}

	// 생성 시간 설정
	now := time.Now()
	page.CreatedAt = now
	page.UpdatedAt = now

	return s.pageRepo.Create(ctx, page)
}

func (s *pageService) GetPageByID(ctx context.Context, id int64) (*models.Page, error) {
	return s.pageRepo.GetByID(ctx, id)
}

func (s *pageService) GetPageBySlug(ctx context.Context, slug string) (*models.Page, error) {
	return s.pageRepo.GetBySlug(ctx, slug)
}

func (s *pageService) UpdatePage(ctx context.Context, page *models.Page) error {
	// 업데이트 시간 설정
	page.UpdatedAt = time.Now()
	return s.pageRepo.Update(ctx, page)
}

func (s *pageService) DeletePage(ctx context.Context, id int64) error {
	// 먼저 페이지 이미지 삭제
	if s.uploadService != nil {
		if err := s.uploadService.DeletePageImages(ctx, id); err != nil {
			// 이미지 삭제 실패는 로깅만 하고 계속 진행
			log.Printf("페이지 이미지 삭제 실패 (ID: %d): %v", id, err)
		}
	}

	// 페이지 삭제
	return s.pageRepo.Delete(ctx, id)
}

func (s *pageService) ListPages(ctx context.Context, onlyActive bool) ([]*models.Page, error) {
	return s.pageRepo.List(ctx, onlyActive)
}
