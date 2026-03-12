package main

import (
	"fmt"
	"microblog/internal/handlers"
	"microblog/internal/logger"
	"microblog/internal/queue"
	"microblog/internal/repository"
	"microblog/internal/service"
	"net/http"
	"os"
	"github.com/jackc/pgx/v5/pgxpool"
	"context"

	_ "net/http/pprof"
)

const (
	ServerPort = ":8080"
)

func main() {
	log := logger.NewLogger(500)
	defer log.Close()

	log.Log("app_start", map[string]any{
		"version": "1.0",
		"port":    ServerPort,
	})

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Log("db_error", map[string]any{
			"error": "DATABASE_URL not set",
		})
		panic("DATABASE_URL not set")
	}

	ctx := context.Background()

	dbPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Log("db_connection_failed", map[string]any{
			"error": err.Error(),
		})
		panic(err)
	}
	defer dbPool.Close()

	log.Log("db_connected", nil)

	userRepo := repository.NewUserRepository(dbPool)
	postRepo := repository.NewPostRepository(dbPool)
	likeRepo := repository.NewLikeRepository(dbPool)

	// storage := storage.NewObjectStorage()
	// log.Log("storage_init", nil)

	userService := service.NewUserService(userRepo, log)
	log.Log("user_service_init", nil)

	postService := service.NewPostService(postRepo, log)
	log.Log("post_service_init", nil)

	likeService := queue.NewLikeService(likeRepo, 1000)
	log.Log("like_service_init", map[string]any{
		"buffer": 1000,
	})

	postService.SetLikeService(likeService)

	likeService.StartWorker()
	log.Log("like_worker_started", nil)

	defer func() {
		log.Log("like_worker_stopping", nil)
		likeService.StopWorker()
	}()

	handlers.SetupRoutes(userService, postService, log)
	log.Log("routes_initialized", nil)

	fmt.Println("Запускаю сервер")

	go func() {
		fmt.Println("Pprof будет доступен по адресу :6060/debug/pprof/")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Log("pprof_error", map[string]any{
				"error": err.Error(),
			})
			fmt.Println("Произошла ошибка:", err.Error())
		}
	}()

	err = http.ListenAndServe(ServerPort, nil)
	if err != nil {
		log.Log("server_error", map[string]any{
			"error": err.Error(),
		})
		fmt.Println("Произошла ошибка:", err.Error())
	}

	log.Log("server_stopped", nil)
}
