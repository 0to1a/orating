package platform

import (
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"

	"project/internal/platform/compiled"
	"project/internal/platform/cron"
	"project/internal/platform/httpx"
	"project/internal/platform/mailer"
)

type Deps struct {
	Router          chi.Router
	API             huma.API
	Scheduler       *cron.Scheduler
	Logger          *slog.Logger
	Queries         *compiled.Queries
	Mailer          mailer.Mailer
	APIKeyMw        httpx.Middleware
	AuthInvalidator func(token string)
}

func (d Deps) WithAPIKeyMw(mw httpx.Middleware) Deps {
	d.APIKeyMw = mw
	return d
}

func (d Deps) WithAuthInvalidator(fn func(string)) Deps {
	d.AuthInvalidator = fn
	return d
}
