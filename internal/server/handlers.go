package server

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"ocr-server/internal/utils"
	"ocr-server/logger"
	"time"
)

type ocrRequest struct {
	ImagePath     string `json:"image_path,omitempty"`
	Base64Content string `json:"image_base64,omitempty"`
}

type ocrResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func (s *Server) handleOCR(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/stats" {
		logger.LogInfo("收到获取服务器状态的请求")
		stats := s.GetStats()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
		return
	}

	if r.Method != http.MethodPost {
		logger.LogInfo("收到不支持的请求方法: %s", r.Method)
		http.Error(w, "不支持的请求方法", http.StatusMethodNotAllowed)
		return
	}

	var req ocrRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.LogInfo("解析 JSON 失败: %v", err)
		http.Error(w, "解析 JSON 失败", http.StatusBadRequest)
		return
	}
	if req.ImagePath != "" {
		_, err := utils.DetectImageFormat(req.ImagePath)
		if err != nil {
			logger.LogError("请求参数非法！: %v", err)
			http.Error(w, "图片上传格式错误", http.StatusNotAcceptable)
			return
		}
	}
	if req.Base64Content != "" && !utils.IsBase64Image(req.Base64Content) {
		logger.LogError("请求参数非法！")
		http.Error(w, "base64图片格式错误", http.StatusNotAcceptable)
		return

	}
	if req.ImagePath == "" && req.Base64Content == "" {
		logger.LogInfo("收到缺少图像数据的请求")
		http.Error(w, "缺少 image_path 或 image_base64 参数", http.StatusNotAcceptable)
		return
	}

	logger.LogInfo("收到 OCR 请求，正在排队处理")
	task := ocrTask{
		ImagePath: req.ImagePath,
		Response:  make(chan ocrResponse, 1),
	}

	if req.Base64Content != "" {
		imageData, err := base64.StdEncoding.DecodeString(req.Base64Content)
		if err != nil {
			logger.LogInfo("无效的 base64 图像数据: %v", err)
			http.Error(w, "无效的 base64 图像数据", http.StatusBadRequest)
			return
		}
		task.ImageData = imageData
	}

	select {
	case s.taskQueue <- task:
		logger.LogInfo("任务队列处理器已启动")
		response := <-task.Response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case <-time.After(10 * time.Second):
		logger.LogInfo("任务队列已满，请求超时")
		http.Error(w, "服务器繁忙，请稍后再试", http.StatusServiceUnavailable)
	}
}
