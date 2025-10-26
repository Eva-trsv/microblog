package main

import (
	"fmt"
	"microblog/internal/handlers"
	"microblog/internal/service"
	"microblog/internal/storage"
	"net/http"
)

const (
	ServerPort = ":9091"
)

func main() {
	storage := &storage.ObjectStorage{}

	userService := service.NewUserService(storage)
	postService := service.NewPostService(storage)

	handlers.SetupRoutes(userService, postService)

	fmt.Println("Запускаю сервер")

	err := http.ListenAndServe(ServerPort, nil)
	if err != nil {
		fmt.Println("Произошла ошибка:", err.Error())
	}
}
