package main

import (
	"fmt"
	"microblog/internal/handlers"
	"microblog/internal/storage"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("The initial page. Начальная страница"))
}

func main() {
	var store = storage.NewObjectStorage()
	handlers.InitHandlers(store)

	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/posts/like", handlers.LikeHandler)
	http.HandleFunc("/posts", handlers.CreatePostHandler)
	http.HandleFunc("/posts/", handlers.GetPostHandler)

	fmt.Println("Запускаю сервер")

	err := http.ListenAndServe(":9091", nil)
	if err != nil {
		fmt.Println("Произошла ошибка:", err.Error())
	}
}
