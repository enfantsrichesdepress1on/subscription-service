package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"subscription-service/internal/config"
	"subscription-service/internal/db"
	"subscription-service/internal/handler"
	"subscription-service/internal/logger"
	"subscription-service/internal/repository"
	"subscription-service/internal/service"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log, err := logger.New(cfg.Log.Level)
	if err != nil {
		panic(err)
	}
	defer log.Sync()
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatal("db connect failed", zap.Error(err))
	}
	defer pool.Close()
	repo := repository.NewSubscriptionRepository(pool)
	svc := service.NewSubscriptionService(repo)
	h := handler.New(svc, log)
	server := &http.Server{Addr: ":" + cfg.HTTP.Port, Handler: h.Routes(), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Info("server started", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server failed", zap.Error(err))
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown failed", zap.Error(err))
	}
}
