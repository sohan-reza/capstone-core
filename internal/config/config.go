package config

import (
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port         string        `mapstructure:"PORT"`
		TimeoutRead  time.Duration `mapstructure:"SERVER_TIMEOUT_READ"`
		TimeoutWrite time.Duration `mapstructure:"SERVER_TIMEOUT_WRITE"`
	} `mapstructure:"SERVER"`

	Database struct {
		DSN string `mapstructure:"DSN"`
	} `mapstructure:"DATABASE"`

	Upload struct {
		Dir              string `mapstructure:"UPLOAD_DIR"`
		MaxUploadSizeMB  int64  `mapstructure:"MAX_UPLOAD_SIZE_MB"`
		AllowedFileTypes string `mapstructure:"ALLOWED_FILE_TYPES"`
	} `mapstructure:"UPLOAD"`

	Plagiarism struct {
		APIEndpoint string `mapstructure:"PLAGIARISM_API_ENDPOINT"`
		Threshold   int    `mapstructure:"PLAGIARISM_THRESHOLD"`
	} `mapstructure:"PLAGIARISM"`
}

func LoadConfig(path string) (*Config, error) {

	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	viper.SetDefault("SERVER.PORT", "8080")
	viper.SetDefault("SERVER.SERVER_TIMEOUT_READ", "15s")
	viper.SetDefault("SERVER.SERVER_TIMEOUT_WRITE", "15s")
	viper.SetDefault("UPLOAD.UPLOAD_DIR", "./uploads")
	viper.SetDefault("UPLOAD.MAX_UPLOAD_SIZE_MB", 100)
	viper.SetDefault("UPLOAD.ALLOWED_FILE_TYPES", ".pdf,.zip,.tar,.rar")

	viper.SetDefault("PLAGIARISM.PLAGIARISM_API_ENDPOINT", "localhost:8081")
	viper.SetDefault("PLAGIARISM.PLAGIARISM_THRESHOLD", 15)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		log.Println("No .env file found, using defaults and environment variables")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(config.Upload.Dir, 0755); err != nil {
		return nil, err
	}

	return &config, nil
}
