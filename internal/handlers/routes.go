package handlers

import (
	"microblog/internal/service"
	"net/http"
)

func SetupRoutes(userService *service.UserService, postService *service.PostService) {
	userHandlers := NewUserHandlers(userService)
	postHandlers := NewPostHandlers(postService)

	http.HandleFunc("/register", userHandlers.RegisterHandler)
	http.HandleFunc("/login", userHandlers.LoginHandler)
	http.HandleFunc("/posts", postHandlers.GetPostHandler)
	http.HandleFunc("/posts/", postHandlers.GetPostHandler)
	http.HandleFunc("/posts/create", postHandlers.CreatePostHandler)
	http.HandleFunc("/like", postHandlers.LikeHandler)
}
