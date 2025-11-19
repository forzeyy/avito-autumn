package models

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusOpen   Status = "OPEN"
	StatusMerged Status = "MERGED"
)

type PullRequest struct {
	ID                uuid.UUID   `json:"pull_request_id"`
	Name              string      `json:"pull_request_name"`
	AuthorID          uuid.UUID   `json:"author_id"`
	Status            Status      `json:"status"`
	AssignedReviewers []uuid.UUID `json:"assigned_reviewers"`
	CreatedAt         *time.Time  `json:"created_at,omitempty"`
	MergedAt          *time.Time  `json:"merged_at,omitempty"`
}
