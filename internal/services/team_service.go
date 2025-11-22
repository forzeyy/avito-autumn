package services

import (
	"context"
	"errors"

	"github.com/forzeyy/avito-autumn/internal/models"
	"github.com/forzeyy/avito-autumn/internal/repos"
)

type TeamService interface {
	CreateTeam(ctx context.Context, team *models.Team) error
	GetTeam(ctx context.Context, teamName string) (*models.Team, error)
}

type teamService struct {
	teamRepo repos.TeamRepo
	userRepo repos.UserRepo
}

func NewTeamService(teamRepo repos.TeamRepo, userRepo repos.UserRepo) TeamService {
	return &teamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (ts *teamService) CreateTeam(ctx context.Context, team *models.Team) error {
	exists, err := ts.teamRepo.IsTeamExists(ctx, team.Name)
	if err != nil {
		return err
	}
	if *exists {
		return errors.New(TEAM_EXISTS)
	}

	err = ts.teamRepo.CreateTeam(ctx, team.Name)
	if err != nil {
		return err
	}

	for _, member := range team.Members {
		user := &models.User{
			ID:       member.UserID,
			Username: member.Username,
			TeamName: team.Name,
			IsActive: member.IsActive,
		}
		err := ts.userRepo.UpsertUser(ctx, user)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ts *teamService) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	team, err := ts.teamRepo.GetTeam(ctx, teamName)
	if err != nil {
		return nil, err
	}
	return team, nil
}
