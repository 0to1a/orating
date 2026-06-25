package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"project/internal/platform/authctx"
	"project/internal/platform/cache"
	"project/internal/platform/compiled"
	"project/internal/platform/httpx"
)

// Middleware resolves session tokens to *authctx.Principal. Cache TTL is
// 5 minutes; logout invalidates eagerly via Module.Invalidate.
type Middleware struct {
	queries *compiled.Queries
	cache   *cache.Cache[string, *authctx.Principal]
}

const sessionCacheTTL = 5 * time.Minute

// Wrap is tolerant: empty Bearer or sk_ prefix passes through to the next
// resolver. Non-sk_ tokens that fail resolution return 401.
func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" || strings.HasPrefix(token, "sk_") {
			next.ServeHTTP(w, r)
			return
		}

		p, err := m.resolve(r.Context(), token)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid session token")
			return
		}

		next.ServeHTTP(w, r.WithContext(authctx.WithPrincipal(r.Context(), p)))
	})
}

func (m *Middleware) resolve(ctx context.Context, token string) (*authctx.Principal, error) {
	if p, ok := m.cache.Get(token); ok {
		return p, nil
	}
	row, err := m.queries.AuthFindUserBySessionToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	p := principalFromRow(row, token)
	m.cache.SetWithTTL(token, p, sessionCacheTTL)
	return p, nil
}

// invalidate is called from logout (via Module.Invalidate) for instant effect.
func (m *Middleware) invalidate(token string) {
	m.cache.Delete(token)
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return ""
	}
	return strings.TrimSpace(h[len(prefix):])
}

func principalFromRow(row compiled.User, token string) *authctx.Principal {
	p := &authctx.Principal{
		UserID: row.ID,
		Email:  row.Email,
		Name:   row.Name,
		Source: authctx.SourceSession,
		Token:  token,
	}
	if row.SelectedCompanyID.Valid {
		p.SelectedCompanyID = row.SelectedCompanyID.Int64
	}
	return p
}
