package handlers

import (
	"microblog/internal/models"
	"net/http"
	"strconv"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Error! Only POST", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	username := r.FormValue("username")
	if email == "" || username == "" {
		http.Error(w, "Email and username are required", http.StatusBadRequest)
		return
	}

	user := models.User{
		Username: username,
		Email:    email,
	}

	if Store == nil {
		http.Error(w, "Error! Store is nil", http.StatusInternalServerError)
		return
	}

	Store.CreateUser(user)
	w.Write([]byte("User created: " + strconv.Itoa(user.ID)))

}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Login endpoint - to be implemented"))
}
