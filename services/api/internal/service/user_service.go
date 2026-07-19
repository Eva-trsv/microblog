package service

import (
	"context"
	"fmt"
	"microblog/services/api/internal/events"
	"microblog/services/api/internal/logger"
	"microblog/services/api/internal/models"
	"regexp"
	"strings"
)

type UserStorage interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id int) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

type UserService struct {
	storage  UserStorage
	log      *logger.Logger
	producer events.Producer
}

func NewUserService(storage UserStorage, log *logger.Logger, producer ...events.Producer) *UserService {
	service := &UserService{
		storage: storage,
		log:     log,
	}
	if len(producer) > 0 {
		service.producer = producer[0]
	}
	return service
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

	if s.producer != nil {
		event := events.NewUserRegisteredEvent(createdUser.ID, createdUser.Email, "")
		if err := s.producer.Publish(context.Background(), events.TopicUserRegistered, event); err != nil {
			s.log.Log("event_publish_error", map[string]any{
				"error":   err.Error(),
				"user_id": createdUser.ID,
			})
			return nil, err
		}
	}

	return createdUser, nil

}

func (s *UserService) SetProducer(producer events.Producer) {
	s.producer = producer
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
