package service_test

import (
	"context"
	"microblog/internal/logger"
	"microblog/internal/repository"
	"microblog/internal/service"
	"testing"
)

type UserTestEnv struct {
	pgContainer *service.PostgresContainer
	ctx         context.Context
	log         *logger.Logger
}

type UserTestServices struct {
	userService *service.UserService
}

type UserServiceSuite struct {
	env      *UserTestEnv
	services *UserTestServices
}

func (s *UserServiceSuite) SetupSuite() {
	ctx := context.Background()

	container, err := service.SetupPostgresContainer(ctx)
	if err != nil {
		panic(err)
	}

	log := logger.NewLogger(10)

	userRepo := repository.NewUserRepository(container.Pool)
	userService := service.NewUserService(userRepo, log)

	s.env = &UserTestEnv{
		pgContainer: container,
		ctx:         ctx,
		log:         log,
	}

	s.services = &UserTestServices{
		userService: userService,
	}
}

func (s *UserServiceSuite) TearDownSuite() {
	if s.env.log != nil {
		s.env.log.Close()
	}
	if s.env.pgContainer != nil {
		_ = s.env.pgContainer.Cleanup(s.env.ctx)
	}
}

func (s *UserServiceSuite) SetupTest() {
	_ = s.env.pgContainer.TruncateTables(s.env.ctx)
}

func TestCreateUser(t *testing.T) {
	s := &UserServiceSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()
	s.SetupTest()

	user, err := s.services.userService.CreateUser("Eva", "testeva@mail.ru")
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
		t.Errorf("expected email 'testeva@mail.ru', got %s", user.Email)
	}
}

func TestGetUserByID(t *testing.T) {
	s := &UserServiceSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()
	s.SetupTest()

	created, err := s.services.userService.CreateUser("User1", "getuser@mail.ru")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	found, err := s.services.userService.GetUserByID(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if found.ID == 0 {
		t.Error("expected non-zero user ID")
	}

	if found.Username == "" {
		t.Error("the username is empty")
	}

	// ❗ у тебя тут был баг — ты сравнивал с "GetUser"
	if found.Username != "User1" {
		t.Errorf("expected username 'User1', got %s", found.Username)
	}

	if found.Email == "" {
		t.Error("the email is empty")
	}

	if found.Email != "getuser@mail.ru" {
		t.Errorf("expected email 'getuser@mail.ru', got %s", found.Email)
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	s := &UserServiceSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()
	s.SetupTest()

	_, err := s.services.userService.GetUserByID(99999)
	if err == nil {
		t.Error("expected error for non-existent user ID")
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	s := &UserServiceSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()
	s.SetupTest()

	_, err := s.services.userService.CreateUser("Eva", "duplicate@mail.ru")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = s.services.userService.CreateUser("Another", "duplicate@mail.ru")
	if err == nil {
		t.Error("expected error for duplicate email, got nil")
	}
}
