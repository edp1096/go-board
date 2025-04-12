// internal/utils/sitemap.go
package utils

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// 사이트맵 XML 구조체
type SitemapURL struct {
	XMLName    xml.Name `xml:"url"`
	Loc        string   `xml:"loc"`
	LastMod    string   `xml:"lastmod,omitempty"`
	ChangeFreq string   `xml:"changefreq,omitempty"`
	Priority   string   `xml:"priority,omitempty"`
}

type Sitemap struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

type SitemapIndex struct {
	XMLName  xml.Name           `xml:"sitemapindex"`
	XMLNS    string             `xml:"xmlns,attr"`
	Sitemaps []SitemapIndexItem `xml:"sitemap"`
}

type SitemapIndexItem struct {
	XMLName xml.Name `xml:"sitemap"`
	Loc     string   `xml:"loc"`
	LastMod string   `xml:"lastmod,omitempty"`
}

// 새 사이트맵 생성
func NewSitemap() *Sitemap {
	return &Sitemap{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  make([]SitemapURL, 0),
	}
}

// 새 사이트맵 인덱스 생성
func NewSitemapIndex() *SitemapIndex {
	return &SitemapIndex{
		XMLNS:    "http://www.sitemaps.org/schemas/sitemap/0.9",
		Sitemaps: make([]SitemapIndexItem, 0),
	}
}

// URL 추가 메서드
func (s *Sitemap) AddURL(loc string, lastMod time.Time, changeFreq string, priority string) {
	url := SitemapURL{
		Loc:        loc,
		ChangeFreq: changeFreq,
		Priority:   priority,
	}

	if !lastMod.IsZero() {
		url.LastMod = lastMod.Format("2006-01-02")
	}

	s.URLs = append(s.URLs, url)
}

// 사이트맵 XML 문자열 생성
func (s *Sitemap) ToXML() (string, error) {
	output, err := xml.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}

	// 불필요한 <script/> 태그 제거
	xmlStr := xml.Header + string(output)
	xmlStr = strings.Replace(xmlStr, "<script></script>", "", -1)
	xmlStr = strings.Replace(xmlStr, "<script/>", "", -1)

	return xmlStr, nil
}

// 사이트맵 파일로 분할
func (s *Sitemap) Split(baseURL string, maxURLsPerFile int) (*SitemapIndex, map[string]*Sitemap, error) {
	if maxURLsPerFile <= 0 {
		maxURLsPerFile = 1000 // 기본값
	}

	// 결과 변수 초기화
	index := NewSitemapIndex()
	sitemaps := make(map[string]*Sitemap)

	// URL이 없는 경우 처리
	if len(s.URLs) == 0 {
		return index, sitemaps, nil
	}

	// 파일 분할 로직
	totalFiles := (len(s.URLs) + maxURLsPerFile - 1) / maxURLsPerFile
	for i := 0; i < totalFiles; i++ {
		// 파일명 생성
		fileName := fmt.Sprintf("sitemap_%d.xml", i+1)

		// 새 사이트맵 객체 생성
		subSitemap := NewSitemap()

		// 시작 및 종료 인덱스 계산
		start := i * maxURLsPerFile
		end := (i + 1) * maxURLsPerFile
		if end > len(s.URLs) {
			end = len(s.URLs)
		}

		// URL 복사
		subSitemap.URLs = append(subSitemap.URLs, s.URLs[start:end]...)

		// 사이트맵 저장
		sitemaps[fileName] = subSitemap

		// 인덱스에 추가
		index.Sitemaps = append(index.Sitemaps, SitemapIndexItem{
			Loc:     strings.TrimRight(baseURL, "/") + "/" + fileName,
			LastMod: time.Now().Format("2006-01-02"),
		})
	}

	return index, sitemaps, nil
}

// 사이트맵 인덱스 XML 문자열 생성
func (si *SitemapIndex) ToXML() (string, error) {
	output, err := xml.MarshalIndent(si, "", "  ")
	if err != nil {
		return "", err
	}

	// 불필요한 태그가 포함되지 않도록 문자열 처리
	xmlStr := xml.Header + string(output)
	xmlStr = strings.Replace(xmlStr, "<script></script>", "", -1)
	xmlStr = strings.Replace(xmlStr, "<script/>", "", -1)

	return xmlStr, nil
}
