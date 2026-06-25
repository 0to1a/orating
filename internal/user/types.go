// Package user — profile + /api/users/me. Read by tygo.
package user

type Profile struct {
	ID                int64  `json:"id"`
	Email             string `json:"email"`
	Name              string `json:"name"`
	SelectedCompanyID int64  `json:"selectedCompanyId"`
	CreatedAt         string `json:"createdAt"`
}

type UpdateProfileRequest struct {
	Name string `json:"name"`
}

// ===== huma Input/Output types =====

type UpdateMeInput struct {
	Body UpdateProfileRequest
}

type MeOutput struct {
	Body Profile
}
