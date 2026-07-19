package events

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type TopicConfig struct {
	UserRegistered string
	PostCreated    string
	PostLiked      string
}

type KafkaProducer struct {
	writer  *kafka.Writer
	topics  TopicConfig
	timeout time.Duration
	retries int
}

func NewKafkaProducer(brokers []string, topics TopicConfig, timeout time.Duration, retries int) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireAll,
		},
		topics:  topics,
		timeout: timeout,
		retries: retries,
	}
}

func (p *KafkaProducer) Publish(ctx context.Context, topic string, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	message := kafka.Message{
		Topic: p.topicName(topic),
		Key:   []byte(eventKey(event)),
		Value: data,
		Time:  event.OccurredAt,
	}

	var lastErr error
	for attempt := 0; 
		attempt < p.retries; 
		attempt++ {
			writeCtx, cancel := context.WithTimeout(ctx, p.timeout)
			lastErr = p.writer.WriteMessages(writeCtx, message)
			cancel()
			if lastErr == nil {
			return nil
			}	
		time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
	}

	return fmt.Errorf("publish kafka event %s to %s: %w", event.EventID, message.Topic, lastErr)
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}

func (p *KafkaProducer) topicName(topic string) string {
	switch topic {
	case TopicUserRegistered:
		return p.topics.UserRegistered
	case TopicPostCreated:
		return p.topics.PostCreated
	case TopicPostLiked:
		return p.topics.PostLiked
	default:
		return topic
	}
}

func eventKey(event Event) string {
	switch payload := event.Payload.(type) {
	case UserRegisteredPayload:
		return strconv.Itoa(payload.UserID)
	case PostCreatedPayload:
		return strconv.Itoa(payload.PostID)
	case PostLikedPayload:
		return strconv.Itoa(payload.PostID)
	default:
		return event.EventID
	}
}
