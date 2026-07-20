package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bitrok/bitrok/server/internal/config"
	"github.com/bitrok/bitrok/server/internal/relay"
	"github.com/bitrok/bitrok/server/internal/store"
)

// startMaintenance bounds persisted request-log growth without tying cleanup
// to request traffic.
func startMaintenance(st store.Store, logRetention time.Duration) (stop func()) {
	ticker := time.NewTicker(1 * time.Hour)
	done := make(chan struct{})
	var wg sync.WaitGroup
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := st.CleanupTunnelLogs(ctx, logRetention); err != nil {
			slog.Warn("tunnel log cleanup failed", "error", err)
		}
	}
	cleanup()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ticker.C:
				cleanup()
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	return func() {
		close(done)
		wg.Wait()
	}
}

// Run bootstraps the server from config and starts listening with graceful shutdown.
func Run(cfg *config.Config) error {
	logger := newLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	if len(cfg.AuthTokens) > 0 {
		slog.Warn("static auth tokens are deprecated; migrate to JWT authentication")
	}

	st, err := store.NewSQLite(cfg.DBPath, cfg.DBMaxOpenConns, cfg.DBMaxIdleConns,
		time.Duration(cfg.DBConnLifetime)*time.Second, time.Duration(cfg.DBConnIdleTime)*time.Second)
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}
	defer st.Close()

	hub := relay.NewHub()
	router, rateLimiter := NewRouter(cfg, st, hub)

	stopMaintenance := startMaintenance(st, time.Duration(cfg.LogRetentionHours)*time.Hour)
	defer stopMaintenance()

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router,
		ReadTimeout:       time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:       time.Duration(cfg.IdleTimeout) * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeout) * time.Second,
	}

	// Start server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		if cfg.TLSCertPath != "" && cfg.TLSKeyPath != "" {
			slog.Info("server listening with TLS", "addr", srv.Addr, "db", cfg.DBPath, "domain", cfg.Domain)
			errCh <- srv.ListenAndServeTLS(cfg.TLSCertPath, cfg.TLSKeyPath)
		} else {
			slog.Warn("Server running without TLS. Set BITROK_TLS_CERT and BITROK_TLS_KEY to enable HTTPS.")
			slog.Info("server listening", "addr", srv.Addr, "db", cfg.DBPath, "domain", cfg.Domain)
			errCh <- srv.ListenAndServe()
		}
	}()

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-sigCh:
		slog.Info("shutdown signal received", "signal", sig.String())
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownTimeout)*time.Second)
	defer cancel()

	if rateLimiter != nil {
		rateLimiter.Stop()
	}

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	slog.Info("server stopped gracefully")
	return nil
}

func newLogger(level string) *slog.Logger {
	var lv slog.Level
	switch level {
	case "debug":
		lv = slog.LevelDebug
	case "warn":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: lv}
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
