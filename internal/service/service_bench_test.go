package service_test

import (
	"context"
	"microblog/internal/logger"
	"microblog/internal/queue"
	"microblog/internal/repository"
	"microblog/internal/service"
	"strconv"
	"strings"
	"testing"
)

const loggerBenchBufferSize = 1000
const likeQueueSize = 1000

type BenchEnv struct {
	container *service.PostgresContainer
	ctx       context.Context
	log       *logger.Logger
}

type BenchServices struct {
	userSvc *service.UserService
	postSvc *service.PostService
	likeSvc *queue.LikeService
}

type BenchData struct {
	userID int
}

type BenchSuite struct {
	env      *BenchEnv
	services *BenchServices
	data     *BenchData
}

func (s *BenchSuite) Setup() {
	ctx := context.Background()

	container, err := service.SetupPostgresContainer(ctx)
	if err != nil {
		panic(err)
	}

	log := logger.NewLogger(loggerBenchBufferSize)

	userRepo := repository.NewUserRepository(container.Pool)
	postRepo := repository.NewPostRepository(container.Pool)
	likeRepo := repository.NewLikeRepository(container.Pool)

	userSvc := service.NewUserService(userRepo, log)
	postSvc := service.NewPostService(postRepo, log)
	likeSvc := queue.NewLikeService(likeRepo, likeQueueSize)

	postSvc.SetLikeService(likeSvc)
	likeSvc.StartWorker()

	user, err := userSvc.CreateUser("BenchUser", "bench@test.com")
	if err != nil {
		panic(err)
	}

	s.env = &BenchEnv{
		container: container,
		ctx:       ctx,
		log:       log,
	}

	s.services = &BenchServices{
		userSvc: userSvc,
		postSvc: postSvc,
		likeSvc: likeSvc,
	}

	s.data = &BenchData{
		userID: user.ID,
	}
}

func (s *BenchSuite) TearDown() {
	if s.env.log != nil {
		s.env.log.Close()
	}
	if s.env.container != nil {
		_ = s.env.container.Cleanup(s.env.ctx)
	}
}

func BenchmarkCreateUser(b *testing.B) {
	s := &BenchSuite{}
	s.Setup()
	defer s.TearDown()

	if err := s.env.container.TruncateTables(s.env.ctx); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		username := "User" + strconv.Itoa(i)
		email := "user" + strconv.Itoa(i) + "@test.com"
		s.services.userSvc.CreateUser(username, email)
	}
}

func BenchmarkCreatePost(b *testing.B) {
	s := &BenchSuite{}
	s.Setup()
	defer s.TearDown()

	if err := s.env.container.TruncateTables(s.env.ctx); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.services.postSvc.CreatePost(s.data.userID, "Мой первый пост")
	}
}

func BenchmarkGetPostByID(b *testing.B) {
	s := &BenchSuite{}
	s.Setup()
	defer s.TearDown()

	if err := s.env.container.TruncateTables(s.env.ctx); err != nil {
		b.Fatal(err)
	}

	post, err := s.services.postSvc.CreatePost(s.data.userID, "Benchmark post")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.services.postSvc.GetPostByID(post.ID)
	}
}

func BenchmarkGetPostsByAuthorID(b *testing.B) {
	s := &BenchSuite{}
	s.Setup()
	defer s.TearDown()

	if err := s.env.container.TruncateTables(s.env.ctx); err != nil {
		b.Fatal(err)
	}

	for i := 0; i < 100; i++ {
		_, err := s.services.postSvc.CreatePost(s.data.userID, "Benchmark post")
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.services.postSvc.GetPostsByAuthorID(s.data.userID)
	}
}

func BenchmarkLargeTextCreateAndRead(b *testing.B) {
	s := &BenchSuite{}
	s.Setup()
	defer s.TearDown()

	if err := s.env.container.TruncateTables(s.env.ctx); err != nil {
		b.Fatal(err)
	}

	largeText := strings.Repeat("A", 2000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		post, err := s.services.postSvc.CreatePost(s.data.userID, largeText)
		if err != nil {
			b.Fatal(err)
		}
		s.services.postSvc.GetPostByID(post.ID)
	}
}

func BenchmarkLikePost(b *testing.B) {
	s := &BenchSuite{}
	s.Setup()
	defer s.TearDown()

	if err := s.env.container.TruncateTables(s.env.ctx); err != nil {
		b.Fatal(err)
	}

	post, err := s.services.postSvc.CreatePost(s.data.userID, "Benchmark like post")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.services.postSvc.LikePost(s.data.userID, post.ID)
	}
}