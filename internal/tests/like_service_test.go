package tests

import (
	"context"
	"microblog/internal/logger"
	"microblog/internal/queue"
	"microblog/internal/repository"
	"microblog/internal/service"
	"microblog/internal/tests/helpers"
	"testing"
	"time"
)

type LikeServiceSuite struct {
	pgContainer *helpers.PostgresContainer
	ctx         context.Context
	userService *service.UserService
	postService *service.PostService
	likeService *queue.LikeService
	log         *logger.Logger
	testUserID  int
	testPostID  int
}

func (s *LikeServiceSuite) SetupSuite() {
	s.ctx = context.Background()

	container, err := helpers.SetupPostgresContainer(s.ctx)
	if err != nil {
		panic(err)
	}
	s.pgContainer = container

	s.log = logger.NewLogger(10)

	userRepo := repository.NewUserRepository(container.Pool)
	postRepo := repository.NewPostRepository(container.Pool)
	likeRepo := repository.NewLikeRepository(container.Pool)

	s.userService = service.NewUserService(userRepo, s.log)
	s.postService = service.NewPostService(postRepo, s.log)
	s.likeService = queue.NewLikeService(likeRepo, 1000)

	s.postService.SetLikeService(s.likeService)
}

func (s *LikeServiceSuite) TearDownSuite() {
	if s.log != nil {
		s.log.Close()
	}
	if s.pgContainer != nil {
		_ = s.pgContainer.Cleanup(s.ctx)
	}
}

func (s *LikeServiceSuite) SetupTest() {
	_ = s.pgContainer.TruncateTables(s.ctx)

	user, err := s.userService.CreateUser("LikeTester", "liketester@test.com")
	if err != nil {
		panic(err)
	}
	s.testUserID = user.ID

	post, err := s.postService.CreatePost(s.testUserID, "Тестовый пост для лайков")
	if err != nil {
		panic(err)
	}
	s.testPostID = post.ID

	s.likeService.StartWorker()
}

func (s *LikeServiceSuite) TearDownTest() {
	s.likeService.StopWorker()
}


func TestLikePost(t *testing.T) {
	s := &LikeServiceSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()
	s.SetupTest()

	msg, err := s.postService.LikePost(s.testUserID, s.testPostID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "like queued" {
		t.Errorf("expected message 'like queued', got %s", msg)
	}

	time.Sleep(50 * time.Millisecond)

	posts, err := s.postService.GetPostsByAuthorID(s.testUserID)
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

	_, err := s.postService.LikePost(s.testUserID, 999)
	if err == nil {
		t.Error("expected error for non-existent post ID")
	}
}

func TestMultipleLikes(t *testing.T) {
	s := &LikeServiceSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()
	s.SetupTest()

	user2, err := s.userService.CreateUser("SecondUser", "second@test.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	post2, err := s.postService.CreatePost(user2.ID, "Пост второго пользователя")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = s.postService.LikePost(post2.ID, user2.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	posts, err := s.postService.GetPostsByAuthorID(user2.ID)
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

	_, err := s.postService.LikePost(s.testPostID, s.testUserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = s.postService.LikePost(s.testPostID, s.testUserID)
	if err != nil {
		t.Fatalf("unexpected error on second like: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	posts, err := s.postService.GetPostsByAuthorID(s.testUserID)
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