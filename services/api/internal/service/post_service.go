package service

import (
	"context"
	"fmt"
	"microblog/services/api/internal/events"
	"microblog/services/api/internal/logger"
	"microblog/services/api/internal/models"
	"microblog/services/api/internal/queue"
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
	GetAll(ctx context.Context) ([]*models.Post, error)
	GetByAuthorID(ctx context.Context, authorID int) ([]*models.Post, error)
	Delete(ctx context.Context, postID int) error
	GetPostByID(ctx context.Context, postID int) (*models.Post, error)
}

type LikeStorage interface {
	AddLike(ctx context.Context, userID, postID int) error
	CountLikes(ctx context.Context, postID int) (int, error)
}

type PostService struct {
	storage     PostStorage
	likeService *queue.LikeService
	log         *logger.Logger
	producer    events.Producer
}

func NewPostService(storage PostStorage, log *logger.Logger, producer ...events.Producer) *PostService {
	service := &PostService{
		storage: storage,
		log:     log,
	}
	if len(producer) > 0 {
		service.producer = producer[0]
	}
	return service
}

func (s *PostService) SetLikeService(likeService *queue.LikeService) {
	s.likeService = likeService
}

func (s *PostService) CreatePost(author_id int, content string, traceID ...string) (*models.Post, error) {
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

	if s.producer != nil { //добавили событие
		event := events.NewPostCreatedEvent(post.ID, author_id, content, optionalTraceID(traceID))
		err = s.producer.Publish(context.Background(), events.TopicPostCreated, event)
		if err != nil {
			s.log.Log("event_publish_error", map[string]any{
				"error":   err.Error(),
				"post_id": post.ID,
			})
			return nil, err
		}
	}

	s.log.Log("post_created", map[string]any{
		"post_id":   post.ID,
		"author_id": author_id})
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

func (s *PostService) GetPosts() ([]*models.Post, error) {
	posts, err := s.storage.GetAll(context.Background())
	if err != nil {
		s.log.Log("posts_get_error", map[string]any{"error": err.Error()})
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}

	s.log.Log("posts_fetched", map[string]any{"count": len(posts)})
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

func (s *PostService) LikePost(postID, userID int) (string, error) {
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

	if s.producer != nil { //добавили событие
		event := events.NewPostLikedEvent(postID, userID, "")
		err := s.producer.Publish(context.Background(), events.TopicPostLiked, event)
		if err != nil {
			s.log.Log("event_publish_error", map[string]any{
				"error":   err.Error(),
				"post_id": postID,
				"user_id": userID,
			})
			return "", err
		}
	}

	s.log.Log("like_queued", map[string]any{"post_id": postID, "user_id": userID})
	return "like queued", nil
}

func (s *PostService) SetProducer(producer events.Producer) {
	s.producer = producer
}

func optionalTraceID(traceID []string) string {
	if len(traceID) == 0 {
		return ""
	}
	return traceID[0]
}
