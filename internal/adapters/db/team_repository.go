package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/penkovgd/pr-reviews/internal/core"
)

func (d *DB) CreateTeam(ctx context.Context, team *core.Team) error {
	query := `INSERT INTO teams (name) VALUES ($1)`
	_, err := d.conn.ExecContext(ctx, query, team.Name)
	if err != nil {
		return fmt.Errorf("create team %s: %w", team.Name, err)
	}
	return nil
}

func (d *DB) GetTeamByName(ctx context.Context, teamName string) (*core.Team, error) {
	var team core.Team
	query := `SELECT * FROM teams WHERE name = $1`
	err := d.conn.GetContext(ctx, &team, query, teamName)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("team %s: %w", teamName, core.ErrTeamNotFound)
		}
		return nil, fmt.Errorf("get team %s: %w", teamName, err)
	}

	usersQuery := `SELECT id, username, team_name, is_active FROM users WHERE team_name = $1`
	var users []core.User
	err = d.conn.SelectContext(ctx, &users, usersQuery, teamName)
	if err != nil {
		return nil, fmt.Errorf("get team %s users: %w", teamName, err)
	}

	return &core.Team{Name: teamName, Members: users}, nil
}
