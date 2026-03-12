package service_test

import (
	"microblog/internal/logger"
	"microblog/internal/queue"
	"microblog/internal/service"
	"microblog/internal/storage"
	"strings"
	"testing"
)

func BenchmarkRegisterUser(b *testing.B) {
	storage := storage.NewObjectStorage()

	log := logger.NewLogger(1000)
	defer log.Close()

	userService := service.NewUserService(storage, log)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = userService.RegisterUser("Eva", "testbenc@mail.ru")
	}
}

func BenchmarkCreatePost(b *testing.B) {
	storage := storage.NewObjectStorage()

	log := logger.NewLogger(1000)
	defer log.Close()

	postService := service.NewPostService(storage, log)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = postService.CreatePost(1, "Мой первый bench")
	}

}

func BenchmarkLargeTextCreateAndRead(b *testing.B) {
	storage := storage.NewObjectStorage()

	log := logger.NewLogger(1000)
	defer log.Close()

	postService := service.NewPostService(storage, log)

	largeText := strings.Repeat("A", 2_000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		post, err := postService.CreatePost(1, largeText)
		if err != nil {
			b.Fatalf("failed to create post: %v", err)
		}

		_, err = postService.GetPostById(post.ID)
		if err != nil {
			b.Fatalf("failed to get post: %v", err)
		}
	}
}

func BenchmarkGetPostById(b *testing.B) {
	storage := storage.NewObjectStorage()

	log := logger.NewLogger(1000)
	defer log.Close()

	postService := service.NewPostService(storage, log)

	post, _ := postService.CreatePost(1, "Benchmark post")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = postService.GetPostById(post.ID)
	}
}

func BenchmarkGetAllPosts(b *testing.B) {
	storage := storage.NewObjectStorage()

	log := logger.NewLogger(1000)
	defer log.Close()

	postService := service.NewPostService(storage, log)

	for i := 0; i < 100; i++ {
		postService.CreatePost(1, "Benchmark post")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = postService.GetAllPosts()
	}
}

func BenchmarkLikePost(b *testing.B) {
	storage := storage.NewObjectStorage()

	log := logger.NewLogger(1000)
	defer log.Close()

	postService := service.NewPostService(storage, log)
	likeService := queue.NewLikeService(storage, 1000)

	postService.SetLikeService(likeService)
	likeService.StartWorker()
	defer likeService.StopWorker()

	post, _ := postService.CreatePost(1, "Benchmark like post")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = postService.LikePost(post.ID)
	}

}
