package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
	firebaseauth "firebase.google.com/go/v4/auth"
	"github.com/danielgtaylor/huma/v2"
	"google.golang.org/api/option"

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

// initFirebase creates a Firebase Auth client from env vars. Returns nil (no
// error) when Firebase is not configured — the endpoint will return 422.
func initFirebase(ctx context.Context) (*firebaseauth.Client, error) {
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		return nil, nil
	}
	var opts []option.ClientOption
	if b64 := os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON"); b64 != "" {
		jsonCreds, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("firebase: decode service account: %w", err)
		}
		opts = append(opts, option.WithCredentialsJSON(jsonCreds))
	}
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID}, opts...)
	if err != nil {
		return nil, fmt.Errorf("firebase init: %w", err)
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase auth client: %w", err)
	}
	return client, nil
}

// RegisterRoutes wires all auth HTTP routes.
func (m *Module) RegisterRoutes(deps platform.Deps) error {
	tmpl, err := parseTemplates()
	if err != nil {
		return fmt.Errorf("auth: parse templates: %w", err)
	}

	fbAuth, err := initFirebase(context.Background())
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	h := &Handler{
		queries:      deps.Queries,
		mailer:       deps.Mailer,
		mw:           m.internalMw,
		cache:        m.tokenCache,
		tmpl:         tmpl,
		firebaseAuth: fbAuth,
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

	huma.Register(deps.API, huma.Operation{
		OperationID: "google-login",
		Method:      http.MethodPost,
		Path:        "/api/auth/google",
		Summary:     "Login with Google via Firebase ID token",
		Tags:        []string{"auth"},
	}, h.handleGoogleLogin)

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
