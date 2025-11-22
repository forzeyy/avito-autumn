package services

import (
	"context"
	"errors"

	"github.com/forzeyy/avito-autumn/internal/models"
	"github.com/forzeyy/avito-autumn/internal/repos"
)

type UserService interface {
	SetUserActive(ctx context.Context, userID string, isActive bool) (*models.User, error)
	GetPRsByReviewer(ctx context.Context, userID string) ([]models.PullRequestShort, error)
}

type userService struct {
	userRepo repos.UserRepo
	prRepo   repos.PRRepo
}

func NewUserService(userRepo repos.UserRepo, prRepo repos.PRRepo) UserService {
	return &userService{
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

func (us *userService) SetUserActive(ctx context.Context, userID string, isActive bool) (*models.User, error) {
	user, err := us.userRepo.SetUserActive(ctx, userID, isActive)
	if err != nil {
		return nil, errors.New(NOT_FOUND)
	}
	return user, nil
}

func (us *userService) GetPRsByReviewer(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	prs, err := us.prRepo.GetPRsByReviewer(ctx, userID)
	if err != nil {
		return nil, errors.New(NOT_FOUND)
	}
	return prs, nil
}
