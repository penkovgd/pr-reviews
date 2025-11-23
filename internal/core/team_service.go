package core

import (
	"context"
	"errors"
	"fmt"
)

type teamService struct {
	teamRepo TeamRepository
	userRepo UserRepository
}

func NewTeamService(teamRepo TeamRepository, userRepo UserRepository) TeamService {
	return &teamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *teamService) CreateTeam(ctx context.Context, team *Team) error {
	t, err := s.teamRepo.GetTeamByName(ctx, team.Name)
	if t != nil {
		return ErrTeamExists
	}

	if !errors.Is(err, ErrTeamNotFound) {
		return fmt.Errorf("check team existence: %w", err)
	}

	if err := s.teamRepo.CreateTeam(ctx, team); err != nil {
		return fmt.Errorf("create team: %w", err)
	}

	for i := range team.Members {
		user := &team.Members[i]
		user.TeamName = team.Name

		if err := s.userRepo.UpsertUser(ctx, user); err != nil {
			return fmt.Errorf("create user %s: %w", user.ID, err)
		}
	}

	return nil
}

func (s *teamService) GetTeam(ctx context.Context, teamName string) (*Team, error) {
	team, err := s.teamRepo.GetTeamByName(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("get team: %w", err)
	}
	return team, nil
}
