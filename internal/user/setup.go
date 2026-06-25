package user

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"project/internal/platform"
	"project/internal/platform/humax"
)

// Setup wires the user domain. No return value — domain doesn't expose
// middleware/handles to other domains (that's only auth + apikey).
func Setup(_ context.Context, deps platform.Deps) {
	h := &Handler{queries: deps.Queries}

	huma.Register(deps.API, huma.Operation{
		OperationID: "get-user-me",
		Method:      http.MethodGet,
		Path:        "/api/users/me",
		Summary:     "Get current user profile",
		Tags:        []string{"user"},
		Security:    humax.BearerAuth(),
	}, h.handleMe)

	huma.Register(deps.API, huma.Operation{
		OperationID: "update-user-me",
		Method:      http.MethodPut,
		Path:        "/api/users/me",
		Summary:     "Update current user profile",
		Tags:        []string{"user"},
		Security:    humax.BearerAuth(),
	}, h.handleUpdateMe)
}
