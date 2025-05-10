package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/edp1096/toy-board/config"
)

// loadConfig는 환경 파일에서 설정을 로드하고 드라이버를 오버라이드합니다
func loadConfig(envFile string, driverOverride string) (*config.Config, error) {
	// 기존 환경 변수 저장
	oldAppEnv := os.Getenv("APP_ENV")
	oldDBDriver := os.Getenv("DB_DRIVER")

	// 환경 변수 설정
	if envFile != "" && envFile != ".env" {
		// 환경 파일 내용 로드
		content, err := os.ReadFile(envFile)
		if err != nil {
			return nil, fmt.Errorf("환경 파일 읽기 실패: %w", err)
		}

		// 환경 변수 파싱 및 설정
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// 주석 처리: # 문자가 있으면 그 앞부분만 사용
			if commentIdx := strings.Index(value, "#"); commentIdx >= 0 {
				value = strings.TrimSpace(value[:commentIdx])
			}
			os.Setenv(key, value)
		}
	}

	// 드라이버 오버라이드
	if driverOverride != "" {
		os.Setenv("DB_DRIVER", driverOverride)
	}

	// 설정 로드
	cfg, err := config.Load()

	// 기존 환경 변수 복원
	os.Setenv("APP_ENV", oldAppEnv)
	os.Setenv("DB_DRIVER", oldDBDriver)

	return cfg, err
}
