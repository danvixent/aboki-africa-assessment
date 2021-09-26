package aboki_africa_assessment

import (
	"context"
	"time"
)

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

type UserPointRepository interface {
	CreditUser(ctx context.Context, userID string, points int64) error
	CreateUserPoint(ctx context.Context, userPoint *UserPoints) error
	GetUserPointsBalance(ctx context.Context, userID string) (int64, error)
	GetUserTotalTransferredPoints(ctx context.Context, userID string) (int64, error)
	CreatePointTransaction(ctx context.Context, txn *Transaction) error
	DebitUser(ctx context.Context, points int64, userID string) error
	TransferPoints(ctx context.Context, senderID string, recipientID string, points int64) error
}
