package httpx

import "net/http"

// Middleware is the single contract for wrapping http.Handler. Domains that
// expose middleware (auth, apikey) implement this and return it via Module.
type Middleware interface {
	Wrap(http.Handler) http.Handler
}

type MiddlewareFunc func(http.Handler) http.Handler

func (f MiddlewareFunc) Wrap(h http.Handler) http.Handler { return f(h) }

// Chain composes left-to-right: Chain(a, b).Wrap(h) → a(b(h)).
func Chain(mws ...Middleware) Middleware {
	return MiddlewareFunc(func(h http.Handler) http.Handler {
		for i := len(mws) - 1; i >= 0; i-- {
			h = mws[i].Wrap(h)
		}
		return h
	})
}

// Passthrough is the no-op middleware. Use as a Deps default before the
// real middleware is built.
var Passthrough Middleware = MiddlewareFunc(func(h http.Handler) http.Handler { return h })
