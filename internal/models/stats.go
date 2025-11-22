package models

type UserStats struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	ReviewCount int    `json:"review_count"`
	IsActive    bool   `json:"is_active"`
}

type StatsResponse struct {
	TotalPRsCreated int         `json:"total_prs_created"`
	ReviewsByUser   []UserStats `json:"reviews_by_user"`
}
