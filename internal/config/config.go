package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"ocr-server/internal/ocr"
	"ocr-server/logger"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Addr             string        `mapstructure:"addr" yaml:"addr" validate:"required"`                                     // 服务器地址
	Port             int           `mapstructure:"port" yaml:"port" validate:"required,min=1,max=65535"`                     // 服务器端口
	OCRExePath       string        `mapstructure:"ocr_exe_path" yaml:"ocr_exe_path"`                                         // OCR 可执行文件路径
	MinProcessors    int           `mapstructure:"min_processors" yaml:"min_processors" validate:"required,min=2"`           // 最小处理器数量
	MaxProcessors    int           `mapstructure:"max_processors" yaml:"max_processors" validate:"required,min=1"`           // 最大处理器数量
	QueueSize        int           `mapstructure:"queue_size" yaml:"queue_size" validate:"required,min=1"`                   // 任务队列大小
	ScaleThreshold   int64         `mapstructure:"scale_threshold" yaml:"scale_threshold" validate:"required,min=0"`         // 扩展处理器阈值
	DegradeThreshold int64         `mapstructure:"degrade_threshold" yaml:"degrade_threshold" validate:"required,min=0"`     // 缩减处理器阈值
	IdleTimeout      time.Duration `mapstructure:"idle_timeout" yaml:"idle_timeout" validate:"required"`                     // 处理器空闲超时时间
	WarmUpCount      int           `mapstructure:"warm_up_count" yaml:"warm_up_count" validate:"required,min=0"`             // 预热处理器数量
	ShutdownTimeout  time.Duration `mapstructure:"shutdown_timeout" yaml:"shutdown_timeout" validate:"required"`             // 优雅关闭超时时间
	LogFilePath      string        `mapstructure:"log_file_path" yaml:"log_file_path" validate:"required"`                   // 日志文件路径名
	LogMaxSize       int           `mapstructure:"log_max_size" yaml:"log_max_size" validate:"required,min=10"`              // 日志文件最大大小（MB）
	LogMaxBackups    int           `mapstructure:"log_max_backups" yaml:"log_max_backups" validate:"required,min=0"`         // 保留的旧日志文件最大数量
	LogMaxAge        int           `mapstructure:"log_max_age" yaml:"log_max_age" validate:"required,min=1"`                 // 保留旧日志文件的最大天数
	LogCompress      bool          `mapstructure:"log_compress" yaml:"log_compress"`                                         // 是否压缩轮转的日志文件
	ThresholdMode    int           `mapstructure:"threshold_mode" yaml:"threshold_mode"`                                     // 阈值模式
	ThresholdValue   int           `mapstructure:"threshold_value" yaml:"threshold_value" validate:"required,min=0,max=255"` // 阈值
}

func LoadConfig() (Config, error) {
	var cfg Config
	setDefaults(&cfg)                                  // 设置默认值
	if err := generateDefaultConfig(cfg); err != nil { // 生成默认配置文件（如果不存在）
		return Config{}, fmt.Errorf("生成默认配置文件错误: %w", err)
	}

	// 读取配置文件
	if err := readConfigFile(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func setDefaults(cfg *Config) {
	cfg.OCRExePath = ocr.GetOCREnginePath()
	cfg.MaxProcessors = runtime.NumCPU()
	cfg.ScaleThreshold = 75
	cfg.DegradeThreshold = 25
	cfg.ShutdownTimeout = 30 * time.Second
	cfg.LogMaxBackups = 3
	cfg.LogMaxAge = 28
	cfg.ThresholdMode = 0
	cfg.ThresholdValue = 100
}

func generateDefaultConfig(cfg Config) error {
	configPath := getConfigFilePath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("序列化默认配置错误: %w", err)
		}
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return fmt.Errorf("创建配置目录错误: %w", err)
		}
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("写入默认配置文件错误: %w", err)
		}
		logger.LogInfo("已生成默认配置文件: %s\n", configPath)
	}
	return nil
}

func readConfigFile(cfg *Config) error {
	viper.SetConfigFile(getConfigFilePath())
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件错误: %w", err)
	}
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("解析配置错误: %w", err)
	}
	return nil
}

func ValidateConfig(cfg *Config) error {
	validate := validator.New()
	return validate.Struct(cfg)
}
func getConfigFilePath() string {
	homeDir := "."
	return filepath.Join(homeDir, ".ocr-server", "config.yaml")
}
