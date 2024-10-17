package utils

import (
	"fmt"
	"image"
	"ocr-server/logger"
	"os"
	"strings"
)

func DetectImageFormat(filePath string) (string, error) {
	file, err := os.Open(filePath) // 打开图像文件
	if err != nil {
		logger.LogError("打开文件失败：", err)
		return "", err
	}
	defer file.Close()

	// 使用 image.DecodeConfig 来检测图像格式
	_, format, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Println("图像格式失败：", err)
		return "", err
	}

	switch format {
	case "jpeg", "jpg":
		return "jpeg", nil
	case "png":
		return "png", nil
	case "gif":
		return "gif", nil
	default:
		return "", fmt.Errorf("未知文件类型！")
	}
}

func IsBase64Image(base64Str string) bool {
	// 常见图片的 Base64 开头部分
	imagePrefixes := []string{"data:image/jpeg;base64,", "data:image/png;base64,", "data:image/gif;base64,"}
	for _, prefix := range imagePrefixes {
		if strings.HasPrefix(base64Str, prefix) {
			return true
		}
	}
	return false
}
