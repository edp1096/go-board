package goboard

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
)

//go:embed web/templates
var templatesFS embed.FS

//go:embed web/static
var staticFS embed.FS

//go:embed migrations/mysql
var MysqlMigrationsFS embed.FS

//go:embed migrations/postgres
var PostgresMigrationsFS embed.FS

// GetTemplatesFS는 템플릿 파일시스템을 반환합니다
func GetTemplatesFS() http.FileSystem {
	// 임베디드된 templates 디렉토리 경로 접근
	subFS, err := fs.Sub(templatesFS, "web/templates")
	if err != nil {
		// log.Printf("템플릿 서브디렉토리 접근 오류: %v", err)
		panic(err)
	}

	// 실제 파일 시스템이 존재하는지 확인하고 우선 사용
	if stat, err := os.Stat("./web/templates"); err == nil && stat.IsDir() {
		// log.Println("실제 템플릿 디렉토리를 사용합니다: ./web/templates")
		return http.Dir("./web/templates")
	}

	// log.Println("임베디드 템플릿을 사용합니다")
	return http.FS(subFS)
}

// GetStaticFS는 정적 파일 파일시스템을 반환합니다
func GetStaticFS() http.FileSystem {
	// 임베디드된 static 디렉토리 경로 접근
	subFS, err := fs.Sub(staticFS, "web/static")
	if err != nil {
		// log.Printf("정적 파일 서브디렉토리 접근 오류: %v", err)
		panic(err)
	}

	// 실제 파일 시스템이 존재하는지 확인하고 우선 사용
	if stat, err := os.Stat("./web/static"); err == nil && stat.IsDir() {
		// log.Println("실제 정적 파일 디렉토리를 사용합니다: ./web/static")
		return http.Dir("./web/static")
	}

	// log.Println("임베디드 정적 파일을 사용합니다")
	return http.FS(subFS)
}

// GetTemplatesDir는 템플릿 디렉토리 경로를 반환합니다
func GetTemplatesDir() string {
	return "."
}
