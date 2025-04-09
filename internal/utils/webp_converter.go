// internal/utils/webp_converter.go
package utils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color/palette"
	stdDraw "image/draw"
	"image/gif"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/webp"
)

// Frame 구조체는 ANMF 청크에서 추출한 정보를 담습니다.
type WebPFrame struct {
	X, Y          int    // 프레임의 좌표 (합성할 때 활용 가능)
	Width, Height int    // 프레임 크기 (원래 값에 1을 더한 값)
	Delay         int    // 프레임 딜레이 (1/100초 단위)
	Data          []byte // 프레임의 raw WebP 이미지 데이터 (VP8/VP8L 청크 등)
}

// ParseAnimatedWebP는 파일을 읽고 RIFF 컨테이너 내의 ANMF 청크들을 파싱하여 WebPFrame 리스트를 반환합니다.
func ParseAnimatedWebP(filePath string) ([]WebPFrame, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 최소 12바이트(RIFF header) 확인: "RIFF" ... "WEBP"
	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
		return nil, errors.New("유효한 WebP 파일이 아닙니다")
	}

	var frames []WebPFrame
	offset := 12
	// RIFF 파일은 각 청크가 8바이트(청크 ID 4바이트 + 크기 4바이트) 단위로 구성됨.
	for offset+8 <= len(data) {
		chunkType := string(data[offset : offset+4])
		chunkSize := binary.LittleEndian.Uint32(data[offset+4 : offset+8])
		chunkDataStart := offset + 8
		chunkDataEnd := chunkDataStart + int(chunkSize)
		if chunkDataEnd > len(data) {
			break
		}

		if chunkType == "ANMF" {
			// ANMF 청크는 적어도 16바이트의 프레임 헤더를 포함해야 함.
			if int(chunkSize) < 16 {
				return nil, fmt.Errorf("ANMF 청크가 너무 작습니다")
			}
			header := data[chunkDataStart : chunkDataStart+16]
			// 3바이트씩 읽어 24비트 정수로 변환 (리틀 엔디언)
			x := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)
			y := int(uint32(header[3]) | uint32(header[4])<<8 | uint32(header[5])<<16)
			// 프레임 폭/높이는 값에 1을 더해 사용
			wMinus1 := int(uint32(header[6]) | uint32(header[7])<<8 | uint32(header[8])<<16)
			hMinus1 := int(uint32(header[9]) | uint32(header[10])<<8 | uint32(header[11])<<16)
			delay := int(binary.LittleEndian.Uint32(header[12:16])) // 1/100초 단위

			frameWidth := wMinus1 + 1
			frameHeight := hMinus1 + 1

			// 나머지 바이트는 실제 프레임 이미지 데이터 (예: VP8 또는 VP8L 청크)
			frameData := data[chunkDataStart+16 : chunkDataEnd]

			frame := WebPFrame{
				X:      x,
				Y:      y,
				Width:  frameWidth,
				Height: frameHeight,
				Delay:  delay,
				Data:   frameData,
			}
			frames = append(frames, frame)
		}
		// 홀수 바이트 패딩 처리
		offset = chunkDataEnd
		if offset%2 == 1 {
			offset++
		}
	}

	if len(frames) == 0 {
		return nil, errors.New("ANMF 청크를 찾을 수 없습니다")
	}
	return frames, nil
}

// ReconstructWebP는 ANMF 프레임의 image 데이터를 유효한 WebP 파일 형식으로 재구성합니다.
func ReconstructWebP(frame WebPFrame) []byte {
	// 유효한 WebP 파일은 "RIFF" + 파일 크기 (frameData+4바이트) + "WEBP" + frameData
	frameData := frame.Data
	fileSize := uint32(len(frameData) + 4) // "WEBP" 문자열 포함
	buf := new(bytes.Buffer)
	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, fileSize)
	buf.WriteString("WEBP")
	buf.Write(frameData)
	return buf.Bytes()
}

