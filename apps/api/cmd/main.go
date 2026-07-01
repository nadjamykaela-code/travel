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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nadjamykaela-code/travel/apps/api/middleware"
	"github.com/nadjamykaela-code/travel/apps/api/routes"
	"github.com/nadjamykaela-code/travel/apps/api/service"
	"github.com/nadjamykaela-code/travel/internal/config"
	"github.com/nadjamykaela-code/travel/internal/firestore"
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

	filterStore := service.NewFilterStoreFirestore(fsClient.Firestore())
	filterSvc := service.NewFilterService(filterStore)

	var authSvc service.AuthService
	authSvc, authErr := service.NewAuthServiceFirebase(context.Background(), cfg.FCMCredentialsPath)
	if authErr != nil {
		slog.Warn("firebase auth not available, auth endpoints will reject all requests", "error", authErr)
		authSvc = service.NewNoopAuthService()
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.Use(middleware.TraceMiddleware())
	router.Use(middleware.MetricsMiddleware())
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Trace-ID"},
		AllowCredentials: true,
	}))

	deps := &routes.Dependencies{
		FilterService: filterSvc,
		AuthService:   authSvc,
	}
	routes.RegisterRoutes(router, deps)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("server starting",
			"port", cfg.Port,
			"cloud_run_service", os.Getenv("K_SERVICE"),
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("shutting down server", "signal", sig.String())

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped gracefully")
}
