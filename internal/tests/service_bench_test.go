package tests

import (
	"context"
	"microblog/internal/logger"
	"microblog/internal/queue"
	"microblog/internal/repository"
	"microblog/internal/service"
	"microblog/internal/tests/helpers"
	"strings"
	"testing"
)

var (
	benchContainer *helpers.PostgresContainer
	benchCtx       = context.Background()
	benchLog       *logger.Logger
	benchUserSvc   *service.UserService
	benchPostSvc   *service.PostService
	benchLikeSvc   *queue.LikeService
	benchUserID    int
)

func init() {

	container, err := helpers.SetupPostgresContainer(benchCtx) // Запускаем контейнер один раз для всех бенчмарков
	if err != nil {
		panic(err)
	}
	benchContainer = container

	benchLog = logger.NewLogger(1000)

	userRepo := repository.NewUserRepository(benchContainer.Pool)
	postRepo := repository.NewPostRepository(benchContainer.Pool)
	likeRepo := repository.NewLikeRepository(benchContainer.Pool)

	benchUserSvc = service.NewUserService(userRepo, benchLog)
	benchPostSvc = service.NewPostService(postRepo, benchLog)
	benchLikeSvc = queue.NewLikeService(likeRepo, 1000)

	benchPostSvc.SetLikeService(benchLikeSvc)
	benchLikeSvc.StartWorker()

	user, err := benchUserSvc.CreateUser("BenchUser", "bench@test.com")
	if err != nil {
		panic(err)
	}
	benchUserID = user.ID
}

func BenchmarkCreateUser(b *testing.B) {
	if err := benchContainer.TruncateTables(benchCtx); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		username := "User" + string(rune(i))
		email := "user" + string(rune(i)) + "@test.com"
		benchUserSvc.CreateUser(username, email)
	}
}

func BenchmarkCreatePost(b *testing.B) {
	if err := benchContainer.TruncateTables(benchCtx); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchPostSvc.CreatePost(benchUserID, "Мой первый пост")
	}
}

func BenchmarkGetPostByID(b *testing.B) {
	if err := benchContainer.TruncateTables(benchCtx); err != nil {
		b.Fatal(err)
	}

	post, err := benchPostSvc.CreatePost(benchUserID, "Benchmark post")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchPostSvc.GetPostByID(post.ID)
	}
}

func BenchmarkGetPostsByAuthorID(b *testing.B) {
	if err := benchContainer.TruncateTables(benchCtx); err != nil {
		b.Fatal(err)
	}

	for i := 0; i < 100; i++ {
		_, err := benchPostSvc.CreatePost(benchUserID, "Benchmark post")
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchPostSvc.GetPostsByAuthorID(benchUserID)
	}
}

func BenchmarkLargeTextCreateAndRead(b *testing.B) {
	if err := benchContainer.TruncateTables(benchCtx); err != nil {
		b.Fatal(err)
	}

	largeText := strings.Repeat("A", 2000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		post, err := benchPostSvc.CreatePost(benchUserID, largeText)
		if err != nil {
			b.Fatal(err)
		}
		benchPostSvc.GetPostByID(post.ID)
	}
}

func BenchmarkLikePost(b *testing.B) {
	if err := benchContainer.TruncateTables(benchCtx); err != nil {
		b.Fatal(err)
	}

	post, err := benchPostSvc.CreatePost(benchUserID, "Benchmark like post")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchPostSvc.LikePost(benchUserID, post.ID)
	}
}
