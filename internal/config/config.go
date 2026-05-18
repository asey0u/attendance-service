package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	AppHost string
	AppPort string

	JWTSecret []byte
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		DBHost:     env("DB_HOST", "localhost"),
		DBPort:     env("DB_PORT", "5432"),
		DBUser:     env("DB_USER", "postgres"),
		DBPassword: env("DB_PASSWORD", ""),
		DBName:     env("DB_NAME", "attendance_db"),
		AppHost:    env("APP_HOST", "0.0.0.0"),
		AppPort:    env("APP_PORT", "8080"),
		JWTSecret:  []byte(env("JWT_SECRET", "")),
	}

	if len(cfg.JWTSecret) == 0 {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}

func (c *Config) Addr() string {
	return c.AppHost + ":" + c.AppPort
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
