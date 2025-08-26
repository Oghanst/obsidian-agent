package property

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	// DefaultLogDir is the default directory for log files.
	DefaultLogDir = "/Users/jianghaojun/Projects/obsidian-agent/agent/logs"
	// DefaultApikey is the default API key for the agent.
	DefaultApikey = "sk-1234567890abcdef1234567890abcdef"
	DefaultLocalServerAddr = "127.0.0.1:8787"
)

type Config struct {
	LogDir string `json:"log_dir"`
	Apikey string `json:"apikey"`
	ServerAddr string `json:"server_addr"`
}

var currentConfig *Config

func LoadDefaultConfig(){
	// 确保LogDir存在
	if _, err := os.Stat(DefaultLogDir); os.IsNotExist(err) {
		if err := os.MkdirAll(DefaultLogDir, 0755); err != nil {
			fmt.Printf("Failed to create log directory: %v\n", err)
		}
	}
	currentConfig = &Config{
		LogDir: DefaultLogDir,
		Apikey: DefaultApikey,
		ServerAddr: DefaultLocalServerAddr,
	}
}

func GetConfig() *Config{
	if currentConfig == nil {
		LoadDefaultConfig()
	}
	return currentConfig
}

// LoadConfig 从指定json文件加载配置
func LoadConfig(filePath string) error {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer jsonFile.Close()
	var config Config
	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&config); err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}
	if config.LogDir == "" {
		config.LogDir = DefaultLogDir
	}
	if config.Apikey == "" {
		config.Apikey = DefaultApikey
	}
	if config.ServerAddr == "" {
		config.ServerAddr = DefaultLocalServerAddr
	}
	currentConfig = &config
	return nil
}