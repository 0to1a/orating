package humax

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"project/internal/platform/authctx"
)

// BearerAuth returns the huma security requirement for Bearer token auth.
func BearerAuth() []map[string][]string {
	return []map[string][]string{{"bearerAuth": {}}}
}

// SecuritySchemes returns the OpenAPI security scheme definitions.
func SecuritySchemes() map[string]*huma.SecurityScheme {
	return map[string]*huma.SecurityScheme{
		"bearerAuth": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "opaque",
		},
	}
}

// RequireAuth extracts the Principal from context. Returns huma 401 if not present.
// Call at the start of every protected handler.
func RequireAuth(ctx context.Context) (authctx.Principal, error) {
	p, ok := authctx.TryPrincipalFromContext(ctx)
	if !ok {
		return authctx.Principal{}, huma.Error401Unauthorized("unauthorized", nil)
	}
	return *p, nil
}

// RequireSelectedCompany extracts the Principal and returns 400 if no company is selected.
// Use for endpoints that are scoped to a selected company.
func RequireSelectedCompany(ctx context.Context) (authctx.Principal, error) {
	p, err := RequireAuth(ctx)
	if err != nil {
		return p, err
	}
	if p.SelectedCompanyID == 0 {
		return p, huma.Error400BadRequest("no selected company", nil)
	}
	return p, nil
}

// AdaptMiddleware is a no-op identity function — chi middleware is already http.Handler → http.Handler.
func AdaptMiddleware(mw func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return mw
}
