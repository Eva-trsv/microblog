package events

import "testing"

func TestEventKey(t *testing.T) {
	tests := []struct {
		name  string
		event Event
		want  string
	}{
		{
			name:  "user registered uses user id",
			event: NewUserRegisteredEvent(42, "user@example.com", ""),
			want:  "42",
		},
		{
			name:  "post created uses post id",
			event: NewPostCreatedEvent(7, 42, "hello", ""),
			want:  "7",
		},
		{
			name:  "post liked uses post id",
			event: NewPostLikedEvent(9, 42, ""),
			want:  "9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := eventKey(tt.event); got != tt.want {
				t.Fatalf("eventKey() = %q, want %q", got, tt.want)
			}
		})
	}
}
