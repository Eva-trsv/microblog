package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"microblog/services/engagement/logger"

	"github.com/segmentio/kafka-go"
)

const (
	eventTypeUserRegistered = "UserRegistered"
	eventTypePostCreated    = "PostCreated"
	eventTypePostLiked      = "PostLiked"
)

type Consumer struct {
	service postLikeHandler
	reader  *kafka.Reader
	log     *logger.Logger
}

type postLikeHandler interface {
	HandleEvent(eventID string) error
	HandlePostLiked(eventID string, postID int) error
}

type Config struct {
	Brokers []string
	GroupID string
	Topics  []string
}

type Event struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
	TraceID   string          `json:"trace_id,omitempty"`
}

type PostLikedPayload struct {
	PostID int `json:"post_id"`
	UserID int `json:"user_id"`
}

type UserRegisteredPayload struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}

type PostCreatedPayload struct {
	PostID int    `json:"post_id"`
	UserID int    `json:"user_id"`
	Title  string `json:"title"`
}

func NewConsumer(service postLikeHandler, log *logger.Logger, cfg Config) *Consumer {
	return &Consumer{
		service: service,
		log:     log,
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     cfg.Brokers,
			GroupID:     cfg.GroupID,
			GroupTopics: cfg.Topics,
			MinBytes:    1,
			MaxBytes:    10e6,
		}),
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	c.log.Log("consumer_started", nil)
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			c.log.Log("consumer_fetch_error", map[string]any{"error": err.Error()})
			continue
		}

		if err := c.HandleMessage(ctx, msg.Topic, msg.Value); err != nil {
			c.log.Log("consumer_handle_error", map[string]any{
				"topic":  msg.Topic,
				"offset": msg.Offset,
				"error":  err.Error(),
			})
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.log.Log("consumer_commit_error", map[string]any{
				"topic":  msg.Topic,
				"offset": msg.Offset,
				"error":  err.Error(),
			})
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func (c *Consumer) HandleMessage(ctx context.Context, topic string, data []byte) error {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}
	if event.EventID == "" {
		return fmt.Errorf("event_id is required")
	}

	switch event.EventType {
	case eventTypeUserRegistered:
		var payload UserRegisteredPayload
		if err := json.Unmarshal(event.Payload, &payload); 
		err != nil {
			return err
		}
		if payload.UserID <= 0 {
			return fmt.Errorf("invalid user_id %s", strconv.Itoa(payload.UserID))
		}
		return c.service.HandleEvent(event.EventID)
	case eventTypePostCreated:
		var payload PostCreatedPayload
		if err := json.Unmarshal(event.Payload, &payload); 
		err != nil {
			return err
		}
		if payload.PostID <= 0 {
			return fmt.Errorf("invalid post_id %s", strconv.Itoa(payload.PostID))
		}
		return c.service.HandleEvent(event.EventID)
	case eventTypePostLiked:
		var payload PostLikedPayload
		if err := json.Unmarshal(event.Payload, &payload); 
		err != nil {
			return err
		}
		if payload.PostID <= 0 {
			return fmt.Errorf("invalid post_id %s", strconv.Itoa(payload.PostID))
		}
		return c.service.HandlePostLiked(event.EventID, payload.PostID)
	default:
		c.log.Log("consumer_event_ignored", map[string]any{
			"topic":      topic,
			"event_type": event.EventType,
			"event_id":   event.EventID,
		})
		return nil
	}
}
