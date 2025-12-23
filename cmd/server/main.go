package main

import (
	"fmt"
	"microblog/internal/handlers"
	"microblog/internal/logger"
	"microblog/internal/queue"
	"microblog/internal/service"
	"microblog/internal/storage"
	"net/http"

	_ "net/http/pprof"
)

const (
	ServerPort = ":9091"
)

func main() {
	log := logger.NewLogger(500)
	defer log.Close()

	log.Log("app_start", map[string]any{
		"version": "1.0",
		"port":    ServerPort,
	})

	storage := storage.NewObjectStorage()
	log.Log("storage_init", nil)

	userService := service.NewUserService(storage, log)
	log.Log("user_service_init", nil)

	postService := service.NewPostService(storage, log)
	log.Log("post_service_init", nil)

	likeService := queue.NewLikeService(storage, 1000)
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

	err := http.ListenAndServe(ServerPort, nil)
	if err != nil {
		log.Log("server_error", map[string]any{
			"error": err.Error(),
		})
		fmt.Println("Произошла ошибка:", err.Error())
	}

	log.Log("server_stopped", nil)
}
