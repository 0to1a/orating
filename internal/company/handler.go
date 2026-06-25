package company

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"project/internal/platform/compiled"
	"project/internal/platform/humax"
)

const pgErrUniqueViolation = "23505"

type Handler struct {
	queries        *compiled.Queries
	invalidateAuth func(token string) // nil-safe; called when user state mutates
}

// ===== company-level =====

func (h *Handler) handleList(ctx context.Context, _ *struct{}) (*CompanyListOutput, error) {
	p, err := humax.RequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := h.queries.CompanyListByUser(ctx, p.UserID)
	if err != nil {
		return nil, mapAuthError(err)
	}
	out := make([]CompanyInfo, 0, len(rows))
	for _, row := range rows {
		out = append(out, CompanyInfo{
			ID:        row.ID,
			Name:      row.Name,
			Role:      row.Role,
			IsOwner:   row.OwnerID == p.UserID,
			CreatedAt: row.CreatedAt.Time.UTC().Format(time.RFC3339),
		})
	}
	return &CompanyListOutput{Body: CompanyListResponse{Companies: out}}, nil
}

func (h *Handler) handleCreate(ctx context.Context, input *CompanyCreateInput) (*CompanyCreateOutput, error) {
	p, err := humax.RequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	req := input.Body
	if strings.TrimSpace(req.Name) == "" {
		return nil, humax.BadRequest("name is required")
	}

	// Owner-only: must already own a company. Non-owners bootstrap via invite.
	owns, err := h.queries.CompanyExistsOwnedBy(ctx, p.UserID)
	if err != nil {
		return nil, mapAuthError(err)
	}
	if !owns {
		return nil, humax.Forbidden("only existing company owners can create new companies")
	}

	c, err := h.queries.CompanyCreate(ctx, compiled.CompanyCreateParams{
		Name:    req.Name,
		OwnerID: p.UserID,
	})
	if err != nil {
		return nil, mapAuthError(err)
	}

	if _, err := h.queries.CompanyAddMember(ctx, compiled.CompanyAddMemberParams{
		CompanyID: c.ID,
		UserID:    p.UserID,
		Role:      RoleAdmin,
	}); err != nil {
		return nil, mapAuthError(err)
	}

	return &CompanyCreateOutput{Body: CompanyCreateResponse{Company: CompanyInfo{
		ID:        c.ID,
		Name:      c.Name,
		Role:      RoleAdmin,
		IsOwner:   true,
		CreatedAt: c.CreatedAt.Time.UTC().Format(time.RFC3339),
	}}}, nil
}

func (h *Handler) handleSelect(ctx context.Context, input *SelectInput) (*struct{}, error) {
	p, err := humax.RequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	companyID := input.ID
	if _, err := h.requireMember(ctx, companyID, p.UserID); err != nil {
		return nil, mapAuthError(err)
	}
	if err := h.queries.CompanySetUserSelectedCompany(ctx, compiled.CompanySetUserSelectedCompanyParams{
		ID:                p.UserID,
		SelectedCompanyID: pgInt8(companyID),
	}); err != nil {
		return nil, mapAuthError(err)
	}
	// Invalidate so the next request resolves with the new selected_company_id.
	if h.invalidateAuth != nil && p.Token != "" {
		h.invalidateAuth(p.Token)
	}
	return nil, nil // 204 No Content
}

func (h *Handler) handleListRoles(ctx context.Context, _ *struct{}) (*RolesOutput, error) {
	_, err := humax.RequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	return &RolesOutput{Body: RolesResponse{Roles: ListRoles()}}, nil
}

// ===== member-level (operate on selected company) =====

func (h *Handler) handleListMembers(ctx context.Context, _ *struct{}) (*MemberListOutput, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}
	companyID := p.SelectedCompanyID
	// Admin-only.
	if err := h.requireAdmin(ctx, companyID, p.UserID); err != nil {
		return nil, mapAuthError(err)
	}
	rows, err := h.queries.CompanyListMembers(ctx, companyID)
	if err != nil {
		return nil, mapAuthError(err)
	}
	out := make([]Member, 0, len(rows))
	for _, row := range rows {
		out = append(out, Member{
			UserID:   row.UserID,
			Email:    row.Email,
			Name:     row.Name,
			Role:     row.Role,
			JoinedAt: row.JoinedAt.Time.UTC().Format(time.RFC3339),
		})
	}
	return &MemberListOutput{Body: MemberListResponse{Members: out}}, nil
}

