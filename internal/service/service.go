package service

import (
	"context"
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
	Create(ctx context.Context, post *models.Post) error
	GetByAuthorID(ctx context.Context, authorID int) ([]*models.Post, error)
	Delete(ctx context.Context, postID int) error
	GetPostByID(ctx context.Context, postID int) (*models.Post, error)
}

type UserStorage interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id int) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

type LikeStorage interface {
	AddLike(ctx context.Context, userID, postID int) error
	CountLikes(ctx context.Context, postID int) (int, error)
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

func (s *PostService) CreatePost(author_id int, content string) (*models.Post, error) {
	if author_id <= 0 {
		return nil, fmt.Errorf("invalid author id")
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

	post := &models.Post{
		AuthorID:  author_id,
		Content:   content,
		LikeCount: 0,
	}

	err := s.storage.Create(context.Background(), post)
	if err != nil {
		s.log.Log("post_create_error", map[string]any{"error": err.Error(), "author_id": author_id})
		return nil, fmt.Errorf("failed to create post: %w", err)
	}
	s.log.Log("post_created", map[string]any{"post_id": post.ID, "author_id": author_id})
	return post, nil
}

func (s *PostService) GetPostsByAuthorID(authorID int) ([]*models.Post, error) {
	posts, err := s.storage.GetByAuthorID(context.Background(), authorID)
	if err != nil {
		s.log.Log("posts_get_error", map[string]any{
			"error":     err.Error(),
			"author_id": authorID,
		})
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}

	s.log.Log("posts_fetched", map[string]any{"count": len(posts), "author_id": authorID})
	return posts, nil
}

func (s *PostService) GetPostByID(postID int) (*models.Post, error) {
	post, err := s.storage.GetPostByID(context.Background(), postID)
	if err != nil {
		s.log.Log("posts_get_error", map[string]any{"error": err.Error(), "post_id": postID})
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}

	s.log.Log("posts_fetched", map[string]any{
		"author_id": post.AuthorID,
		"post_id":   postID,
	})
	return post, nil
}

func (s *PostService) DeletePost(postID int) error {
	if postID <= 0 {
		return fmt.Errorf("invalid post ID")
	}

	err := s.storage.Delete(context.Background(), postID)
	if err != nil {
		s.log.Log("post_delete_error", map[string]any{"error": err.Error(), "post_id": postID})
		return fmt.Errorf("failed to delete post: %w", err)
	}

	s.log.Log("post_deleted", map[string]any{"post_id": postID})
	return nil
}

func (s *PostService) LikePost(userID, postID int) (string, error) {
	if postID <= 0 || userID <= 0 {
		err := fmt.Errorf("invalid post ID")
		s.log.Log("like_error", map[string]any{"error": err.Error(), "post_id": postID, "user_id": userID})
		return "", fmt.Errorf("invalid post ID")
	}

	if s.likeService == nil {
		err := fmt.Errorf("like service not configured")
		s.log.Log("like_error", map[string]any{"error": err.Error()})
		return "", err
	}

	s.likeService.EnqueueLike(userID, postID)
	s.log.Log("like_queued", map[string]any{"post_id": postID, "user_id": userID})
	return "like queued", nil
}

// USER SERVICE

func (s *UserService) CreateUser(username, email string) (*models.User, error) {
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

	existingUser, err := s.storage.GetByEmail(context.Background(), email)
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

	user := &models.User{
		Username: username,
		Email:    email,
	}

	err = s.storage.Create(context.Background(), user)
	if err != nil {
		s.log.Log("user_register_error", map[string]any{"error": err.Error(), "username": username})
		return nil, err
	}

	createdUser, err := s.storage.GetByEmail(context.Background(), email)
	if err != nil {
		s.log.Log("user_register_error", map[string]any{"error": err.Error(), "emai": email})
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

	user, err := s.storage.GetByEmail(context.Background(), email)
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

func (s *UserService) GetUserByID(id int) (*models.User, error) {
	if id <= 0 {
		err := fmt.Errorf("invalid user id")
		s.log.Log("user_get_error", map[string]any{"error": err.Error(), "id": id})
		return nil, err
	}

	user, err := s.storage.GetByID(context.Background(), id)
	if err != nil {
		s.log.Log("user_get_error", map[string]any{
			"error": err.Error(),
			"id":    id,
		})
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	if user == nil {
		err := fmt.Errorf("user not found")
		s.log.Log("user_get_error", map[string]any{
			"error": err.Error(),
			"id":    id,
		})
		return nil, err
	}

	s.log.Log("user_fetched", map[string]any{"id": id})
	return user, nil
}
