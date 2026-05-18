package config

import (
	"log"
	"os"

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
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
