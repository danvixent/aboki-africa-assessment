package postgres

import (
	"context"
	app "github.com/danvixent/aboki-africa-assessment"
	"github.com/danvixent/aboki-africa-assessment/errors"
	"strings"
	"time"
)

type UserPointsRepository struct {
	client *Client
}

func NewUserPointsRepository(client *Client) *UserPointsRepository {
	return &UserPointsRepository{client: client}
}

func (u *UserPointsRepository) CreditUser(ctx context.Context, userID string, points int64) error {
	tx, err := u.client.GetTx(ctx)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "UPDATE user_points SET points = points + $1 WHERE user_id = $2", points, userID)
	return err
}

func (u *UserPointsRepository) CreateUserPoint(ctx context.Context, userPoint *app.UserPoints) error {
	tx, err := u.client.GetTx(ctx)
	if err != nil {
		return err
	}

	row := tx.QueryRow(ctx, "INSERT INTO user_points (user_id, points) VALUES ($1,$2) RETURNING id, created_at, updated_at", userPoint.UserID, userPoint.Points)

	return row.Scan(&userPoint.ID, &userPoint.CreatedAt, &userPoint.UpdatedAt)
}

func (u *UserPointsRepository) GetUserPointsBalance(ctx context.Context, userID string) (int64, error) {
	tx, err := u.client.GetTx(ctx)
	if err != nil {
		return 0, err
	}

	var balance int64
	row := tx.QueryRow(ctx, "SELECT points from user_points WHERE user_id = $1 AND deleted_at IS NULL", userID)
	if err := row.Scan(&balance); err != nil {
		return 0, err
	}
	return balance, nil
}

func (u *UserPointsRepository) GetUserTotalTransferredPoints(ctx context.Context, userID string) (int64, error) {
	tx, err := u.client.GetTx(ctx)
	if err != nil {
		return 0, err
	}

	var balance int64
	row := tx.QueryRow(ctx, "SELECT SUM(points) from transactions WHERE user_id = $1 AND deleted_at IS NULL", userID)
	if err = row.Scan(&balance); err != nil {
		if strings.Contains(err.Error(), "can't scan into dest[0]") {
			return 0, nil
		}
		return 0, err
	}
	return balance, nil
}

func (u *UserPointsRepository) CreatePointTransaction(ctx context.Context, txn *app.Transaction) error {
	tx, err := u.client.GetTx(ctx)
	if err != nil {
		return err
	}

	txn.CreatedAt = time.Now()
	txn.UpdatedAt = time.Now()

	row := tx.QueryRow(ctx, "INSERT INTO transactions (user_id, recipient_user_id, points) VALUES ($1,$2,$3) RETURNING id, created_at, updated_at", txn.UserID, txn.RecipientUserID, txn.Points)
	return row.Scan(&txn.ID, &txn.CreatedAt, &txn.UpdatedAt)
}

func (u *UserPointsRepository) DebitUser(ctx context.Context, points int64, userID string) error {
	tx, err := u.client.GetTx(ctx)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "UPDATE user_points SET points = points - $1 WHERE user_id = $2 AND deleted_at IS NULL", points, userID)
	return err
}

func (u *UserPointsRepository) TransferPoints(ctx context.Context, senderID string, recipientID string, points int64) error {
	err := u.DebitUser(ctx, points, senderID)
	if err != nil {
		return errors.Wrap(err, "debit user failed")
	}

	err = u.CreditUser(ctx, recipientID, points)
	if err != nil {
		return errors.Wrap(err, "credit user failed")
	}
	return nil
}
