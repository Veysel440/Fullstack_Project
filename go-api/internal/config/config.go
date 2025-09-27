package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port              string
	DBURL             string
	CORSOrigins       []string
	RateLimitRPS      int
	RateLimitBurst    int
	SentryDSN         string
	SentryEnv         string
	JWTAccessSecret   string
	JWTRefreshSecret  string
	JWTAccessTTLMin   int
	JWTRefreshTTLDays int
	RedisAddr         string
	RedisPassword     string
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func fromEnvOrFile(key, def string) string {
	if p := os.Getenv(key + "_FILE"); p != "" {
		if b, err := os.ReadFile(p); err == nil {
			v := strings.TrimSpace(string(b))
			if v != "" {
				return v
			}
		}
	}
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func FromEnv() Config {
	return Config{
		Port:              getenv("PORT", "8080"),
		DBURL:             getenv("DB_URL", "postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable"),
		CORSOrigins:       strings.Split(getenv("CORS_ORIGINS", "http://localhost,http://localhost:3000,http://localhost:3001"), ","),
		RateLimitRPS:      getenvInt("RATE_LIMIT_RPS", 5),
		RateLimitBurst:    getenvInt("RATE_LIMIT_BURST", 10),
		SentryDSN:         os.Getenv("SENTRY_DSN"),
		SentryEnv:         getenv("SENTRY_ENV", "dev"),
		JWTAccessSecret:   fromEnvOrFile("JWT_ACCESS_SECRET", "change-this"),
		JWTRefreshSecret:  fromEnvOrFile("JWT_REFRESH_SECRET", "change-this-too"),
		JWTAccessTTLMin:   getenvInt("JWT_ACCESS_TTL_MIN", 15),
		JWTRefreshTTLDays: getenvInt("JWT_REFRESH_TTL_DAYS", 7),
		RedisAddr:         getenv("REDIS_ADDR", "redis:6379"),
		RedisPassword:     fromEnvOrFile("REDIS_PASSWORD", ""),
	}
}
