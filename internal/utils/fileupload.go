package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"maps"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

// 허용된 이미지 MIME 타입 목록
var AllowedImageTypes = map[string]bool{
	"image/jpeg":    true,
	"image/png":     true,
	"image/gif":     true,
	"image/webp":    true,
	"image/svg+xml": true,
}

// 허용된 파일 MIME 타입 목록 (필요에 따라 확장)
var AllowedFileTypes = map[string]bool{
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	"text/plain":                   true,
	"text/plain; charset=utf-8":    true,
	"text/csv":                     true,
	"application/zip":              true,
	"application/x-zip-compressed": true,
}

// 파일 업로드 설정 구조체
type UploadConfig struct {
	BasePath       string          // 기본 저장 경로
	MaxSize        int64           // 최대 파일 크기 (바이트)
	AllowedTypes   map[string]bool // 허용된 MIME 타입
	UniqueFilename bool            // 고유 파일명 생성 여부
}

// 업로드된 파일 정보
type UploadedFile struct {
	OriginalName string
	StorageName  string
	Path         string
	Size         int64
	MimeType     string
	URL          string
	ThumbnailURL string
	IsImage      bool
}

// 파일 업로드 처리 함수
func UploadFile(file *multipart.FileHeader, config UploadConfig, uploadDir string) (*UploadedFile, error) {
	// 파일 크기 확인
	if file.Size > config.MaxSize {
		maxSizeMB := config.MaxSize / (1024 * 1024)
		fileSizeMB := float64(file.Size) / float64(1024*1024)
		return nil, fmt.Errorf("파일 크기가 허용 한도를 초과했습니다: %.2fMB (최대 허용: %dMB)", fileSizeMB, maxSizeMB)
	}

	// 파일 열기
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("파일 열기 실패: %w", err)
	}
	defer src.Close()

	// 파일명 준비
	originalName := filepath.Base(file.Filename)
	ext := filepath.Ext(originalName)

	// MIME 타입 확인
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("파일 읽기 실패: %w", err)
	}

	// 파일 포인터 처음으로 되돌리기
	src.Seek(0, io.SeekStart)

	// MIME 타입 감지
	mimeType := http.DetectContentType(buffer)
	mimeType = strings.SplitN(mimeType, ";", 2)[0]

	// WebP 파일 확장자 명시적 확인 (.webp 확장자이지만 MIME 타입이 제대로 감지되지 않는 경우)
	if ext == ".webp" {
		// fmt.Printf("WebP 파일 감지: %s (원래 감지된 MIME: %s)\n", file.Filename, mimeType)
		mimeType = "image/webp"
	}

	// MIME 타입 확인
	if !config.AllowedTypes[mimeType] {
		return nil, fmt.Errorf("허용되지 않는 파일 형식입니다: %s", mimeType)
	}

	var storageName string
	if config.UniqueFilename {
		// 고유 파일명 생성
		storageName = uuid.New().String() + ext
	} else {
		storageName = originalName
	}

	basePath := config.BasePath
	if !filepath.IsAbs(basePath) && !strings.HasPrefix(basePath, uploadDir) {
		// 상대 경로인 경우 uploadDir을 기준으로 설정
		if strings.HasPrefix(basePath, "uploads/") {
			basePath = filepath.Join(uploadDir, basePath[8:])
		} else if strings.HasPrefix(basePath, "./uploads/") {
			basePath = filepath.Join(uploadDir, basePath[10:])
		}
	}

	// 저장 경로 생성
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("디렉토리 생성 실패: %w", err)
	}

	fullPath := filepath.Join(basePath, storageName)

	// 파일 저장
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("파일 생성 실패: %w", err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("파일 저장 실패: %w", err)
	}

	// 결과 반환
	isImage := AllowedImageTypes[mimeType]

	// URL 경로 생성 - 항상 uploads/를 포함하도록 설정
	relativePath := strings.TrimPrefix(config.BasePath, uploadDir)
	relativePath = strings.TrimPrefix(relativePath, "/")
	urlPath := filepath.ToSlash(filepath.Join("/uploads", relativePath, storageName))

	// 업로드된 파일 객체 생성
	uploadedFile := &UploadedFile{
		OriginalName: originalName,
		StorageName:  storageName,
		Path:         fullPath,
		Size:         file.Size,
		MimeType:     mimeType,
		URL:          urlPath,
		IsImage:      isImage,
	}

	// 이미지인 경우 썸네일 URL 설정
	if isImage {
		// 썸네일 URL 생성
		uploadedFile.ThumbnailURL = GetThumbnailURL(urlPath)
	}

	return uploadedFile, nil
}

// 여러 파일 업로드 처리 함수
func UploadFiles(files []*multipart.FileHeader, config UploadConfig, uploadDir string) ([]*UploadedFile, error) {
	var uploadedFiles []*UploadedFile

	for _, file := range files {
		uploadedFile, err := UploadFile(file, config, uploadDir)
		if err != nil {
			// 오류 발생 시 이미 업로드한 파일들 삭제
			for _, f := range uploadedFiles {
				os.Remove(f.Path)
			}
			return nil, err
		}
		uploadedFiles = append(uploadedFiles, uploadedFile)
	}

	return uploadedFiles, nil
}

