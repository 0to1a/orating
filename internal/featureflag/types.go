// Package featureflag — per-company FE-gated feature toggles.
//
// FE-gated only: BE never enforces these. Management is centralized to
// admins of the Default Company (id=1); other companies can read their own
// flags but can't toggle. Snapshot cache refreshed hourly; toggles push
// directly for instant effect.
package featureflag

// FlagMap is the consumer-facing flat view. Unknown keys → false on the FE.
type FlagMap = map[string]bool

type FlagListResponse struct {
	Flags FlagMap `json:"flags"`
}

// AdminFlag is a flag with its company name + timestamp, for admin views.
type AdminFlag struct {
	ID          int64  `json:"id"`
	CompanyID   int64  `json:"companyId"`
	CompanyName string `json:"companyName"`
	FlagKey     string `json:"flagKey"`
	Enabled     bool   `json:"enabled"`
	UpdatedAt   string `json:"updatedAt"`
}

// CompanyOption is the (id, name) pair for the multi-select target picker.
type CompanyOption struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// AdminContextResponse bundles the admin page's data into one fetch.
type AdminContextResponse struct {
	Flags         []AdminFlag     `json:"flags"`
	Companies     []CompanyOption `json:"companies"`
	SuggestedKeys []string        `json:"suggestedKeys"`
}

// UpsertRequest applies a flag to one or more companies.
// AllCompanies=true ignores CompanyIDs and targets every active company.
type UpsertRequest struct {
	CompanyIDs   []int64 `json:"companyIds"`
	AllCompanies bool    `json:"allCompanies"`
	FlagKey      string  `json:"flagKey"`
	Enabled      bool    `json:"enabled"`
}

type UpsertResponse struct {
	Flags []AdminFlag `json:"flags"`
}

// ===== huma Input/Output types =====

type FlagListOutput struct {
	Body FlagListResponse
}

type AdminContextOutput struct {
	Body AdminContextResponse
}

type UpsertInput struct {
	Body UpsertRequest
}

type UpsertOutput struct {
	Body UpsertResponse
}
