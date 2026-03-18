package tests

import (
	"context"
	"microblog/internal/logger"
	"microblog/internal/queue"
	"microblog/internal/repository"
	"microblog/internal/service"
	"microblog/internal/tests/helpers"
	"strconv"
	"sync"
	"testing"
	"time"
)

var (
	raceContainer *helpers.PostgresContainer
	raceCtx       = context.Background()
	raceLog       *logger.Logger
	raceUserSvc   *service.UserService
	racePostSvc   *service.PostService
	raceLikeSvc   *queue.LikeService
	raceUserID    int
)

func init() {

	container, err := helpers.SetupPostgresContainer(raceCtx) // Запускаем контейнер один раз для всех тестов
	if err != nil {
		panic(err)
	}
	raceContainer = container

	raceLog = logger.NewLogger(1000)

	userRepo := repository.NewUserRepository(raceContainer.Pool)
	postRepo := repository.NewPostRepository(raceContainer.Pool)
	likeRepo := repository.NewLikeRepository(raceContainer.Pool)

	raceUserSvc = service.NewUserService(userRepo, raceLog)
	racePostSvc = service.NewPostService(postRepo, raceLog)
	raceLikeSvc = queue.NewLikeService(likeRepo, 1000)

	racePostSvc.SetLikeService(raceLikeSvc)

	user, err := raceUserSvc.CreateUser("RaceUser", "race@test.com")
	if err != nil {
		panic(err)
	}
	raceUserID = user.ID
}

func TestRegisterUserRace(t *testing.T) {
	if err := raceContainer.TruncateTables(raceCtx); err != nil {
		t.Fatal(err)
	}

	raceLikeSvc.StartWorker()
	defer raceLikeSvc.StopWorker()

	var wg sync.WaitGroup
	numUsers := 100

	for i := 0; i < numUsers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			raceUserSvc.CreateUser("User"+strconv.Itoa(i), "user"+strconv.Itoa(i)+"@mail.com")
		}(i)
	}

	wg.Wait()
}

func TestCreatePostRace(t *testing.T) {
	if err := raceContainer.TruncateTables(raceCtx); err != nil {
		t.Fatal(err)
	}

	raceLikeSvc.StartWorker()
	defer raceLikeSvc.StopWorker()

	var wg sync.WaitGroup
	numPosts := 100

	for i := 0; i < numPosts; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			racePostSvc.CreatePost(raceUserID, "пост"+strconv.Itoa(i))
		}(i)
	}

	wg.Wait()
}

func TestLikePostRace(t *testing.T) {
	if err := raceContainer.TruncateTables(raceCtx); err != nil {
		t.Fatal(err)
	}

	raceLikeSvc.StartWorker()
	defer raceLikeSvc.StopWorker()

	post, err := racePostSvc.CreatePost(raceUserID, "Race test post")
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	numLikes := 100

	for i := 0; i < numLikes; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			racePostSvc.LikePost(raceUserID, post.ID)
		}()
	}

	wg.Wait()

	time.Sleep(100 * time.Millisecond)

	updatedPost, err := racePostSvc.GetPostByID(post.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updatedPost.LikeCount != numLikes {
		t.Errorf("expected %d likes, got %d", numLikes, updatedPost.LikeCount)
	}
}

func TestGetPostByIDRace(t *testing.T) {
	if err := raceContainer.TruncateTables(raceCtx); err != nil {
		t.Fatal(err)
	}

	raceLikeSvc.StartWorker()
	defer raceLikeSvc.StopWorker()

	post, err := racePostSvc.CreatePost(raceUserID, "Race test post")
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	numReaders := 100

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := racePostSvc.GetPostByID(post.ID)
			if err != nil {
				t.Errorf("failed to get post: %v", err)
			}
		}()
	}

	wg.Wait()
}
