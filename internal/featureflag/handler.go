package featureflag

import (
	"context"
	"strings"
	"time"

	"project/internal/platform/cache"
	"project/internal/platform/compiled"
	"project/internal/platform/humax"
)

// Handler holds the private snapshot cache. Refresh: hourly cron + direct
// push from admin toggles for instant effect.
type Handler struct {
	queries *compiled.Queries
	cache   *cache.Cache[int64, FlagMap]
}

// ===== public consumption =====

// handleList returns the FlagMap for the caller's selected company. Empty
// map if no selection or no flags configured.
func (h *Handler) handleList(ctx context.Context, _ *struct{}) (*FlagListOutput, error) {
	p, err := humax.RequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	flags, _ := h.cache.Get(p.SelectedCompanyID)
	if flags == nil {
		flags = FlagMap{}
	}
	return &FlagListOutput{Body: FlagListResponse{Flags: flags}}, nil
}

// ===== admin management =====

// handleAdminContext returns everything the admin page needs in one fetch:
// flags joined with company names, companies for multi-select, distinct
// keys for autocomplete suggestions.
func (h *Handler) handleAdminContext(ctx context.Context, _ *struct{}) (*AdminContextOutput, error) {
	p, err := humax.RequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.checkAdmin(ctx, p.SelectedCompanyID, p.UserID); err != nil {
		return nil, err
	}

	flagRows, err := h.queries.FFAdminListAll(ctx)
	if err != nil {
		return nil, err
	}
	flags := make([]AdminFlag, 0, len(flagRows))
	for _, row := range flagRows {
		flags = append(flags, AdminFlag{
			ID:          row.ID,
			CompanyID:   row.CompanyID,
			CompanyName: row.CompanyName,
			FlagKey:     row.FlagKey,
			Enabled:     row.Enabled,
			UpdatedAt:   row.UpdatedAt.Time.UTC().Format(time.RFC3339),
		})
	}

	companyRows, err := h.queries.FFListCompaniesAdmin(ctx)
	if err != nil {
		return nil, err
	}
	companies := make([]CompanyOption, 0, len(companyRows))
	for _, row := range companyRows {
		companies = append(companies, CompanyOption{ID: row.ID, Name: row.Name})
	}

	keys, err := h.queries.FFListDistinctKeys(ctx)
	if err != nil {
		return nil, err
	}
	if keys == nil {
		keys = []string{}
	}

	return &AdminContextOutput{Body: AdminContextResponse{
		Flags:         flags,
		Companies:     companies,
		SuggestedKeys: keys,
	}}, nil
}

// handleUpsert: multi-company upsert. AllCompanies=true expands to every
// active company; otherwise CompanyIDs lists the targets. Pushes per-company
// cache updates so consumers see the change without waiting for cron.
func (h *Handler) handleUpsert(ctx context.Context, input *UpsertInput) (*UpsertOutput, error) {
	p, err := humax.RequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.checkAdmin(ctx, p.SelectedCompanyID, p.UserID); err != nil {
		return nil, err
	}

	req := input.Body
	if strings.TrimSpace(req.FlagKey) == "" {
		return nil, humax.BadRequest("flagKey is required")
	}

	// Resolve target company IDs.
	targets := req.CompanyIDs
	if req.AllCompanies {
		rows, err := h.queries.FFListCompaniesAdmin(ctx)
		if err != nil {
			return nil, err
		}
		targets = make([]int64, 0, len(rows))
		for _, row := range rows {
			targets = append(targets, row.ID)
		}
	}
	if len(targets) == 0 {
		return nil, humax.BadRequest("no target companies specified")
	}

	// Fetch company names once for the response payload.
	companyRows, err := h.queries.FFListCompaniesAdmin(ctx)
	if err != nil {
		return nil, err
	}
	companyName := map[int64]string{}
	for _, c := range companyRows {
		companyName[c.ID] = c.Name
	}

	out := make([]AdminFlag, 0, len(targets))
	for _, companyID := range targets {
		row, err := h.queries.FFUpsert(ctx, compiled.FFUpsertParams{
			CompanyID: companyID,
			FlagKey:   req.FlagKey,
			Enabled:   req.Enabled,
		})
		if err != nil {
			return nil, err
		}
		h.touchCache(companyID, req.FlagKey, req.Enabled)
		out = append(out, AdminFlag{
			ID:          row.ID,
			CompanyID:   row.CompanyID,
			CompanyName: companyName[row.CompanyID],
			FlagKey:     row.FlagKey,
			Enabled:     row.Enabled,
			UpdatedAt:   row.UpdatedAt.Time.UTC().Format(time.RFC3339),
		})
	}

	return &UpsertOutput{Body: UpsertResponse{Flags: out}}, nil
}

func (h *Handler) touchCache(companyID int64, key string, enabled bool) {
	flags, _ := h.cache.Get(companyID)
	next := FlagMap{}
	for k, v := range flags {
		next[k] = v
	}
	next[key] = enabled
	h.cache.Set(companyID, next)
}

// Feature flags are managed from the Default Company only — the "control"
// company. Other admins can read their own flags but can't write.
const controlCompanyID int64 = 1

// checkAdmin: principal must be admin of the control company.
func (h *Handler) checkAdmin(ctx context.Context, selectedCompanyID, userID int64) error {
	if selectedCompanyID != controlCompanyID {
		return humax.Forbidden("feature flags can only be managed from the default company")
	}
	isAdmin, err := h.queries.FFIsAdmin(ctx, compiled.FFIsAdminParams{
		CompanyID: controlCompanyID,
		UserID:    userID,
	})
	if err != nil {
		return err
	}
	if !isAdmin {
		return humax.Forbidden("admin role required")
	}
	return nil
}

// refreshCache rebuilds the snapshot from DB. Cron hourly + once at startup.
func (h *Handler) refreshCache(ctx context.Context) error {
	rows, err := h.queries.FFListAll(ctx)
	if err != nil {
		return err
	}
	byCompany := map[int64]FlagMap{}
	for _, row := range rows {
		m, ok := byCompany[row.CompanyID]
		if !ok {
			m = FlagMap{}
			byCompany[row.CompanyID] = m
		}
		m[row.FlagKey] = row.Enabled
	}
	h.cache.Replace(byCompany)
	return nil
}
