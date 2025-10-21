package handlers

import (
	"microblog/internal/storage"
	"microblog/internal/service"
)


var (
	PostService *service.PostService
	UserService *service.UserService
)

func InitHandlers(store storage.Storage) {
	PostService = service.NewPostService(store)
	UserService = service.NewUserService(store)
}
