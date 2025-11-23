package core

import (
	"context"
	"fmt"
)

type userService struct {
	userRepo UserRepository
	prRepo   PullRequestRepository
}

func NewUserService(userRepo UserRepository, prRepo PullRequestRepository) UserService {
	return &userService{
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

func (s *userService) SetUserActive(ctx context.Context, userID string, isActive bool) (*User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	user.IsActive = isActive
	if err := s.userRepo.UpsertUser(ctx, user); err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}

	return user, nil
}

func (s *userService) GetUserReviewRequests(ctx context.Context, userID string) ([]*PullRequest, error) {
	_, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	prs, err := s.prRepo.GetPRsByReviewer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user review requests: %w", err)
	}

	return prs, nil
}
