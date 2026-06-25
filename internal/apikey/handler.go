package apikey

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"project/internal/platform/compiled"
	"project/internal/platform/humax"
)

const (
	tokenRandomBytes = 24 // 48 hex chars after `sk_` prefix
	prefixVisibleLen = 8  // "sk_" + 5 hex chars, shown in the UI
)

type Handler struct {
	queries *compiled.Queries
}

func (h *Handler) handleList(ctx context.Context, _ *struct{}) (*APIKeyListOutput, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := h.queries.APIKeyListByCompany(ctx, p.SelectedCompanyID)
	if err != nil {
		return nil, err
	}
	out := make([]APIKeyInfo, 0, len(rows))
	for _, row := range rows {
		out = append(out, toInfo(row, ""))
	}
	return &APIKeyListOutput{Body: APIKeyListResponse{APIKeys: out}}, nil
}

func (h *Handler) handleCreate(ctx context.Context, input *APIKeyCreateInput) (*APIKeyCreateOutput, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}

	req := input.Body
	if strings.TrimSpace(req.Name) == "" {
		return nil, humax.BadRequest("name is required")
	}

	token, err := generateToken()
	if err != nil {
		return nil, err
	}
	hash := hashToken(token)
	prefix := token[:prefixVisibleLen]

	row, err := h.queries.APIKeyCreate(ctx, compiled.APIKeyCreateParams{
		Hash:      hash,
		Prefix:    prefix,
		Name:      req.Name,
		CompanyID: p.SelectedCompanyID,
		CreatedBy: p.UserID,
	})
	if err != nil {
		return nil, err
	}

	return &APIKeyCreateOutput{Body: APIKeyCreateResponse{APIKey: toInfo(row, token)}}, nil
}

func (h *Handler) handleRevoke(ctx context.Context, input *RevokeInput) (*struct{}, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}

	row, err := h.queries.APIKeyGetByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.NotFound("api key not found")
		}
		return nil, err
	}
	// Scope: only allowed to revoke keys belonging to selected company.
	if row.CompanyID != p.SelectedCompanyID {
		return nil, humax.Forbidden("api key belongs to another company")
	}
	if err := h.queries.APIKeyRevoke(ctx, input.ID); err != nil {
		return nil, err
	}
	return nil, nil // 204 No Content
}

// ===== utils =====

func generateToken() (string, error) {
	buf := make([]byte, tokenRandomBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return tokenPrefix + hex.EncodeToString(buf), nil
}

func toInfo(row compiled.ApiKey, token string) APIKeyInfo {
	info := APIKeyInfo{
		ID:        row.ID,
		Name:      row.Name,
		Prefix:    row.Prefix,
		Token:     token,
		CreatedAt: row.CreatedAt.Time.UTC().Format(time.RFC3339),
	}
	if row.LastUsedAt.Valid {
		info.LastUsedAt = row.LastUsedAt.Time.UTC().Format(time.RFC3339)
	}
	if row.RevokedAt.Valid {
		info.RevokedAt = row.RevokedAt.Time.UTC().Format(time.RFC3339)
	}
	return info
}
