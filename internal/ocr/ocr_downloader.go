package ocr

import (
	"fmt"
	"github.com/bodgit/sevenzip"
	"github.com/cenkalti/backoff/v4"
	"io"
	"net/http"
	"net/url"
	"ocr-server/logger"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ocrDownloadURL = "https://github.com/hiroi-sora/PaddleOCR-json/releases/download/v1.4.1/PaddleOCR-json_v1.4.1_windows_x64.7z"
	ocrFileName    = "PaddleOCR-json.7z"
	ocrExeName     = "PaddleOCR-json_v1.4.1/PaddleOCR-json.exe"
	resDir         = "res"
)

func EnsureOCREngine() (string, error) {
	ocrPath := filepath.Join(resDir, ocrExeName)

	if _, err := os.Stat(ocrPath); err == nil {
		fmt.Println("OCR 引擎已存在。")
		return ocrPath, nil
	}

	logger.LogInfo("未找到 OCR 引擎。开始下载过程...")
	if err := downloadOCRWithRetry(); err != nil {
		return "", fmt.Errorf("下载 OCR 引擎失败: %w", err)
	}

	if err := extractArchive(); err != nil {
		return "", fmt.Errorf("提取 OCR 引擎失败: %w", err)
	}

	logger.LogInfo("OCR 引擎安装成功。")
	return ocrPath, nil
}

func downloadOCRWithRetry() error {
	var proxyURL string
	logger.LogInfo("输入代理 URL (留空则直接下载): ")
	fmt.Scanln(&proxyURL)
	if proxyURL == "" {
		fmt.Println("代理 URL，开始下载...")
	}
	if _, err := os.Stat(resDir); os.IsNotExist(err) {
		if err := os.MkdirAll(resDir, 0755); err != nil {
			return fmt.Errorf("创建 res 目录失败: %w", err)
		}
	}

	client := &http.Client{}
	if proxyURL != "" {
		proxyURLParsed, err := url.Parse(proxyURL)
		if err != nil {
			return fmt.Errorf("无效的代理 URL: %w", err)
		}
		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURLParsed)}
	}

	filePath := filepath.Join(resDir, ocrFileName)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	resumePos := fileInfo.Size()
	operation := func() error {
		req, err := http.NewRequest("GET", ocrDownloadURL, nil)
		if err != nil {
			return fmt.Errorf("创建请求失败: %w", err)
		}

		if resumePos > 0 {
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumePos))
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("发送请求失败: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
			return fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
		}

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}
		return nil
	}

	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = 5 * time.Minute

	err = backoff.Retry(operation, backOff)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	logger.LogInfo("\n下载成功完成。")
	return nil
}
func extractArchive() error {
	entries, err := os.ReadDir(resDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return err
	}
	var ocrFileName string
	for _, entry := range entries {
		if !entry.IsDir() {
			if strings.HasSuffix(entry.Name(), "7z") {
				ocrFileName = entry.Name()
			}
		}
	}
	archivePath := filepath.Join(resDir, ocrFileName)
	err = unZip(archivePath)
	if err != nil {
		logger.LogError("提取失败: %v\n\n", err)
	}

	go func() {
		time.Sleep(10 * time.Second)
		// Remove the archive file after extraction
		if err := os.Remove(archivePath); err != nil {
			logger.LogError("警告: 删除文件失败: %v\n", err)
		}
	}()

	logger.LogInfo("提取成功完成。")
	return nil
}
func unZip(archive string) error {
	r, err := sevenzip.OpenReader(archive)
	if err != nil {
		return fmt.Errorf("打开压缩文件失败: %w", err)
	}
	defer r.Close()
	for _, file := range r.File {
		path := filepath.Join(resDir, file.Name)
		if !file.FileInfo().IsDir() {
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				return err
			}
			rc, err := file.Open()
			if err != nil {
				f.Close()
				return err
			}
			_, err = io.Copy(f, rc)
			rc.Close()
			f.Close()
			if err != nil {
				return err
			}
		} else {
			os.MkdirAll(path, os.ModePerm)
		}
	}
	return nil
}
func IsOCREngineInstalled() bool {
	_, err := os.Stat(filepath.Join(resDir, ocrExeName))
	return err == nil
}

func GetOCREnginePath() string {
	return filepath.Join(resDir, ocrExeName)
}

// ProgressReader is a custom io.Reader that reports progress
type ProgressReader struct {
	Reader     io.Reader
	Total      int64
	Current    int64
	OnProgress func(int64)
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Current += int64(n)
	if pr.OnProgress != nil {
		pr.OnProgress(pr.Current)
	}
	return n, err
}
