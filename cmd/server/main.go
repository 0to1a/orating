// Package main is the server entrypoint: load config, open pgxpool, run
// migrations, build Deps, wire each domain via *.Setup, mount the SPA, then
// start. Route/cache/cron details belong in domain setup.go, not here.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"

	"project/cmd/config"
	"project/internal/apikey"
	"project/internal/auth"
	"project/internal/company"
	"project/internal/featureflag"
	"project/internal/platform"
	"project/internal/rating"
	"project/internal/platform/compiled"
	"project/internal/platform/cron"
	"project/internal/platform/frontend"
	"project/internal/platform/httpx"
	"project/internal/platform/humax"
	"project/internal/platform/mailer"
	"project/internal/platform/migration"
	"project/internal/user"
)

func main() {
	cfg := config.Load()
	logger := newLogger(cfg.Env)
	slog.SetDefault(logger)

	if err := runMigrations(cfg.DatabaseURL, logger); err != nil {
		logger.Error("migration failed", "error", err)
		os.Exit(1)
	}

	bgCtx, bgCancel := context.WithCancel(context.Background())
	defer bgCancel()

	pool, err := pgxpool.New(bgCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("pgxpool open failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(bgCtx); err != nil {
		logger.Error("db ping failed", "error", err)
		os.Exit(1)
	}

	queries := compiled.New(pool)
	mail := mailer.New(cfg.ResendAPIKey, cfg.MailFrom)

	r := chi.NewRouter()
	scheduler := cron.New(logger)

	// Global middleware: recovery, logging, CORS.
	r.Use(middleware.Recoverer)
	r.Use(httpx.RequestLogger(logger))
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Build token-resolver middlewares BEFORE humachi.New registers docs routes.
	// chi panics if r.Use is called after any route is registered.
	authMod, err := auth.NewModule(bgCtx, queries)
	if err != nil {
		logger.Error("auth module init failed", "error", err)
		os.Exit(1)
	}
	apikeyMod := apikey.NewModule(queries)
	r.Use(apikeyMod.Middleware.Wrap) // tolerant: enriches ctx, never rejects
	r.Use(authMod.Middleware.Wrap)   // tolerant: enriches ctx, rejects only bad tokens

	// Create huma API (registers /docs + /openapi.json routes).
	humaConfig := huma.DefaultConfig("Project API", "1.0.0")
	humaConfig.Components.SecuritySchemes = humax.SecuritySchemes()
	api := humachi.New(r, humaConfig)

	deps := platform.Deps{
		Router:    r,
		API:       api,
		Scheduler: scheduler,
		Logger:    logger,
		Queries:   queries,
		Pool:      pool,
		Mailer:    mail,
		APIKeyMw:  apikeyMod.Middleware,
	}
	deps = deps.WithAuthInvalidator(authMod.Invalidate)

	// Register all domain routes.
	if err := authMod.RegisterRoutes(deps); err != nil {
		logger.Error("auth route registration failed", "error", err)
		os.Exit(1)
	}
	apikeyMod.RegisterRoutes(deps)
	user.Setup(bgCtx, deps)
	company.Setup(bgCtx, deps)
	featureflag.Setup(bgCtx, deps)
	rating.Setup(bgCtx, deps)

	// Health check.
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// SPA catch-all (must be last).
	r.Handle("/*", frontend.Handler())

	scheduler.Start(bgCtx)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("http server listening", "port", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("shutting down")
	bgCancel()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown error", "error", err)
	}
	scheduler.Wait()
	logger.Info("shutdown complete")
}

func newLogger(env string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	if env == "prod" {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}

// runMigrations applies all embedded SQL migrations. Idempotent.
func runMigrations(databaseURL string, logger *slog.Logger) error {
	d, err := iofs.New(migration.FS, "sql")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, databaseURL)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	logger.Info("migrations applied")
	return nil
}
