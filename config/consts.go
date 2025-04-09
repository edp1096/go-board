package config

// 파일 업로드 관련 상수
const (
	// 기본 파일 업로드 제한
	DefaultUploadSizeMB      = 10 // 일반 파일 업로드 크기 제한 (MB)
	DefaultImageUploadSizeMB = 20 // 이미지 업로드 크기 제한 (MB)
	DefaultBodyLimitMB       = 20 // HTTP 요청 본문 크기 제한 (MB)

	// 바이트 단위 변환 상수
	BytesPerMB = 1024 * 1024

	// 바이트 단위 파일 업로드 제한
	DefaultUploadSize      = DefaultUploadSizeMB * BytesPerMB      // 일반 파일 업로드 (바이트)
	DefaultImageUploadSize = DefaultImageUploadSizeMB * BytesPerMB // 이미지 업로드 (바이트)
	DefaultBodyLimit       = DefaultBodyLimitMB * BytesPerMB       // HTTP 요청 본문 (바이트)
)
