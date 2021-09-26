package postgres

import (
	"context"
	app "github.com/danvixent/aboki-africa-assessment"
	"github.com/jackc/pgx/v4"
)

// UserResource helps to access the database atomically, it should only be used once
type UserResource struct {
	client *Client
}

var defaultOptions = pgx.TxOptions{
	IsoLevel:       pgx.ReadCommitted,
	DeferrableMode: pgx.NotDeferrable,
	AccessMode:     pgx.ReadWrite,
}

func NewUserRepository(client *Client) *UserResource {
	return &UserResource{client: client}
}

func (u *UserResource) CreateUser(ctx context.Context, user *app.User) error {
	tx, err := u.client.GetTx(ctx)
	if err != nil {
		return err
	}

	row := tx.QueryRow(ctx,
		"INSERT INTO users (name, email, referral_code) VALUES($1, $2, $3) RETURNING id, created_at, updated_at",
		user.Name, user.Email, user.ReferralCode)

	err = row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserResource) FindUserByID(ctx context.Context, id string) error {
	tx, err := u.client.GetTx(ctx)
	if err != nil {
		return err
	}

	row := tx.QueryRow(ctx, "SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL", id)

	user := &app.User{}
	err = row.Scan(&user.ID, &user.Name, &user.Email, &user.ReferralCode, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserResource) FindUserByReferralCode(ctx context.Context, code string) (*app.User, error) {
	tx, err := u.client.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	row := tx.QueryRow(ctx, "SELECT * FROM users WHERE referral_code = $1 AND deleted_at IS NULL", code)

	user := &app.User{}
	err = row.Scan(&user.ID, &user.Name, &user.Email, &user.ReferralCode, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}