// ResizeImage는 주어진 이미지를 최대 maxWidth×maxHeight 크기에 맞게, 원본 비율을 유지하며 리사이즈합니다.
// maxWidth나 maxHeight가 0이면 다른 값에 맞춰 비율을 유지합니다.
// 두 값이 모두 0이면 원본 이미지를 그대로 반환합니다.
func ResizeImage(src image.Image, maxWidth, maxHeight int) image.Image {
	bounds := src.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// 둘 다 0이면 원본 반환
	if maxWidth == 0 && maxHeight == 0 {
		return src
	}

	// 하나만 0이면 다른 값에 맞춰 비율 계산
	if maxWidth == 0 {
		// 너비가 0이면 높이에 맞춰 비율 계산
		maxWidth = int(float64(srcWidth) * float64(maxHeight) / float64(srcHeight))
	} else if maxHeight == 0 {
		// 높이가 0이면 너비에 맞춰 비율 계산
		maxHeight = int(float64(srcHeight) * float64(maxWidth) / float64(srcWidth))
	}

	// 안전장치: 최소 크기 보장
	if maxWidth < 1 {
		maxWidth = 1
	}
	if maxHeight < 1 {
		maxHeight = 1
	}

	// 실제 리사이징할 크기 계산
	scale := 1.0
	scaleW := float64(maxWidth) / float64(srcWidth)
	scaleH := float64(maxHeight) / float64(srcHeight)

	// 더 작은 비율을 선택하여 이미지가 지정된 크기 내에 맞도록 함
	scale = min(scaleW, scaleH)

	newWidth := int(float64(srcWidth) * scale)
	newHeight := int(float64(srcHeight) * scale)

	// 최소 크기 보장
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	xdraw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, bounds, xdraw.Over, nil)
	return dst
}

// SelectFrames는 전체 프레임 중 최대 maxFrames 개만 균등하게 선택합니다.
func SelectFrames(frames []image.Image, delays []int, maxFrames int) ([]image.Image, []int) {
	total := len(frames)
	if total <= maxFrames {
		return frames, delays
	}
	var selFrames []image.Image
	var selDelays []int
	interval := float64(total-1) / float64(maxFrames-1)
	for i := 0; i < maxFrames; i++ {
		index := int(float64(i)*interval + 0.5) // 반올림하여 선택
		if index >= total {
			index = total - 1
		}
		selFrames = append(selFrames, frames[index])
		selDelays = append(selDelays, delays[index])
	}
	return selFrames, selDelays
}

// ConvertWebPToJPG는 WebP 이미지(애니메이션 포함)를 JPG로 변환합니다.
func ConvertWebPToJPG(webpPath, jpgPath string, maxWidth, maxHeight int, quality int) (string, error) {
	// 파일이 존재하는지 확인
	if _, err := os.Stat(webpPath); os.IsNotExist(err) {
		return "", fmt.Errorf("WebP 파일이 존재하지 않습니다: %s", webpPath)
	}

	// 파일 열기
	file, err := os.Open(webpPath)
	if err != nil {
		return "", fmt.Errorf("WebP 파일 열기 실패: %w", err)
	}
	defer file.Close()

	// // 파일 전체 읽기
	// // fileData, err := io.ReadAll(file)
	// _, err = io.ReadAll(file)
	// if err != nil {
	// 	return "", fmt.Errorf("WebP 파일 읽기 실패: %w", err)
	// }

	var img image.Image

	// 애니메이션 WebP인지 확인
	isAnimated, _ := IsAnimatedWebP(webpPath)
	if isAnimated {
		// 애니메이션 WebP 처리
		frames, err := ParseAnimatedWebP(webpPath)
		if err != nil || len(frames) == 0 {
			return "", fmt.Errorf("WebP 애니메이션 파싱 실패: %w", err)
		}

		// 첫 번째 프레임만 사용
		webpData := ReconstructWebP(frames[0])

		img, err = webp.Decode(bytes.NewReader(webpData))
		if err != nil {
			// 재구성에 실패한 경우 원본 파일에서 디코딩 시도
			file.Seek(0, io.SeekStart)
			img, err = webp.Decode(file)
			if err != nil {
				return "", fmt.Errorf("WebP 디코딩 실패: %w", err)
			}
		}
	} else {
		// 일반 WebP 이미지
		file.Seek(0, io.SeekStart)
		img, err = webp.Decode(file)
		if err != nil {
			return "", fmt.Errorf("WebP 디코딩 실패: %w", err)
		}
	}

	// 이미지 리사이징
	if maxWidth > 0 || maxHeight > 0 {
		img = ResizeImage(img, maxWidth, maxHeight)
	}

	// JPG 파일 경로 생성
	if jpgPath == "" {
		ext := filepath.Ext(webpPath)
		jpgPath = webpPath[:len(webpPath)-len(ext)] + ".jpg"
	}

	// 디렉토리 생성
	dir := filepath.Dir(jpgPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("디렉토리 생성 실패: %w", err)
	}

	// JPG 파일 저장
	outFile, err := os.Create(jpgPath)
	if err != nil {
		return "", fmt.Errorf("JPG 파일 생성 실패: %w", err)
	}
	defer outFile.Close()

	if quality <= 0 || quality > 100 {
		quality = 85 // 기본 품질
	}

	err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: quality})
	if err != nil {
		return "", fmt.Errorf("JPG 인코딩 실패: %w", err)
	}

	// 파일 크기 확인 (디버깅용)
	fileInfo, _ := os.Stat(jpgPath)
	if fileInfo != nil {
		fmt.Printf("생성된 JPG 파일 크기: %d 바이트\n", fileInfo.Size())
	}

	return jpgPath, nil
}

