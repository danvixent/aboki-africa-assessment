package postgres

import (
	"context"
	app "github.com/danvixent/aboki-africa-assessment"
	"github.com/jackc/pgx"
	"time"
)

// UserReferralResource helps to access the database atomically, it should only be used once
type UserReferralResource struct {
	tx pgx.Tx
}

func NewUserReferralResource(tx pgx.Tx) *UserReferralResource {
	return &UserReferralResource{tx: tx}
}

func (u *UserReferralResource) CreateUserReferral(ctx context.Context, referral *app.UserReferral) error {
	referral.CreatedAt = time.Now()
	referral.UpdatedAt = time.Now()

	row := u.tx.QueryRow(ctx, "INSERT INTO user_referrals (referrer_id, referee_id, created_at, updated_at) VALUES ('$1','$2','$3','$4') RETURNING id",
		referral.ReferrerID, referral.RefereeID, referral.CreatedAt, referral.UpdatedAt)

	err := row.Scan(&referral.ID)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserReferralResource) GetUnpaidUserReferralCount(ctx context.Context, userID string) (int64, error) {
	var count int64

	row := u.tx.QueryRow(ctx, "SELECT COUNT(*) FROM user_referrals WHERE referrer_id = $1 AND paid_out = false", userID)

	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (u *UserReferralResource) MarkPendingReferralsAsPaid(ctx context.Context, referrerID string) error {
	_, err := u.tx.Exec(ctx, "UPDATE user_referrals SET paid_out = true WHERE referrer_id = $1", referrerID)
	return err
}

func (u *UserReferralResource) GetUserReferrer(ctx context.Context, userID string) (*app.User, error) {
	row := u.tx.QueryRow(ctx, "SELECT * FROM users WHERE id IN( SELECT referrer_id FROM user_referrals WHERE referee_id = $1 AND deleted_at IS NULL)", userID)

	user := &app.User{}
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.ReferralCode, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}
