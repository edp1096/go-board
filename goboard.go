package goboard

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

//go:embed web/templates
var templatesFS embed.FS

//go:embed web/static
var staticFS embed.FS

//go:embed migrations/mysql
var MysqlMigrationsFS embed.FS

//go:embed migrations/postgres
var PostgresMigrationsFS embed.FS

// HybridFileSystem은 실제 파일시스템과 임베디드 파일시스템을 결합합니다
type HybridFileSystem struct {
	embeddedFS http.FileSystem
	realPath   string
	useReal    bool
}

// Open 메서드는 먼저 실제 파일을 확인하고, 없으면 임베디드 파일을 엽니다
func (h *HybridFileSystem) Open(name string) (http.File, error) {
	if h.useReal {
		f, err := os.Open(filepath.Join(h.realPath, name))
		if err == nil {
			return f, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
		// 파일이 없는 경우 임베디드 파일로 폴백
	}
	return h.embeddedFS.Open(name)
}

// GetTemplatesFS는 하이브리드 템플릿 파일시스템을 반환합니다
func GetTemplatesFS() http.FileSystem {
	subFS, err := fs.Sub(templatesFS, "web/templates")
	if err != nil {
		panic(err)
	}

	// 실제 템플릿 디렉토리 검사
	realPath := "./web/templates"
	useReal := false

	if _, err := os.Stat(realPath); err == nil {
		useReal = true
		log.Printf("실제 템플릿 디렉토리를 사용합니다: %s", realPath)
	} else {
		log.Printf("임베디드 템플릿을 사용합니다 (실제 디렉토리 없음: %s)", realPath)
	}

	return &HybridFileSystem{
		embeddedFS: http.FS(subFS),
		realPath:   realPath,
		useReal:    useReal,
	}
}

// GetStaticFS는 하이브리드 정적 파일 파일시스템을 반환합니다
func GetStaticFS() http.FileSystem {
	subFS, err := fs.Sub(staticFS, "web/static")
	if err != nil {
		panic(err)
	}

	// 실제 정적 파일 디렉토리 검사
	realPath := "./web/static"
	useReal := false

	if _, err := os.Stat(realPath); err == nil {
		useReal = true
		log.Printf("실제 정적 파일 디렉토리를 사용합니다: %s", realPath)
	} else {
		log.Printf("임베디드 정적 파일을 사용합니다 (실제 디렉토리 없음: %s)", realPath)
	}

	return &HybridFileSystem{
		embeddedFS: http.FS(subFS),
		realPath:   realPath,
		useReal:    useReal,
	}
}

// GetTemplatesDir는 템플릿 디렉토리 경로를 반환합니다
func GetTemplatesDir() string {
	return "." // 이미 서브디렉토리로 접근하므로 현재 디렉토리를 반환
}
