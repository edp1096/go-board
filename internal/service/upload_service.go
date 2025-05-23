// internal/service/upload_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/edp1096/toy-board/config"
	"github.com/edp1096/toy-board/internal/models"
	"github.com/edp1096/toy-board/internal/repository"
	"github.com/edp1096/toy-board/internal/utils"
)

// UploadService는 파일 업로드 관련 기능을 제공합니다
type UploadService interface {
	SaveAttachments(ctx context.Context, boardID, postID, userID int64, files []*utils.UploadedFile) ([]*models.Attachment, error)
	GetAttachmentsByPostID(ctx context.Context, boardID, postID int64) ([]*models.Attachment, error)
	GetAttachmentByID(ctx context.Context, id int64) (*models.Attachment, error)
	DeleteAttachment(ctx context.Context, id int64) error
	DeleteAttachmentsByPostID(ctx context.Context, boardID, postID int64) error
	IncrementDownloadCount(ctx context.Context, id int64) error
	DeletePageImages(ctx context.Context, pageID int64) error
}

type uploadService struct {
	attachmentRepo repository.AttachmentRepository
	config         *config.Config
}

func NewUploadService(attachmentRepo repository.AttachmentRepository, cfg *config.Config) UploadService {
	return &uploadService{
		attachmentRepo: attachmentRepo,
		config:         cfg,
	}
}

