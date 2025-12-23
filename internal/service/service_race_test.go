package service_test

import (
	"microblog/internal/logger"
	"microblog/internal/queue"
	"microblog/internal/service"
	"microblog/internal/storage"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestRegisterUserRace(t *testing.T) {
	storage := storage.NewObjectStorage()

	log := logger.NewLogger(1000)
	defer log.Close()

	userService := service.NewUserService(storage, log)

	var wg sync.WaitGroup
	numUsers := 100

	for i := 0; i < numUsers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _ = userService.RegisterUser("User"+strconv.Itoa(i), "user"+strconv.Itoa(i)+"@mail.com")
		}(i)
	}

	wg.Wait()

}

func TestLikePostRace(t *testing.T) {
	storage := storage.NewObjectStorage()

	log := logger.NewLogger(1000)
	defer log.Close()

	postService := service.NewPostService(storage, log)
	likeService := queue.NewLikeService(storage, 1000)

	postService.SetLikeService(likeService)
	likeService.StartWorker()
	defer likeService.StopWorker()

	post, _ := postService.CreatePost("Eva", "Race test post")

	var wg sync.WaitGroup
	numLikes := 100

	for i := 0; i < numLikes; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = postService.LikePost(post.ID)
		}()
	}

	wg.Wait()

	time.Sleep(50 * time.Millisecond)

	updatedPost, _ := postService.GetPostById(post.ID)
	if updatedPost.LikeCount != numLikes {
		t.Errorf("expected %d likes, got %d", numLikes, updatedPost.LikeCount)
	}
}

func TestCreatePostRace(t *testing.T) {
	storage := storage.NewObjectStorage()

	log := logger.NewLogger(1000)
	defer log.Close()

	postService := service.NewPostService(storage, log)

	var wg sync.WaitGroup
	numPosts := 100

	for i := 0; i < numPosts; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _ = postService.CreatePost("User"+strconv.Itoa(i), "пост"+strconv.Itoa(i))
		}(i)
	}
	wg.Wait()
}

func TestGetPostByIdRace(t *testing.T) {
	storage := storage.NewObjectStorage()

	log := logger.NewLogger(1000)
	defer log.Close()

	postService := service.NewPostService(storage, log)

	post, _ := postService.CreatePost("Eva", "Race test post")

	var wg sync.WaitGroup
	numReaders := 100

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := postService.GetPostById(post.ID)
			if err != nil {
				t.Errorf("failed to get post: %v", err)
			}
		}()
	}

	wg.Wait()
}
