package aboki_africa_assessment

import (
	"context"
	"time"
)

type User struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	ReferralCode string     `json:"referral_code"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

type UserReferral struct {
	ID         string     `json:"id"`
	ReferrerID string     `json:"referrer_id"` // ID of the user whose referral code was used
	RefereeID  string     `json:"referee_id"`  // ID of the user who was referred
	PaidOut    bool       `json:"paid_out"`    // has this referral bonus being paid out to the referrer
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at"`
}

type ReferredUserTransactionBonus struct {
	ID         string     `json:"id"`
	ReferrerID string     `json:"referrer_id"` // ID of the user whose referral code was used
	RefereeID  string     `json:"referee_id"`  // ID of the user who was referred
	PaidOut    bool       `json:"paid_out"`    // has this referred user transaction bonus been paid out to the referrer
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at"`
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	FindUserByID(ctx context.Context, id string) error
	FindUserByReferralCode(ctx context.Context, code string) (*User, error)
}

type UserReferralRepository interface {
	CreateUserReferral(ctx context.Context, referral *UserReferral) error
	GetUnpaidUserReferralCount(ctx context.Context, userID string) (int64, error)
	MarkPendingReferralsAsPaid(ctx context.Context, referrerID string) error
	GetUserReferrer(ctx context.Context, userID string) (*User, error)
	CreateReferredUserTransactionBonus(ctx context.Context, referral *ReferredUserTransactionBonus) error
	GetUnpaidReferredUserTransactionBonus(ctx context.Context, userID string) ([]*ReferredUserTransactionBonus, error)
	PayReferralsTransactionsBonuses(ctx context.Context, ids []string) error
}
