package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                string
	ProjectID           string
	SkyscannerAPIKey    string
	SkyscannerBaseURL   string
	SendGridAPIKey      string
	FCMCredentialsPath  string
	MaxRequestsPerMonth int
	RequestTimeout      time.Duration
	ShutdownTimeout     time.Duration
	LogLevel            string
	RateLimitPerMin     int
	SkyscannerTimeout   time.Duration
	SkyscannerRetries   int
	CbMaxRequests       int
	CbInterval          time.Duration
	CbTimeout           time.Duration
}

func Load() *Config {
	port := getEnv("PORT", "8080")

	maxReqs := 100
	if v, err := strconv.Atoi(getEnv("SKYSCANNER_MAX_REQ", "100")); err == nil && v > 0 {
		maxReqs = v
	}

	reqTimeout := 30 * time.Second
	if v, err := time.ParseDuration(getEnv("REQUEST_TIMEOUT", "30s")); err == nil {
		reqTimeout = v
	}

	shutdownTimeout := 10 * time.Second
	if v, err := time.ParseDuration(getEnv("SHUTDOWN_TIMEOUT", "10s")); err == nil {
		shutdownTimeout = v
	}

	rateLimit := 60
	if v, err := strconv.Atoi(getEnv("RATE_LIMIT_PER_MIN", "60")); err == nil && v > 0 {
		rateLimit = v
	}

	skyTimeout := 10 * time.Second
	if v, err := time.ParseDuration(getEnv("SKYSCANNER_TIMEOUT", "10s")); err == nil {
		skyTimeout = v
	}

	skyRetries := 2
	if v, err := strconv.Atoi(getEnv("SKYSCANNER_RETRIES", "2")); err == nil && v >= 0 {
		skyRetries = v
	}

	cbMaxReqs := 3
	if v, err := strconv.Atoi(getEnv("CB_MAX_REQUESTS", "3")); err == nil && v > 0 {
		cbMaxReqs = v
	}

	cbInterval := 60 * time.Second
	if v, err := time.ParseDuration(getEnv("CB_INTERVAL", "60s")); err == nil {
		cbInterval = v
	}

	cbTimeout := 30 * time.Second
	if v, err := time.ParseDuration(getEnv("CB_TIMEOUT", "30s")); err == nil {
		cbTimeout = v
	}

	return &Config{
		Port:                port,
		ProjectID:           os.Getenv("GCP_PROJECT_ID"),
		SkyscannerAPIKey:    os.Getenv("SKYSCANNER_API_KEY"),
		SkyscannerBaseURL:   getEnv("SKYSCANNER_BASE_URL", "https://partners.api.skyscanner.net/apiservices/v3"),
		SendGridAPIKey:      os.Getenv("SENDGRID_API_KEY"),
		FCMCredentialsPath:  os.Getenv("FCM_CREDENTIALS"),
		MaxRequestsPerMonth: maxReqs,
		RequestTimeout:      reqTimeout,
		ShutdownTimeout:     shutdownTimeout,
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		RateLimitPerMin:     rateLimit,
		SkyscannerTimeout:   skyTimeout,
		SkyscannerRetries:   skyRetries,
		CbMaxRequests:       cbMaxReqs,
		CbInterval:          cbInterval,
		CbTimeout:           cbTimeout,
	}
}

func (c *Config) Validate() error {
	if c.ProjectID == "" {
		return fmt.Errorf("GCP_PROJECT_ID is required")
	}
	if c.SkyscannerAPIKey == "" {
		return fmt.Errorf("SKYSCANNER_API_KEY is required")
	}
	if c.Port == "" {
		return fmt.Errorf("PORT is required")
	}
	if c.MaxRequestsPerMonth <= 0 {
		return fmt.Errorf("SKYSCANNER_MAX_REQ must be positive")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
