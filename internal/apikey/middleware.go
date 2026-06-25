package apikey

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"project/internal/platform/authctx"
	"project/internal/platform/compiled"
	"project/internal/platform/httpx"
)

const (
	tokenPrefix      = "sk_"
	lastUsedThrottle = time.Minute
)

type Middleware struct {
	queries *compiled.Queries

	// last_used_at update is async + throttled to one DB write per key per minute.
	mu       sync.Mutex
	lastUsed map[string]int64
}

func newMiddleware(queries *compiled.Queries) *Middleware {
	return &Middleware{
		queries:  queries,
		lastUsed: make(map[string]int64),
	}
}

// Wrap is tolerant: tokens without the sk_ prefix pass through to the
// next resolver in the chain (e.g. session auth).
func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if !strings.HasPrefix(token, tokenPrefix) {
			next.ServeHTTP(w, r)
			return
		}
		hash := hashToken(token)
		row, err := m.queries.APIKeyResolve(r.Context(), hash)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httpx.WriteError(w, http.StatusUnauthorized, "invalid api key")
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, "internal error")
			return
		}

		m.maybeTouchLastUsed(row.ApiKeyID, hash)

		p := &authctx.Principal{
			UserID: row.UserID,
			Email:  row.UserEmail,
			Name:   row.UserName,
			Source: authctx.SourceAPIKey,
			Token:  hash,
		}
		if row.SelectedCompanyID.Valid {
			p.SelectedCompanyID = row.SelectedCompanyID.Int64
		}

		next.ServeHTTP(w, r.WithContext(authctx.WithPrincipal(r.Context(), p)))
	})
}

func (m *Middleware) maybeTouchLastUsed(keyID int64, hash string) {
	now := time.Now().UnixNano()
	m.mu.Lock()
	last := m.lastUsed[hash]
	if now-last < int64(lastUsedThrottle) {
		m.mu.Unlock()
		return
	}
	m.lastUsed[hash] = now
	m.mu.Unlock()

	go func() {
		_ = m.queries.APIKeyTouchLastUsed(context.Background(), keyID)
	}()
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return ""
	}
	return strings.TrimSpace(h[len(prefix):])
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
