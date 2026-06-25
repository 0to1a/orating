package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"project/internal/platform/authctx"
	"project/internal/platform/cache"
	"project/internal/platform/compiled"
	"project/internal/platform/humax"
	"project/internal/platform/mailer"
)

const (
	otpValidity     = 10 * time.Minute
	sessionTokenLen = 32 // bytes — 64 hex chars
	sessionTTL      = 30 * 24 * time.Hour
	localhostOTP    = "123456"
)

type Handler struct {
	queries *compiled.Queries
	mailer  mailer.Mailer
	mw      *Middleware // for cache invalidation on logout
	cache   *cache.Cache[string, *authctx.Principal]
	tmpl    *template.Template
}

// ===== handlers =====

func (h *Handler) handleLoginRequest(ctx context.Context, input *LoginRequestInput) (*struct{}, error) {
	req := input.Body
	if !isValidEmail(req.Email) {
		return nil, humax.BadRequest("invalid email")
	}

	user, err := h.queries.AuthFindUserByEmail(ctx, strings.ToLower(req.Email))
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		// First-login auto-creates the user.
		user, err = h.queries.AuthCreateUser(ctx, compiled.AuthCreateUserParams{
			Email: strings.ToLower(req.Email),
			Name:  emailLocalPart(req.Email),
		})
		if err != nil {
			return nil, err
		}
	}

	otp, err := h.generateAndStoreOTP(ctx, user.ID, req.Email)
	if err != nil {
		return nil, err
	}

	// @localhost emails skip the send — they accept hardcoded OTP 123456.
	if !isLocalhostEmail(req.Email) {
		if err := h.sendOTPEmail(ctx, req.Email, otp); err != nil {
			return nil, err
		}
	}

	return nil, nil // 204 No Content
}

func (h *Handler) handleLoginVerify(ctx context.Context, input *LoginVerifyInput) (*LoginVerifyOutput, error) {
	req := input.Body
	if req.OTP == "" || req.Email == "" {
		return nil, humax.BadRequest("email and otp are required")
	}

	user, err := h.queries.AuthFindUserByEmail(ctx, strings.ToLower(req.Email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.BadRequest("invalid otp")
		}
		return nil, err
	}

	if !h.verifyOTP(user, req) {
		return nil, humax.BadRequest("invalid otp")
	}

	token, err := generateSessionToken()
	if err != nil {
		return nil, err
	}

	if _, err := h.queries.AuthCreateSession(ctx, compiled.AuthCreateSessionParams{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: pgxTimestamp(time.Now().Add(sessionTTL)),
	}); err != nil {
		return nil, err
	}
	_ = h.queries.AuthClearOTP(ctx, user.ID)

	// Auto-select first company if none selected (common after invite flow).
	if !user.SelectedCompanyID.Valid {
		if firstCompanyID, err := h.queries.AuthFirstCompanyForUser(ctx, user.ID); err == nil {
			_ = h.queries.AuthSetSelectedCompany(ctx, compiled.AuthSetSelectedCompanyParams{
				ID:                user.ID,
				SelectedCompanyID: pgtype.Int8{Int64: firstCompanyID, Valid: true},
			})
			user.SelectedCompanyID = pgtype.Int8{Int64: firstCompanyID, Valid: true}
		}
	}

	p := principalFromRow(user, token)
	h.cache.SetWithTTL(token, p, sessionCacheTTL)

	return &LoginVerifyOutput{Body: VerifyResponse{
		Token: token,
		Profile: AuthProfile{
			ID:                p.UserID,
			Email:             p.Email,
			Name:              p.Name,
			SelectedCompanyID: p.SelectedCompanyID,
		},
	}}, nil
}

func (h *Handler) handleLogout(ctx context.Context, _ *struct{}) (*struct{}, error) {
	p, err := humax.RequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if p.Token != "" {
		_ = h.queries.AuthDeleteSession(ctx, p.Token)
		h.mw.invalidate(p.Token)
	}
	return nil, nil // 204 No Content
}

func (h *Handler) handleMe(ctx context.Context, _ *struct{}) (*AuthMeOutput, error) {
	p, err := humax.RequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	return &AuthMeOutput{Body: AuthProfile{
		ID:                p.UserID,
		Email:             p.Email,
		Name:              p.Name,
		SelectedCompanyID: p.SelectedCompanyID,
	}}, nil
}

// ===== helpers =====

func (h *Handler) verifyOTP(user compiled.User, req VerifyRequest) bool {
	if isLocalhostEmail(req.Email) && req.OTP == localhostOTP {
		return true
	}
	if !user.OtpCode.Valid || !user.OtpExpiresAt.Valid {
		return false
	}
	if time.Now().After(user.OtpExpiresAt.Time) {
		return false
	}
	return user.OtpCode.String == req.OTP
}

func (h *Handler) generateAndStoreOTP(ctx context.Context, userID int64, email string) (string, error) {
	otp := localhostOTP
	if !isLocalhostEmail(email) {
		var err error
		otp, err = generateNumericOTP(6)
		if err != nil {
			return "", err
		}
	}
	expires := time.Now().Add(otpValidity)
	if err := h.queries.AuthSetOTP(ctx, compiled.AuthSetOTPParams{
		ID:           userID,
		OtpCode:      pgxText(otp),
		OtpExpiresAt: pgxTimestamp(expires),
	}); err != nil {
		return "", err
	}
	return otp, nil
}

func (h *Handler) sendOTPEmail(ctx context.Context, to, otp string) error {
	var body bytes.Buffer
	if err := h.tmpl.ExecuteTemplate(&body, "otp.html", struct {
		OTP          string
		MinutesValid int
	}{OTP: otp, MinutesValid: int(otpValidity.Minutes())}); err != nil {
		return err
	}
	return h.mailer.Send(ctx, mailer.Message{
		To:      to,
		Subject: "Your login code",
		HTML:    body.String(),
		Text:    fmt.Sprintf("Your login code: %s (expires in %d minutes)", otp, int(otpValidity.Minutes())),
	})
}

// ===== utils =====

func generateNumericOTP(digits int) (string, error) {
	max := byte(10)
	buf := make([]byte, digits)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, digits)
	for i, b := range buf {
		out[i] = '0' + (b % max)
	}
	return string(out), nil
}

func generateSessionToken() (string, error) {
	buf := make([]byte, sessionTokenLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func emailLocalPart(email string) string {
	at := strings.Index(email, "@")
	if at <= 0 {
		return email
	}
	return email[:at]
}

func isLocalhostEmail(email string) bool {
	return strings.HasSuffix(strings.ToLower(email), "@localhost")
}

func isValidEmail(s string) bool {
	// Permissive: accepts single-label domains like user@localhost for dev/test.
	s = strings.TrimSpace(s)
	at := strings.Index(s, "@")
	return at > 0 && at < len(s)-1
}

func pgxText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func pgxTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t, Valid: true}
}
