package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

// Config stores all app configuration loaded from environment variables.
type Config struct {
	AppPort         string
	SessionTTL      time.Duration
	SessionTTLInt   int
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	RedisPingTimout time.Duration
}

// MustLoad loads config from env and exits on invalid values.
func MustLoad() Config {
	redisHost := mustGetEnv("REDIS_HOST")
	redisPort := mustGetEnv("REDIS_PORT")
	ttlSeconds := mustGetEnvInt("APP_USER_SESSION_TTL")

	return Config{
		AppPort:         mustGetEnv("APP_PORT"),
		SessionTTL:      time.Duration(ttlSeconds) * time.Second,
		SessionTTLInt:   ttlSeconds,
		RedisAddr:       redisHost + ":" + redisPort,
		RedisPassword:   os.Getenv("REDIS_PASSWORD"),
		RedisDB:         mustGetEnvInt("REDIS_DB"),
		RedisPingTimout: 3 * time.Second,
	}
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s is required", key)
	}
	return value
}

func mustGetEnvInt(key string) int {
	value := mustGetEnv(key)
	n, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("%s must be a valid integer: %v", key, err)
	}
	return n
}
