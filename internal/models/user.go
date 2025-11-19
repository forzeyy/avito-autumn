package models

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	TeamName string    `json:"team_name"`
	IsActive bool      `json:"is_active"`
}
