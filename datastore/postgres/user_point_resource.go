package postgres

import (
	"context"
	"github.com/jackc/pgx"
)

// UserPointsResource helps to access the database atomically, it should only be used once
type UserPointsResource struct {
	tx pgx.Tx
}

func NewUserPointsResource(tx pgx.Tx) *UserPointsResource {
	return &UserPointsResource{tx: tx}
}

func (u *UserPointsResource) CreditUser(ctx context.Context, userID string, points int64) error {
	_, err := u.tx.Exec(ctx, "UPDATE user_points SET balance = balance + $1", points)
	return err
}
