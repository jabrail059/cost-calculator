package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	defaultServerAddr      = ":8080"
	defaultFrontendAPIAddr = ":3000"
	defaultDBHost          = "localhost"
	defaultDBPort          = "5432"
	defaultDBUser          = "postgres"
	defaultDBName          = "proddb"
	defaultDBSSLMode       = "disable"
	defaultWebDir          = "./web"
	defaultUploadMaxMemory = int64(10 << 20)
	defaultOneCTimeout     = 30 * time.Second
)

type Config struct {
	ServerAddr      string
	FrontendAPIAddr string
	DatabaseURL     string
	JWTSecret       string
	OneCURL         string
	OneCTimeout     time.Duration
	UploadMaxMemory int64
	WebDir          string
}

var current Config

func Load() Config {
	cfg := Config{
		ServerAddr:      env("SERVER_ADDR", defaultServerAddr),
		FrontendAPIAddr: env("FRONTEND_API_ADDR", defaultFrontendAPIAddr),
		DatabaseURL:     databaseURL(),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		OneCURL:         os.Getenv("ONEC_URL"),
		OneCTimeout:     durationEnv("ONEC_TIMEOUT", defaultOneCTimeout),
		UploadMaxMemory: int64Env("UPLOAD_MAX_MEMORY", defaultUploadMaxMemory),
		WebDir:          env("WEB_DIR", defaultWebDir),
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET not set")
	}
	current = cfg
	return cfg
}

func Current() Config {
	return current
}

func GetJWTSecret() []byte {
	secret := current.JWTSecret
	if secret == "" {
		secret = os.Getenv("JWT_SECRET")
	}
	if secret == "" {
		log.Fatal("JWT_SECRET not set")
	}
	return []byte(secret)
}

func databaseURL() string {
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		return dsn
	}

	user := env("DB_USER", defaultDBUser)
	password := os.Getenv("DB_PASSWORD")
	host := env("DB_HOST", defaultDBHost)
	port := env("DB_PORT", defaultDBPort)
	name := env("DB_NAME", defaultDBName)
	sslMode := env("DB_SSLMODE", defaultDBSSLMode)

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   host + ":" + port,
		Path:   "/" + name,
	}
	q := u.Query()
	q.Set("sslmode", sslMode)
	u.RawQuery = q.Encode()
	return u.String()
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		log.Fatal(fmt.Errorf("%s has invalid duration %q: %w", key, value, err))
	}
	return parsed
}

func int64Env(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Fatal(fmt.Errorf("%s has invalid integer %q: %w", key, value, err))
	}
	return parsed
}
