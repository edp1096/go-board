package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// toSnakeCase는 캐멀 케이스 문자열을 스네이크 케이스로 변환합니다
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// parseTimeString은 다양한 포맷의 시간 문자열을 파싱합니다
func parseTimeString(timeStr string) (time.Time, error) {
	// 일반적인 시간 포맷 목록
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
		time.RFC3339Nano,
	}

	var t time.Time
	var err error

	// 모든 포맷 시도
	for _, format := range formats {
		t, err = time.Parse(format, timeStr)
		if err == nil {
			return t, nil
		}
	}

	// 실패 시 마지막 오류 반환
	return time.Time{}, err
}

// tryBase64Decode는 문자열이 Base64로 인코딩되었는지 확인하고 디코딩을 시도합니다
func tryBase64Decode(s string) (string, error) {
	// Base64 디코딩 시도
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return s, err
	}

	// 결과가 유효한 UTF-8 문자열인지 확인
	if !utf8.Valid(bytes) {
		return s, fmt.Errorf("invalid UTF-8 sequence")
	}

	return string(bytes), nil
}

// parseTags는 bun 태그를 파싱하여 맵으로 반환합니다
func parseTags(tag string) map[string]string {
	tagMap := make(map[string]string)

	// 태그 파싱
	parts := strings.SplitSeq(tag, ",")
	for part := range parts {
		// column:name 형식 처리
		if strings.Contains(part, ":") {
			kv := strings.SplitN(part, ":", 2)
			tagMap[kv[0]] = kv[1]
		} else if strings.Contains(part, "=") {
			// column=name 형식 처리
			kv := strings.SplitN(part, "=", 2)
			tagMap[kv[0]] = kv[1]
		} else {
			// 단일 태그 처리 (pk, notnull 등)
			tagMap[part] = "true"
		}
	}

	return tagMap
}

// quoteTableName은 데이터베이스 드라이버에 따라 테이블 이름을 인용 부호로 묶습니다
func quoteTableName(driver, tableName string) string {
	switch driver {
	case "postgres":
		// PostgreSQL에서는 항상 따옴표로 감싸기
		return fmt.Sprintf("\"%s\"", tableName)
	case "mysql", "mariadb":
		// MySQL에서는 항상 백틱으로 감싸기
		return fmt.Sprintf("`%s`", tableName)
	case "sqlite":
		// SQLite에서는 특수 문자가 있으면 항상 따옴표로 감싸기
		if strings.ContainsAny(tableName, "-.") || strings.Contains(tableName, " ") {
			return fmt.Sprintf("\"%s\"", tableName)
		}
		return tableName
	default:
		return tableName
	}
}

// quoteColumnName은 컬럼 이름을 인용 부호로 묶습니다
func quoteColumnName(columnName string) string {
	return columnName
}

// disableForeignKeyConstraints는 대상 데이터베이스의 외래 키 제약 조건을 일시적으로 비활성화합니다
func disableForeignKeyConstraints(config *DataMigrationConfig) error {
	ctx := context.Background()

	switch config.TargetDBConfig.DBDriver {
	case "mysql", "mariadb":
		_, err := config.TargetDB.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0;")
		if err != nil {
			return fmt.Errorf("MySQL 외래 키 제약 비활성화 실패: %w", err)
		}
	case "postgres":
		_, err := config.TargetDB.ExecContext(ctx, "SET session_replication_role = 'replica';")
		if err != nil {
			return fmt.Errorf("PostgreSQL 외래 키 제약 비활성화 실패: %w", err)
		}
	case "sqlite":
		_, err := config.TargetDB.ExecContext(ctx, "PRAGMA foreign_keys = OFF;")
		if err != nil {
			return fmt.Errorf("SQLite 외래 키 제약 비활성화 실패: %w", err)
		}
	}

	return nil
}

// enableForeignKeyConstraints는 대상 데이터베이스의 외래 키 제약 조건을 다시 활성화합니다
func enableForeignKeyConstraints(config *DataMigrationConfig) error {
	ctx := context.Background()

	switch config.TargetDBConfig.DBDriver {
	case "mysql", "mariadb":
		_, err := config.TargetDB.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1;")
		if err != nil {
			return fmt.Errorf("MySQL 외래 키 제약 재활성화 실패: %w", err)
		}
	case "postgres":
		_, err := config.TargetDB.ExecContext(ctx, "SET session_replication_role = 'origin';")
		if err != nil {
			return fmt.Errorf("PostgreSQL 외래 키 제약 재활성화 실패: %w", err)
		}
	case "sqlite":
		_, err := config.TargetDB.ExecContext(ctx, "PRAGMA foreign_keys = ON;")
		if err != nil {
			return fmt.Errorf("SQLite 외래 키 제약 재활성화 실패: %w", err)
		}
	}

	return nil
}
