package handlers

import (
	"encoding/json"
	"io"
	"microblog/internal/logger"
	"microblog/internal/service"
	"net/http"
	"strconv"
	"strings"
)

type PostHandlers struct {
	postService *service.PostService
	log         *logger.Logger
}

func NewPostHandlers(postService *service.PostService, log *logger.Logger) *PostHandlers {
	if postService == nil {
		panic("PostHandlers: postService cannot be nil")
	}
	if log == nil {
		panic("PostHandlers: log cannot be nil")
	}
	return &PostHandlers{
		postService: postService,
		log:         log,
	}
}

func (p *PostHandlers) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		p.log.Log("http_method_not_allowed", map[string]any{
			"method": r.Method,
		})
		http.Error(w, "Error! Only POST", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		p.log.Log("http_request_body_read_error", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var request struct {
		AuthorID int    `json:"author"`
		Content  string `json:"content"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		p.log.Log("http_invalid_json", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	post, err := p.postService.CreatePost(request.AuthorID, request.Content)
	if err != nil {
		p.log.Log("create_post_failed", map[string]any{
			"author": request.AuthorID,
			"error":  err.Error(),
		})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p.log.Log("http_post_created", map[string]any{
		"post_id":   post.ID,
		"author_id": post.AuthorID,
	})

	response := map[string]interface{}{
		"message":   "The post created",
		"post_id":   post.ID,
		"author_id": post.AuthorID,
		"content":   post.Content,
		"likes":     post.LikeCount,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		p.log.Log("http_response_marshal_error", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)

}

func (p *PostHandlers) LikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		p.log.Log("http_method_not_allowed", map[string]any{
			"method": r.Method,
		})
		http.Error(w, "Error! Only POST", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) != 3 {
		p.log.Log("http_invalid_like_path", map[string]any{
			"path": path,
		})
		http.Error(w, "Invalid path. Use /like/{id}", http.StatusBadRequest)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		p.log.Log("http_invalid_user_id", map[string]any{"value": userIDStr})
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(parts[2])
	if err != nil || postID <= 0 {
		p.log.Log("http_invalid_like_id", map[string]any{
			"value": parts[2],
		})
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := p.postService.GetPostByID(postID)
	if err != nil {
		p.log.Log("get_post_failed", map[string]any{
			"post_id": postID,
			"error":   err.Error(),
		})
		http.Error(w, "Failed to get post data", http.StatusInternalServerError)
		return
	}

	msg, err := p.postService.LikePost(postID, userID)
	if err != nil {
		p.log.Log("like_failed", map[string]any{
			"post_id": postID,
			"error":   err.Error(),
		})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p.log.Log("like_queued", map[string]any{
		"post_id": postID,
	})

	response := map[string]interface{}{
		"message":   msg,
		"post_id":   postID,
		"likes":     post.LikeCount,
		"author_id": post.AuthorID,
		"content":   post.Content,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		p.log.Log("http_response_marshal_error", map[string]any{
			"post_id": postID,
			"error":   err.Error(),
		})
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

	p.log.Log("http_like_response_sent", map[string]any{
		"post_id": postID,
		"likes":   post.LikeCount,
	})
}

func (p *PostHandlers) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		p.log.Log("http_method_not_allowed", map[string]any{
			"method": r.Method,
		})
		http.Error(w, "Error! Only DELETE", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		p.log.Log("http_missing_post_id", nil)
		http.Error(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(idStr)
	if err != nil || postID <= 0 {
		p.log.Log("http_invalid_post_id", map[string]any{
			"value": idStr,
		})
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	err = p.postService.DeletePost(postID)
	if err != nil {
		p.log.Log("delete_post_failed", map[string]any{
			"post_id": postID,
			"error":   err.Error(),
		})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p.log.Log("post_deleted", map[string]any{
		"post_id": postID,
	})

	response := map[string]interface{}{
		"message": "delete",
		"post_id": postID,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		p.log.Log("http_response_marshal_error", map[string]any{
			"post_id": postID,
			"error":   err.Error(),
		})
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func (p *PostHandlers) GetByAuthorIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		p.log.Log("http_method_not_allowed", map[string]any{
			"method": r.Method,
		})
		http.Error(w, "Error! Only GET", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) != 4 || parts[1] != "authors" || parts[3] != "posts" {
		p.log.Log("http_invalid_author_posts_path", map[string]any{
			"path": path,
		})
		http.Error(w, "Invalid path. Use /authors/{id}/posts", http.StatusBadRequest)
		return
	}

	authorID, err := strconv.Atoi(parts[2])
	if err != nil || authorID <= 0 {
		p.log.Log("http_invalid_author_id", map[string]any{
			"value": parts[2],
		})
		http.Error(w, "Invalid author ID", http.StatusBadRequest)
		return
	}

	posts, err := p.postService.GetPostsByAuthorID(authorID)
	if err != nil {
		p.log.Log("get_posts_by_author_failed", map[string]any{
			"author_id": authorID,
			"error":     err.Error(),
		})
		http.Error(w, "Failed to get posts", http.StatusInternalServerError)
		return
	}

	p.log.Log("posts_by_author_fetched", map[string]any{
		"author_id": authorID,
		"count":     len(posts),
	})

	response := map[string]interface{}{
		"message":   "Posts retrieved successfully",
		"author_id": authorID,
		"count":     len(posts),
		"posts":     posts,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		p.log.Log("http_response_marshal_error", map[string]any{
			"author_id": authorID,
			"error":     err.Error(),
		})
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

}

func (p *PostHandlers) GetPostByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		p.log.Log("http_method_not_allowed", map[string]any{
			"method": r.Method,
		})
		http.Error(w, "Error! Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := p.postService.GetPostByID(postID)
	if err != nil {
		p.log.Log("get_post_failed", map[string]any{
			"post_id": postID,
			"error":   err.Error(),
		})
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"post_id":    post.ID,
		"author_id":  post.AuthorID,
		"content":    post.Content,
		"like_count": post.LikeCount,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		p.log.Log("http_response_marshal_error", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
