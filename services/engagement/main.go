package main

import (
	"context"
	"fmt"
	"microblog/services/engagement/config"
	"microblog/services/engagement/consumer"
	"microblog/services/engagement/handlers"
	"microblog/services/engagement/logger"
	"microblog/services/engagement/repository"
	engagementService "microblog/services/engagement/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const timeDb = 5

func main() {
	log := logger.NewLogger(500)
	defer log.Close()

	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Log("db_error", map[string]any{"error": "DATABASE_URL not set"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeDb*time.Second)
	defer cancel()

	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Log("db_connection_failed", map[string]any{"error": err.Error()})
		panic(err)
	}
	defer dbPool.Close()

	repo := repository.NewRepository(dbPool)
	service := engagementService.NewService(repo)
	statsHandler := handlers.NewStatsHandler(service, log)

	consumerCtx, stopConsumer := context.WithCancel(context.Background())
	kafkaConsumer := consumer.NewConsumer(service, log, consumer.Config{
		Brokers: cfg.KafkaBrokers,
		GroupID: cfg.KafkaGroupID,
		Topics: []string{
			cfg.KafkaTopicUserRegistered,
			cfg.KafkaTopicPostCreated,
			cfg.KafkaTopicPostLiked,
		},
	})
	defer kafkaConsumer.Close()

	go func() {
		if err := kafkaConsumer.Start(consumerCtx); err != nil {
			log.Log("consumer_stopped_with_error", map[string]any{"error": err.Error()})
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/stats/posts/", statsHandler.GetStats)
	mux.HandleFunc("/stats/posts", statsHandler.GetStats)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:    cfg.ServerPort,
		Handler: mux,
	}
	serverErr := make(chan error, 1)
	go func() {
		fmt.Println("Engagement server listening on", cfg.ServerPort)
		serverErr <- server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case <-stop:
		stopConsumer()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Log("server_shutdown_error", map[string]any{"error": err.Error()})
		}
	case err := <-serverErr:
		stopConsumer()
		if err != nil && err != http.ErrServerClosed {
			log.Log("server_error", map[string]any{"error": err.Error()})
		}
	}

	log.Log("server_stopped", nil)
}
