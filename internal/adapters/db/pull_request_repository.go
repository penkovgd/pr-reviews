package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/penkovgd/pr-reviews/internal/core"
)

func (d *DB) CreatePR(ctx context.Context, pr *core.PullRequest) error {
	tx, err := d.conn.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			d.log.Error("transaction rollback", "error", err)
		}
	}()

	query := `INSERT INTO pull_requests (id, name, author_id, status, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.ExecContext(ctx, query, pr.ID, pr.Name, pr.AuthorID, pr.Status, pr.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert pull request %s: %w", pr.ID, err)
	}

	reviewerQuery := `INSERT INTO pull_request_reviewers (pull_request_id, user_id) VALUES ($1, $2)`
	for _, reviewerID := range pr.AssignedReviewers {
		_, err := tx.ExecContext(ctx, reviewerQuery, pr.ID, reviewerID)
		if err != nil {
			return fmt.Errorf("assign reviewer %s to PR %s: %w", reviewerID, pr.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit pull request creation: %w", err)
	}
	return nil
}

func (d *DB) GetPRByID(ctx context.Context, prID string) (*core.PullRequest, error) {
	var pr core.PullRequest

	query := `SELECT id, name, author_id, status, created_at, merged_at FROM pull_requests WHERE id = $1`
	if err := d.conn.GetContext(ctx, &pr, query, prID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("pull request %s: %w", prID, core.ErrPRNotFound)
		}
		return nil, fmt.Errorf("get pull request %s: %w", prID, err)
	}

	var reviewers []string
	reviewerQuery := `SELECT user_id FROM pull_request_reviewers WHERE pull_request_id = $1`
	if err := d.conn.SelectContext(ctx, &reviewers, reviewerQuery, prID); err != nil {
		return nil, fmt.Errorf("get reviewers for PR %s: %w", prID, err)
	}

	pr.AssignedReviewers = reviewers
	return &pr, nil
}

func (d *DB) GetPRsByReviewer(ctx context.Context, userID string) ([]*core.PullRequest, error) {
	var user core.User
	userQuery := `SELECT * FROM users WHERE id = $1`
	if err := d.conn.GetContext(ctx, &user, userQuery, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user %s: %w", userID, core.ErrPRNotFound)
		}
		return nil, fmt.Errorf("check user exists %s: %w", userID, err)
	}

	query := `
		SELECT pr.id, pr.name, pr.author_id, pr.status, pr.created_at, pr.merged_at
		FROM pull_requests pr
		JOIN pull_request_reviewers prr ON pr.id = prr.pull_request_id
		WHERE prr.user_id = $1
		ORDER BY pr.created_at DESC
		`

	var prs []*core.PullRequest
	if err := d.conn.SelectContext(ctx, &prs, query, userID); err != nil {
		return nil, fmt.Errorf("get pull requests for reviewer %s: %w", userID, err)
	}

	for _, pr := range prs {
		var reviewers []string
		reviewerQuery := `SELECT user_id FROM pull_request_reviewers WHERE pull_request_id = $1`
		if err := d.conn.SelectContext(ctx, &reviewers, reviewerQuery, pr.ID); err != nil {
			return nil, fmt.Errorf("get reviewers for PR %s: %w", pr.ID, err)
		}
		pr.AssignedReviewers = reviewers
	}

	return prs, nil
}

func (d *DB) UpdatePR(ctx context.Context, pr *core.PullRequest) error {
	tx, err := d.conn.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			d.log.Error("transaction rollback", "error", err)
		}
	}()

	query := `UPDATE pull_requests SET name = $1, author_id = $2, status = $3, merged_at = $4 WHERE id = $5`
	result, err := tx.ExecContext(ctx, query, pr.Name, pr.AuthorID, pr.Status, pr.MergedAt, pr.ID)
	if err != nil {
		return fmt.Errorf("update pull request %s: %w", pr.ID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected for PR %s: %w", pr.ID, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("pull request %s: %w", pr.ID, core.ErrPRNotFound)
	}

	// if status MERGED, doesn't update reviewers
	if pr.Status == core.StatusMerged {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit pull request update: %w", err)
		}
		return nil
	}

	// delete old reviewers and insert new reviewers
	deleteQuery := `DELETE FROM pull_request_reviewers WHERE pull_request_id = $1`
	_, err = tx.ExecContext(ctx, deleteQuery, pr.ID)
	if err != nil {
		return fmt.Errorf("delete old reviewers for PR %s: %w", pr.ID, err)
	}

	insertQuery := `INSERT INTO pull_request_reviewers (pull_request_id, user_id) VALUES ($1, $2)`
	for _, reviewerID := range pr.AssignedReviewers {
		_, err := tx.ExecContext(ctx, insertQuery, pr.ID, reviewerID)
		if err != nil {
			return fmt.Errorf("insert reviewer %s for PR %s: %w", reviewerID, pr.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit pull request update with reviewers: %w", err)
	}
	return nil
}
