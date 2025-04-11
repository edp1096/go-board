package goboard

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed .env.example
var envExampleFS embed.FS

//go:embed web/templates
var templatesFS embed.FS

//go:embed web/static
var staticFS embed.FS

//go:embed migrations/mysql
var MysqlMigrationsFS embed.FS

//go:embed migrations/postgres
var PostgresMigrationsFS embed.FS

//go:embed migrations/sqlite
var SQLiteMigrationsFS embed.FS

// GetEnvExampleContent는 .env.example 파일의 내용을 반환합니다
func GetEnvExampleContent() ([]byte, error) {
	return envExampleFS.ReadFile(".env.example")
}

// GetWebContentDirs는 웹 콘텐츠 디렉토리의 경로와 임베디드 파일 시스템을 반환합니다.
func GetWebContentDirs() map[string]embed.FS {
	return map[string]embed.FS{
		"web/templates": templatesFS,
		"web/static":    staticFS,
	}
}

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

// exportWebContent는 embed 웹 콘텐츠를 지정된 경로로 내보냅니다
func ExportWebContent(destPath string) error {
	fmt.Printf("웹 콘텐츠를 %s 경로로 내보냅니다...\n", destPath)

	// 웹 콘텐츠 디렉토리와 파일 시스템 가져오기
	contentDirs := GetWebContentDirs()

	// 각 콘텐츠 디렉토리 처리
	for dirPath, embedFS := range contentDirs {
		// 대상 디렉토리 경로 계산 (web/templates -> templates, web/static -> static)
		relativePath := strings.TrimPrefix(dirPath, "web/")
		targetPath := filepath.Join(destPath, relativePath)

		fmt.Printf("%s 디렉토리를 %s로 내보냅니다...\n", dirPath, targetPath)

		// 대상 디렉토리 생성
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return fmt.Errorf("디렉토리 생성 실패: %w", err)
		}

		// 디렉토리 내용 내보내기
		err := fs.WalkDir(embedFS, dirPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// 상대 경로 생성 (예: web/templates/layouts/base.html -> layouts/base.html)
			relPath, err := filepath.Rel(dirPath, path)
			if err != nil {
				return fmt.Errorf("상대 경로 생성 실패: %w", err)
			}

			// 루트 디렉토리는 건너뛰기
			if relPath == "." {
				return nil
			}

			destFilePath := filepath.Join(targetPath, relPath)

			if d.IsDir() {
				// 디렉토리 생성
				if err := os.MkdirAll(destFilePath, 0755); err != nil {
					return fmt.Errorf("디렉토리 생성 실패(%s): %w", destFilePath, err)
				}
			} else {
				// 파일 내용 읽기
				data, err := embedFS.ReadFile(path)
				if err != nil {
					return fmt.Errorf("파일 읽기 실패(%s): %w", path, err)
				}

				// 파일 쓰기
				if err := os.WriteFile(destFilePath, data, 0644); err != nil {
					return fmt.Errorf("파일 쓰기 실패(%s): %w", destFilePath, err)
				}
				fmt.Printf("  파일 내보내기: %s/%s\n", relativePath, relPath)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("%s 디렉토리 내보내기 실패: %w", dirPath, err)
		}
	}

	return nil
}

// ExportEnvExample은 .env.example 파일만 지정된 경로로 내보냅니다
func ExportEnvExample(destPath string) error {
	fmt.Printf(".env.example 파일을 %s 경로로 내보냅니다...\n", destPath)

	// .env.example 파일 내용 가져오기
	envData, err := GetEnvExampleContent()
	if err != nil {
		return fmt.Errorf(".env.example 파일 읽기 실패: %w", err)
	}

	// 대상 디렉토리가 없으면 생성
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %w", err)
	}

	// 파일 쓰기
	if err := os.WriteFile(destPath, envData, 0644); err != nil {
		return fmt.Errorf(".env.example 파일 쓰기 실패: %w", err)
	}

	fmt.Printf(".env.example 파일이 %s에 성공적으로 내보내졌습니다\n", destPath)
	return nil
}
