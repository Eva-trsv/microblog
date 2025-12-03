package main

import (
	"fmt"
	"microblog/internal/handlers"
	"microblog/internal/service"
	"microblog/internal/storage"
	"net/http"
	"microblog/internal/queue"
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

	err := http.ListenAndServe(ServerPort, nil)
	if err != nil {
		fmt.Println("Произошла ошибка:", err.Error())
	}
}
