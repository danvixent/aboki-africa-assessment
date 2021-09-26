package postgres

import (
	"context"
	app "github.com/danvixent/aboki-africa-assessment"
	"github.com/jackc/pgx"
	"time"
)

// UserPointsResource helps to access the database atomically, it should only be used once
type UserPointsResource struct {
	tx pgx.Tx
}

func NewUserPointsResource(tx pgx.Tx) *UserPointsResource {
	return &UserPointsResource{tx: tx}
}

func (u *UserPointsResource) CreditUser(ctx context.Context, userID string, points int64) error {
	_, err := u.tx.Exec(ctx, "UPDATE user_points SET points = points + $1 WHERE user_id = $2", points, userID)
	return err
}

func (u *UserPointsResource) GetUserPointsBalance(ctx context.Context, userID string) (int64, error) {
	var balance int64
	row := u.tx.QueryRow(ctx, "SELECT balance from user_points WHERE user_id = $1", userID)
	if err := row.Scan(&balance); err != nil {
		return 0, err
	}
	return balance, nil
}

func (u *UserPointsResource) GetUserTotalTransferredPoints(ctx context.Context, userID string) (int64, error) {
	var balance int64
	row := u.tx.QueryRow(ctx, "SELECT SUM(points) from transactions WHERE user_id = $1 AND deleted_at IS NULL", userID)
	if err := row.Scan(&balance); err != nil {
		return 0, err
	}
	return balance, nil
}

func (u *UserPointsResource) CreatePointTransaction(ctx context.Context, txn *app.Transaction) error {
	txn.CreatedAt = time.Now()
	txn.UpdatedAt = time.Now()

	row := u.tx.QueryRow(ctx, "INSERT INTO transactions (user_id, recipient_user_id, points, created_at, updated_at) RETURNING id", txn.UserID, txn.RecipientUserID, txn.Points, txn.CreatedAt, txn.UpdatedAt)
	return row.Scan(&txn.ID)
}

func (u *UserPointsResource) DebitUser(ctx context.Context, points int64, userID string) error {
	_, err := u.tx.Exec(ctx, "UPDATE user_points SET points = points - $1 WHERE user_id = $2", points, userID)
	return err
}
