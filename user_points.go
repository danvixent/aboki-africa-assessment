package aboki_africa_assessment

import "time"

type UserPoints struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Balance   int64      `json:"balance"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type UserTransferredPoints struct {
	UserID                 string `json:"user_id"`
	TotalTransferredPoints int64  `json:"total_transferred_points"`
}
