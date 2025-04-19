// internal/utils/convert_utils.go
package utils

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

// InterfaceToString converts an any to string safely
func InterfaceToString(val any) string {
	if val == nil {
		return ""
	}

	switch t := val.(type) {
	case string:
		return t
	case sql.NullString:
		return t.String
	case []byte:
		return string(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

// InterfaceToInt64 converts an any to int64 safely
func InterfaceToInt64(val any) int64 {
	if val == nil {
		return 0
	}

	switch v := val.(type) {
	case int64:
		return v
	case sql.NullInt64:
		return v.Int64
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case sql.NullFloat64:
		return int64(v.Float64)
	case uint64:
		return int64(v)
	case uint:
		return int64(v)
	case uint32:
		return int64(v)
	case uint16:
		return int64(v)
	case uint8:
		return int64(v)
	case string:
		id, _ := strconv.ParseInt(v, 10, 64)
		return id
	case sql.NullString:
		id, _ := strconv.ParseInt(v.String, 10, 64)
		return id
	case []byte:
		s := string(v)
		id, _ := strconv.ParseInt(s, 10, 64)
		return id
	default:
		// Last resort, try string conversion
		numStr := fmt.Sprintf("%v", v)
		id, err := strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			return 0
		}
		return id
	}
}

// InterfaceToInt converts an any to int safely
func InterfaceToInt(val any) int {
	return int(InterfaceToInt64(val))
}

// InterfaceToBool converts an any to bool safely
func InterfaceToBool(val any) bool {
	if val == nil {
		return false
	}

	switch v := val.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0
	case string:
		return v == "true" || v == "1" || v == "yes" || v == "y" || v == "on"
	default:
		// 마지막 수단으로 문자열 변환 시도
		s := fmt.Sprintf("%v", v)
		return s == "true" || s == "1" || s == "yes" || s == "y" || s == "on"
	}
}

// InterfaceToTime converts an any to time.Time safely
func InterfaceToTime(val any, defaultTime time.Time) time.Time {
	if val == nil {
		return defaultTime
	}

	switch t := val.(type) {
	case time.Time:
		return t
	case string:
		parsed, err := time.Parse("2006-01-02 15:04:05", t)
		if err != nil {
			return defaultTime
		}
		return parsed
	case []byte:
		parsed, err := time.Parse("2006-01-02 15:04:05", string(t))
		if err != nil {
			return defaultTime
		}
		return parsed
	default:
		return defaultTime
	}
}
