package config

import "os"

type Config struct {
	DatabaseURL  string
	HTTPPort     string
	ResendAPIKey string
	MailFrom     string
	Env          string // "dev" | "prod"
}

func Load() *Config {
	return &Config{
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://user:pwd@localhost:5432/dbname?sslmode=disable"),
		HTTPPort:     getEnv("HTTP_PORT", "8080"),
		ResendAPIKey: getEnv("RESEND_API_KEY", ""),
		MailFrom:     getEnv("MAIL_FROM", "noreply@mg.lavorus.com"),
		Env:          getEnv("ENV", "dev"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
