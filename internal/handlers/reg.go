package handlers

import (
	"encoding/json"
	"net/http"
	
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Error! Only POST", http.StatusMethodNotAllowed)
		return
	}

	if UserService == nil {
		http.Error(w, "Service not initialized", http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	username := r.FormValue("username")
	
	user, err := UserService.RegisterUser(username, email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "User registered successfully",
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}


func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Login endpoint - to be implemented"))
}
