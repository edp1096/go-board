// internal/utils/whois_ip.go
package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	// IP WHOIS 캐시
	ipWhoisCache      = make(map[string]ipWhoisCacheEntry)
	ipWhoisCacheMutex sync.RWMutex
	ipWhoisCacheTTL   = 24 * time.Hour
)

// IP WHOIS 캐시 항목
type ipWhoisCacheEntry struct {
	Data    *IPWhoisInfo
	Expires time.Time
}

// IPWhoisInfo는 IP WHOIS 정보를 담는 구조체
type IPWhoisInfo struct {
	IPAddress    string   `json:"ipAddress"`
	Network      string   `json:"network"`
	Country      string   `json:"country"`
	City         string   `json:"city"`
	Organization string   `json:"organization"`
	ISP          string   `json:"isp"`
	ASN          string   `json:"asn"`
	Status       []string `json:"status,omitempty"`
	Query        string   `json:"query"`
	RawData      string   `json:"rawData"`
}

// GetIPWhois는 IP 주소의 WHOIS 정보를 조회합니다
func GetIPWhois(ipAddress string) (*IPWhoisInfo, error) {
	// 캐시 확인
	ipWhoisCacheMutex.RLock()
	entry, exists := ipWhoisCache[ipAddress]
	ipWhoisCacheMutex.RUnlock()

	// 캐시에 있고 만료되지 않았으면 사용
	if exists && time.Now().Before(entry.Expires) {
		return entry.Data, nil
	}

	// 내부/프라이빗 IP 여부 확인
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return nil, fmt.Errorf("유효하지 않은 IP 주소: %s", ipAddress)
	}

	if isPrivateIP(ip) {
		return &IPWhoisInfo{
			IPAddress:    ipAddress,
			Network:      "Private Network",
			Country:      "Local",
			Organization: "Private IP Range",
			Status:       []string{"Private IP"},
			Query:        ipAddress,
			RawData:      "Private IP address in reserved range",
		}, nil
	}

	// RDAP API로 조회 (ARIN, APNIC)
	info, err := getRDAPInfo(ipAddress)
	if err == nil {
		// 캐싱 및 반환
		ipWhoisCacheMutex.Lock()
		ipWhoisCache[ipAddress] = ipWhoisCacheEntry{
			Data:    info,
			Expires: time.Now().Add(ipWhoisCacheTTL),
		}
		ipWhoisCacheMutex.Unlock()
		return info, nil
	}

	// RDAP 실패 시 IP API로 대체
	info, err = getIPAPIInfo(ipAddress)
	if err == nil {
		// 캐싱 및 반환
		ipWhoisCacheMutex.Lock()
		ipWhoisCache[ipAddress] = ipWhoisCacheEntry{
			Data:    info,
			Expires: time.Now().Add(ipWhoisCacheTTL),
		}
		ipWhoisCacheMutex.Unlock()
		return info, nil
	}

	return nil, fmt.Errorf("IP WHOIS 정보를 가져올 수 없습니다: %w", err)
}

// RDAP API로 IP 정보 조회
func getRDAPInfo(ipAddress string) (*IPWhoisInfo, error) {
	// RDAP API URL (ARIN API를 기본으로 사용)
	url := fmt.Sprintf("https://rdap.arin.net/registry/ip/%s", ipAddress)

	// HTTP 요청 생성 및 타임아웃 설정
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 요청 수행
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("RDAP 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	// 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RDAP 응답 오류 (상태 코드: %d)", resp.StatusCode)
	}

	// 응답 읽기
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("RDAP 응답 읽기 실패: %w", err)
	}

	// 원본 데이터 보존
	rawData := string(body)

	// JSON 파싱
	var rdapResp map[string]interface{}
	if err := json.Unmarshal(body, &rdapResp); err != nil {
		return nil, fmt.Errorf("RDAP 응답 파싱 실패: %w", err)
	}

	// IPWhoisInfo 구조체로 변환
	info := &IPWhoisInfo{
		IPAddress: ipAddress,
		Query:     ipAddress,
		RawData:   rawData,
	}

	// 네트워크 정보 추출
	if startAddr, ok := rdapResp["startAddress"].(string); ok {
		if endAddr, ok := rdapResp["endAddress"].(string); ok {
			info.Network = fmt.Sprintf("%s - %s", startAddr, endAddr)
		}
	} else if handle, ok := rdapResp["handle"].(string); ok {
		info.Network = handle
	}

	// 국가 정보 추출
	if country, ok := rdapResp["country"].(string); ok {
		info.Country = country
	}

	// 조직 정보 추출
	if name, ok := rdapResp["name"].(string); ok {
		info.Organization = name
	} else if entities, ok := rdapResp["entities"].([]interface{}); ok && len(entities) > 0 {
		if entity, ok := entities[0].(map[string]interface{}); ok {
			if vcardArray, ok := entity["vcardArray"].([]interface{}); ok && len(vcardArray) > 1 {
				if vcardProps, ok := vcardArray[1].([]interface{}); ok {
					for _, prop := range vcardProps {
						if propArray, ok := prop.([]interface{}); ok && len(propArray) > 3 {
							if propName, ok := propArray[0].(string); ok && propName == "fn" {
								if orgName, ok := propArray[3].(string); ok {
									info.Organization = orgName
								}
							}
						}
					}
				}
			}
		}
	}

	return info, nil
}

