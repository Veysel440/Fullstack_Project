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
	SentryDSN         string
	SentryEnv         string
	RateLimitRPS      int
	RateLimitBurst    int
	JWTAccessSecret   string
	JWTRefreshSecret  string
	JWTAccessTTLMin   int
	JWTRefreshTTLDays int
	RedisAddr         string
	KafkaBrokers      string
	KafkaTopicItems   string
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
func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	out := []string{}
	for _, p := range strings.Split(s, ",") {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func FromEnv() Config {
	return Config{
		Port:              getenv("PORT", "8080"),
		DBURL:             getenv("DB_URL", ""),
		CORSOrigins:       splitCSV(getenv("CORS_ORIGINS", "")),
		SentryDSN:         getenv("SENTRY_DSN", ""),
		SentryEnv:         getenv("SENTRY_ENV", "dev"),
		RateLimitRPS:      getenvInt("RATE_LIMIT_RPS", 5),
		RateLimitBurst:    getenvInt("RATE_LIMIT_BURST", 10),
		JWTAccessSecret:   getenv("JWT_ACCESS_SECRET", ""),
		JWTRefreshSecret:  getenv("JWT_REFRESH_SECRET", ""),
		JWTAccessTTLMin:   getenvInt("JWT_ACCESS_TTL_MIN", 15),
		JWTRefreshTTLDays: getenvInt("JWT_REFRESH_TTL_DAYS", 7),
		RedisAddr:         getenv("REDIS_ADDR", ""),
		KafkaBrokers:      getenv("KAFKA_BROKERS", ""),
		KafkaTopicItems:   getenv("KAFKA_TOPIC_ITEMS", "items"),
	}
}
