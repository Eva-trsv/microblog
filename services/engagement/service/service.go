package service

import (
	"context"
	"microblog/services/engagement/repository"
)

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) HandlePostLiked(eventID string, postID int) error {
	ctx := context.Background()
	processed, err := s.repo.IsProcessed(ctx, eventID)
	if err != nil {
		return err
	}

	if processed {
		return nil
	}

	if err := s.repo.IncrementLike(ctx, postID); err != nil {
		return err
	}

	return s.repo.SaveProcessed(ctx, eventID)
}

func (s *Service) GetPostStats(postID int) (int, error) {
	return s.repo.GetLikes(context.Background(), postID)
}