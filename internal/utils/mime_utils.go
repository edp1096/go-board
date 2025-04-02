// internal/utils/mime_utils.go
package utils

import (
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// MIME 타입 매핑 맵
var MimeTypes = map[string]string{
	// 이미지
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
	".svg":  "image/svg+xml",
	".ico":  "image/x-icon",
	".bmp":  "image/bmp",
	".tiff": "image/tiff",
	".tif":  "image/tiff",

	// 문서
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",

	// 텍스트
	".txt":  "text/plain",
	".csv":  "text/csv",
	".html": "text/html",
	".css":  "text/css",
	".js":   "application/javascript",
	".json": "application/json",
	".xml":  "application/xml",

	// 압축 파일
	".zip": "application/zip",
	".gz":  "application/gzip",
	".tar": "application/x-tar",
	".rar": "application/vnd.rar",
	".7z":  "application/x-7z-compressed",

	// 오디오
	".mp3":  "audio/mpeg",
	".wav":  "audio/wav",
	".ogg":  "audio/ogg",
	".flac": "audio/flac",
	".aac":  "audio/aac",

	// 비디오
	".mp4":  "video/mp4",
	".webm": "video/webm",
	".avi":  "video/x-msvideo",
	".mpeg": "video/mpeg",
	".mov":  "video/quicktime",
}

// GetMimeType은 파일 경로에서 MIME 타입을 추출합니다
func GetMimeType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	if mime, ok := MimeTypes[ext]; ok {
		return mime
	}
	// 기본값은 바이너리 데이터로 설정
	return "application/octet-stream"
}

// SetMimeTypeHeader는 Fiber 컨텍스트에 MIME 타입 헤더를 설정합니다
func SetMimeTypeHeader(c *fiber.Ctx, filePath string) {
	mimeType := GetMimeType(filePath)
	c.Set("Content-Type", mimeType)
}

// IsImage는 파일이 이미지인지 확인합니다
func IsImage(filePath string) bool {
	mime := GetMimeType(filePath)
	return strings.HasPrefix(mime, "image/")
}
