// Package authctx provides Principal — the request-scoped caller identity
// shared across resolvers (session auth, api key, future SSO/OAuth).
//
// Lives in platform/ because multiple domains produce it; consumer domains
// read via PrincipalFromContext without depending on which resolver wrote it.
package authctx

import "context"

type Source uint8

const (
	SourceUnknown Source = iota
	SourceSession
	SourceAPIKey
)

func (s Source) String() string {
	switch s {
	case SourceSession:
		return "session"
	case SourceAPIKey:
		return "apikey"
	default:
		return "unknown"
	}
}

// Principal is a snapshot of the caller. Fields are populated by the
// middleware that resolved the request; reads in downstream handlers don't
// reflect mid-request DB mutations.
type Principal struct {
	UserID            int64
	Email             string
	Name              string
	SelectedCompanyID int64 // 0 if no company selected

	Source Source
	Token  string // credential used; useful for logout/revoke
}

type ctxKey int

const principalKey ctxKey = iota

func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

// PrincipalFromContext panics if no principal is attached — that's a wiring
// bug. Use TryPrincipalFromContext for optionally-authenticated routes.
func PrincipalFromContext(ctx context.Context) *Principal {
	p, ok := ctx.Value(principalKey).(*Principal)
	if !ok {
		panic("authctx: PrincipalFromContext called on unauthenticated request")
	}
	return p
}

func TryPrincipalFromContext(ctx context.Context) (*Principal, bool) {
	p, ok := ctx.Value(principalKey).(*Principal)
	return p, ok
}
