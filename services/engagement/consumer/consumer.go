package consumer

import (
	"context"
	"encoding/json"
	"microblog/services/engagement/service"
)

type Consumer struct {
	service *service.Service
}

func NewConsumer(service *service.Service) *Consumer {
	return &Consumer{
		service: service,
	}
}

// универсальный обработчик события
func (c *Consumer) HandleMessage(ctx context.Context, topic string, data []byte) error {

	switch topic {

	case "post_liked":
		var event struct {
			EventID string `json:"event_id"`
			PostID  int    `json:"post_id"`
		}

		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}

		return c.service.HandlePostLiked(event.EventID, event.PostID)

	default:
		return nil
	}
}