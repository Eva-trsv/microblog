package service_test

import (
	"microblog/internal/logger"
	"microblog/internal/queue"
	"microblog/internal/service"
	"microblog/internal/storage"
	"testing"
	"time"
)

// USER SERVICE

func TestRegisterUser(t *testing.T) {
	storage := storage.NewObjectStorage()

	log := logger.NewLogger(10)
	defer log.Close()

	userService := service.NewUserService(storage, log)

	user, err := userService.RegisterUser("Eva", "testeva@mail.ru")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.ID == 0 {
		t.Error("expected non-zero user ID")
	}

	if user.Username == "" {
		t.Error("the username is empty")
	}

	if user.Username != "Eva" {
		t.Errorf("expected username 'Eva', got %s", user.Username)
	}

	if user.Email == "" {
		t.Error("the email is empty")
	}

	if user.Email != "testeva@mail.ru" {
		t.Errorf("expected email 'testeva@mail.ru', got %s", user.Username)
	}

}

// POST SERVICE

func TestCreatePost(t *testing.T) {
	storage := storage.NewObjectStorage()

	log := logger.NewLogger(10)
	defer log.Close()

	postService := service.NewPostService(storage, log)

	post, err := postService.CreatePost("Eva", "Мой первый пост")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if post.Author == "" {
		t.Error("the author is empty")
	}

	if post.Author != "Eva" {
		t.Errorf("expected username 'Eva', got %s", post.Author)
	}

	if post.Content == "" {
		t.Error("the content is empty")
	}

	if post.Content != "Мой первый пост" {
		t.Errorf("expected content 'Мой первый пост', got %s", post.Content)
	}

	if post.ID == 0 {
		t.Error("expected non-zero post ID")
	}
	if post.LikeCount != 0 {
		t.Error("expected 0 likes initially")
	}

}

func TestGetAllPosts(t *testing.T) {
	storage := storage.NewObjectStorage()

	log := logger.NewLogger(10)
	defer log.Close()

	postService := service.NewPostService(storage, log)

	postService.CreatePost("Eva", "Мой первый пост")
	postService.CreatePost("Alice", "Мой dnjhjq пост")

	posts, err := postService.GetAllPosts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(posts) != 2 {
		t.Errorf("expected 2 posts, got %d", len(posts))
	}
}

func TestGetPostById(t *testing.T) {
	storage := storage.NewObjectStorage()

	log := logger.NewLogger(10)
	defer log.Close()

	postService := service.NewPostService(storage, log)

	post, _ := postService.CreatePost("Eva", "Тестовый пост")

	got, err := postService.GetPostById(post.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != post.ID || got.Author != "Eva" || got.Content != "Тестовый пост" {
		t.Errorf("post data mismatch: got %+v", got)
	}

	_, err = postService.GetPostById(9999)
	if err == nil {
		t.Error("expected error for non-existent post ID")
	}
}

func TestLikePost(t *testing.T) {
	storage := storage.NewObjectStorage()

	log := logger.NewLogger(10)
	defer log.Close()

	postService := service.NewPostService(storage, log)
	likeService := queue.NewLikeService(storage, 1000)

	postService.SetLikeService(likeService)
	likeService.StartWorker()
	defer likeService.StopWorker()

	post, _ := postService.CreatePost("Eva", "Test post")

	msg, err := postService.LikePost(post.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "like queued" {
		t.Errorf("expected message 'like queued', got %s", msg)
	}

	time.Sleep(10 * time.Millisecond)

	updatedPost, _ := postService.GetPostById(post.ID)
	if updatedPost.LikeCount != 1 {
		t.Errorf("expected 1 like, got %d", updatedPost.LikeCount)
	}
}
