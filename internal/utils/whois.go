// internal/utils/whois.go
package utils

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	// WHOIS 서버 목록
	whoisServers = map[string]string{
		"com":  "whois.verisign-grs.com",
		"net":  "whois.verisign-grs.com",
		"org":  "whois.pir.org",
		"info": "whois.afilias.net",
		"io":   "whois.nic.io",
		// 더 많은 TLD 추가 가능
	}

	// 캐시 TTL (24시간)
	whoisCacheTTL = 24 * time.Hour

	// WHOIS 결과 캐시
	whoisCache      = make(map[string]whoisCacheEntry)
	whoisCacheMutex sync.RWMutex
)

// WHOIS 캐시 항목
type whoisCacheEntry struct {
	Data    string
	Expires time.Time
}

// WhoisInfo는 WHOIS 조회 결과를 담는 구조체
type WhoisInfo struct {
	Domain       string   `json:"domain"`
	Registrar    string   `json:"registrar"`
	RegistrarURL string   `json:"registrarUrl"`
	CreatedDate  string   `json:"createdDate"`
	UpdatedDate  string   `json:"updatedDate"`
	ExpiryDate   string   `json:"expiryDate"`
	NameServers  []string `json:"nameServers"`
	Status       []string `json:"status"`
	RawData      string   `json:"rawData"`
}

// GetWhoisInfo는 도메인의 WHOIS 정보를 조회합니다
func GetWhoisInfo(domain string) (*WhoisInfo, error) {
	// 캐시 확인
	whoisCacheMutex.RLock()
	entry, exists := whoisCache[domain]
	whoisCacheMutex.RUnlock()

	// 캐시에 있고 만료되지 않았으면 사용
	if exists && time.Now().Before(entry.Expires) {
		return parseWhoisData(domain, entry.Data)
	}

	// 도메인에서 TLD 추출
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("잘못된 도메인 형식: %s", domain)
	}

	tld := parts[len(parts)-1]

	// WHOIS 서버 찾기
	server, ok := whoisServers[tld]
	if !ok {
		// IANA WHOIS 서버 사용
		server = "whois.iana.org"
	}

	// WHOIS 서버에 연결
	conn, err := net.DialTimeout("tcp", server+":43", 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("WHOIS 서버 연결 실패: %w", err)
	}
	defer conn.Close()

	// 쿼리 전송
	_, err = conn.Write([]byte(domain + "\r\n"))
	if err != nil {
		return nil, fmt.Errorf("WHOIS 쿼리 전송 실패: %w", err)
	}

	// 타임아웃 설정
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// 응답 읽기
	rawData, err := io.ReadAll(conn)
	if err != nil {
		return nil, fmt.Errorf("WHOIS 응답 읽기 실패: %w", err)
	}

	rawWhois := string(rawData)

	// 결과 캐싱
	whoisCacheMutex.Lock()
	whoisCache[domain] = whoisCacheEntry{
		Data:    rawWhois,
		Expires: time.Now().Add(whoisCacheTTL),
	}
	whoisCacheMutex.Unlock()

	// 파싱 및 결과 반환
	return parseWhoisData(domain, rawWhois)
}

// parseWhoisData는 원시 WHOIS 데이터를 파싱합니다
func parseWhoisData(domain, rawData string) (*WhoisInfo, error) {
	info := &WhoisInfo{
		Domain:      domain,
		RawData:     rawData,
		NameServers: []string{},
		Status:      []string{},
	}

	// 각 줄을 처리
	lines := strings.Split(rawData, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 콜론으로 구분된 키-값 쌍
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 주요 정보 추출
		switch {
		case containsAnyCase(key, "registrar"):
			info.Registrar = value
		case containsAnyCase(key, "registrar url"):
			info.RegistrarURL = value
		case containsAnyCase(key, "creation date", "created on", "registration date"):
			info.CreatedDate = value
		case containsAnyCase(key, "updated date", "last updated"):
			info.UpdatedDate = value
		case containsAnyCase(key, "expiry date", "expiration date"):
			info.ExpiryDate = value
		case containsAnyCase(key, "name server", "nameserver"):
			if value != "" {
				info.NameServers = append(info.NameServers, value)
			}
		case containsAnyCase(key, "status"):
			if value != "" {
				info.Status = append(info.Status, value)
			}
		}
	}

	return info, nil
}

// containsAnyCase는 문자열이 대소문자 구분 없이 하나 이상의 부분 문자열을 포함하는지 확인합니다
func containsAnyCase(s string, substrs ...string) bool {
	lowerS := strings.ToLower(s)
	for _, sub := range substrs {
		if strings.Contains(lowerS, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}
