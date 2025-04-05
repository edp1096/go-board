// internal/utils/string_utils.go
package utils

import "strings"

// Split은 문자열을 구분자로 분할합니다.
func Split(s, sep string) []string {
	if s == "" {
		return []string{}
	}

	result := strings.Split(s, sep)

	// 공백 제거
	for i, v := range result {
		result[i] = strings.TrimSpace(v)
	}

	// 빈 문자열 제거
	cleanResult := make([]string, 0, len(result))
	for _, v := range result {
		if v != "" {
			cleanResult = append(cleanResult, v)
		}
	}

	return cleanResult
}
