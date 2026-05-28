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
	LogicGRPC LogicGRPCConfig `yaml:"logic_grpc"`
	JWT       JWTConfig       `yaml:"jwt"`
	CORS      CORSConfig      `yaml:"cors"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type LogicGRPCConfig struct {
	Address string `yaml:"address"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
}

type CORSConfig struct {
	AllowOrigins []string `yaml:"allow_origins"`
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
		Server:    ServerConfig{Host: "0.0.0.0", Port: 8080},
		LogicGRPC: LogicGRPCConfig{Address: "127.0.0.1:50051"},
		CORS: CORSConfig{
			AllowOrigins: []string{"http://localhost:5173", "http://localhost:5174"},
		},
	}
}

func applyEnv(cfg *Config) {
	setString(&cfg.JWT.Secret, "JWT_SECRET")
	setString(&cfg.LogicGRPC.Address, "LOGIC_GRPC_ADDRESS")
	setString(&cfg.Server.Host, "SERVER_HOST")
	setInt(&cfg.Server.Port, "SERVER_PORT")
	if v := strings.TrimSpace(os.Getenv("CORS_ALLOW_ORIGINS")); v != "" {
		cfg.CORS.AllowOrigins = strings.Split(v, ",")
	}
}

func (c *Config) validate() error {
	if c.JWT.Secret == "" {
		return fmt.Errorf("missing required env: JWT_SECRET (copy .env.example to .env)")
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

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
