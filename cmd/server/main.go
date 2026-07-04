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

	"github.com/abubakar508/dispatch-ops/internal/config"
	"github.com/abubakar508/dispatch-ops/internal/geo"
	"github.com/abubakar508/dispatch-ops/internal/geocode"
	"github.com/abubakar508/dispatch-ops/internal/handlers"
	"github.com/abubakar508/dispatch-ops/internal/logging"
	"github.com/abubakar508/dispatch-ops/internal/optimizer"
	"github.com/abubakar508/dispatch-ops/internal/osrm"
	"github.com/abubakar508/dispatch-ops/internal/services"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server exited with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := logging.New(cfg.LogLevel, cfg.Environment)
	slog.SetDefault(logger)

	osrmClient := osrm.NewHTTPClient(cfg.OSRMBaseURL, cfg.OSRMTimeout)
	solver := optimizer.NewLocalSearchSolver()
	indexer := geo.NewIndexer(cfg.H3Resolution)
	routeService := services.NewRouteService(osrmClient, solver, indexer)
	geocoder := geocode.NewNominatimClient(cfg.NominatimBaseURL, cfg.NominatimAgent, cfg.GeocodeTimeout)
	handler := handlers.New(routeService, geocoder, logger)

	router := handlers.NewRouter(handler, cfg.AllowedOrigin)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  90 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("starting api server",
			slog.String("port", cfg.Port),
			slog.String("osrm", cfg.OSRMBaseURL),
			slog.String("environment", cfg.Environment),
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return err
	case sig := <-shutdown:
		logger.Info("shutdown signal received", slog.String("signal", sig.String()))
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			_ = server.Close()
			return err
		}
		logger.Info("server stopped cleanly")
	}

	return nil
}
