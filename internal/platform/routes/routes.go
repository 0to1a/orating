// Package routes registers all domain routes onto deps.API.
// It is used by cmd/gen-spec to produce an OpenAPI spec without a live DB.
package routes

import (
	"context"

	"project/internal/apikey"
	"project/internal/auth"
	"project/internal/company"
	"project/internal/featureflag"
	"project/internal/platform"
	"project/internal/rating"
	"project/internal/user"
)

// RegisterAll registers all domain routes on deps.API.
// deps.Queries may be nil — this is safe for schema generation (gen-spec)
// because handlers are registered but never executed during spec output.
// deps.Scheduler must be non-nil (use cron.New(nil) for gen-spec).
func RegisterAll(ctx context.Context, deps platform.Deps) error {
	_, err := auth.Setup(ctx, deps)
	if err != nil {
		return err
	}
	apikey.Setup(ctx, deps)
	user.Setup(ctx, deps)
	company.Setup(ctx, deps)
	featureflag.Setup(ctx, deps)
	rating.Setup(ctx, deps)
	return nil
}
