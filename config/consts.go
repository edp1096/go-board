package config

// 파일 업로드 관련 상수
const (
	// 기본 파일 업로드 제한
	DefaultUploadSizeKB      = 10240 // 일반 파일 업로드 크기 제한 (KiB)
	DefaultMediaUploadSizeKB = 20480 // 이미지 업로드 크기 제한 (KiB)
	DefaultBodyLimitKB       = 5000  // HTTP 요청 본문 크기 제한 (KiB)

	// 바이트 단위 변환 상수
	BytesPerKB = 1024
	BytesPerMB = 1024 * 1024

	// 기본 업로드 디렉토리
	DefaultUploadDir = "./uploads"

	// 바이트 단위 파일 업로드 제한
	DefaultUploadSize      = DefaultUploadSizeKB * BytesPerKB      // 일반 파일 업로드 (바이트)
	DefaultMediaUploadSize = DefaultMediaUploadSizeKB * BytesPerKB // 이미지 업로드 (바이트)
	DefaultBodyLimit       = DefaultBodyLimitKB * BytesPerKB       // HTTP 요청 본문 (바이트)
)
