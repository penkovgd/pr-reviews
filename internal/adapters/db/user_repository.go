package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/penkovgd/pr-reviews/internal/core"
)

func (d *DB) UpsertUser(ctx context.Context, user *core.User) error {
	query := `
        INSERT INTO users (id, username, team_name, is_active) 
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) DO UPDATE SET
            username = EXCLUDED.username,
            team_name = EXCLUDED.team_name,
            is_active = EXCLUDED.is_active
    `
	_, err := d.conn.ExecContext(ctx, query, user.ID, user.Username, user.TeamName, user.IsActive)
	if err != nil {
		return fmt.Errorf("create or update user %s: %w", user.ID, err)
	}
	return nil
}

func (d *DB) GetUserByID(ctx context.Context, userID string) (*core.User, error) {
	var user core.User
	query := `SELECT id, username, team_name, is_active FROM users WHERE id = $1`
	err := d.conn.GetContext(ctx, &user, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user %s: %w", userID, core.ErrUserNotFound)
		}
		return nil, fmt.Errorf("get user %s: %w", userID, err)
	}
	return &user, nil
}

func (d *DB) GetUsersByTeam(ctx context.Context, teamName string) ([]*core.User, error) {
	var users []*core.User
	query := `SELECT id, username, team_name, is_active FROM users WHERE team_name = $1`
	err := d.conn.SelectContext(ctx, &users, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("get users for team %s: %w", teamName, err)
	}
	return users, nil
}
