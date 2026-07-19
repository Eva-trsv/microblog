package service_test

import (
	"context"
	"microblog/services/api/internal/logger"
	"microblog/services/api/internal/repository"
	"microblog/services/api/internal/service"
	"testing"
)

type PostTestEnv struct {
	pgContainer *service.PostgresContainer
	ctx         context.Context
	log         *logger.Logger
}

type PostTestServices struct {
	userService *service.UserService
	postService *service.PostService
}

type PostTestData struct {
	testUserID int
}

type PostServiceSuite struct {
	env      *PostTestEnv
	services *PostTestServices
	data     *PostTestData
}

func (s *PostServiceSuite) SetupSuite() {
	ctx := context.Background()

	container, err := service.SetupPostgresContainer(ctx)
	if err != nil {
		panic(err)
	}

	log := logger.NewLogger(10)

	userRepo := repository.NewUserRepository(container.Pool)
	postRepo := repository.NewPostRepository(container.Pool)

	userService := service.NewUserService(userRepo, log)
	postService := service.NewPostService(postRepo, log)

	s.env = &PostTestEnv{
		pgContainer: container,
		ctx:         ctx,
		log:         log,
	}

	s.services = &PostTestServices{
		userService: userService,
		postService: postService,
	}
}

func (s *PostServiceSuite) TearDownSuite() {
	if s.env.log != nil {
		s.env.log.Close()
	}
	if s.env.pgContainer != nil {
		_ = s.env.pgContainer.Cleanup(s.env.ctx)
	}
}

func (s *PostServiceSuite) SetupTest() {
	_ = s.env.pgContainer.TruncateTables(s.env.ctx)

	user, err := s.services.userService.CreateUser("TestAuthor", "author@test.com")
	if err != nil {
		panic(err)
	}

	s.data = &PostTestData{
		testUserID: user.ID,
	}
}

func TestCreatePost(t *testing.T) {
	s := &PostServiceSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()
	s.SetupTest()

	post, err := s.services.postService.CreatePost(s.data.testUserID, "Мой первый пост")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if post.AuthorID == 0 {
		t.Error("the author is empty")
	}

	if post.AuthorID != s.data.testUserID {
		t.Errorf("expected authorID %d, got %d", s.data.testUserID, post.AuthorID)
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

func TestGetPostByAuthorID(t *testing.T) {
	s := &PostServiceSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()
	s.SetupTest()

	post1, err := s.services.postService.CreatePost(s.data.testUserID, "Пост 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	post2, err := s.services.postService.CreatePost(s.data.testUserID, "Пост 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	posts, err := s.services.postService.GetPostsByAuthorID(s.data.testUserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(posts) != 2 {
		t.Errorf("expected 2 posts, got %d", len(posts))
	}

	if posts[0].AuthorID != s.data.testUserID {
		t.Error("author ID mismatch")
	}
	if posts[0].Content != post1.Content && posts[0].Content != post2.Content {
		t.Error("post content mismatch")
	}

	emptyPosts, err := s.services.postService.GetPostsByAuthorID(99999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(emptyPosts) != 0 {
		t.Error("expected empty posts for non-existent author")
	}
}

func TestDeletePost(t *testing.T) {
	s := &PostServiceSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()
	s.SetupTest()

	post1, err := s.services.postService.CreatePost(s.data.testUserID, "Мой первый пост")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	post2, err := s.services.postService.CreatePost(s.data.testUserID, "Мой второй пост")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.services.postService.DeletePost(post1.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	posts, err := s.services.postService.GetPostsByAuthorID(s.data.testUserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(posts) != 1 {
		t.Errorf("expected 1 post after deletion, got %d", len(posts))
	}

	if posts[0].ID != post2.ID {
		t.Error("wrong post remained after deletion")
	}
}

func TestDeletePost_NotFound(t *testing.T) {
	s := &PostServiceSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()
	s.SetupTest()

	err := s.services.postService.DeletePost(99999)
	if err == nil {
		t.Error("expected error for non-existent post ID")
	}
}