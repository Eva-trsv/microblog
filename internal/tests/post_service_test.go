package tests

import (
	"context"
	"microblog/internal/logger"
	"microblog/internal/repository"
	"microblog/internal/service"
	"microblog/internal/tests/helpers"
	"testing"
)

type PostServiceSuite struct {
	pgContainer *helpers.PostgresContainer
	ctx         context.Context
	userService *service.UserService
	postService *service.PostService
	log         *logger.Logger
	testUserID  int
}

func (s *PostServiceSuite) SetupSuite() {
	s.ctx = context.Background()

	container, err := helpers.SetupPostgresContainer(s.ctx)
	if err != nil {
		panic(err)
	}
	s.pgContainer = container

	s.log = logger.NewLogger(10)

	userRepo := repository.NewUserRepository(container.Pool)
	postRepo := repository.NewPostRepository(container.Pool)

	s.userService = service.NewUserService(userRepo, s.log)
	s.postService = service.NewPostService(postRepo, s.log)
}

func (s *PostServiceSuite) TearDownSuite() {
	if s.log != nil {
		s.log.Close()
	}
	if s.pgContainer != nil {
		_ = s.pgContainer.Cleanup(s.ctx)
	}
}

func (s *PostServiceSuite) SetupTest() {
	_ = s.pgContainer.TruncateTables(s.ctx) //очищаем табл

	user, err := s.userService.CreateUser("TestAuthor", "author@test.com")
	if err != nil {
		panic(err)
	}
	s.testUserID = user.ID
}

func TestCreatePost(t *testing.T) {
	s := &PostServiceSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()
	s.SetupTest()

	post, err := s.postService.CreatePost(s.testUserID, "Мой первый пост")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if post.AuthorID == 0 {
		t.Error("the author is empty")
	}

	if post.AuthorID != s.testUserID {
		t.Errorf("expected authorID %d, got %d", s.testUserID, post.AuthorID)
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

	post1, err := s.postService.CreatePost(s.testUserID, "Пост 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	post2, err := s.postService.CreatePost(s.testUserID, "Пост 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	posts, err := s.postService.GetPostsByAuthorID(s.testUserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(posts) != 2 {
		t.Errorf("expected 2 posts, got %d", len(posts))
	}

	if posts[0].AuthorID != s.testUserID {
		t.Error("author ID mismatch")
	}
	if posts[0].Content != post1.Content && posts[0].Content != post2.Content {
		t.Error("post content mismatch")
	}

	emptyPosts, err := s.postService.GetPostsByAuthorID(99999)
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

	post1, err := s.postService.CreatePost(s.testUserID, "Мой первый пост")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	post2, err := s.postService.CreatePost(s.testUserID, "Мой второй пост")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.postService.DeletePost(post1.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	posts, err := s.postService.GetPostsByAuthorID(s.testUserID)
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

	err := s.postService.DeletePost(99999)
	if err == nil {
		t.Error("expected error for non-existent post ID")
	}
}
