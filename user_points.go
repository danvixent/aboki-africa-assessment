package aboki_africa_assessment

import "time"

type UserPoints struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Points    int64      `json:"balance"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type Transaction struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	RecipientUserID string     `json:"recipient_user_id"`
	Points          int64      `json:"points"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
}
