package core

import "errors"

var (
	// Team errors
	ErrTeamExists   = errors.New("team already exists")
	ErrTeamNotFound = errors.New("team not found")

	// User errors
	ErrUserNotFound  = errors.New("user not found")
	ErrUserNotActive = errors.New("user is not active")

	// PullRequest errors
	ErrPRExists   = errors.New("pull request already exists")
	ErrPRNotFound = errors.New("pull request not found")
	ErrPRMerged   = errors.New("cannot modify merged pull request")

	// Reviewer assignment errors
	ErrReviewerNotAssigned = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate         = errors.New("no active replacement candidate in team")
)
