package common

import (
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	configOnce sync.Once
	config     *Config
)

func NewConfig() *Config {
	configOnce.Do(func() {
		content, err := os.ReadFile("config.yaml")
		if err != nil {
			panic(err)
		}

		config = &Config{}
		err = yaml.Unmarshal(content, config)
		if err != nil {
			panic(err)
		}
	})
	return config
}

type Config struct {
	Server *ServerConfig `yaml:"server"`
	DB     *DBConfig     `yaml:"db"`
	Minio  *MinioConfig  `yaml:"minio"`
}

type ServerConfig struct {
	PublicAddr      string        `yaml:"publicAddr"`      // 公网地址
	PrivateAddr     string        `yaml:"privateAddr"`     // 内网地址
	UploadTimeout   time.Duration `yaml:"uploadTimeout"`   // 上传URL有效时间
	DownloadTimeout time.Duration `yaml:"downloadTimeout"` // 下载URL有效时间
}

type DBConfig struct {
	Type            string        `yaml:"type"`
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	DBName          string        `yaml:"dbname"`
	MaxOpenConns    int           `yaml:"maxOpenConns"`
	MaxIdleConns    int           `yaml:"maxIdleConns"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime"`
}

type MinioConfig struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"accessKey"`
	SecretKey string `yaml:"secretKey"`
	BucketID  string `yaml:"bucketID"`
}
