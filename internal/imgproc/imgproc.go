package imgproc

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
)

// ThresholdMode 表示阈值模式
type ThresholdMode int

const (
	// ThreshBinary uses a fixed threshold value
	ThreshBinary ThresholdMode = iota
	// ThreshOtsu uses Otsu's method to determine the threshold
	ThreshOtsu
)

// ToGrayscale 将图像转换为灰度
func ToGrayscale(img image.Image) *image.Gray {
	bounds := img.Bounds()
	grayImg := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			oldColor := img.At(x, y)
			r, g, b, _ := oldColor.RGBA()
			gray := uint8((0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 256.0)
			grayImg.Set(x, y, color.Gray{Y: gray})
		}
	}

	return grayImg
}

// Threshold 将二进制阈值应用于灰度图像
func Threshold(img *image.Gray, thresh uint8, mode ThresholdMode) *image.Gray {
	bounds := img.Bounds()
	binaryImg := image.NewGray(bounds)

	if mode == ThreshOtsu {
		thresh = otsuThreshold(img)
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if img.GrayAt(x, y).Y > thresh {
				binaryImg.Set(x, y, color.White)
			} else {
				binaryImg.Set(x, y, color.Black)
			}
		}
	}

	return binaryImg
}

// otsuThreshold 使用 Otsu 的方法计算最佳阈值
func otsuThreshold(img *image.Gray) uint8 {
	histogram := make([]int, 256)
	bounds := img.Bounds()
	totalPixels := (bounds.Max.X - bounds.Min.X) * (bounds.Max.Y - bounds.Min.Y)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			histogram[img.GrayAt(x, y).Y]++
		}
	}

	sum := 0
	for i := 0; i < 256; i++ {
		sum += i * histogram[i]
	}

	sumB := 0
	wB := 0
	wF := 0
	varMax := 0.0
	threshold := 0

	for i := 0; i < 256; i++ {
		wB += histogram[i]
		if wB == 0 {
			continue
		}
		wF = totalPixels - wB
		if wF == 0 {
			break
		}
		sumB += i * histogram[i]
		mB := float64(sumB) / float64(wB)
		mF := float64(sum-sumB) / float64(wF)
		varBetween := float64(wB) * float64(wF) * (mB - mF) * (mB - mF)
		if varBetween > varMax {
			varMax = varBetween
			threshold = i
		}
	}

	return uint8(threshold)
}

// ProcessImage 将灰度和阈值应用于图像
func ProcessImage(img image.Image, thresh uint8, mode ThresholdMode) *image.Gray {
	grayImg := ToGrayscale(img)
	return Threshold(grayImg, thresh, mode)
}

// DecodeBase64Image 解码 base64 编码的图像
func DecodeBase64Image(b64 string) (image.Image, error) {
	reader := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(b64))
	img, _, err := image.Decode(reader)
	return img, err
}

// EncodeToBase64 将图像编码为 base64
func EncodeToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// BytesToImage 将字节切片转换为 image.Image
func BytesToImage(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return img, nil
}

// GrayImageToPNGBytes 将 image.Gray 格式图片转换为 PNG 格式字节切片
func GrayImageToPNGBytes(img *image.Gray) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GrayImageToBytes 将 image.Gray 格式图片转换为对应格式字节切片
func GrayImageToBytes(img *image.Gray, format string) ([]byte, error) {
	var err error
	buf := new(bytes.Buffer)
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(buf, img, nil)
	case "png":
		err = png.Encode(buf, img)
	case "gif":
		err = gif.Encode(buf, img, nil)
	default:
		err = fmt.Errorf("图像格式不符合！只识别jpeg、png、gif图像")
	}

	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
