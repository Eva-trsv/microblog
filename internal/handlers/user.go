package handlers

import (
	"encoding/json"
	"microblog/internal/service"
	"net/http"
)

type UserHandlers struct {
	userService *service.UserService
}

func NewUserHandlers(userService *service.UserService) *UserHandlers {
	if userService == nil {
		panic("UserHandlers: userService cannot be nil")
	}
	return &UserHandlers{userService: userService}
}

func (u *UserHandlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Error! Only POST", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	username := r.FormValue("username")

	user, err := u.userService.RegisterUser(username, email)
	if err != nil {
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
