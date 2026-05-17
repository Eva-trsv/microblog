package main

import (
	"context"
	"fmt"
	"microblog/services/api/internal/config"
	"microblog/services/api/internal/events"
	"microblog/services/api/internal/handlers"
	"microblog/services/api/internal/logger"
	"microblog/services/api/internal/queue"
	apiRepository "microblog/services/api/internal/repository"
	"microblog/services/api/internal/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	_ "net/http/pprof"
)

const (
	ServerPort    = ":8080"
	timeDb        = 5
	PprofPort     = ":6060"
	chLikeService = 1000
	timeServe = 10
)

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

	producer := events.NewKafkaProducer(
		cfg.KafkaBrokers,
		events.TopicConfig{
			UserRegistered: cfg.KafkaTopicUserRegistered,
			PostCreated:    cfg.KafkaTopicPostCreated,
			PostLiked:      cfg.KafkaTopicPostLiked,
		},
		cfg.KafkaWriteTimeout,
		cfg.KafkaRetries,
	)
	defer producer.Close()

	userRepo := apiRepository.NewUserRepository(dbPool)
	postRepo := apiRepository.NewPostRepository(dbPool)
	likeRepo := apiRepository.NewLikeRepository(dbPool)

	userService := service.NewUserService(userRepo, log, producer)
	postService := service.NewPostService(postRepo, log, producer)

	likeService := queue.NewLikeService(likeRepo, chLikeService)
	postService.SetLikeService(likeService)
	likeService.StartWorker()
	defer func() {
		log.Log("like_worker_stopping", nil)
		likeService.StopWorker()
	}()

	handlers.SetupRoutes(userService, postService, log)
	log.Log("routes_initialized", nil)

	go func() {
		fmt.Println("Pprof available at :6060/debug/pprof/")
		if err := http.ListenAndServe(PprofPort, nil); err != nil {
			log.Log("pprof_error", map[string]any{"error": err.Error()})
		}
	}()

	server := &http.Server{Addr: ServerPort}
	serverErr := make(chan error, 1)
	go func() {
		fmt.Println("API server listening on", ServerPort)
		serverErr <- server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case <-stop:
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), timeServe*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Log("server_shutdown_error", map[string]any{"error": err.Error()})
		}
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Log("server_error", map[string]any{"error": err.Error()})
		}
	}

	log.Log("server_stopped", nil)
}
