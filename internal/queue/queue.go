package queue

import (
	"microblog/internal/storage"
)

type LikeService struct {
	storage   *storage.ObjectStorage
	likeQueue chan int
}

func NewLikeService(storage *storage.ObjectStorage, queueSize int) *LikeService {
	return &LikeService{
		storage:   storage,
		likeQueue: make(chan int),
	}
}

func (s *LikeService) EnqueueLike(postID int) {
	s.likeQueue <- postID
}

func (s *LikeService) StartWorker() {
	go func() {
		for postID := range s.likeQueue {
			s.storage.LikePost(postID)
		}
	}()
}

func (s *LikeService) StopWorker() {
	close(s.likeQueue)
}
