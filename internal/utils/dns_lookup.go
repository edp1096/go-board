// internal/utils/dns_lookup.go
package utils

import (
	"context"
	"net"
	"sync"
	"time"
)

var (
	// DNS 조회 결과 캐시
	ptrCache   = make(map[string]cacheDNSEntry)
	hostCache  = make(map[string]cacheIPsEntry)
	cacheMutex sync.RWMutex

	// 캐시 TTL (12시간)
	dnsExpiration = 12 * time.Hour
)

// DNS 조회 결과를 캐싱하기 위한 구조체
type cacheDNSEntry struct {
	hostname string
	expires  time.Time
}

// IP 주소 목록을 캐싱하기 위한 구조체
type cacheIPsEntry struct {
	ips     []string
	expires time.Time
}

// LookupPTR은 IP 주소에 대한 역DNS(PTR) 조회를 수행합니다
// 캐싱을 통해 중복 조회를 방지합니다
func LookupPTR(ipAddr string) (string, error) {
	if ipAddr == "" || ipAddr == "unknown" {
		return "", nil
	}

	// 캐시 확인
	cacheMutex.RLock()
	entry, exists := ptrCache[ipAddr]
	cacheMutex.RUnlock()

	// 캐시에 있고 만료되지 않았으면 반환
	if exists && time.Now().Before(entry.expires) {
		return entry.hostname, nil
	}

	// DNS 조회 수행
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	hostnames, err := net.DefaultResolver.LookupAddr(ctx, ipAddr)

	result := ""
	if err == nil && len(hostnames) > 0 {
		result = hostnames[0]
	}

	// 결과 캐싱 (에러 발생 시에도 빈 문자열 캐싱)
	cacheMutex.Lock()
	ptrCache[ipAddr] = cacheDNSEntry{
		hostname: result,
		expires:  time.Now().Add(dnsExpiration),
	}
	cacheMutex.Unlock()

	return result, err
}

// LookupHost는 호스트명에 대한 정DNS(A/AAAA) 조회를 수행합니다
func LookupHost(hostname string) ([]string, error) {
	if hostname == "" {
		return nil, nil
	}

	// 캐시 확인
	cacheMutex.RLock()
	entry, exists := hostCache[hostname]
	cacheMutex.RUnlock()

	// 캐시에 있고 만료되지 않았으면 반환
	if exists && time.Now().Before(entry.expires) {
		return entry.ips, nil
	}

	// DNS 조회 수행
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ips, err := net.DefaultResolver.LookupHost(ctx, hostname)

	// 결과 캐싱 (에러 발생 시에도 빈 배열 캐싱)
	result := []string{}
	if err == nil {
		result = ips
	}

	cacheMutex.Lock()
	hostCache[hostname] = cacheIPsEntry{
		ips:     result,
		expires: time.Now().Add(dnsExpiration),
	}
	cacheMutex.Unlock()

	return result, err
}

// GetDomainInfo는 도메인에 대한 정보를 가져옵니다 (단일 함수에서 정방향 DNS 조회 수행)
func GetDomainInfo(domain string) []string {
	if domain == "" || domain == "direct" || domain == "unknown" {
		return nil
	}

	ips, _ := LookupHost(domain)
	return ips
}
