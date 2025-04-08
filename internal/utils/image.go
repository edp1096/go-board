// internal/utils/image.go
package utils

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"golang.org/x/image/webp"
)

// 썸네일 크기 상수
const (
	GalleryThumbnailWidth  = 300 // 갤러리 게시판용 썸네일 너비
	GalleryThumbnailHeight = 200 // 갤러리 게시판용 썸네일 높이 (비율 유지로 0 사용 가능)
	ContentThumbnailWidth  = 800 // 본문 이미지용 썸네일 너비
	ContentThumbnailHeight = 0   // 본문 이미지용 썸네일 높이 (0은 비율 유지)
)

// GenerateThumbnail 주어진 이미지 파일에 대한 썸네일을 생성합니다
// WebP 이미지의 경우 JPEG 형식으로 변환하여 썸네일을 생성합니다
func GenerateThumbnail(imagePath string, maxWidth, maxHeight int) (string, error) {
	// 파일 확장자 확인
	ext := strings.ToLower(filepath.Ext(imagePath))

	var src image.Image
	var err error

	if ext == ".webp" {
		// WebP 이미지는 별도 처리
		src, err = loadWebPImage(imagePath)
		if err != nil {
			return "", fmt.Errorf("WebP 이미지 로드 실패: %w", err)
		}
	} else {
		// 다른 이미지 포맷은 imaging 라이브러리 사용
		src, err = imaging.Open(imagePath)
		if err != nil {
			return "", fmt.Errorf("이미지 열기 실패: %w", err)
		}
	}

	// 이미지 리사이징 (비율 유지)
	var thumbnail *image.NRGBA
	if maxHeight == 0 {
		// 너비만 지정된 경우 (비율 유지)
		thumbnail = imaging.Resize(src, maxWidth, 0, imaging.Lanczos)
	} else if maxWidth == 0 {
		// 높이만 지정된 경우 (비율 유지)
		thumbnail = imaging.Resize(src, 0, maxHeight, imaging.Lanczos)
	} else {
		// 너비와 높이 모두 지정된 경우 (비율 유지하며 맞춤)
		thumbnail = imaging.Fit(src, maxWidth, maxHeight, imaging.Lanczos)
	}

	// 썸네일 저장 경로 생성
	dir, filename := filepath.Split(imagePath)
	thumbsDir := filepath.Join(dir, "thumbs")

	// thumbs 디렉토리 생성
	if err := os.MkdirAll(thumbsDir, 0755); err != nil {
		return "", fmt.Errorf("썸네일 디렉토리 생성 실패: %w", err)
	}

	// WebP의 경우 저장 형식을 JPEG으로 변경 (파일명도 변경)
	if ext == ".webp" {
		baseFilename := filename[:len(filename)-len(ext)]
		filename = baseFilename + ".jpg"
	}

	thumbnailPath := filepath.Join(thumbsDir, filename)

	// 썸네일 저장
	if ext == ".webp" {
		// WebP -> JPEG으로 변환하여 저장
		out, err := os.Create(thumbnailPath)
		if err != nil {
			return "", fmt.Errorf("썸네일 파일 생성 실패: %w", err)
		}
		defer out.Close()

		err = jpeg.Encode(out, thumbnail, &jpeg.Options{Quality: 90})
		if err != nil {
			return "", fmt.Errorf("JPEG 썸네일 저장 실패: %w", err)
		}
	} else {
		// 그 외 포맷은 기존 방식으로 저장
		err = imaging.Save(thumbnail, thumbnailPath)
		if err != nil {
			return "", fmt.Errorf("썸네일 저장 실패: %w", err)
		}
	}

	return thumbnailPath, nil
}

// loadWebPImage는 WebP 이미지를 로드합니다
func loadWebPImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("파일 열기 실패: %w", err)
	}
	defer file.Close()

	img, err := webp.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("WebP 디코딩 실패: %w", err)
	}

	return img, nil
}

// GetThumbnailPath 원본 이미지 경로에 대한 썸네일 경로를 반환합니다
func GetThumbnailPath(imagePath string) string {
	dir, filename := filepath.Split(imagePath)
	return filepath.Join(dir, "thumbs", filename)
}

// GetThumbnailURL 원본 이미지 URL에 대한 썸네일 URL을 반환합니다
func GetThumbnailURL(imageURL string) string {
	if imageURL == "" {
		return ""
	}

	// URL 경로에서 thumbs 디렉토리 추가
	urlParts := strings.Split(imageURL, "/")
	if len(urlParts) < 2 {
		return imageURL
	}

	// 파일명과 확장자 처리
	filename := urlParts[len(urlParts)-1]

	// WebP 파일인 경우 JPG 확장자로 변경
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".webp" {
		baseFilename := filename[:len(filename)-len(ext)]
		filename = baseFilename + ".jpg"
	}

	// 경로 재구성 (마지막 요소 앞에 thumbs 추가)
	newParts := append(urlParts[:len(urlParts)-1], "thumbs", filename)
	return strings.Join(newParts, "/")
}

// IsImageFile 파일 경로가 이미지 파일인지 확인합니다
func IsImageFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	default:
		return false
	}
}

// ProcessContentImages HTML 내용에서 이미지 태그를 찾아 썸네일 처리를 합니다
func ProcessContentImages(htmlContent string) string {
	// // 간단한 구현 - 실제로는 HTML 파서 사용 권장
	// // 이미지 태그 찾기 (<img src="...">)
	// imgTagRegex := `<img[^>]+src="([^"]+)"[^>]*>`

	// // 이 함수는 서버 사이드에서 HTML 내용을 분석하고
	// // 원본 이미지 URL을 썸네일 URL로 바꾸는 로직을 구현해야 합니다.
	// // 여기서는 클라이언트 측에서 JavaScript로 처리하도록 하겠습니다.

	return htmlContent
}

// EnsureThumbnail 이미지 경로에 썸네일이 없으면 생성합니다
func EnsureThumbnail(imagePath string) error {
	// 원본 파일이 존재하는지 확인
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return fmt.Errorf("원본 이미지가 존재하지 않습니다: %s", imagePath)
	}

	// 이미지 파일인지 확인
	if !IsImageFile(imagePath) {
		return fmt.Errorf("이미지 파일이 아닙니다: %s", imagePath)
	}

	// 썸네일 경로 생성
	thumbPath := GetThumbnailPath(imagePath)

	// 썸네일이 이미 존재하는지 확인
	if _, err := os.Stat(thumbPath); err == nil {
		// 이미 존재하면 아무것도 하지 않음
		return nil
	}

	// 없으면 썸네일 생성
	// 갤러리용 썸네일 생성
	_, err := GenerateThumbnail(imagePath, GalleryThumbnailWidth, GalleryThumbnailHeight)
	if err != nil {
		return fmt.Errorf("갤러리 썸네일 생성 실패: %w", err)
	}

	// 컨텐츠용 썸네일 생성
	_, err = GenerateThumbnail(imagePath, ContentThumbnailWidth, ContentThumbnailHeight)
	if err != nil {
		return fmt.Errorf("컨텐츠 썸네일 생성 실패: %w", err)
	}

	return nil
}
