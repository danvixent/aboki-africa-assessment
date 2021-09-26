package handler

type UserRequest struct {
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	ReferralCode *string `json:"referral_code"`
}

type TransferPointsRequest struct {
	UserID          string `json:"user_id"`
	RecipientUserID string `json:"recipient_user_id"`
	Points          int64  `json:"points"`
}
