package aboki_africa_assessment

import "time"

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

type UserReferralTransactionTotal struct {
	ID         string     `json:"id"`
	ReferrerID string     `json:"referrer_id"` // ID of the user whose referral code was used
	RefereeID  string     `json:"referee_id"`  // ID of the user who was referred
	Points     int64      `json:"points"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at"`
}
