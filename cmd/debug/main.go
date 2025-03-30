// cmd/debug/main.go 파일 생성
package main

import (
	"fmt"

	goboard "go-board"
)

func main() {
	// 파일 시스템 가져오기
	templatesFS := goboard.GetTemplatesFS()

	// 임베디드 파일 확인
	fmt.Println("=== 템플릿 파일 확인 ===")

	// http.FileSystem을 통해 특정 경로가 있는지 확인
	paths := []string{
		"board/list.html",
		"board/posts.html",
		"layouts/base.html",
		"partials/header.html",
	}

	for _, path := range paths {
		file, err := templatesFS.Open(path)
		if err != nil {
			fmt.Printf("파일을 찾을 수 없음: %s - 오류: %v\n", path, err)
			continue
		}

		// 파일 정보 확인
		stat, err := file.Stat()
		if err != nil {
			fmt.Printf("파일 정보 읽기 실패: %s - 오류: %v\n", path, err)
			file.Close()
			continue
		}

		fmt.Printf("파일 찾음: %s - 크기: %d bytes\n", path, stat.Size())
		file.Close()
	}
}
