package company

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"project/internal/platform"
	"project/internal/platform/humax"
)

// Setup wires the company domain. No return value — doesn't expose
// middleware/handles to other domains.
func Setup(_ context.Context, deps platform.Deps) {
	h := &Handler{
		queries:        deps.Queries,
		invalidateAuth: deps.AuthInvalidator,
	}

	// Cross-company / bootstrap routes — selected company optional.
	huma.Register(deps.API, huma.Operation{
		OperationID: "list-companies",
		Method:      http.MethodGet,
		Path:        "/api/companies",
		Summary:     "List companies for the current user",
		Tags:        []string{"company"},
		Security:    humax.BearerAuth(),
	}, h.handleList)

	huma.Register(deps.API, huma.Operation{
		OperationID:   "create-company",
		Method:        http.MethodPost,
		Path:          "/api/companies",
		Summary:       "Create a new company",
		Tags:          []string{"company"},
		Security:      humax.BearerAuth(),
		DefaultStatus: http.StatusCreated,
	}, h.handleCreate)

	huma.Register(deps.API, huma.Operation{
		OperationID:   "select-company",
		Method:        http.MethodPost,
		Path:          "/api/companies/{id}/select",
		Summary:       "Select active company",
		Tags:          []string{"company"},
		Security:      humax.BearerAuth(),
		DefaultStatus: http.StatusNoContent,
	}, h.handleSelect)

	huma.Register(deps.API, huma.Operation{
		OperationID: "list-company-roles",
		Method:      http.MethodGet,
		Path:        "/api/companies/roles",
		Summary:     "List available company roles",
		Tags:        []string{"company"},
		Security:    humax.BearerAuth(),
	}, h.handleListRoles)

	// Tenant-scoped routes — require selected company (enforced in-handler via RequireSelectedCompany).
	huma.Register(deps.API, huma.Operation{
		OperationID: "list-company-members",
		Method:      http.MethodGet,
		Path:        "/api/companies/members",
		Summary:     "List members of the selected company (admin only)",
		Tags:        []string{"company"},
		Security:    humax.BearerAuth(),
	}, h.handleListMembers)

	huma.Register(deps.API, huma.Operation{
		OperationID:   "invite-company-member",
		Method:        http.MethodPost,
		Path:          "/api/companies/invite",
		Summary:       "Invite a user to the selected company",
		Tags:          []string{"company"},
		Security:      humax.BearerAuth(),
		DefaultStatus: http.StatusCreated,
	}, h.handleInvite)

	huma.Register(deps.API, huma.Operation{
		OperationID:   "update-company-member-role",
		Method:        http.MethodPut,
		Path:          "/api/companies/members/{userId}/role",
		Summary:       "Update a member's role in the selected company",
		Tags:          []string{"company"},
		Security:      humax.BearerAuth(),
		DefaultStatus: http.StatusNoContent,
	}, h.handleUpdateRole)

	huma.Register(deps.API, huma.Operation{
		OperationID:   "remove-company-member",
		Method:        http.MethodDelete,
		Path:          "/api/companies/members/{userId}",
		Summary:       "Remove a member from the selected company",
		Tags:          []string{"company"},
		Security:      humax.BearerAuth(),
		DefaultStatus: http.StatusNoContent,
	}, h.handleRemoveMember)
}
