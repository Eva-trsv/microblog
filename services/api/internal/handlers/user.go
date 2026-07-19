package handlers

import (
	"encoding/json"
	"io"
	"microblog/services/api/internal/logger"
	"microblog/services/api/internal/service"
	"net/http"
	"strconv"
)

type UserHandlers struct {
	userService *service.UserService
	log         *logger.Logger
}

func NewUserHandlers(userService *service.UserService, log *logger.Logger) *UserHandlers {
	if userService == nil {
		panic("UserHandlers: userService cannot be nil")
	}
	if log == nil {
		panic("UserHandlers: log cannot be nil")
	}

	return &UserHandlers{
		userService: userService,
		log:         log}
}

func (u *UserHandlers) CreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		u.log.Log("http_method_not_allowed", map[string]any{
			"method": r.Method,
		})
		http.Error(w, "Error! Only POST", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		u.log.Log("http_request_body_read_error", map[string]any{
			"method": r.Method,
		})
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var request struct {
		Email    string `json:"email"`
		Username string `json:"username"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		u.log.Log("http_invalid_json", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := u.userService.CreateUser(request.Username, request.Email)
	if err != nil {
		u.log.Log("register_user_failed", map[string]any{
			"author": request.Username,
			"error":  err.Error(),
		})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"message":  "User registered successfully",
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		u.log.Log("http_response_marshal_error", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)
}

func (u *UserHandlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Login endpoint - to be implemented"))
}

func (u *UserHandlers) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		u.log.Log("http_method_not_allowed", map[string]any{
			"method": r.Method,
		})
		http.Error(w, "Error! Only GET", http.StatusMethodNotAllowed)
		return
	}

	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	user, err := u.userService.GetUserByID(id)
	if err != nil {
		u.log.Log("get_user_by_id_failed", map[string]any{
			"id":    id,
			"error": err.Error(),
		})
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		u.log.Log("http_response_marshal_error", map[string]any{"error": err.Error()})
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (u *UserHandlers) GetUserByEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		u.log.Log("http_method_not_allowed", map[string]any{
			"method": r.Method,
		})
		http.Error(w, "Error! Only GET", http.StatusMethodNotAllowed)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Missing email parameter", http.StatusBadRequest)
		return
	}

	user, err := u.userService.GetUserByEmail(email)
	if err != nil {
		u.log.Log("get_user_by_email_failed", map[string]any{
			"email": email,
			"error": err.Error(),
		})
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		u.log.Log("http_response_marshal_error", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
