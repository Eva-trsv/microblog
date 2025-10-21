package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func GetPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Error! Only GET", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path

	if path == "/posts" {
		getAllPosts(w, r)
		return
	}

	if strings.HasPrefix(path, "/posts/") {
		getPostByID(w, r, path)
		return
	}

	http.Error(w, "Invalid path", http.StatusBadRequest)
}

func getAllPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := PostService.GetAllPosts()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Posts retrieved successfully",
		"count":   len(posts),
		"posts":   posts,
	})
}

func getPostByID(w http.ResponseWriter, r *http.Request, path string) {
	idStr := strings.TrimPrefix(path, "/posts/")
	if idStr == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := PostService.GetPostById(postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Post retrieved successfully",
		"post":    post,
	})
}

func LikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Error! Only POST", http.StatusMethodNotAllowed)
		return
	}

	if PostService == nil {
		http.Error(w, "Storage not initialized", http.StatusInternalServerError)
		return
	}

	postIDStr := r.FormValue("post_id")
	if postIDStr == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	_, err = PostService.LikePost(postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	post, err := PostService.GetPostById(postID)
	if err != nil {
		http.Error(w, "Failed to get updated post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Post liked successfully",
		"post_id": postID,
		"likes":   post.Like,
		"author":  post.Author,
		"content": post.Content,
	})
}
