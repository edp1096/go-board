// internal/utils/fileupload.go
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	IsImage      bool
}

// 랜덤 문자열 생성 (고유 파일명 용)
func generateRandomString(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// 파일 업로드 처리 함수
func UploadFile(file *multipart.FileHeader, config UploadConfig) (*UploadedFile, error) {
	// 파일 크기 확인
	if file.Size > config.MaxSize {
		return nil, errors.New("파일 크기가 허용 한도를 초과했습니다")
	}

	// 파일 열기
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("파일 열기 실패: %w", err)
	}
	defer src.Close()

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

	// MIME 타입 확인
	if !config.AllowedTypes[mimeType] {
		return nil, fmt.Errorf("허용되지 않는 파일 형식입니다: %s", mimeType)
	}

	// 파일명 준비
	originalName := filepath.Base(file.Filename)
	ext := filepath.Ext(originalName)
	// nameWithoutExt := strings.TrimSuffix(originalName, ext)

	var storageName string
	if config.UniqueFilename {
		// // 고유 파일명 생성
		// storageName = fmt.Sprintf("%s_%s%s",
		// 	generateRandomString(8),
		// 	strings.ReplaceAll(nameWithoutExt, " ", "_"),
		// 	ext)
		// storageName = base64.URLEncoding.EncodeToString([]byte(storageName))
		storageName = fmt.Sprintf("%s%s", uuid.New().String(), ext)
	} else {
		storageName = originalName
	}

	// 저장 경로 생성
	if err := os.MkdirAll(config.BasePath, 0755); err != nil {
		return nil, fmt.Errorf("디렉토리 생성 실패: %w", err)
	}

	fullPath := filepath.Join(config.BasePath, storageName)

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

	return &UploadedFile{
		OriginalName: originalName,
		StorageName:  storageName,
		Path:         fullPath,
		Size:         file.Size,
		MimeType:     mimeType,
		URL:          filepath.Join("/uploads", storageName),
		IsImage:      isImage,
	}, nil
}

// 여러 파일 업로드 처리 함수
func UploadFiles(files []*multipart.FileHeader, config UploadConfig) ([]*UploadedFile, error) {
	var uploadedFiles []*UploadedFile

	for _, file := range files {
		uploadedFile, err := UploadFile(file, config)
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

// 이미지 업로드 헬퍼 함수
func UploadImages(files []*multipart.FileHeader, basePath string, maxSize int64) ([]*UploadedFile, error) {
	config := UploadConfig{
		BasePath:       basePath,
		MaxSize:        maxSize,
		AllowedTypes:   AllowedImageTypes,
		UniqueFilename: true,
	}

	return UploadFiles(files, config)
}

// 일반 파일 업로드 헬퍼 함수
func UploadAttachments(files []*multipart.FileHeader, basePath string, maxSize int64) ([]*UploadedFile, error) {
	normalizedPath := norm.NFC.String(basePath)
	config := UploadConfig{
		BasePath:       normalizedPath,
		MaxSize:        maxSize,
		AllowedTypes:   AllowedFileTypes,
		UniqueFilename: true,
	}

	return UploadFiles(files, config)
}
