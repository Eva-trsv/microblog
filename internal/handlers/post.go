package handlers

import (
	"encoding/json"
	"io"
	"microblog/internal/service"
	"net/http"
	"strconv"
	"strings"
)

type PostHandlers struct {
	postService *service.PostService
}

func NewPostHandlers(postService *service.PostService) *PostHandlers {
	if postService == nil {
		panic("PostHandlers: postService cannot be nil")
	}
	return &PostHandlers{postService: postService}
}

func (p *PostHandlers) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Error! Only POST", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var request struct {
		Author  string `json:"author"`
		Content string `json:"content"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	post, err := p.postService.CreatePost(request.Author, request.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"message": "The post created",
		"post_id": post.ID,
		"author":  post.Author,
		"content": post.Content,
		"likes":   post.LikeCount,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)

}

func (p *PostHandlers) getAllPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := p.postService.GetAllPosts()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "Posts retrieved successfully",
		"count":   len(posts),
		"posts":   posts,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)

}

func (p *PostHandlers) getPostByID(w http.ResponseWriter, r *http.Request, path string) {
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

	post, err := p.postService.GetPostById(postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"message": "Post retrieved successfully",
		"post":    post,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)
}

func (p *PostHandlers) GetPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Error! Only GET", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path

	if path == "/posts" {
		p.getAllPosts(w, r)
		return
	}

	if strings.HasPrefix(path, "/posts/") {
		p.getPostByID(w, r, path)
		return
	}

	http.Error(w, "Invalid path", http.StatusBadRequest)
}

func (p *PostHandlers) LikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Error! Only POST", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) != 3 {
		http.Error(w, "Invalid path. Use /like/{id}", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(parts[2])
	if err != nil || postID <= 0 {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	msg, err := p.postService.LikePost(postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	post, err := p.postService.GetPostById(postID)
	if err != nil {
		http.Error(w, "Failed to get post data", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": msg,
		"post_id": postID,
		"likes":   post.LikeCount,
		"author":  post.Author,
		"content": post.Content,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
