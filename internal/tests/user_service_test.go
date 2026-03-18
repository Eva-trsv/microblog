package tests

import (
	"context"
	"microblog/internal/logger"
	"microblog/internal/repository"
	"microblog/internal/service"
	"microblog/internal/tests/helpers"
	"testing"
)

type UserServiceSuite struct {
	pgContainer *helpers.PostgresContainer
	ctx         context.Context
	userService *service.UserService
	log         *logger.Logger
}

func (s *UserServiceSuite) SetupSuite() {
	s.ctx = context.Background()

	container, err := helpers.SetupPostgresContainer(s.ctx) // Запускаем контейнер с PostgreSQL
	if err != nil {
		panic(err)
	}
	s.pgContainer = container

	s.log = logger.NewLogger(10)

	userRepo := repository.NewUserRepository(container.Pool)
	s.userService = service.NewUserService(userRepo, s.log)
}

func (s *UserServiceSuite) TearDownSuite() { //Закрывает логгер и удаляет контейнер
	if s.log != nil {
		s.log.Close()
	}
	if s.pgContainer != nil {
		_ = s.pgContainer.Cleanup(s.ctx)
	}
}

func (s *UserServiceSuite) SetupTest() { // очистка таблицы перед каждым тестом

	_ = s.pgContainer.TruncateTables(s.ctx)
}

func TestCreateUser(t *testing.T) {
	s := &UserServiceSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()
	s.SetupTest()

	user, err := s.userService.CreateUser("Eva", "testeva@mail.ru")
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

	created, err := s.userService.CreateUser("User1", "getuser@mail.ru")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	found, err := s.userService.GetUserByID(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if found.ID == 0 {
		t.Error("expected non-zero user ID")
	}

	if found.Username == "" {
		t.Error("the username is empty")
	}

	if found.Username != "GetUser" {
		t.Errorf("expected username 'GetUser', got %s", found.Username)
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

	_, err := s.userService.GetUserByID(99999)
	if err == nil {
		t.Error("expected error for non-existent user ID")
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	s := &UserServiceSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()
	s.SetupTest()

	_, err := s.userService.CreateUser("Eva", "duplicate@mail.ru")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = s.userService.CreateUser("Another", "duplicate@mail.ru")
	if err == nil {
		t.Error("expected error for duplicate email, got nil")
	}
}
