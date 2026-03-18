package handlers

import (
	"microblog/internal/logger"
	"microblog/internal/service"
	"net/http"
)

func SetupRoutes(userService *service.UserService, postService *service.PostService, log *logger.Logger) {
	userHandlers := NewUserHandlers(userService, log)
	postHandlers := NewPostHandlers(postService, log)

	http.HandleFunc("/register", userHandlers.CreateHandler)
	http.HandleFunc("/login", userHandlers.LoginHandler)
	http.HandleFunc("/users", userHandlers.GetUserByIDHandler)
	http.HandleFunc("/users/email", userHandlers.GetUserByEmailHandler)
	http.HandleFunc("/authors/", postHandlers.GetByAuthorIDHandler)
	http.HandleFunc("/posts/", postHandlers.DeleteHandler)
	http.HandleFunc("/posts/create", postHandlers.CreatePostHandler)
	http.HandleFunc("/like/", postHandlers.LikeHandler)
	http.HandleFunc("/posts/get", postHandlers.GetPostByIDHandler)
}
