package apikey

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"project/internal/platform"
	"project/internal/platform/compiled"
	"project/internal/platform/httpx"
	"project/internal/platform/humax"
)

type Module struct {
	Middleware  httpx.Middleware
	internalMw *Middleware
}

// NewModule creates apikey middleware without registering any routes.
// Call r.Use(m.Middleware.Wrap) before calling RegisterRoutes.
func NewModule(queries *compiled.Queries) *Module {
	mw := newMiddleware(queries)
	return &Module{Middleware: mw, internalMw: mw}
}

// RegisterRoutes wires all apikey HTTP routes.
func (m *Module) RegisterRoutes(deps platform.Deps) {
	h := &Handler{queries: deps.Queries}

	huma.Register(deps.API, huma.Operation{
		OperationID: "list-api-keys",
		Method:      http.MethodGet,
		Path:        "/api/apikeys",
		Summary:     "List API keys for selected company",
		Tags:        []string{"apikeys"},
		Security:    humax.BearerAuth(),
	}, h.handleList)

	huma.Register(deps.API, huma.Operation{
		OperationID:   "create-api-key",
		Method:        http.MethodPost,
		Path:          "/api/apikeys",
		Summary:       "Create a new API key",
		Tags:          []string{"apikeys"},
		Security:      humax.BearerAuth(),
		DefaultStatus: http.StatusCreated,
	}, h.handleCreate)

	huma.Register(deps.API, huma.Operation{
		OperationID:   "revoke-api-key",
		Method:        http.MethodPost,
		Path:          "/api/apikeys/{id}/revoke",
		Summary:       "Revoke an API key",
		Tags:          []string{"apikeys"},
		Security:      humax.BearerAuth(),
		DefaultStatus: http.StatusNoContent,
	}, h.handleRevoke)
}

// Setup is a convenience wrapper used by gen-spec where middleware order doesn't matter.
func Setup(_ context.Context, deps platform.Deps) *Module {
	mod := NewModule(deps.Queries)
	mod.RegisterRoutes(deps)
	return mod
}
