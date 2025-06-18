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
		Host     string `mapstructure:"DB_HOST"`
		Port     string `mapstructure:"DB_PORT"`
		User     string `mapstructure:"DB_USER"`
		Password string `mapstructure:"DB_PASSWORD"`
		Name     string `mapstructure:"DB_NAME"`
		SSLMode  string `mapstructure:"DB_SSLMODE"`
	} `mapstructure:"DATABASE"`

	AWS struct {
		Region          string `mapstructure:"AWS_REGION"`
		BucketName      string `mapstructure:"AWS_BUCKET_NAME"`
		AccessKeyID     string `mapstructure:"AWS_ACCESS_KEY_ID"`
		SecretAccessKey string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	} `mapstructure:"AWS"`

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

	// Database defaults
	viper.SetDefault("DATABASE.DB_HOST", "localhost")
	viper.SetDefault("DATABASE.DB_PORT", "5432")
	viper.SetDefault("DATABASE.DB_USER", "postgres")
	viper.SetDefault("DATABASE.DB_PASSWORD", "postgres")
	viper.SetDefault("DATABASE.DB_NAME", "capstone")
	viper.SetDefault("DATABASE.DB_SSLMODE", "disable")

	// AWS defaults
	viper.SetDefault("AWS.AWS_REGION", "us-east-1")
	viper.SetDefault("AWS.AWS_BUCKET_NAME", "your-bucket-name")
	viper.SetDefault("AWS.AWS_ACCESS_KEY_ID", "")
	viper.SetDefault("AWS.AWS_SECRET_ACCESS_KEY", "")

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
