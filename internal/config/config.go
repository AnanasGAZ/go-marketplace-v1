package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

const defaultJWTTokenTTL = "15m"

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	JWTSecret   string
	JWTTokenTTL time.Duration
}

func Load() (Config, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 || strings.TrimSpace(jwtSecret) == "" {
		return Config{}, errors.New("JWT_SECRET must be set and contain at least 32 bytes")
	}

	jwtTokenTTL, err := time.ParseDuration(envOrDefault("JWT_TTL", defaultJWTTokenTTL))
	if err != nil {
		return Config{}, fmt.Errorf("JWT_TTL must be a valid duration: %w", err)
	}
	if jwtTokenTTL <= 0 {
		return Config{}, errors.New("JWT_TTL must be positive")
	}

	return Config{
		HTTPAddr:    envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseURL: envOrDefault("DATABASE_URL", "postgres://marketplace:marketplace_dev@localhost:5432/marketplace?sslmode=disable"),
		JWTSecret:   jwtSecret,
		JWTTokenTTL: jwtTokenTTL,
	}, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
