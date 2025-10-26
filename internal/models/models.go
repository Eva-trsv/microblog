package models

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type Post struct {
	ID      int    `json:"id"`
	Author  string `json:"author"`
	Content string `json:"content"`
	LikeCount    int    `json:"like"`
}
