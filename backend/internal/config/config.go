package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Stripe   StripeConfig
	Storage  StorageConfig
	Email    EmailConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	Port            string
	Environment     string
	AllowedOrigins  []string
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type JWTConfig struct {
	Secret                string
	AccessTokenDuration   time.Duration
	RefreshTokenDuration  time.Duration
}

type StripeConfig struct {
	SecretKey      string
	WebhookSecret  string
	TrialDays      int
}

type StorageConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	UseSSL          bool
}

type EmailConfig struct {
	APIKey     string
	FromEmail  string
	FromName   string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:            getEnv("PORT", "8080"),
			Environment:     getEnv("ENVIRONMENT", "development"),
			AllowedOrigins:  []string{getEnv("FRONTEND_URL", "http://localhost:3000")},
			ShutdownTimeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			Name:            getEnv("DB_NAME", "postgres"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: 10 * time.Minute,
		},
		JWT: JWTConfig{
			Secret:               getEnv("JWT_SECRET", "your-secret-key-change-this"),
			AccessTokenDuration:  15 * time.Minute,
			RefreshTokenDuration: 7 * 24 * time.Hour,
		},
		Stripe: StripeConfig{
			SecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
			WebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
			TrialDays:     getEnvAsInt("STRIPE_TRIAL_DAYS", 14),
		},
		Storage: StorageConfig{
			Endpoint:        getEnv("STORAGE_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("STORAGE_ACCESS_KEY_ID", "minioadmin"),
			SecretAccessKey: getEnv("STORAGE_SECRET_ACCESS_KEY", "minioadmin"),
			Bucket:          getEnv("STORAGE_BUCKET", "saas-files"),
			UseSSL:          getEnvAsBool("STORAGE_USE_SSL", false),
		},
		Email: EmailConfig{
			APIKey:    getEnv("RESEND_API_KEY", ""),
			FromEmail: getEnv("EMAIL_FROM", "noreply@example.com"),
			FromName:  getEnv("EMAIL_FROM_NAME", "SaaS Blueprint"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
