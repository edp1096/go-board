// internal/utils/convert_utils.go
package utils

import (
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
	case int:
		return int64(v)
	case float64:
		return int64(v)
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