// IP-API.com을 사용한 IP 정보 조회 (RDAP 대체용)
func getIPAPIInfo(ipAddress string) (*IPWhoisInfo, error) {
	// IP-API.com URL
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,country,city,isp,org,as,query", ipAddress)

	// HTTP 요청 생성 및 타임아웃 설정
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 요청 수행
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("IP-API 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	// 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IP-API 응답 오류 (상태 코드: %d)", resp.StatusCode)
	}

	// 응답 읽기
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("IP-API 응답 읽기 실패: %w", err)
	}

	// 원본 데이터 보존
	rawData := string(body)

	// JSON 파싱
	var apiResp map[string]interface{}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("IP-API 응답 파싱 실패: %w", err)
	}

	// 상태 확인
	if status, ok := apiResp["status"].(string); !ok || status != "success" {
		errMsg := "알 수 없는 오류"
		if msg, ok := apiResp["message"].(string); ok {
			errMsg = msg
		}
		return nil, fmt.Errorf("IP-API 오류: %s", errMsg)
	}

	// IPWhoisInfo 구조체로 변환
	info := &IPWhoisInfo{
		IPAddress: ipAddress,
		Query:     ipAddress,
		RawData:   rawData,
	}

	// 정보 추출
	if country, ok := apiResp["country"].(string); ok {
		info.Country = country
	}

	if city, ok := apiResp["city"].(string); ok {
		info.City = city
	}

	if isp, ok := apiResp["isp"].(string); ok {
		info.ISP = isp
	}

	if org, ok := apiResp["org"].(string); ok {
		info.Organization = org
	} else if isp, ok := apiResp["isp"].(string); ok {
		info.Organization = isp
	}

	if asn, ok := apiResp["as"].(string); ok {
		info.ASN = asn
		// ASN에서 네트워크 정보 추출
		parts := strings.Split(asn, " ")
		if len(parts) > 0 {
			info.Network = parts[0]
		}
	}

	return info, nil
}

// isPrivateIP는 IP 주소가 사설(내부) 네트워크에 속하는지 확인합니다
func isPrivateIP(ip net.IP) bool {
	// IPv4 사설 범위 확인
	if ip4 := ip.To4(); ip4 != nil {
		// 로컬호스트: 127.0.0.0/8
		if ip4[0] == 127 {
			return true
		}
		// 클래스 A 사설망: 10.0.0.0/8
		if ip4[0] == 10 {
			return true
		}
		// 클래스 B 사설망: 172.16.0.0/12
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return true
		}
		// 클래스 C 사설망: 192.168.0.0/16
		if ip4[0] == 192 && ip4[1] == 168 {
			return true
		}
		// 링크 로컬 주소: 169.254.0.0/16
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}
	}

	// IPv6 사설 범위 확인
	// 로컬호스트: ::1/128
	if ip.IsLoopback() {
		return true
	}
	// 유니크 로컬 주소: fc00::/7
	if len(ip) == 16 && (ip[0] == 0xfc || ip[0] == 0xfd) {
		return true
	}

	return false
}
