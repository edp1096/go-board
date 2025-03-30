// internal/utils/template_utils.go

package utils

import (
	"fmt"
	"path/filepath"
	"strings"

	"maps"

	"github.com/gofiber/fiber/v2"
)

// 페이지 스크립트 경로를 관리하는 유틸리티 함수

// GetPageScriptPath는 주어진 템플릿 이름에 대응하는 페이지 스크립트 경로를 생성합니다.
func GetPageScriptPath(templateName string) string {
	// 템플릿 이름에서 경로와 확장자 제거
	baseName := filepath.Base(templateName)
	ext := filepath.Ext(baseName)
	scriptName := baseName[:len(baseName)-len(ext)]

	// 템플릿 이름에 대응하는 JS 파일 경로 생성
	// 네이밍 규칙: 템플릿 경로의 '/'는 '-'로 변환 (예: admin/board_create -> admin-board-create)
	scriptPath := fmt.Sprintf("/static/js/pages/%s.js", scriptName)

	return scriptPath
}

// JSPath는 템플릿 경로에 기반한 JS 파일 경로를 생성합니다.
func JSPath(path string) string {
	// 경로를 처리하여 적절한 JS 파일 경로 반환
	// 예: "board/create" -> "/static/js/pages/board-create.js"
	parts := strings.Split(path, "/")
	filename := strings.Join(parts, "-")

	return fmt.Sprintf("/static/js/pages/%s.js", filename)
}

// templateMap은 템플릿별 기본 데이터를 관리합니다.
func TemplateMap(title string, data map[string]any) map[string]any {
	result := make(map[string]any)

	// 기본 데이터 설정
	result["title"] = title

	// 추가 데이터 복사
	maps.Copy(result, data)

	return result
}

// MergeUserData는 템플릿 데이터에 사용자 정보를 병합합니다.
func MergeUserData(c *fiber.Ctx, data fiber.Map) fiber.Map {
	// 사용자 정보 가져오기
	user := c.Locals("user")

	// 사용자 정보가 있으면 데이터에 추가
	if user != nil {
		data["user"] = user
	}

	// CSRF 토큰 추가
	if csrf := c.Locals("csrf"); csrf != nil {
		data["csrf"] = csrf
	}

	return data
}
