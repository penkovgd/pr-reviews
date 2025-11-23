package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/penkovgd/pr-reviews/internal/adapters/db"
	"github.com/penkovgd/pr-reviews/internal/adapters/rest"
	"github.com/penkovgd/pr-reviews/internal/config"
	"github.com/penkovgd/pr-reviews/internal/core"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()
	cfg := config.MustLoad(configPath)

	log := mustMakeLogger(cfg.LogLevel)

	if err := run(cfg, log); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	log.Info("starting server")
	log.Debug("debug messages are enabled")

	// database adapter
	db, err := db.New(log, cfg.DBUrl)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	if err := db.Migrate(); err != nil {
		return fmt.Errorf("migrate db: %w", err)
	}

	// services
	teamService := core.NewTeamService(db, db)
	userService := core.NewUserService(db, db)
	prService := core.NewPullRequestService(db, db)

	// rest adapter
	mux := http.NewServeMux()
	// Teams
	mux.Handle("POST /team/add", rest.NewAddTeamHandler(log, teamService))
	mux.Handle("GET /team/get", rest.NewGetTeamHandler(log, teamService))
	// Users
	mux.Handle("POST /users/setIsActive", rest.NewSetUserActiveHandler(log, userService))
	mux.Handle("GET /users/getReview", rest.NewGetUserReviewHandler(log, userService))
	// PullRequests
	mux.Handle("POST /pullRequest/create", rest.NewCreatePRHandler(log, prService))
	mux.Handle("POST /pullRequest/merge", rest.NewMergePRHandler(log, prService))
	mux.Handle("POST /pullRequest/reassign", rest.NewReassignReviewerHandler(log, prService))
	// bonus: statistics
	mux.Handle("GET /stats/user-assignments", rest.NewUserAssignmentStatsHandler(log, db))

	server := http.Server{
		Addr:        cfg.HTTPConfig.Address,
		ReadTimeout: cfg.HTTPConfig.Timeout,
		Handler:     mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Error("erroneous shutdown", "error", err)
		}
	}()

	log.Info("running HTTP server", "address", cfg.HTTPConfig.Address)
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server closed unexpectedly: %w", err)
		}
	}
	return nil
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		panic("unknown log level: " + logLevel)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
