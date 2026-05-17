package consumer

import (
	"context"
	"encoding/json"
	"testing"

	"microblog/services/engagement/logger"
)

type fakePostLikeHandler struct {
	events     []string
	likeEvents []handledPostLike
}

type handledPostLike struct {
	eventID string
	postID  int
}

func (f *fakePostLikeHandler) HandleEvent(eventID string) error {
	f.events = append(f.events, eventID)
	return nil
}

func (f *fakePostLikeHandler) HandlePostLiked(eventID string, postID int) error {
	f.likeEvents = append(f.likeEvents, handledPostLike{
		eventID: eventID,
		postID:  postID,
	})
	return nil
}

func TestHandleMessageProcessesKnownEvents(t *testing.T) {
	log := logger.NewLogger(10)
	defer log.Close()

	handler := &fakePostLikeHandler{}
	consumer := &Consumer{
		service: handler,
		log:     log,
	}

	userRegistered := mustJSON(t, Event{
		EventID:   "event-registered",
		EventType: eventTypeUserRegistered,
		Payload:   mustJSON(t, UserRegisteredPayload{UserID: 7, Email: "user@example.com"}),
	})
	if err := consumer.HandleMessage(context.Background(), "user_registered", userRegistered); err != nil {
		t.Fatalf("HandleMessage(UserRegistered) returned error: %v", err)
	}
	if len(handler.events) != 1 {
		t.Fatalf("expected 1 handled event, got %d", len(handler.events))
	}
	if handler.events[0] != "event-registered" {
		t.Fatalf("unexpected handled event: %s", handler.events[0])
	}

	postCreated := mustJSON(t, Event{
		EventID:   "event-created",
		EventType: "PostCreated",
		Payload:   mustJSON(t, map[string]any{"post_id": 10}),
	})
	if err := consumer.HandleMessage(context.Background(), "post_created", postCreated); err != nil {
		t.Fatalf("HandleMessage(PostCreated) returned error: %v", err)
	}
	if len(handler.events) != 2 {
		t.Fatalf("expected 1 handled event, got %d", len(handler.events))
	}
	if handler.events[1] != "event-created" {
		t.Fatalf("unexpected handled event: %s", handler.events[1])
	}

	postLiked := mustJSON(t, Event{
		EventID:   "event-liked",
		EventType: eventTypePostLiked,
		Payload:   mustJSON(t, PostLikedPayload{PostID: 10, UserID: 20}),
	})
	if err := consumer.HandleMessage(context.Background(), "post_liked", postLiked); err != nil {
		t.Fatalf("HandleMessage(PostLiked) returned error: %v", err)
	}
	if len(handler.likeEvents) != 1 {
		t.Fatalf("expected 1 handled like event, got %d", len(handler.likeEvents))
	}
	if handler.likeEvents[0].eventID != "event-liked" || handler.likeEvents[0].postID != 10 {
		t.Fatalf("unexpected handled like event: %+v", handler.likeEvents[0])
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}
	return data
}
