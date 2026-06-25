package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"project/internal/platform"
	"project/internal/platform/authctx"
	"project/internal/platform/cache"
	"project/internal/platform/compiled"
	"project/internal/platform/httpx"
	"project/internal/platform/humax"
)

// Module exposes the middleware and cache invalidator for use by main.go.
type Module struct {
	Middleware httpx.Middleware
	Invalidate func(token string)
	internalMw *Middleware
	tokenCache *cache.Cache[string, *authctx.Principal]
}

// NewModule creates auth middleware without registering any routes.
// Call r.Use(m.Middleware.Wrap) before calling RegisterRoutes.
func NewModule(ctx context.Context, queries *compiled.Queries) (*Module, error) {
	tokenCache := cache.New[string, *authctx.Principal](ctx)
	mw := &Middleware{queries: queries, cache: tokenCache}
	return &Module{
		Middleware: mw,
		Invalidate: mw.invalidate,
		internalMw: mw,
		tokenCache: tokenCache,
	}, nil
}

// RegisterRoutes wires all auth HTTP routes.
func (m *Module) RegisterRoutes(deps platform.Deps) error {
	tmpl, err := parseTemplates()
	if err != nil {
		return fmt.Errorf("auth: parse templates: %w", err)
	}

	h := &Handler{
		queries: deps.Queries,
		mailer:  deps.Mailer,
		mw:      m.internalMw,
		cache:   m.tokenCache,
		tmpl:    tmpl,
	}

	huma.Register(deps.API, huma.Operation{
		OperationID:   "login-request",
		Method:        http.MethodPost,
		Path:          "/api/auth/login-request",
		Summary:       "Request OTP login",
		Tags:          []string{"auth"},
		DefaultStatus: http.StatusNoContent,
	}, h.handleLoginRequest)

	huma.Register(deps.API, huma.Operation{
		OperationID: "login-verify",
		Method:      http.MethodPost,
		Path:        "/api/auth/login-verify",
		Summary:     "Verify OTP and issue session token",
		Tags:        []string{"auth"},
	}, h.handleLoginVerify)

	huma.Register(deps.API, huma.Operation{
		OperationID:   "logout",
		Method:        http.MethodPost,
		Path:          "/api/auth/logout",
		Summary:       "Logout",
		Tags:          []string{"auth"},
		Security:      humax.BearerAuth(),
		DefaultStatus: http.StatusNoContent,
	}, h.handleLogout)

	huma.Register(deps.API, huma.Operation{
		OperationID: "get-me",
		Method:      http.MethodGet,
		Path:        "/api/auth/me",
		Summary:     "Get current user profile",
		Tags:        []string{"auth"},
		Security:    humax.BearerAuth(),
	}, h.handleMe)

	return nil
}

// Setup is a convenience wrapper used by gen-spec where middleware order doesn't matter.
func Setup(ctx context.Context, deps platform.Deps) (*Module, error) {
	mod, err := NewModule(ctx, deps.Queries)
	if err != nil {
		return nil, err
	}
	return mod, mod.RegisterRoutes(deps)
}
