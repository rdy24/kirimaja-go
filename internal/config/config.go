package config

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	DatabaseURL       string
	JWTSecret         string
	JWTExpiresIn      string
	OpenCageAPIKey    string
	MidtransServerKey string
	MidtransEnv       string
	RedisURL          string
	SMTPHost          string
	SMTPPort          string
	SMTPUser          string
	SMTPPass          string
	SMTPSender        string
	FrontendURL       string
	PublicDir         string
	AppEnv            string

	MinIOEndpoint       string
	MinIOPublicEndpoint string
	MinIOAccessKey      string
	MinIOSecretKey      string
	MinIOBucket         string
	MinIOUseSSL         bool
	MinIOPresignExpiry  time.Duration
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}
	return &Config{
		Port:              getEnv("PORT", "3000"),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		JWTSecret:         getEnv("JWT_SECRET_KEY", ""),
		JWTExpiresIn:      getEnv("JWT_EXPIRES_IN", "24h"),
		OpenCageAPIKey:    getEnv("OPENCAGE_API_KEY", ""),
		MidtransServerKey: getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransEnv:       getEnv("MIDTRANS_ENV", "sandbox"),
		RedisURL:          getEnv("REDIS_URL", "localhost:6379"),
		SMTPHost:          getEnv("SMTP_HOST", ""),
		SMTPPort:          getEnv("SMTP_PORT", "587"),
		SMTPUser:          getEnv("SMTP_USER", ""),
		SMTPPass:          getEnv("SMTP_PASS", ""),
		SMTPSender:        getEnv("SMTP_EMAIL_SENDER", ""),
		FrontendURL:       getEnv("FRONTEND_URL", "http://localhost:5173"),
		PublicDir:         getEnv("PUBLIC_DIR", "./public"),
		AppEnv:            getEnv("APP_ENV", "development"),

		MinIOEndpoint:       getEnv("MINIO_ENDPOINT", ""),
		MinIOPublicEndpoint: getEnv("MINIO_PUBLIC_ENDPOINT", ""),
		MinIOAccessKey:      getEnv("MINIO_ACCESS_KEY", ""),
		MinIOSecretKey:      getEnv("MINIO_SECRET_KEY", ""),
		MinIOBucket:         getEnv("MINIO_BUCKET", "kirimaja"),
		MinIOUseSSL:         getEnv("MINIO_USE_SSL", "false") == "true",
		MinIOPresignExpiry:  parseDuration(getEnv("MINIO_PRESIGN_EXPIRY", "15m"), 15*time.Minute),
	}
}

func parseDuration(s string, fallback time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	return fallback
}

// Validate fails fast on missing critical configuration. An empty JWT secret
// signs every token with []byte("") (silent total auth bypass); an empty DB
// URL or Midtrans key breaks core flows. Better to crash on boot than serve.
func (c *Config) Validate() {
	var missing []string
	if c.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET_KEY")
	}
	if c.MidtransServerKey == "" {
		missing = append(missing, "MIDTRANS_SERVER_KEY")
	}
	// Cutover to MinIO is full — without it uploads/QR/PDF break entirely.
	if c.MinIOEndpoint == "" {
		missing = append(missing, "MINIO_ENDPOINT")
	}
	if c.MinIOAccessKey == "" {
		missing = append(missing, "MINIO_ACCESS_KEY")
	}
	if c.MinIOSecretKey == "" {
		missing = append(missing, "MINIO_SECRET_KEY")
	}
	if len(missing) > 0 {
		log.Fatalf("config: required environment variables not set: %s", strings.Join(missing, ", "))
	}
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
