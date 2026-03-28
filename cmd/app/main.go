package main

import (
	"context"
	"log"
	"net/http"

	goredis "github.com/redis/go-redis/v9"

	"ndbx-app/internal/config"
	redisrepo "ndbx-app/internal/repository/redis"
	sessionsvc "ndbx-app/internal/service/session"
	httptransport "ndbx-app/internal/transport/http"
)

func main() {
	cfg := config.MustLoad()

	redisClient := goredis.NewClient(&goredis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("failed to close redis client: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RedisPingTimout)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis ping failed: %v", err)
	}

	store := redisrepo.NewSessionStore(redisClient)
	service := sessionsvc.NewService(store, sessionsvc.CryptoSIDGenerator{}, sessionsvc.SystemClock{}, cfg.SessionTTL)
	handler := httptransport.NewHandler(service, cfg.SessionTTLInt)

	mux := http.NewServeMux()
	handler.Register(mux)

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: mux,
	}

	log.Printf("listening on :%s", cfg.AppPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
