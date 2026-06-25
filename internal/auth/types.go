// Package auth — OTP-based login + Bearer token middleware.
package auth

type LoginRequest struct {
	Email string `json:"email"`
}

type VerifyRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type VerifyResponse struct {
	Token   string      `json:"token"`
	Profile AuthProfile `json:"profile"`
}

// AuthProfile mirrors the user shape; defined here (not imported from user)
// to keep cross-domain layering strict. Named AuthProfile to avoid an
// OpenAPI schema name collision with user.Profile.
type AuthProfile struct {
	ID                int64  `json:"id"`
	Email             string `json:"email"`
	Name              string `json:"name"`
	SelectedCompanyID int64  `json:"selectedCompanyId"`
}

// ===== huma Input/Output types =====

type LoginRequestInput struct {
	Body LoginRequest
}

type LoginVerifyInput struct {
	Body VerifyRequest
}

type LoginVerifyOutput struct {
	Body VerifyResponse
}

type AuthMeOutput struct {
	Body AuthProfile
}

type GoogleLoginRequest struct {
	IDToken string `json:"idToken"`
}

type GoogleLoginInput struct {
	Body GoogleLoginRequest
}
