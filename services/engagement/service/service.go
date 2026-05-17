package service

import (
	"context"
	"microblog/services/engagement/repository"
)

type EngService struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) *EngService {
	return &EngService{
		repo: repo,
	}
}

func (s *EngService) HandleEvent(eventID string) error {
	return s.repo.SaveProcessed(context.Background(), eventID)
}

func (s *EngService) HandlePostLiked(eventID string, postID int) error {
	_, err := s.repo.ProcessPostLiked(context.Background(), eventID, postID)
	return err
}

func (s *EngService) GetPostStats(postID int) (int, error) {
	return s.repo.GetLikes(context.Background(), postID)
}
