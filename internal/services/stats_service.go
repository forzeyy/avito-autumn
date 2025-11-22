package services

import (
	"context"

	"github.com/forzeyy/avito-autumn/internal/models"
	"github.com/forzeyy/avito-autumn/internal/repos"
)

type StatsService interface {
	GetStats(ctx context.Context) (*models.StatsResponse, error)
}

type statsService struct {
	prRepo   repos.PRRepo
	userRepo repos.UserRepo
}

func NewStatsService(prRepo repos.PRRepo, userRepo repos.UserRepo) StatsService {
	return &statsService{
		prRepo:   prRepo,
		userRepo: userRepo,
	}
}

func (ss *statsService) GetStats(ctx context.Context) (*models.StatsResponse, error) {
	totalPRs, err := ss.prRepo.GetTotalPRCount(ctx)
	if err != nil {
		return nil, err
	}

	userStats, err := ss.prRepo.GetReviewCountByUser(ctx)
	if err != nil {
		return nil, err
	}

	return &models.StatsResponse{
		TotalPRsCreated: totalPRs,
		ReviewsByUser:   userStats,
	}, nil
}
