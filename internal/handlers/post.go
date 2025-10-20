package handlers

import (
	"encoding/json"
	"microblog/internal/models"
	"net/http"
)

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Error! Only POST", http.StatusMethodNotAllowed)
		return
	}

	author := r.FormValue("author")
	content := r.FormValue("content")
	if author == "" || content == "" {
		http.Error(w, "Author and content are required", http.StatusBadRequest)
		return
	}

	post := models.Post{
		Author:  author,
		Content: content,
		Like:    0,
	}

	if Store == nil {
		http.Error(w, "Error! Store is nil", http.StatusInternalServerError)
		return
	}

	Store.CreatePost(post)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Post created successfully",
		"post_id": post.ID,
		"author":  post.Author,
		"content": post.Content,
		"likes":   post.Like,
	})
}
