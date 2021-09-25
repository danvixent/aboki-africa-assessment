package postgres

import (
	"context"
	app "github.com/danvixent/aboki-africa-assessment"
	"github.com/jackc/pgx"
	"time"
)

// UserResource helps to access the database atomically, it should only be used once
type UserResource struct {
	tx pgx.Tx
}

var defaultOptions = pgx.TxOptions{
	IsoLevel:       pgx.ReadCommitted,
	DeferrableMode: pgx.Deferrable,
	AccessMode:     pgx.ReadWrite,
}

func NewUserResource(tx pgx.Tx) *UserResource {
	return &UserResource{tx: tx}
}

func (u *UserResource) CreateUser(ctx context.Context, user *app.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	row := u.tx.QueryRow(ctx,
		"INSERT INTO users (name, email, referral_code, created_at, updated_at) VALUES('$1','$2','$3','$4','$5') RETURNING id",
		user.Name, user.Email, user.ReferralCode, user.CreatedAt, user.UpdatedAt)

	err := row.Scan(&user.ID)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserResource) FindUserByID(ctx context.Context, id string) error {
	row := u.tx.QueryRow(ctx, "SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL", id)

	user := &app.User{}
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.ReferralCode, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserResource) FindUserByReferralCode(ctx context.Context, code string) (*app.User, error) {
	row := u.tx.QueryRow(ctx, "SELECT * FROM users WHERE referral_code = $1 AND deleted_at IS NULL", code)

	user := &app.User{}
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.ReferralCode, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}
