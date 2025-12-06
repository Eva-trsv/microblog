package main

import (
	"fmt"
	"microblog/internal/handlers"
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
	storage := storage.NewObjectStorage()

	userService := service.NewUserService(storage)
	postService := service.NewPostService(storage)
	likeService := queue.NewLikeService(storage, 1000)

	postService.SetLikeService(likeService)

	likeService.StartWorker()
	defer likeService.StopWorker()

	handlers.SetupRoutes(userService, postService)

	fmt.Println("Запускаю сервер")

	go func() {
		fmt.Println("Pprof будет доступен по адресу :6060/debug/pprof/")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			fmt.Println("Произошла ошибка:", err.Error())
		}
	}()

	err := http.ListenAndServe(ServerPort, nil)
	if err != nil {
		fmt.Println("Произошла ошибка:", err.Error())
	}
}
