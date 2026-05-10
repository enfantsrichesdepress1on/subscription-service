package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTP     HTTPConfig     `yaml:"http"`
	Database DatabaseConfig `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
}

type HTTPConfig struct {
	Port string `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"sslmode"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

func Load() (Config, error) {
	_ = godotenv.Load()
	cfg := Config{
		HTTP:     HTTPConfig{Port: "8080"},
		Database: DatabaseConfig{Host: "localhost", Port: "5432", User: "postgres", Password: "postgres", Name: "subscriptions", SSLMode: "disable"},
		Log:      LogConfig{Level: "info"},
	}
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return Config{}, err
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return Config{}, err
		}
	}
	override(&cfg.HTTP.Port, "HTTP_PORT")
	override(&cfg.Database.Host, "DB_HOST")
	override(&cfg.Database.Port, "DB_PORT")
	override(&cfg.Database.User, "DB_USER")
	override(&cfg.Database.Password, "DB_PASSWORD")
	override(&cfg.Database.Name, "DB_NAME")
	override(&cfg.Database.SSLMode, "DB_SSLMODE")
	override(&cfg.Log.Level, "LOG_LEVEL")
	if _, err := strconv.Atoi(cfg.HTTP.Port); err != nil {
		return Config{}, fmt.Errorf("invalid HTTP_PORT: %w", err)
	}
	return cfg, nil
}

func override(dst *string, key string) {
	if value := os.Getenv(key); value != "" {
		*dst = value
	}
}

func (c Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.Name, c.Database.SSLMode)
}
