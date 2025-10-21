package handlers

import (
	"encoding/json"
	"net/http"
)

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Error! Only POST", http.StatusMethodNotAllowed)
		return
	}

	if PostService == nil {
		http.Error(w, "Service not initialized", http.StatusInternalServerError)
		return
	}

	author := r.FormValue("author")
	content := r.FormValue("content")

	post, err := PostService.CreatePost(author, content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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
