package core

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type pullRequestService struct {
	prRepo   PullRequestRepository
	userRepo UserRepository
}

func NewPullRequestService(prRepo PullRequestRepository, userRepo UserRepository) PullRequestService {
	return &pullRequestService{
		prRepo:   prRepo,
		userRepo: userRepo,
	}
}

func (s *pullRequestService) CreatePR(ctx context.Context, prID, prName, authorID string) (*PullRequest, error) {
	existingPR, err := s.prRepo.GetPRByID(ctx, prID)
	if err == nil && existingPR != nil {
		return nil, ErrPRExists
	}
	if err != nil && !errors.Is(err, ErrPRNotFound) {
		return nil, fmt.Errorf("check PR existence: %w", err)
	}

	author, err := s.userRepo.GetUserByID(ctx, authorID)
	if err != nil {
		return nil, fmt.Errorf("get author: %w", err)
	}
	if !author.IsActive {
		return nil, ErrUserNotActive
	}

	reviewerIDs, err := s.assignReviewers(ctx, authorID)
	if err != nil {
		return nil, fmt.Errorf("assign reviewers: %w", err)
	}

	now := time.Now()
	pr := &PullRequest{
		ID:                prID,
		Name:              prName,
		AuthorID:          authorID,
		Status:            StatusOpen,
		AssignedReviewers: reviewerIDs,
		CreatedAt:         &now,
	}

	if err := s.prRepo.CreatePR(ctx, pr); err != nil {
		return nil, fmt.Errorf("create PR: %w", err)
	}

	return pr, nil
}

func (s *pullRequestService) assignReviewers(ctx context.Context, authorID string) ([]string, error) {
	author, err := s.userRepo.GetUserByID(ctx, authorID)
	if err != nil {
		return nil, fmt.Errorf("get author: %w", err)
	}

	teamUsers, err := s.userRepo.GetUsersByTeam(ctx, author.TeamName)
	if err != nil {
		return nil, fmt.Errorf("get team users: %w", err)
	}

	var candidates []string
	for _, user := range teamUsers {
		if user.IsActive && user.ID != authorID {
			candidates = append(candidates, user.ID)
		}
	}

	return s.selectRandomReviewers(candidates, 2), nil
}

func (s *pullRequestService) selectRandomReviewers(candidates []string, count int) []string {
	if len(candidates) == 0 {
		return nil
	}

	if count > len(candidates) {
		count = len(candidates)
	}

	shuffled := make([]string, len(candidates))
	copy(shuffled, candidates)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:count]
}

func (s *pullRequestService) MergePR(ctx context.Context, prID string) (*PullRequest, error) {
	pr, err := s.prRepo.GetPRByID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("get PR: %w", err)
	}

	if pr.Status == StatusMerged {
		return pr, nil
	}

	pr.Status = StatusMerged
	now := time.Now()
	pr.MergedAt = &now

	if err := s.prRepo.UpdatePR(ctx, pr); err != nil {
		return nil, fmt.Errorf("update PR: %w", err)
	}

	return pr, nil
}

func (s *pullRequestService) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*ReviewReassignment, error) {
	pr, err := s.prRepo.GetPRByID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("get PR: %w", err)
	}

	if pr.Status == StatusMerged {
		return nil, ErrPRMerged
	}

	found := false
	for _, reviewer := range pr.AssignedReviewers {
		if reviewer == oldUserID {
			found = true
			break
		}
	}
	if !found {
		return nil, ErrReviewerNotAssigned
	}

	excludeUsers := make([]string, len(pr.AssignedReviewers))
	copy(excludeUsers, pr.AssignedReviewers)
	excludeUsers = append(excludeUsers, pr.AuthorID)

	newReviewerID, err := s.findReplacement(ctx, oldUserID, excludeUsers)
	if err != nil {
		return nil, fmt.Errorf("find replacement: %w", err)
	}

	newReviewers := make([]string, 0, len(pr.AssignedReviewers))
	for _, reviewer := range pr.AssignedReviewers {
		if reviewer == oldUserID {
			newReviewers = append(newReviewers, newReviewerID)
		} else {
			newReviewers = append(newReviewers, reviewer)
		}
	}

	pr.AssignedReviewers = newReviewers

	if err := s.prRepo.UpdatePR(ctx, pr); err != nil {
		return nil, fmt.Errorf("update PR: %w", err)
	}

	return &ReviewReassignment{
		PR:            pr,
		NewReviewerID: newReviewerID,
	}, nil
}

func (s *pullRequestService) findReplacement(ctx context.Context, oldReviewerID string, excludeUsers []string) (string, error) {
	oldReviewer, err := s.userRepo.GetUserByID(ctx, oldReviewerID)
	if err != nil {
		return "", fmt.Errorf("get old reviewer: %w", err)
	}

	teamUsers, err := s.userRepo.GetUsersByTeam(ctx, oldReviewer.TeamName)
	if err != nil {
		return "", fmt.Errorf("get team users: %w", err)
	}

	excludeSet := make(map[string]bool)
	for _, userID := range excludeUsers {
		excludeSet[userID] = true
	}

	var candidates []string
	for _, user := range teamUsers {
		if user.IsActive && !excludeSet[user.ID] {
			candidates = append(candidates, user.ID)
		}
	}

	if len(candidates) == 0 {
		return "", ErrNoCandidate
	}

	return candidates[rand.Intn(len(candidates))], nil
}
