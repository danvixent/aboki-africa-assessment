package handler

type UserRequest struct {
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	ReferralCode *string `json:"referral_c_ode"`
}
