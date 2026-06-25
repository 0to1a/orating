package featureflag

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"project/internal/platform"
	"project/internal/platform/cache"
	"project/internal/platform/cron"
	"project/internal/platform/humax"
)

func Setup(ctx context.Context, deps platform.Deps) {
	c := cache.New[int64, FlagMap](ctx)
	h := &Handler{queries: deps.Queries, cache: c}

	// Hourly refresh; the first run fires at Start so the cache is warm
	// before the first request arrives.
	deps.Scheduler.Register(cron.Job{
		Name:     "featureflag-refresh",
		Interval: time.Hour,
		Run:      h.refreshCache,
	})

	// Any authenticated user can read their own company's flags.
	huma.Register(deps.API, huma.Operation{
		OperationID: "list-feature-flags",
		Method:      http.MethodGet,
		Path:        "/api/feature-flags",
		Summary:     "List feature flags for the current user's selected company",
		Tags:        []string{"feature-flag"},
		Security:    humax.BearerAuth(),
	}, h.handleList)

	// Admin endpoints — gated in-handler by control-company + admin role.
	huma.Register(deps.API, huma.Operation{
		OperationID: "admin-list-feature-flags",
		Method:      http.MethodGet,
		Path:        "/api/feature-flags/admin",
		Summary:     "Admin: get full context (flags, companies, keys)",
		Tags:        []string{"feature-flag"},
		Security:    humax.BearerAuth(),
	}, h.handleAdminContext)

	huma.Register(deps.API, huma.Operation{
		OperationID: "admin-upsert-feature-flags",
		Method:      http.MethodPost,
		Path:        "/api/feature-flags/admin",
		Summary:     "Admin: upsert a flag across one or more companies",
		Tags:        []string{"feature-flag"},
		Security:    humax.BearerAuth(),
	}, h.handleUpsert)
}
