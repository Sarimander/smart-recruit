package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	MySQL     MySQLConfig     `yaml:"mysql"`
	JWT       JWTConfig       `yaml:"jwt"`
	OSS       OSSConfig       `yaml:"oss"`
	DashScope DashScopeConfig `yaml:"dashscope"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type MySQLConfig struct {
	DSN string `yaml:"dsn"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type OSSConfig struct {
	Endpoint              string `yaml:"endpoint"`
	AccessKeyID           string `yaml:"access_key_id"`
	AccessKeySecret       string `yaml:"access_key_secret"`
	BucketName            string `yaml:"bucket_name"`
	UploadExpireSeconds   int64  `yaml:"upload_expire_seconds"`
	DownloadExpireSeconds int64  `yaml:"download_expire_seconds"`
}

type DashScopeConfig struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
}

func Load(path string) (*Config, error) {
	cfg := defaultConfig()
	if path != "" {
		if data, err := os.ReadFile(path); err == nil {
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return nil, fmt.Errorf("parse config: %w", err)
			}
		}
	}
	applyEnv(&cfg)
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func defaultConfig() Config {
	return Config{
		Server: ServerConfig{Host: "0.0.0.0", Port: 50051},
		JWT:    JWTConfig{ExpireHours: 72},
		OSS: OSSConfig{
			UploadExpireSeconds:   900,
			DownloadExpireSeconds: 300,
		},
		DashScope: DashScopeConfig{
			BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
			Model:   "qwen-plus",
		},
	}
}

func applyEnv(cfg *Config) {
	setString(&cfg.MySQL.DSN, "MYSQL_DSN")
	setString(&cfg.JWT.Secret, "JWT_SECRET")
	setInt(&cfg.JWT.ExpireHours, "JWT_EXPIRE_HOURS")
	setString(&cfg.OSS.Endpoint, "OSS_ENDPOINT")
	setString(&cfg.OSS.AccessKeyID, "OSS_ACCESS_KEY_ID")
	setString(&cfg.OSS.AccessKeySecret, "OSS_ACCESS_KEY_SECRET")
	setString(&cfg.OSS.BucketName, "OSS_BUCKET_NAME")
	setInt64(&cfg.OSS.UploadExpireSeconds, "OSS_UPLOAD_EXPIRE_SECONDS")
	setInt64(&cfg.OSS.DownloadExpireSeconds, "OSS_DOWNLOAD_EXPIRE_SECONDS")
	setString(&cfg.DashScope.APIKey, "DASHSCOPE_API_KEY")
	setString(&cfg.DashScope.BaseURL, "DASHSCOPE_BASE_URL")
	setString(&cfg.DashScope.Model, "DASHSCOPE_MODEL")
	setString(&cfg.Server.Host, "SERVER_HOST")
	setInt(&cfg.Server.Port, "SERVER_PORT")
}

func (c *Config) validate() error {
	missing := []string{}
	require := func(ok bool, name string) {
		if !ok {
			missing = append(missing, name)
		}
	}
	require(c.MySQL.DSN != "", "MYSQL_DSN")
	require(c.JWT.Secret != "", "JWT_SECRET")
	require(c.OSS.Endpoint != "", "OSS_ENDPOINT")
	require(c.OSS.AccessKeyID != "", "OSS_ACCESS_KEY_ID")
	require(c.OSS.AccessKeySecret != "", "OSS_ACCESS_KEY_SECRET")
	require(c.OSS.BucketName != "", "OSS_BUCKET_NAME")
	require(c.DashScope.APIKey != "", "DASHSCOPE_API_KEY")
	if len(missing) > 0 {
		return fmt.Errorf("missing required env: %s (copy .env.example to .env)", strings.Join(missing, ", "))
	}
	return nil
}

func setString(target *string, key string) {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		*target = v
	}
}

func setInt(target *int, key string) {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*target = n
		}
	}
}

func setInt64(target *int64, key string) {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			*target = n
		}
	}
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