// SaveAttachments는 업로드된 파일 정보를 데이터베이스에 저장합니다
func (s *uploadService) SaveAttachments(ctx context.Context, boardID, postID, userID int64, files []*utils.UploadedFile) ([]*models.Attachment, error) {
	var attachments []*models.Attachment

	for _, file := range files {
		// 경로에서 항상 슬래시(/)를 사용하도록 수정
		downloadURL := filepath.ToSlash(file.URL)

		// 썸네일 URL 처리
		thumbnailURL := ""
		if file.IsImage && file.ThumbnailURL != "" {
			thumbnailURL = filepath.ToSlash(file.ThumbnailURL)
		}

		attachment := &models.Attachment{
			BoardID:     boardID,
			PostID:      postID,
			UserID:      userID,
			FileName:    file.OriginalName,
			FilePath:    file.Path,
			StorageName: file.StorageName,
			FileSize:    file.Size,
			MimeType:    file.MimeType,
			IsImage:     file.IsImage,
			// DownloadURL:   file.URL,
			DownloadURL:   downloadURL,
			ThumbnailURL:  thumbnailURL,
			DownloadCount: 0,
		}

		err := s.attachmentRepo.Create(ctx, attachment)
		if err != nil {
			// 오류 발생 시 이미 저장된 첨부 파일 삭제
			for _, a := range attachments {
				s.attachmentRepo.Delete(ctx, a.ID)
			}
			return nil, fmt.Errorf("첨부 파일 정보 저장 실패: %w", err)
		}

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

// GetAttachmentsByPostID는 게시물의 첨부 파일 목록을 조회합니다
func (s *uploadService) GetAttachmentsByPostID(ctx context.Context, boardID, postID int64) ([]*models.Attachment, error) {
	attachments, err := s.attachmentRepo.GetByPostID(ctx, boardID, postID)
	if err != nil {
		return nil, err
	}

	// 각 첨부 파일에 대해 경로 조정
	for _, attachment := range attachments {
		if attachment != nil && !filepath.IsAbs(attachment.FilePath) && !strings.HasPrefix(attachment.FilePath, s.config.UploadDir) {
			// uploads/ 또는 ./uploads/로 시작하는지 확인
			if strings.HasPrefix(attachment.FilePath, "uploads/") || strings.HasPrefix(attachment.FilePath, "uploads\\") {
				attachment.FilePath = filepath.Join(s.config.UploadDir, attachment.FilePath[8:])
			} else if strings.HasPrefix(attachment.FilePath, "./uploads/") || strings.HasPrefix(attachment.FilePath, ".\\uploads\\") {
				attachment.FilePath = filepath.Join(s.config.UploadDir, attachment.FilePath[10:])
			} else {
				attachment.FilePath = filepath.Join(s.config.UploadDir, attachment.FilePath)
			}
		}
	}

	return attachments, nil
}

// GetAttachmentByID는 첨부 파일 정보를 조회합니다
func (s *uploadService) GetAttachmentByID(ctx context.Context, id int64) (*models.Attachment, error) {
	attachment, err := s.attachmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 상대경로로 저장된 FilePath에 config의 UploadDir prefix 추가
	// 절대 경로인 경우는 그대로 유지
	if attachment != nil && !filepath.IsAbs(attachment.FilePath) && !strings.HasPrefix(attachment.FilePath, s.config.UploadDir) {
		// uploads/ 또는 ./uploads/로 시작하는지 확인
		if strings.HasPrefix(attachment.FilePath, "uploads/") || strings.HasPrefix(attachment.FilePath, "uploads\\") {
			attachment.FilePath = filepath.Join(s.config.UploadDir, attachment.FilePath[8:])
		} else if strings.HasPrefix(attachment.FilePath, "./uploads/") || strings.HasPrefix(attachment.FilePath, ".\\uploads\\") {
			attachment.FilePath = filepath.Join(s.config.UploadDir, attachment.FilePath[10:])
		} else {
			attachment.FilePath = filepath.Join(s.config.UploadDir, attachment.FilePath)
		}
	}

	return attachment, nil
}

// DeleteAttachment는 첨부 파일을 삭제합니다
func (s *uploadService) DeleteAttachment(ctx context.Context, id int64) error {
	// 첨부 파일 정보 조회
	attachment, err := s.attachmentRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("첨부 파일을 찾을 수 없습니다")
	}

	// 파일 시스템에서 삭제
	if err := os.Remove(attachment.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("파일 삭제 실패: %w", err)
	}

	// 디렉토리가 비어있으면 삭제
	dir := filepath.Dir(attachment.FilePath)
	if isEmpty, _ := isDirEmpty(dir); isEmpty {
		os.Remove(dir)
	}

	// 데이터베이스에서 삭제
	return s.attachmentRepo.Delete(ctx, id)
}

// DeleteAttachmentsByPostID는 게시물의 모든 첨부 파일을 삭제합니다
func (s *uploadService) DeleteAttachmentsByPostID(ctx context.Context, boardID, postID int64) error {
	// 게시물의 첨부 파일 목록 조회
	attachments, err := s.attachmentRepo.GetByPostID(ctx, boardID, postID)
	if err != nil {
		return err
	}

	var postDir string

	// 각 첨부 파일 삭제
	for _, attachment := range attachments {
		// 게시물 디렉토리 저장 (나중에 디렉토리 정리에 사용)
		if postDir == "" && len(attachment.FilePath) > 0 {
			filePath := filepath.Clean(attachment.FilePath)
			attachmentDir := filepath.Dir(filePath) // ~posts/postID/attachments
			postDir = filepath.Dir(attachmentDir)   // ~posts/postID
		}

		// 1. 원본 파일 시스템에서 삭제
		if err := os.Remove(attachment.FilePath); err != nil && !os.IsNotExist(err) {
			// 오류 로깅만 하고 계속 진행
			fmt.Printf("파일 삭제 실패: %v\n", err)
		}

		// 2. 썸네일 파일 삭제 (이미지인 경우)
		if attachment.IsImage {
			// 썸네일 경로 계산
			dir := filepath.Dir(attachment.FilePath)
			filename := filepath.Base(attachment.FilePath)
			thumbsDir := filepath.Join(dir, "thumbs")

			// 일반 썸네일
			thumbnailPath := filepath.Join(thumbsDir, filename)
			if err := os.Remove(thumbnailPath); err != nil && !os.IsNotExist(err) {
				fmt.Printf("썸네일 삭제 실패: %v\n", err)
			}

			// WebP 파일인 경우 JPG 썸네일도 삭제
			if strings.HasSuffix(strings.ToLower(filename), ".webp") {
				baseFilename := filename[:len(filename)-5] // .webp 제거
				jpgThumbnailPath := filepath.Join(thumbsDir, baseFilename+".jpg")
				if err := os.Remove(jpgThumbnailPath); err != nil && !os.IsNotExist(err) {
					fmt.Printf("JPG 썸네일 삭제 실패: %v\n", err)
				}
			}
		}
	}

	// 디렉토리 정리 시도
	if postDir != "" {
		// 1. 디렉토리 경로 계산
		dirs := []string{}

		// attachments 디렉토리 및 thumbs 하위 디렉토리 추가
		attachmentsDir := filepath.Join(postDir, "attachments")
		thumbsDir := filepath.Join(attachmentsDir, "thumbs")
		dirs = append(dirs, thumbsDir, attachmentsDir)

		// 추가적인 게시판 특성에 따른 디렉토리 정리 (예: 갤러리 게시판의 medias(image, video) 디렉토리)
		imagesDir := filepath.Join(postDir, "medias")
		imagesThumbsDir := filepath.Join(imagesDir, "thumbs")
		dirs = append(dirs, imagesThumbsDir, imagesDir)

		// 마지막으로 게시물 디렉토리 자체 추가
		dirs = append(dirs, postDir)

		// 2. 비어있는 디렉토리 삭제 시도 (하위 디렉토리부터)
		for _, dir := range dirs {
			isEmpty, err := isDirEmpty(dir)
			if err == nil && isEmpty {
				if err := os.Remove(dir); err != nil && !os.IsNotExist(err) {
					fmt.Printf("디렉토리 삭제 실패 (%s): %v\n", dir, err)
				}
			}
		}
	}

	// 데이터베이스에서 모든 첨부 파일 정보 삭제
	return s.attachmentRepo.DeleteByPostID(ctx, boardID, postID)
}

// IncrementDownloadCount는 다운로드 카운트를 증가시킵니다
func (s *uploadService) IncrementDownloadCount(ctx context.Context, id int64) error {
	return s.attachmentRepo.IncrementDownloadCount(ctx, id)
}

// isDirEmpty는 디렉토리가 비어있는지 확인합니다
func isDirEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// DeletePageImages는 페이지 관련 모든 이미지를 삭제합니다
func (s *uploadService) DeletePageImages(ctx context.Context, pageID int64) error {
	// 페이지 이미지 디렉토리 경로 계산
	pageDir := filepath.Join(s.config.UploadDir, "pages", strconv.FormatInt(pageID, 10))

	// 디렉토리 존재 여부 확인
	if _, err := os.Stat(pageDir); os.IsNotExist(err) {
		// 디렉토리가 없으면 삭제할 것도 없음
		return nil
	}

	// 비동기 삭제 처리를 위한 고루틴 시작
	go func() {
		// 10초 지연 후 삭제 시도 (브라우저가 파일 참조를 해제할 시간을 줌)
		time.Sleep(10 * time.Second)

		// 디렉토리 내 파일 트리 순회 (RemoveAll 대신 파일별로 처리)
		err := filepath.Walk(pageDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // 오류 무시하고 계속 진행
			}

			if !info.IsDir() {
				// 파일만 삭제 시도
				err := os.Remove(path)
				if err != nil {
					fmt.Printf("파일 삭제 실패 (%s): %v\n", path, err)
				}
			}
			return nil
		})

		if err != nil {
			fmt.Printf("페이지 이미지 디렉토리 순회 중 오류: %v\n", err)
		}

		// 순회가 끝난 후 다시 한 번 디렉토리 삭제 시도
		// 디렉토리를 아래에서 위로 삭제 시도 (빈 디렉토리부터)
		dirs := []string{}

		filepath.Walk(pageDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() && path != pageDir {
				dirs = append(dirs, path)
			}
			return nil
		})

		// 디렉토리를 깊이 우선으로 정렬 (더 깊은 경로가 먼저 삭제되도록)
		sort.Slice(dirs, func(i, j int) bool {
			return len(dirs[i]) > len(dirs[j])
		})

		// 디렉토리 삭제 시도
		for _, dir := range dirs {
			os.Remove(dir) // 오류 무시
		}

		// 마지막으로 루트 디렉토리 삭제 시도
		os.Remove(pageDir)
	}()

	// 성공으로 반환 (실제 삭제는 비동기로 처리됨)
	return nil
}
