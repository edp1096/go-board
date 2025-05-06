package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileHandler 파일 처리 클래스
type FileHandler struct {
	config *Config
}

// NewFileHandler FileHandler 생성자
func NewFileHandler(config *Config) *FileHandler {
	return &FileHandler{
		config: config,
	}
}

// ProcessFile XE 파일을 Go-Board 형식으로 변환하여 복사
func (h *FileHandler) ProcessFile(xeFile XEFile, boardID, postID int64) (string, error) {
	// 가능한 파일 경로 목록 생성 (XE의 여러 가능한 저장 방식 대응)
	possiblePaths := []string{
		// 1. 계산된 업로드 경로 사용
		filepath.Join(h.config.XEUploadPath, xeFile.UploadPath, xeFile.FileName),

		// 2. module_srl만으로 경로 구성 시도
		filepath.Join(h.config.XEUploadPath, "attach", getFileType(xeFile.FileName), fmt.Sprintf("%d", xeFile.ModuleSrl), xeFile.FileName),

		// 3. 단순 경로 시도 (이미지/바이너리 구분)
		filepath.Join(h.config.XEUploadPath, "attach", "images", xeFile.FileName),
		filepath.Join(h.config.XEUploadPath, "attach", "binaries", xeFile.FileName),

		// 4. 파일명만으로 전체 검색
		filepath.Join(h.config.XEUploadPath, xeFile.FileName),
		xeFile.FileName,
	}

	// 실제 파일 찾기
	var xeFilePath string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			xeFilePath = path
			break
		}
	}

	// 파일을 찾지 못한 경우
	if xeFilePath == "" {
		return "", fmt.Errorf("파일을 찾을 수 없습니다: %s (시도 경로: %v)", xeFile.FileName, possiblePaths)
	}

	// 파일 확장자 추출
	ext := filepath.Ext(xeFile.OriginName)

	// UUID 생성하여 고유한 파일명 생성
	uuid, err := generateUUID()
	if err != nil {
		return "", fmt.Errorf("UUID 생성 실패: %w", err)
	}

	// UUID + 원본 확장자로 새 파일명 구성
	fileName := uuid + ext

	// 대상 디렉토리 경로 생성
	targetDir := filepath.Join(
		h.config.TargetUploadPath,
		"boards",
		fmt.Sprintf("%d", boardID),
		"posts",
		fmt.Sprintf("%d", postID),
		"attachments",
	)

	// 대상 디렉토리 생성
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("대상 디렉토리 생성 실패: %w", err)
	}

	// 대상 파일 경로 생성
	targetPath := filepath.Join(targetDir, fileName)

	// 파일 복사
	if err := copyFile(xeFilePath, targetPath); err != nil {
		return "", fmt.Errorf("파일 복사 실패: %w", err)
	}

	// 이미지 파일인 경우 썸네일 생성
	if xeFile.IsImage {
		if err := createThumbnails(targetPath, targetDir); err != nil {
			// 썸네일 생성 실패는 로그만 남기고 계속 진행
			fmt.Printf("썸네일 생성 실패 (%s): %v\n", fileName, err)
		}
	}

	// 상대 경로 반환 (URL 생성용)
	urlPath := fmt.Sprintf("/uploads/boards/%d/posts/%d/attachments/%s", boardID, postID, fileName)

	// 경로 구분자를 URL 형식으로 변환
	urlPath = strings.ReplaceAll(urlPath, string(os.PathSeparator), "/")
	return urlPath, nil
}

// UUID 생성 함수
func generateUUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// RFC 4122 형식으로 UUID 생성
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	return uuid, nil
}

// 파일명 정리 함수
func sanitizeFileName(fileName string) string {
	// 파일 확장자 추출
	ext := filepath.Ext(fileName)
	name := strings.TrimSuffix(fileName, ext)

	// 특수문자 및 공백 처리
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.' || r >= 0x3131 && r <= 0xD79D { // 한글 범위
			return r
		}
		return '_'
	}, name)

	// 이름이 비어있으면 기본값 설정
	if name == "" {
		name = "file"
	}

	return name + ext
}

// 고유한 파일 경로 보장
func ensureUniqueFilePath(path string) string {
	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	name := strings.TrimSuffix(filepath.Base(path), ext)

	counter := 1
	result := path

	for {
		if _, err := os.Stat(result); os.IsNotExist(err) {
			// 파일이 존재하지 않으면 현재 경로 사용
			break
		}
		// 중복된 파일이 있으면 숫자 붙이기
		result = filepath.Join(dir, fmt.Sprintf("%s_%d%s", name, counter, ext))
		counter++
	}

	return result
}

// copyFile 파일 복사 함수
func copyFile(src, dst string) error {
	// 원본 파일 열기
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("원본 파일 열기 실패: %w", err)
	}
	defer srcFile.Close()

	// 파일 정보 확인
	stat, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("파일 정보 조회 실패: %w", err)
	}

	// 디렉토리 경로 추출 및 생성
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("대상 디렉토리 생성 실패: %w", err)
	}

	// 대상 파일 생성
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("대상 파일 생성 실패: %w", err)
	}
	defer dstFile.Close()

	// 파일 권한 설정
	if err := os.Chmod(dst, stat.Mode()); err != nil {
		return fmt.Errorf("파일 권한 설정 실패: %w", err)
	}

	// 파일 내용 복사
	buffer := make([]byte, 1024*1024) // 1MB 버퍼
	for {
		n, err := srcFile.Read(buffer)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("파일 읽기 실패: %w", err)
		}

		if _, err := dstFile.Write(buffer[:n]); err != nil {
			return fmt.Errorf("파일 쓰기 실패: %w", err)
		}
	}

	return nil
}

// createThumbnails 썸네일 생성 함수
func createThumbnails(imagePath, targetDir string) error {
	// 썸네일 디렉토리 생성
	thumbDir := filepath.Join(targetDir, "thumbs")
	if err := os.MkdirAll(thumbDir, 0755); err != nil {
		return fmt.Errorf("썸네일 디렉토리 생성 실패: %w", err)
	}

	// 이미지 파일명 추출
	fileName := filepath.Base(imagePath)

	// 썸네일 경로 생성
	thumbPath := filepath.Join(thumbDir, fileName)

	// 여기서는 실제 썸네일 생성 로직 대신 원본 파일 복사로 대체
	// 실제 구현에서는 이미지 라이브러리를 사용하여 썸네일 생성
	if err := copyFile(imagePath, thumbPath); err != nil {
		return fmt.Errorf("썸네일 파일 복사 실패: %w", err)
	}

	return nil
}
