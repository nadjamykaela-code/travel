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

	"github.com/joho/godotenv"
	"github.com/nadjamykaela-code/travel/apps/worker/job"
	"github.com/nadjamykaela-code/travel/apps/worker/service"
	"github.com/nadjamykaela-code/travel/internal/config"
	"github.com/nadjamykaela-code/travel/internal/firestore"
	"github.com/nadjamykaela-code/travel/pkg/clients"
	"github.com/nadjamykaela-code/travel/pkg/notifications"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		slog.Warn(".env file not found, using system environment variables")
	}

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fsClient, err := firestore.New(ctx, cfg.ProjectID, cfg.FCMCredentialsPath)
	if err != nil {
		slog.Error("failed to connect to Firestore", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := fsClient.Close(); err != nil {
			slog.Error("firestore close error", "error", err)
		}
	}()
	slog.Info("firestore connected", "project_id", cfg.ProjectID)

	firestoreDB := fsClient.Firestore()

	maxReqs := cfg.MaxRequestsPerMonth
	if maxReqs <= 0 {
		maxReqs = 100
	}

	filterRepo := service.NewFilterRepo(firestoreDB)
	indicativeSearcher := clients.NewSkyscannerSearcher(cfg.SkyscannerAPIKey, cfg.SkyscannerBaseURL)
	livePricingClient := clients.NewLivePricingClient(cfg.SkyscannerAPIKey, cfg.SkyscannerBaseURL)
	searcher := clients.NewFailoverSearcher(livePricingClient, indicativeSearcher)
	engine := service.NewEngine()
	limiter := service.NewLimiter(maxReqs)

	var pushNotifier *notifications.FCMNotifier
	if cfg.FCMCredentialsPath != "" {
		var err error
		pushNotifier, err = notifications.NewFCMNotifier(cfg.FCMCredentialsPath)
		if err != nil {
			slog.Warn("FCM not available", "error", err)
		}
	}

	emailClient := notifications.NewSendGrid(cfg.SendGridAPIKey)
	notifier := service.NewNotifier(emailClient, pushNotifier, logger)
	history := service.NewHistoryStore(firestoreDB)

	runner := job.NewRunner(filterRepo, searcher, engine, notifier, history, limiter, logger)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("POST /run", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		slog.InfoContext(r.Context(), "worker run triggered")

		runCtx, runCancel := context.WithTimeout(r.Context(), 10*time.Minute)
		defer runCancel()

		if err := runner.Run(runCtx); err != nil {
			slog.ErrorContext(r.Context(), "worker run failed", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status":"error","message":"worker run failed"}`))
			return
		}

		slog.InfoContext(r.Context(), "worker run completed",
			"duration", time.Since(start).String(),
		)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.Handle("GET /metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("worker HTTP server starting",
			"port", cfg.Port,
			"cloud_run_service", os.Getenv("K_SERVICE"),
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	slog.Info("shutting down worker", "signal", sig.String())
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
	slog.Info("worker stopped")
}