func (h *Handler) handleInvite(ctx context.Context, input *InviteInput) (*InviteOutput, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}
	companyID := p.SelectedCompanyID
	if err := h.requireAdmin(ctx, companyID, p.UserID); err != nil {
		return nil, mapAuthError(err)
	}

	req := input.Body
	if !isValidEmail(req.Email) {
		return nil, humax.BadRequest("invalid email")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, humax.BadRequest("name is required")
	}
	if !IsValidRole(req.Role) {
		return nil, humax.BadRequest("invalid role")
	}

	// Find or create user
	user, err := h.queries.CompanyFindUserByEmail(ctx, strings.ToLower(req.Email))
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, mapAuthError(err)
		}
		user, err = h.queries.CompanyCreateUser(ctx, compiled.CompanyCreateUserParams{
			Email: strings.ToLower(req.Email),
			Name:  req.Name,
		})
		if err != nil {
			if isUniqueViolation(err) {
				// Race: concurrent request just created this user. Re-fetch.
				user, err = h.queries.CompanyFindUserByEmail(ctx, strings.ToLower(req.Email))
				if errors.Is(err, pgx.ErrNoRows) {
					// Email taken by a deactivated account.
					return nil, humax.Unprocessable("an account with this email exists but is deactivated")
				}
			}
			if err != nil {
				return nil, mapAuthError(err)
			}
		}
	}

	cm, err := h.queries.CompanyAddMember(ctx, compiled.CompanyAddMemberParams{
		CompanyID: companyID,
		UserID:    user.ID,
		Role:      req.Role,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, humax.Conflict("user already a member")
		}
		return nil, mapAuthError(err)
	}

	return &InviteOutput{Body: InviteResponse{Member: Member{
		UserID:   user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Role:     req.Role,
		JoinedAt: cm.CreatedAt.Time.UTC().Format(time.RFC3339),
	}}}, nil
}

func (h *Handler) handleRemoveMember(ctx context.Context, input *RemoveMemberInput) (*struct{}, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}
	companyID := p.SelectedCompanyID
	if err := h.requireAdmin(ctx, companyID, p.UserID); err != nil {
		return nil, mapAuthError(err)
	}
	userID := input.UserID
	// Prevent removing the company owner
	c, err := h.queries.CompanyGetByID(ctx, companyID)
	if err != nil {
		return nil, mapAuthError(err)
	}
	if c.OwnerID == userID {
		return nil, humax.Forbidden("cannot remove company owner")
	}
	if err := h.queries.CompanyRemoveMember(ctx, compiled.CompanyRemoveMemberParams{
		CompanyID: companyID,
		UserID:    userID,
	}); err != nil {
		return nil, mapAuthError(err)
	}
	return nil, nil // 204 No Content
}

func (h *Handler) handleUpdateRole(ctx context.Context, input *UpdateRoleInput) (*struct{}, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}
	companyID := p.SelectedCompanyID
	if err := h.requireAdmin(ctx, companyID, p.UserID); err != nil {
		return nil, mapAuthError(err)
	}
	if !IsValidRole(input.Body.Role) {
		return nil, humax.BadRequest("invalid role")
	}
	if err := h.queries.CompanyUpdateMemberRole(ctx, compiled.CompanyUpdateMemberRoleParams{
		CompanyID: companyID,
		UserID:    input.UserID,
		Role:      input.Body.Role,
	}); err != nil {
		return nil, mapAuthError(err)
	}
	return nil, nil // 204 No Content
}

// ===== guards =====

func (h *Handler) requireMember(ctx context.Context, companyID, userID int64) (compiled.CompanyMember, error) {
	cm, err := h.queries.CompanyGetMembership(ctx, compiled.CompanyGetMembershipParams{
		CompanyID: companyID,
		UserID:    userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return cm, errNotMember
		}
		return cm, err
	}
	return cm, nil
}

func (h *Handler) requireAdmin(ctx context.Context, companyID, userID int64) error {
	cm, err := h.requireMember(ctx, companyID, userID)
	if err != nil {
		return err
	}
	if cm.Role != RoleAdmin {
		return errNotAdmin
	}
	return nil
}

// ===== sentinel errors =====
var (
	errNotMember = errors.New("not a member of this company")
	errNotAdmin  = errors.New("admin role required")
)

// mapAuthError maps sentinel errors to the correct huma HTTP errors.
// For other errors, it returns them as-is (huma will render as 500).
func mapAuthError(err error) error {
	switch {
	case errors.Is(err, errNotAdmin), errors.Is(err, errNotMember):
		return humax.Forbidden(err.Error())
	default:
		return err
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgErrUniqueViolation
}

func isValidEmail(s string) bool {
	s = strings.TrimSpace(s)
	at := strings.Index(s, "@")
	return at > 0 && at < len(s)-1
}

// pgInt8 wraps int64 as pgtype.Int8 (nullable bigint).
func pgInt8(n int64) pgtype.Int8 {
	return pgtype.Int8{Int64: n, Valid: true}
}
