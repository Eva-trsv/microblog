package service

import (
	"fmt"
	"microblog/internal/logger"
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
	log         *logger.Logger
}

type UserService struct {
	storage UserStorage
	log     *logger.Logger
}

func NewUserService(storage UserStorage, log *logger.Logger) *UserService {
	return &UserService{
		storage: storage,
		log:     log,
	}
}

func NewPostService(storage PostStorage, log *logger.Logger) *PostService {
	return &PostService{
		storage: storage,
		log:     log,
	}
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
		s.log.Log("post_create_error", map[string]any{"error": err.Error(), "author": author})
		return nil, fmt.Errorf("failed to create post: %w", err)
	}
	s.log.Log("post_created", map[string]any{"post_id": createdPost.ID, "author": author})
	return createdPost, nil
}

func (s *PostService) GetPostById(id int) (*models.Post, error) {
	if id <= 0 {
		err := fmt.Errorf("invalid post ID")
		s.log.Log("post_get_error", map[string]any{"error": err.Error(), "post_id": id})
		return nil, err
	}

	post, err := s.storage.GetPostById(id)
	if err != nil {
		s.log.Log("post_get_error", map[string]any{"error": err.Error(), "post_id": id})
		return nil, fmt.Errorf("post not found: %w", err)
	}

	s.log.Log("post_fetched", map[string]any{"post_id": id})
	return post, nil
}

func (s *PostService) GetAllPosts() ([]models.Post, error) {
	posts, err := s.storage.GetPosts()
	if err != nil {
		s.log.Log("posts_get_error", map[string]any{"error": err.Error()})
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}

	s.log.Log("posts_fetched", map[string]any{"count": len(posts)})
	return posts, nil
}

func (s *PostService) LikePost(postID int) (string, error) {
	if postID <= 0 {
		err := fmt.Errorf("invalid post ID")
		s.log.Log("like_error", map[string]any{"error": err.Error(), "post_id": postID})
		return "", fmt.Errorf("invalid post ID")
	}

	_, err := s.storage.GetPostById(postID)
	if err != nil {
		s.log.Log("like_error", map[string]any{"error": err.Error(), "post_id": postID})
		return "", fmt.Errorf("post not found: %w", err)
	}

	if s.likeService == nil {
		err := fmt.Errorf("like service not configured")
		s.log.Log("like_error", map[string]any{"error": err.Error()})
		return "", err
	}

	s.likeService.EnqueueLike(postID)
	s.log.Log("like_queued", map[string]any{"post_id": postID})
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
		err := fmt.Errorf("invalid email format")
		s.log.Log("user_register_error", map[string]any{"error": err.Error(), "email": email})
		return nil, err
	}

	existingUser, err := s.storage.GetUserByEmail(email)
	if err != nil {
		s.log.Log("user_register_error", map[string]any{"error": err.Error(), "email": email})
		return nil, fmt.Errorf("error checking email: %w", err)
	}
	if existingUser != nil {
		s.log.Log("user_register_error", map[string]any{
			"reason": "email already registered",
			"email":  email,
		})
		return nil, fmt.Errorf("email already registered")
	}

	user := models.User{
		Username: username,
		Email:    email,
	}

	err = s.storage.CreateUser(user)
	if err != nil {
		s.log.Log("user_register_error", map[string]any{"error": err.Error(), "username": username})
	}

	createdUser, err := s.storage.GetUserByEmail(email)
	if err != nil {
		s.log.Log("user_register_error", map[string]any{"error": err.Error(), "emai": username})
		return nil, fmt.Errorf("failed to get created user: %w", err)
	}

	return createdUser, nil

}

func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		err := fmt.Errorf("email is required")
		s.log.Log("user_get_error", map[string]any{"error": err.Error()})
		return nil, err
	}

	user, err := s.storage.GetUserByEmail(email)
	if err != nil {
		s.log.Log("user_get_error", map[string]any{"error": err.Error(), "email": email})
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	if user == nil {
		err := fmt.Errorf("user not found")
		s.log.Log("user_get_error", map[string]any{"error": err.Error(), "email": email})
		return nil, err
	}

	s.log.Log("user_fetched", map[string]any{"email": email})
	return user, nil
}