// ConvertWebPToGIF는 WebP 애니메이션을 GIF로 변환합니다.
// maxFrames는 GIF에 포함할 최대 프레임 수입니다.
func ConvertWebPToGIF(webpPath, gifPath string, maxWidth, maxHeight, maxFrames int) (string, error) {
	// 파일이 존재하는지 확인
	if _, err := os.Stat(webpPath); os.IsNotExist(err) {
		return "", fmt.Errorf("WebP 파일이 존재하지 않습니다: %s", webpPath)
	}

	// 프레임 파싱
	frames, err := ParseAnimatedWebP(webpPath)
	if err != nil {
		return "", fmt.Errorf("WebP 애니메이션 파싱 실패: %w", err)
	}

	var decodedImages []image.Image
	var frameDelays []int

	// 모든 프레임 디코딩 및 리사이징
	for i, frame := range frames {
		webpData := ReconstructWebP(frame)
		img, err := webp.Decode(bytes.NewReader(webpData))
		if err != nil {
			fmt.Printf("프레임 %d 디코딩 오류: %v\n", i, err)
			continue
		}

		// 리사이징
		if maxWidth > 0 || maxHeight > 0 {
			img = ResizeImage(img, maxWidth, maxHeight)
		}

		decodedImages = append(decodedImages, img)
		frameDelays = append(frameDelays, frame.Delay)
	}

	if len(decodedImages) == 0 {
		return "", errors.New("디코딩된 유효 프레임이 없습니다")
	}

	// 프레임 샘플링: 최대 maxFrames 프레임만 선택
	if maxFrames <= 0 {
		maxFrames = 10 // 기본값
	}

	selectedImages, selectedDelays := SelectFrames(decodedImages, frameDelays, maxFrames)

	// GIF 경로 생성
	if gifPath == "" {
		// 같은 위치에 같은 이름으로 확장자만 변경
		ext := filepath.Ext(webpPath)
		gifPath = webpPath[:len(webpPath)-len(ext)] + ".gif"
	}

	// 디렉토리 생성
	dir := filepath.Dir(gifPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("디렉토리 생성 실패: %w", err)
	}

	// GIF 애니메이션 생성
	outGIF := &gif.GIF{}
	for i, img := range selectedImages {
		bounds := img.Bounds()
		paletted := image.NewPaletted(bounds, palette.Plan9)
		stdDraw.Draw(paletted, bounds, img, bounds.Min, stdDraw.Over)
		outGIF.Image = append(outGIF.Image, paletted)
		outGIF.Delay = append(outGIF.Delay, selectedDelays[i]/10) // 1/100초 -> 1/10초 단위로 변환
	}

	outFile, err := os.Create(gifPath)
	if err != nil {
		return "", fmt.Errorf("GIF 파일 생성 실패: %w", err)
	}
	defer outFile.Close()

	if err := gif.EncodeAll(outFile, outGIF); err != nil {
		return "", fmt.Errorf("GIF 인코딩 실패: %w", err)
	}

	return gifPath, nil
}

// IsAnimatedWebP는 WebP 파일이 애니메이션인지 확인합니다.
func IsAnimatedWebP(filePath string) (bool, error) {
	frames, err := ParseAnimatedWebP(filePath)
	if err != nil {
		// 파싱 오류가 발생했다면 애니메이션이 아닌 것으로 간주
		return false, nil
	}
	return len(frames) > 1, nil
}
