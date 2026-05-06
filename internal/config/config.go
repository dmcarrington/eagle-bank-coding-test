package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	Port      string
	DBPath    string
	JWTSecret []byte
	JWTTTL    time.Duration
}

func Load() (Config, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return Config{}, errors.New("JWT_SECRET environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./eagle.db"
	}

	ttlStr := os.Getenv("JWT_TTL")
	ttl := 24 * time.Hour
	if ttlStr != "" {
		parsed, err := time.ParseDuration(ttlStr)
		if err != nil {
			return Config{}, errors.New("invalid JWT_TTL: must be a valid duration (e.g. 24h)")
		}
		ttl = parsed
	}

	return Config{
		Port:      port,
		DBPath:    dbPath,
		JWTSecret: []byte(secret),
		JWTTTL:    ttl,
	}, nil
}
