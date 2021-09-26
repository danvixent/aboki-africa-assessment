package handler

type UserRequest struct {
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	ReferralCode *string `json:"referral_c_ode"`
}

type TransferPointsRequest struct {
	UserID          string `json:"user_id"`
	RecipientUserID string `json:"recipient_email"`
	Points          int64  `json:"points"`
}
