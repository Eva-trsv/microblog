package events

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Event — универсальное событие, которое будет отправляться в Kafka
type Event struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	OccurredAt time.Time `json:"occurred_at"`
	Producer   string    `json:"producer"`           // кто отправил (api)
	Payload    any       `json:"payload"`            // данные события
	TraceID    string    `json:"trace_id,omitempty"` //  для логов
}

const (
	EventTypeUserRegistered = "UserRegistered"
	EventTypePostCreated    = "PostCreated"
	EventTypePostLiked      = "PostLiked"

	TopicUserRegistered = "user_registered"
	TopicPostCreated    = "post_created"
	TopicPostLiked      = "post_liked"
)

type UserRegisteredPayload struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}

type PostCreatedPayload struct {
	PostID int    `json:"post_id"`
	UserID int    `json:"user_id"`
	Title  string `json:"title"`
}

type PostLikedPayload struct {
	PostID int `json:"post_id"`
	UserID int `json:"user_id"`
}

func NewUserRegisteredEvent(userID int, email, traceID string) Event {
	return Event{
		EventID:    uuid.New().String(),
		EventType:  EventTypeUserRegistered,
		OccurredAt: time.Now().UTC(),
		Producer:   "api",
		Payload: UserRegisteredPayload{
			UserID: userID,
			Email:  email,
		},
		TraceID: traceID,
	}
}

func NewPostCreatedEvent(postID, userID int, title, traceID string) Event {
	return Event{
		EventID:    uuid.New().String(),
		EventType:  EventTypePostCreated,
		OccurredAt: time.Now().UTC(),
		Producer:   "api",
		Payload: PostCreatedPayload{
			PostID: postID,
			UserID: userID,
			Title:  title,
		},
		TraceID: traceID,
	}
}

func NewPostLikedEvent(postID, userID int, traceID string) Event {
	return Event{
		EventID:    uuid.New().String(),
		EventType:  EventTypePostLiked,
		OccurredAt: time.Now().UTC(),
		Producer:   "api",
		Payload: PostLikedPayload{
			PostID: postID,
			UserID: userID,
		},
		TraceID: traceID,
	}
}

// Producer — интерфейс отправки событий (Kafka реализация будет позже)
type Producer interface {
	Publish(ctx context.Context, topic string, event Event) error
}
