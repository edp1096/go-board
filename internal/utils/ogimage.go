// internal/utils/ogimage.go
package utils

import (
	"fmt"
	"image/color"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	toyboard "github.com/edp1096/toy-board"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
)

var (
	// 오픈그래프 이미지 기본 크기 (픽셀)
	OGImageWidth  = 1200
	OGImageHeight = 630

	// 배경색 후보 (헥스 코드)
	backgroundColors = []string{
		"#1d4ed8", // 파란색
		"#047857", // 초록색
		"#7e22ce", // 보라색
		"#be123c", // 빨간색
		"#ca8a04", // 노란색
		"#0f766e", // 청록색
		"#334155", // 회색
	}
)

// OGImageOptions는 OG 이미지 생성 옵션을 정의합니다.
type OGImageOptions struct {
	Title           string    // 제목
	Author          string    // 작성자
	Date            time.Time // 작성일
	BackgroundColor string    // 배경색 (헥스 코드, 비어있으면 랜덤 선택)
	FontPath        string    // 폰트 경로 (비어있으면 기본 폰트 사용)
	OutputPath      string    // 출력 경로
	SiteName        string    // 사이트 이름
}

// GetRandomBackgroundColor는 배경색 목록에서 랜덤하게 하나를 선택합니다.
func GetRandomBackgroundColor() string {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	return backgroundColors[r.Intn(len(backgroundColors))]
}

// CreateOGImage는 오픈그래프 이미지를 생성합니다.
func CreateOGImage(opts OGImageOptions) (string, error) {
	// 배경색이 지정되지 않은 경우 랜덤 선택
	if opts.BackgroundColor == "" {
		opts.BackgroundColor = GetRandomBackgroundColor()
	}

	// 폰트 로드
	fontData, err := toyboard.GetFontContent("NotoSansKR-Bold.ttf")
	if err != nil {
		return "", fmt.Errorf("폰트 파일을 읽을 수 없습니다: %w", err)
	}

	font, err := truetype.Parse(fontData)
	if err != nil {
		return "", fmt.Errorf("폰트를 파싱할 수 없습니다: %w", err)
	}

	// 이미지 생성
	dc := gg.NewContext(OGImageWidth, OGImageHeight)

	// 배경색 설정
	bgColor, err := parseHexColor(opts.BackgroundColor)
	if err != nil {
		return "", fmt.Errorf("배경색을 파싱할 수 없습니다: %w", err)
	}
	dc.SetColor(bgColor)
	dc.DrawRectangle(0, 0, float64(OGImageWidth), float64(OGImageHeight))
	dc.Fill()

	// 텍스트 색상 (흰색)
	dc.SetColor(color.White)

	// 제목 그리기
	titleFontFace := truetype.NewFace(font, &truetype.Options{Size: 60})
	dc.SetFontFace(titleFontFace)

	// 제목 줄바꿈 처리
	titleLines := wrapText(opts.Title, dc, float64(OGImageWidth)-120)
	titleLineHeight := 70.0

	// 제목 위치 계산 (중앙 정렬)
	titleY := float64(OGImageHeight)/2 - float64(len(titleLines))*titleLineHeight/2

	// 제목 그리기
	for i, line := range titleLines {
		y := titleY + float64(i)*titleLineHeight
		dc.DrawStringAnchored(line, float64(OGImageWidth)/2, y, 0.5, 0.5)
	}

	// 작성자 정보 그리기
	authorFontFace := truetype.NewFace(font, &truetype.Options{Size: 30})
	dc.SetFontFace(authorFontFace)

	authorInfo := opts.Author
	if !opts.Date.IsZero() {
		authorInfo += " · " + opts.Date.Format("2006-01-02")
	}

	dc.DrawStringAnchored(authorInfo, 60, float64(OGImageHeight)-60, 0, 0.5)

	// 사이트 이름 그리기 (우측 하단)
	if opts.SiteName != "" {
		dc.DrawStringAnchored(opts.SiteName, float64(OGImageWidth)-60, float64(OGImageHeight)-60, 1, 0.5)
	}

	// 출력 디렉토리 생성
	outputDir := filepath.Dir(opts.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("출력 디렉토리 생성 실패: %w", err)
	}

	// 이미지 저장
	if err := dc.SavePNG(opts.OutputPath); err != nil {
		return "", fmt.Errorf("이미지 저장 실패: %w", err)
	}

	return opts.OutputPath, nil
}

// HexColorToRGBA는 헥스 컬러 코드를 RGBA 색상으로 변환합니다.
func parseHexColor(hex string) (color.Color, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return nil, fmt.Errorf("유효하지 않은 헥스 색상 코드: %s", hex)
	}

	var r, g, b uint8
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return color.RGBA{r, g, b, 255}, nil
}

// wrapText는 텍스트를 지정된 폭에 맞게 줄바꿈합니다.
func wrapText(text string, dc *gg.Context, width float64) []string {
	var lines []string
	words := strings.Fields(text)

	if len(words) == 0 {
		return lines
	}

	currentLine := words[0]
	for _, word := range words[1:] {
		// 현재 줄 + 새 단어의 폭 계산
		w, _ := dc.MeasureString(currentLine + " " + word)
		if w <= width {
			// 폭 내에 있으면 현재 줄에 추가
			currentLine += " " + word
		} else {
			// 폭을 초과하면 새 줄 시작
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	// 마지막 줄 추가
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// CreateOGImageFromText는 간단히 텍스트로부터 OG 이미지를 생성합니다.
func CreateOGImageFromText(title, author string, date time.Time, outputPath string, siteName string) (string, error) {
	return CreateOGImage(OGImageOptions{
		Title:           title,
		Author:          author,
		Date:            date,
		BackgroundColor: GetRandomBackgroundColor(),
		OutputPath:      outputPath,
		SiteName:        siteName,
	})
}
