package service

import (
	"fmt"
	"microblog/internal/models"
	"microblog/internal/storage"
	"regexp"
	"strings"
)

type PostService struct {
	storage storage.Storage
}

type UserService struct {
	storage storage.Storage
}

func NewUserService(storage storage.Storage) *UserService {
	return &UserService{storage: storage}
}

func NewPostService(storage storage.Storage) *PostService {
	return &PostService{storage: storage}
}

func (s *PostService) CreatePost(author, content string) (*models.Post, error) {
	author = strings.TrimSpace(author)
	if author == "" {
		return nil, fmt.Errorf("author is required")
	}
	if len(author) < 2 {
		return nil, fmt.Errorf("the author cannot be less than 2 characters")
	}
	if len(author) > 50 {
		return nil, fmt.Errorf("author name too long (max 50 characters)")
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}
	if len(content) < 1 {
		return nil, fmt.Errorf("content too short")
	}
	if len(content) > 2000 {
		return nil, fmt.Errorf("content too long (max 2000 characters)")
	}

	post := models.Post{
		Author:  author,
		Content: content,
		Like:    0,
	}

	err := s.storage.CreatePost(post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	createdPost, err := s.storage.GetPostById(post.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created post: %w", err)
	}

	return createdPost, nil
}

func (s *PostService) GetPostById(id int) (*models.Post, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid post ID")
	}

	post, err := s.storage.GetPostById(id)
	if err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}

	return post, nil
}

func (s *PostService) GetAllPosts() ([]models.Post, error) {
	posts, err := s.storage.GetPosts()
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}

	return posts, nil
}

func (s *PostService) LikePost(postID int) (*models.Post, error) {
	if postID < 0 {
		return nil, fmt.Errorf("invalid post ID")
	}

	_, err := s.storage.GetPostById(postID)
	if err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}

	err = s.storage.LikePost(postID)
	if err != nil {
		return nil, fmt.Errorf("failed to like post: %w", err)
	}

	updatedPost, err := s.storage.GetPostById(postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated post: %w", err)
	}

	return updatedPost, nil
}

func (s *UserService) RegisterUser(username, email string) (*models.User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if len(username) < 2 {
		return nil, fmt.Errorf("username must be at least 2 characters")
	}
	if len(username) > 50 {
		return nil, fmt.Errorf("username too long (max 50 characters)")
	}

	email = strings.TrimSpace(email)
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return nil, fmt.Errorf("invalid email format")
	}

	existingUser, err := s.storage.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("error checking email: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("email already registered")
	}

	user := models.User{
		Username: username,
		Email:    email,
	}

	err = s.storage.CreateUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	createdUser, err := s.storage.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get created user: %w", err)
	}

	return createdUser, nil

}

func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	user, err := s.storage.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}