// UploadImages 헬퍼 함수
func UploadImages(files []*multipart.FileHeader, basePath string, maxSize int64, uploadDir string) ([]*UploadedFile, error) {
	config := UploadConfig{
		BasePath:       basePath,
		MaxSize:        maxSize,
		AllowedTypes:   AllowedImageTypes,
		UniqueFilename: true,
	}

	uploadedFiles, err := UploadFiles(files, config, uploadDir)
	if err != nil {
		return nil, err
	}

	// 이미지인 경우 썸네일 생성
	for _, file := range uploadedFiles {
		if file.IsImage {
			// 갤러리 및 컨텐츠용 썸네일 생성
			_, err := GenerateThumbnail(file.Path, GalleryThumbnailWidth, GalleryThumbnailHeight)
			if err != nil {
				// 썸네일 생성 실패 시 로그 기록하고 계속 진행
				fmt.Printf("갤러리 썸네일 생성 실패 (%s): %v\n", file.OriginalName, err)
			}

			// 컨텐츠용 썸네일도 생성 (너비만 지정하여 비율 유지)
			_, err = GenerateThumbnail(file.Path, ContentThumbnailWidth, 0)
			if err != nil {
				fmt.Printf("컨텐츠 썸네일 생성 실패 (%s): %v\n", file.OriginalName, err)
			}

			// 썸네일 URL은 이미 UploadFile 함수에서 설정됨
		}
	}

	return uploadedFiles, nil
}

// 일반 파일 업로드 헬퍼 함수
func UploadAttachments(files []*multipart.FileHeader, basePath string, maxSize int64, uploadDir string) ([]*UploadedFile, error) {
	normalizedPath := norm.NFC.String(basePath)
	config := UploadConfig{
		BasePath:       normalizedPath,
		MaxSize:        maxSize,
		AllowedTypes:   AllowedFileTypes,
		UniqueFilename: true,
	}

	uploadedFiles, err := UploadFiles(files, config, uploadDir)
	if err != nil {
		return nil, err
	}

	// 이미지 파일인 경우 썸네일 생성 추가
	for _, file := range uploadedFiles {
		if file.IsImage {
			// 갤러리 및 컨텐츠용 썸네일 생성
			_, err := GenerateThumbnail(file.Path, GalleryThumbnailWidth, GalleryThumbnailHeight)
			if err != nil {
				// 썸네일 생성 실패 시 로그 기록하고 계속 진행
				fmt.Printf("갤러리 썸네일 생성 실패 (%s): %v\n", file.OriginalName, err)
			}

			// 컨텐츠용 썸네일도 생성 (너비만 지정하여 비율 유지)
			_, err = GenerateThumbnail(file.Path, ContentThumbnailWidth, 0)
			if err != nil {
				fmt.Printf("컨텐츠 썸네일 생성 실패 (%s): %v\n", file.OriginalName, err)
			}

			// 썸네일 URL 설정
			file.ThumbnailURL = GetThumbnailURL(file.URL)
		}
	}

	return uploadedFiles, nil
}

// 갤러리용 (대표이미지) 파일 업로드 헬퍼 함수 (이미지와 파일 타입 모두 허용)
func UploadGalleryFiles(files []*multipart.FileHeader, basePath string, maxSize int64, uploadDir string) ([]*UploadedFile, error) {
	normalizedPath := norm.NFC.String(basePath)

	// 이미지와 파일 타입을 합친 허용 타입 맵 생성
	combinedTypes := make(map[string]bool)

	// 이미지 타입 복사
	maps.Copy(combinedTypes, AllowedImageTypes)

	// 파일 타입 복사
	maps.Copy(combinedTypes, AllowedFileTypes)

	config := UploadConfig{
		BasePath:       normalizedPath,
		MaxSize:        maxSize,
		AllowedTypes:   combinedTypes,
		UniqueFilename: true,
	}

	uploadedFiles, err := UploadFiles(files, config, uploadDir)
	if err != nil {
		return nil, err
	}

	// 이미지 파일인 경우 썸네일 생성 추가
	for _, file := range uploadedFiles {
		if file.IsImage {
			// 갤러리 및 컨텐츠용 썸네일 생성
			_, err := GenerateThumbnail(file.Path, GalleryThumbnailWidth, GalleryThumbnailHeight)
			if err != nil {
				// 썸네일 생성 실패 시 로그 기록하고 계속 진행
				fmt.Printf("갤러리 썸네일 생성 실패 (%s): %v\n", file.OriginalName, err)
			}

			// 컨텐츠용 썸네일도 생성 (너비만 지정하여 비율 유지)
			_, err = GenerateThumbnail(file.Path, ContentThumbnailWidth, 0)
			if err != nil {
				fmt.Printf("컨텐츠 썸네일 생성 실패 (%s): %v\n", file.OriginalName, err)
			}

			// 썸네일 URL 설정
			// fmt.Println("썸네일 URL 설정:", file.URL)
			file.ThumbnailURL = GetThumbnailURL(file.URL)
		}
	}

	return uploadedFiles, nil
}

// MoveFile은 파일을 소스에서 타겟으로 이동합니다
func MoveFile(source, target string) error {
	// 먼저 파일 복사 시도
	err := CopyFile(source, target)
	if err != nil {
		return err
	}

	// 성공하면 원본 파일 삭제
	return os.Remove(source)
}

// CopyFile은 파일을 소스에서 타겟으로 복사합니다
func CopyFile(source, target string) error {
	// 소스 파일 열기
	s, err := os.Open(source)
	if err != nil {
		return err
	}
	defer s.Close()

	// 타겟 파일 생성
	t, err := os.Create(target)
	if err != nil {
		return err
	}
	defer t.Close()

	// 파일 복사
	_, err = io.Copy(t, s)
	if err != nil {
		return err
	}

	// 성공적으로 복사되면 파일 권한 설정
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	return os.Chmod(target, info.Mode())
}
