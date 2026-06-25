package user

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"project/internal/platform/compiled"
	"project/internal/platform/humax"
)

type Handler struct {
	queries *compiled.Queries
}

func (h *Handler) handleMe(ctx context.Context, _ *struct{}) (*MeOutput, error) {
	p, err := humax.RequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	row, err := h.queries.UserGetByID(ctx, p.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.NotFound("user not found")
		}
		return nil, err
	}
	return &MeOutput{Body: toProfile(row)}, nil
}

func (h *Handler) handleUpdateMe(ctx context.Context, input *UpdateMeInput) (*MeOutput, error) {
	p, err := humax.RequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	req := input.Body
	if strings.TrimSpace(req.Name) == "" {
		return nil, humax.BadRequest("name is required")
	}
	if err := h.queries.UserUpdateName(ctx, compiled.UserUpdateNameParams{
		ID:   p.UserID,
		Name: req.Name,
	}); err != nil {
		return nil, err
	}
	row, err := h.queries.UserGetByID(ctx, p.UserID)
	if err != nil {
		return nil, err
	}
	return &MeOutput{Body: toProfile(row)}, nil
}

func toProfile(u compiled.User) Profile {
	p := Profile{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt.Time.UTC().Format(time.RFC3339),
	}
	if u.SelectedCompanyID.Valid {
		p.SelectedCompanyID = u.SelectedCompanyID.Int64
	}
	return p
}
