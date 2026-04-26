package queue

import (
	"context"
	"microblog/internal/repository"
	"log"
)

type LikeService struct {
	likeRepo  *repository.LikeRepository
	likeQueue chan likeTask
}

type likeTask struct {
	UserID int
	PostID int
}

func NewLikeService(likeRepo *repository.LikeRepository, queueSize int) *LikeService {
	return &LikeService{
		likeRepo:  likeRepo,
		likeQueue: make(chan likeTask, queueSize),
	}
}

func (s *LikeService) EnqueueLike(userID, postID int) {
	s.likeQueue <- likeTask{UserID: userID, PostID: postID}
}

func (s *LikeService) StartWorker() {
	go func() {
		for task := range s.likeQueue {

			ctx := context.Background()

			tx, err := s.likeRepo.DB().Begin(ctx)
			if err != nil {
				log.Println("failed to begin tx:", err)
				continue
			}
			defer tx.Rollback(ctx)

			err = s.likeRepo.AddLike(ctx, tx, task.UserID, task.PostID)
			if err != nil {
				_ = tx.Rollback(ctx)
				log.Println("failed to add like:", err)
				continue
			}

			err = tx.Commit(ctx)
			if err != nil {
				log.Println("failed to commit tx:", err)
			}
		}
	}()
}

func (s *LikeService) StopWorker() {
	close(s.likeQueue)
}
