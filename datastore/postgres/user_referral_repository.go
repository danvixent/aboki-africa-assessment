package postgres

import (
	"context"
	app "github.com/danvixent/aboki-africa-assessment"
	"time"
)

type UserReferralRepository struct {
	client *Client
}

func NewUserReferralRepository(client *Client) *UserReferralRepository {
	return &UserReferralRepository{client: client}
}

func (u *UserReferralRepository) CreateUserReferral(ctx context.Context, referral *app.UserReferral) error {
	tx, err := u.client.GetTx(ctx)
	if err != nil {
		return err
	}

	row := tx.QueryRow(ctx, "INSERT INTO user_referrals (referrer_id, referee_id) VALUES ($1,$2) RETURNING id, created_at, updated_at",
		referral.ReferrerID, referral.RefereeID)

	err = row.Scan(&referral.ID, &referral.CreatedAt, &referral.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserReferralRepository) GetUnpaidUserReferralCount(ctx context.Context, userID string) (int64, error) {
	tx, err := u.client.GetTx(ctx)
	if err != nil {
		return 0, err
	}
	var count int64

	row := tx.QueryRow(ctx, "SELECT COUNT(*) FROM user_referrals WHERE referrer_id = $1 AND paid_out = false AND deleted_at IS NULL", userID)

	if err = row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (u *UserReferralRepository) MarkPendingReferralsAsPaid(ctx context.Context, referrerID string) error {
	tx, err := u.client.GetTx(ctx)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "UPDATE user_referrals SET paid_out = true WHERE referrer_id = $1 AND deleted_at IS NULL", referrerID)
	return err
}

func (u *UserReferralRepository) GetUserReferrer(ctx context.Context, userID string) (*app.User, error) {
	tx, err := u.client.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	row := tx.QueryRow(ctx, "SELECT * FROM users WHERE id IN( SELECT referrer_id FROM user_referrals WHERE referee_id = $1 AND deleted_at IS NULL)", userID)

	user := &app.User{}
	err = row.Scan(&user.ID, &user.Name, &user.Email, &user.ReferralCode, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserReferralRepository) CreateReferredUserTransactionBonus(ctx context.Context, referral *app.ReferredUserTransactionBonus) error {
	tx, err := u.client.GetTx(ctx)
	if err != nil {
		return err
	}

	referral.CreatedAt = time.Now()
	referral.UpdatedAt = time.Now()

	row := tx.QueryRow(ctx, "INSERT INTO referred_user_transaction_bonuses (referrer_id, referee_id, created_at, updated_at) VALUES ('$1','$2','$3','$4') RETURNING id",
		referral.ReferrerID, referral.RefereeID, referral.CreatedAt, referral.UpdatedAt)

	err = row.Scan(&referral.ID)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserReferralRepository) GetUnpaidReferredUserTransactionBonus(ctx context.Context, userID string) ([]*app.ReferredUserTransactionBonus, error) {
	tx, err := u.client.GetTx(ctx)
	if err != nil {
		return nil, err
	}

	bonuses := []*app.ReferredUserTransactionBonus{}

	rows, err := tx.Query(ctx, "SELECT * FROM referred_user_transaction_bonuses WHERE referrer_id = $1 AND paid_out = false AND deleted_at IS NULL LIMIT 3", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		bonus := &app.ReferredUserTransactionBonus{}
		err = rows.Scan(&bonus.ID, &bonus.ReferrerID, &bonus.ReferrerID, &bonus.RefereeID, &bonus.PaidOut, &bonus.CreatedAt, &bonus.UpdatedAt)
		if err != nil {
			return nil, err
		}
		bonuses = append(bonuses, bonus)
	}

	return bonuses, nil
}

func (u *UserReferralRepository) PayReferralsTransactionsBonuses(ctx context.Context, ids []string) error {
	tx, err := u.client.GetTx(ctx)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "UPDATE referred_user_transaction_bonuses SET paid_out = true WHERE id IN ($1) AND deleted_at IS NULL", ids)
	return err
}
