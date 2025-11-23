package core

import (
	"context"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, team *Team) error
	GetTeamByName(ctx context.Context, teamName string) (*Team, error)
}

type UserRepository interface {
	UpsertUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, userID string) (*User, error)
	GetUsersByTeam(ctx context.Context, teamName string) ([]*User, error)
}

type PullRequestRepository interface {
	CreatePR(ctx context.Context, pr *PullRequest) error
	GetPRByID(ctx context.Context, prID string) (*PullRequest, error)
	GetPRsByReviewer(ctx context.Context, userID string) ([]*PullRequest, error)
	UpdatePR(ctx context.Context, pr *PullRequest) error
}

type TeamService interface {
	CreateTeam(ctx context.Context, team *Team) error
	GetTeam(ctx context.Context, teamName string) (*Team, error)
}

type UserService interface {
	SetUserActive(ctx context.Context, userID string, isActive bool) (*User, error)
	GetUserReviewRequests(ctx context.Context, userID string) ([]*PullRequest, error)
}

type PullRequestService interface {
	CreatePR(ctx context.Context, prID, prName, authorID string) (*PullRequest, error)
	MergePR(ctx context.Context, prID string) (*PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (*ReviewReassignment, error)
}

type Statistics interface {
	GetUserAssignmentStats(ctx context.Context) (map[string]int, error)
}
