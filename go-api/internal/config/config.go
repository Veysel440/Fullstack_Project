package config

import "os"

type Config struct {
	Port, DBUser, DBPass, DBHost, DBPort, DBName, SSLMode, CORSOrigins string
}

func FromEnv() Config {
	return Config{
		Port:        get("PORT", "8080"),
		DBUser:      get("DB_USER", "app"),
		DBPass:      get("DB_PASS", "app"),
		DBHost:      get("DB_HOST", "postgres"),
		DBPort:      get("DB_PORT", "5432"),
		DBName:      get("DB_NAME", "appdb"),
		SSLMode:     get("SSL_MODE", "disable"),
		CORSOrigins: get("CORS_ORIGINS", "*"),
	}
}
func get(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
