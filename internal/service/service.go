package service

import (
	"fmt"
	"microblog/internal/models"
	"microblog/internal/queue"
	"regexp"
	"strings"
)

const (
	MinAuthorLength   = 2
	MaxAuthorLength   = 50
	MinContentLength  = 1
	MaxContentLength  = 2000
	MinUsernameLength = 2
	MaxUsernameLength = 50
)

type PostStorage interface {
	CreatePost(post models.Post) (*models.Post, error)
	GetPosts() ([]models.Post, error)
	GetPostById(id int) (*models.Post, error)
	LikePost(postID int) error
}

type UserStorage interface {
	CreateUser(user models.User) error
	GetUserByEmail(email string) (*models.User, error)
}

type PostService struct {
	storage     PostStorage
	likeService *queue.LikeService
}

type UserService struct {
	storage UserStorage
}

func NewUserService(storage UserStorage) *UserService {
	return &UserService{storage: storage}
}

func NewPostService(storage PostStorage) *PostService {
	return &PostService{storage: storage}
}

func (s *PostService) SetLikeService(likeService *queue.LikeService) {
	s.likeService = likeService
}

func (s *PostService) CreatePost(author, content string) (*models.Post, error) {
	author = strings.TrimSpace(author)
	if author == "" {
		return nil, fmt.Errorf("author is required")
	}
	if len(author) < MinAuthorLength {
		return nil, fmt.Errorf("author name must be at least %d characters", MinAuthorLength)
	}
	if len(author) > MaxAuthorLength {
		return nil, fmt.Errorf("author name too long (max %d characters)", MaxAuthorLength)
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}
	if len(content) < MinContentLength {
		return nil, fmt.Errorf("content too short")
	}
	if len(content) > MaxContentLength {
		return nil, fmt.Errorf("content too long (max %d characters)", MaxContentLength)
	}

	post := models.Post{
		Author:    author,
		Content:   content,
		LikeCount: 0,
	}

	createdPost, err := s.storage.CreatePost(post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
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

	result := *post
	return &result, nil
}

func (s *PostService) GetAllPosts() ([]models.Post, error) {
	posts, err := s.storage.GetPosts()
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}

	return posts, nil
}

func (s *PostService) LikePost(postID int) (string, error) {
	if postID <= 0 {
		return "", fmt.Errorf("invalid post ID")
	}

	_, err := s.storage.GetPostById(postID)
	if err != nil {
		return "", fmt.Errorf("post not found: %w", err)
	}

	if s.likeService == nil {
		return "", fmt.Errorf("like service not configured")
	}

	s.likeService.EnqueueLike(postID)
	return "like queued", nil
}

// USER SERVICE

func (s *UserService) RegisterUser(username, email string) (*models.User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if len(username) < MinUsernameLength {
		return nil, fmt.Errorf("username must be at least %d characters", MinUsernameLength)
	}
	if len(username) > MaxUsernameLength {
		return nil, fmt.Errorf("username too long (max %d characters)", MaxUsernameLength)
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
