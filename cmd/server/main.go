package main

import (
	"fmt"
	"microblog/internal/handlers"
	"microblog/internal/logger"
	"microblog/internal/queue"
	"microblog/internal/repository"
	"microblog/internal/service"
	"microblog/internal/config"
	"net/http"
	"github.com/jackc/pgx/v5/pgxpool"
	"context"
	"time"

	_ "net/http/pprof"
)

const (
	ServerPort = ":8080"
	timeDb = 5
	PprofPort = ":6060"	//время подключения к бд
	chLikeService = 1000
)

func main() {


	log := logger.NewLogger(500)
	defer log.Close()

	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Log("db_error", map[string]any{
			"error": "DATABASE_URL not set",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeDb * time.Second)
	defer cancel()

	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
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


	userService := service.NewUserService(userRepo, log)
	log.Log("user_service_init", nil)

	

	likeService := queue.NewLikeService(likeRepo, chLikeService)
	log.Log("like_service_init", map[string]any{
		"buffer": 1000,
	})
	
	postService := service.NewPostService(postRepo, log)
	log.Log("post_service_init", nil)
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
		if err := http.ListenAndServe(PprofPort, nil); err != nil {
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
