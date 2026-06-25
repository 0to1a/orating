package authctx

import (
	"net/http"

	"project/internal/platform/httpx"
)

// RequireAuth returns 401 if no upstream resolver attached a Principal.
// Place last in the chain after all tolerant resolvers.
var RequireAuth httpx.Middleware = httpx.MiddlewareFunc(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := TryPrincipalFromContext(r.Context()); !ok {
			httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
})

// RequireSelectedCompany 400s if Principal has no SelectedCompanyID.
// Chain after RequireAuth.
var RequireSelectedCompany httpx.Middleware = httpx.MiddlewareFunc(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := PrincipalFromContext(r.Context())
		if p.SelectedCompanyID == 0 {
			httpx.WriteError(w, http.StatusBadRequest, "no selected company")
			return
		}
		next.ServeHTTP(w, r)
	})
})
