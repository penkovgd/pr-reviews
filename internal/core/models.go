package core

import (
	"time"
)

type User struct {
	ID       string `db:"id"`
	Username string `db:"username"`
	TeamName string `db:"team_name"`
	IsActive bool   `db:"is_active"`
}

type Team struct {
	Name    string `db:"name"`
	Members []User
}

type PullRequestStatus string

const (
	StatusOpen   PullRequestStatus = "OPEN"
	StatusMerged PullRequestStatus = "MERGED"
)

type PullRequest struct {
	ID                string            `db:"id"`
	Name              string            `db:"name"`
	AuthorID          string            `db:"author_id"`
	Status            PullRequestStatus `db:"status"`
	CreatedAt         *time.Time        `db:"created_at"`
	MergedAt          *time.Time        `db:"merged_at"`
	AssignedReviewers []string
}

type ReviewReassignment struct {
	PR            *PullRequest
	NewReviewerID string
}
