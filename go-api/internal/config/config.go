package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port, DBUser, DBPass, DBHost, DBPort, DBName, SSLMode, CORSOrigins string
	RateLimitRPS                                                       float64
	RateLimitBurst                                                     int
	SentryDSN                                                          string
	SentryEnv                                                          string
	JWTAccessSecret                                                    string
	JWTRefreshSecret                                                   string
	JWTAccessTTLMin                                                    int
	JWTRefreshTTLDays                                                  int
}

func FromEnv() Config {
	return Config{
		Port:              get("PORT", "8080"),
		DBUser:            get("DB_USER", "app"),
		DBPass:            get("DB_PASS", "app"),
		DBHost:            get("DB_HOST", "postgres"),
		DBPort:            get("DB_PORT", "5432"),
		DBName:            get("DB_NAME", "appdb"),
		SSLMode:           get("SSL_MODE", "disable"),
		CORSOrigins:       get("CORS_ORIGINS", "*"),
		RateLimitRPS:      getFloat("RATE_LIMIT_RPS", 5),
		RateLimitBurst:    getInt("RATE_LIMIT_BURST", 10),
		SentryDSN:         get("SENTRY_DSN", ""),
		SentryEnv:         get("SENTRY_ENV", "dev"),
		JWTAccessSecret:   get("JWT_ACCESS_SECRET", "dev-access"),
		JWTRefreshSecret:  get("JWT_REFRESH_SECRET", "dev-refresh"),
		JWTAccessTTLMin:   getInt("JWT_ACCESS_TTL_MIN", 15),
		JWTRefreshTTLDays: getInt("JWT_REFRESH_TTL_DAYS", 7),
	}
}

func get(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func getInt(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return d
}
func getFloat(k string, d float64) float64 {
	if v := os.Getenv(k); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return d
}
