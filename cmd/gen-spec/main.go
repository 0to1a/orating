// Package main generates openapi.json from the registered huma routes.
// It does not require a database connection — handlers are registered for
// schema introspection only and are never executed.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	"project/internal/platform"
	"project/internal/platform/cron"
	"project/internal/platform/humax"
	"project/internal/platform/routes"
)

func main() {
	outPath := "openapi.json"
	if len(os.Args) > 1 {
		outPath = os.Args[1]
	}

	r := chi.NewRouter()
	humaConfig := huma.DefaultConfig("Project API", "1.0.0")
	humaConfig.Components.SecuritySchemes = humax.SecuritySchemes()
	api := humachi.New(r, humaConfig)

	// Scheduler must be non-nil because featureflag.Setup calls Register on it.
	// Start is never called here, so no goroutines are launched.
	scheduler := cron.New(nil)

	deps := platform.Deps{
		Router:    r,
		API:       api,
		Scheduler: scheduler,
		// Queries, Mailer, Logger, etc. are nil/zero — handlers never execute.
	}

	ctx := context.Background()
	if err := routes.RegisterAll(ctx, deps); err != nil {
		fmt.Fprintf(os.Stderr, "error registering routes: %v\n", err)
		os.Exit(1)
	}

	spec, err := json.MarshalIndent(api.OpenAPI(), "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling spec: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outPath, spec, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", outPath, err)
		os.Exit(1)
	}

	fmt.Printf("wrote %s\n", outPath)
}
