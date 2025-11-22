package services

import (
	"context"
	"errors"
	"math/rand"

	"github.com/forzeyy/avito-autumn/internal/models"
	"github.com/forzeyy/avito-autumn/internal/repos"
)

type PRService interface {
	CreatePR(ctx context.Context, prID string, prName string, authorID string) (*models.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*models.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*models.PullRequest, string, error)
}

type prService struct {
	prRepo   repos.PRRepo
	userRepo repos.UserRepo
}

func NewPRService(prRepo repos.PRRepo, userRepo repos.UserRepo) PRService {
	return &prService{
		prRepo:   prRepo,
		userRepo: userRepo,
	}
}

func (prs *prService) CreatePR(ctx context.Context, prID string, prName string, authorID string) (*models.PullRequest, error) {
	_, err := prs.prRepo.GetPRByID(ctx, prID)
	if err == nil {
		return nil, errors.New("PR_EXISTS")
	}

	author, err := prs.userRepo.GetUser(ctx, authorID)
	if err != nil {
		return nil, errors.New("NOT_FOUND")
	}

	teamUsers, err := prs.userRepo.GetActiveUsersByTeam(ctx, author.TeamName)
	if err != nil {
		return nil, errors.New("NOT_FOUND")
	}

	var reviewers []string
	for _, user := range teamUsers {
		if user.ID != authorID && user.IsActive {
			reviewers = append(reviewers, user.ID)
		}
	}

	if len(reviewers) > 2 {
		rand.Shuffle(len(reviewers), func(i, j int) {
			reviewers[i], reviewers[j] = reviewers[j], reviewers[i]
		})
		reviewers = reviewers[:2]
	}

	newPR := &models.PullRequest{
		ID:                prID,
		Name:              prName,
		AuthorID:          authorID,
		Status:            models.StatusOpen,
		AssignedReviewers: reviewers,
	}

	err = prs.prRepo.CreatePR(ctx, newPR)
	if err != nil {
		return nil, err
	}

	return newPR, err
}

func (prs *prService) MergePR(ctx context.Context, prID string) (*models.PullRequest, error) {
	pr, err := prs.prRepo.GetPRByID(ctx, prID)
	if err != nil {
		return nil, errors.New("NOT_FOUND")
	}

	if pr.Status == models.StatusMerged {
		return pr, nil
	}

	pr.Status = models.StatusMerged
	updatedPR, err := prs.prRepo.UpdatePRStatus(ctx, prID, models.StatusMerged)
	if err != nil {
		return nil, err
	}

	return updatedPR, nil
}

func (prs *prService) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*models.PullRequest, string, error) {
	if prID == "" || oldReviewerID == "" {
		return nil, "", errors.New("INVALID_INPUT")
	}

	pr, err := prs.prRepo.GetPRByID(ctx, prID)
	if err != nil {
		return nil, "", errors.New("NOT_FOUND")
	}

	if pr.Status == models.StatusMerged {
		return nil, "", errors.New("PR_MERGED")
	}

	isAssigned := false
	for _, rev := range pr.AssignedReviewers {
		if rev == oldReviewerID {
			isAssigned = true
			break
		}
	}
	if !isAssigned {
		return nil, "", errors.New("NOT_ASSIGNED")
	}

	oldReviewer, err := prs.userRepo.GetUser(ctx, oldReviewerID)
	if err != nil {
		return nil, "", errors.New("NOT_FOUND")
	}

	candidates, err := prs.userRepo.GetActiveUsersByTeam(ctx, oldReviewer.TeamName)
	if err != nil {
		return nil, "", errors.New("NO_CANDIDATE")
	}

	var available []string
	for _, user := range candidates {
		if user.ID == oldReviewerID || user.ID == pr.AuthorID {
			continue
		}

		// на всякий случай
		isAlreadyAssigned := false
		for _, assigned := range pr.AssignedReviewers {
			if assigned == user.ID && assigned != oldReviewerID {
				isAlreadyAssigned = true
				break
			}
		}
		if !isAlreadyAssigned {
			available = append(available, user.ID)
		}
	}

	if len(available) == 0 {
		return nil, "", errors.New("NO_CANDIDATE")
	}

	newReviewerID := available[rand.Intn(len(available))]

	err = prs.prRepo.ReplaceReviewer(ctx, prID, oldReviewerID, newReviewerID)
	if err != nil {
		return nil, "", err
	}

	updPR, err := prs.prRepo.GetPRByID(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	return updPR, newReviewerID, nil
}
