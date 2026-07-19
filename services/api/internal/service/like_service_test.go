package service_test

import (
	"context"
	"microblog/services/api/internal/logger"
	"microblog/services/api/internal/queue"
	"microblog/services/api/internal/repository"
	"microblog/services/api/internal/service"
	"testing"
	"time"
)

type LikeTestEnv struct {
	pgContainer *service.PostgresContainer
	ctx         context.Context
	log         *logger.Logger
}

type LikeTestServices struct {
	userService *service.UserService
	postService *service.PostService
	likeService *queue.LikeService
}

type LikeTestData struct {
	testUserID int
	testPostID int
}

type LikeServiceSuite struct {
	env      *LikeTestEnv
	services *LikeTestServices
	data     *LikeTestData
}

const loggerBufferSize = 10

func (s *LikeServiceSuite) SetupSuite() {
	ctx := context.Background()

	container, err := service.SetupPostgresContainer(ctx)
	if err != nil {
		panic(err)
	}

	log := logger.NewLogger(loggerBufferSize)

	userRepo := repository.NewUserRepository(container.Pool)
	postRepo := repository.NewPostRepository(container.Pool)
	likeRepo := repository.NewLikeRepository(container.Pool)

	userService := service.NewUserService(userRepo, log)
	postService := service.NewPostService(postRepo, log)
	likeService := queue.NewLikeService(likeRepo, 1000)

	postService.SetLikeService(likeService)

	s.env = &LikeTestEnv{
		pgContainer: container,
		ctx:         ctx,
		log:         log,
	}

	s.services = &LikeTestServices{
		userService: userService,
		postService: postService,
		likeService: likeService,
	}
}

func (s *LikeServiceSuite) TearDownSuite() {
	if s.env.log != nil {
		s.env.log.Close()
	}
	if s.env.pgContainer != nil {
		_ = s.env.pgContainer.Cleanup(s.env.ctx)
	}
}

func (s *LikeServiceSuite) SetupTest() {
	_ = s.env.pgContainer.TruncateTables(s.env.ctx)

	user, err := s.services.userService.CreateUser("LikeTester", "liketester@test.com")
	if err != nil {
		panic(err)
	}

	post, err := s.services.postService.CreatePost(user.ID, "Тестовый пост для лайков")
	if err != nil {
		panic(err)
	}

	s.data = &LikeTestData{
		testUserID: user.ID,
		testPostID: post.ID,
	}

	s.services.likeService.StartWorker()
}

func (s *LikeServiceSuite) TearDownTest() {
	s.services.likeService.StopWorker()
}

func TestLikePost(t *testing.T) {
	s := &LikeServiceSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()
	s.SetupTest()

	msg, err := s.services.postService.LikePost(s.data.testUserID, s.data.testPostID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "like queued" {
		t.Errorf("expected message 'like queued', got %s", msg)
	}

	time.Sleep(50 * time.Millisecond)

	posts, err := s.services.postService.GetPostsByAuthorID(s.data.testUserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(posts) == 0 {
		t.Fatal("no posts found")
	}

	if posts[0].LikeCount != 1 {
		t.Errorf("expected 1 like, got %d", posts[0].LikeCount)
	}
}

func TestLikePost_InvalidPost(t *testing.T) {
	s := &LikeServiceSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()
	s.SetupTest()

	_, err := s.services.postService.LikePost(s.data.testUserID, 999)
	if err == nil {
		t.Error("expected error for non-existent post ID")
	}
}

func TestMultipleLikes(t *testing.T) {
	s := &LikeServiceSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()
	s.SetupTest()

	user2, err := s.services.userService.CreateUser("SecondUser", "second@test.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	post2, err := s.services.postService.CreatePost(user2.ID, "Пост второго пользователя")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = s.services.postService.LikePost(user2.ID, post2.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	posts, err := s.services.postService.GetPostsByAuthorID(user2.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(posts) == 0 {
		t.Fatal("no posts found")
	}

	if posts[0].LikeCount != 1 {
		t.Errorf("expected 1 like, got %d", posts[0].LikeCount)
	}
}

func TestLikePost_Twice(t *testing.T) {
	s := &LikeServiceSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()
	s.SetupTest()

	_, err := s.services.postService.LikePost(s.data.testUserID, s.data.testPostID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = s.services.postService.LikePost(s.data.testUserID, s.data.testPostID)
	if err != nil {
		t.Fatalf("unexpected error on second like: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	posts, err := s.services.postService.GetPostsByAuthorID(s.data.testUserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(posts) == 0 {
		t.Fatal("no posts found")
	}

	if posts[0].LikeCount != 1 {
		t.Errorf("expected 1 like, got %d", posts[0].LikeCount)
	}
}