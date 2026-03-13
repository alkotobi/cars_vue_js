package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/yourname/tunnel/internal/client"
	"github.com/yourname/tunnel/internal/config"
	"github.com/yourname/tunnel/internal/log"
)

func main() {
	domain := flag.String("domain", config.BaseDomain, "Base domain (e.g. merhab.com)")
	control := flag.Int("control", config.DefaultControlPort, "Control port (direct TCP)")
	local := flag.String("local", "", "Local URL or dist/ directory")
	subdomain := flag.String("subdomain", "", "Tunnel subdomain (e.g. cars)")
	logLevel := flag.String("log-level", "info", "debug|info|warn|error")
	version := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *version {
		fmt.Println(config.Version)
		return
	}
	if *local == "" {
		fmt.Fprintln(os.Stderr, "-local is required (URL or dist/ directory)")
		os.Exit(1)
	}

	baseDomain := strings.TrimSpace(*domain)
	if baseDomain == "" {
		fmt.Fprintln(os.Stderr, "-domain must not be empty")
		os.Exit(1)
	}

	cfg := config.ClientDefaults()
	cfg.ControlHost = baseDomain
	cfg.ControlPort = *control
	cfg.Subdomain = strings.TrimSpace(*subdomain)
	cfg.LogLevel = *logLevel
	if err := cfg.Validate(false); err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	level := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	logger := log.New(os.Stdout, level)

	// Resolve local target: URL or dist directory served via DistServer.
	localArg := strings.TrimSpace(*local)
	var (
		localURL string
		ds       *client.DistServer
	)
	if strings.HasPrefix(localArg, "http://") || strings.HasPrefix(localArg, "https://") {
		localURL = localArg
	} else {
		var err error
		ds, err = client.NewDistServer(localArg, logger)
		if err != nil {
			logger.Error("cannot serve dist directory", "err", err)
			os.Exit(1)
		}
		if err := ds.Start(); err != nil {
			logger.Error("dist server start failed", "err", err)
			os.Exit(1)
		}
		localURL = ds.URL()
		logger.Info("serving dist", "dir", localArg, "url", localURL)
	}
	cfg.LocalURL = localURL

	fullDomain := baseDomain
	if cfg.Subdomain != "" {
		fullDomain = cfg.Subdomain + "." + baseDomain
	}
	controlAddr := fmt.Sprintf("%s:%d", baseDomain, cfg.ControlPort)

	fmt.Printf("╔══════════════════════════════════════════════╗\n")
	fmt.Printf("║  Tunnel Client v%s                        ║\n", config.Version)
	fmt.Printf("║  Domain:    %-30s ║\n", fullDomain)
	fmt.Printf("║  Control:   %-30s ║\n", controlAddr)
	fmt.Printf("║  Local:     %-30s ║\n", cfg.LocalURL)
	fmt.Printf("║  Note: control connects directly, not nginx ║\n")
	fmt.Printf("║  Waiting for registration...               ║\n")
	fmt.Printf("╚══════════════════════════════════════════════╝\n")

	tc := client.New(cfg, logger)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var shutdownCtx context.Context
	var cancelShutdown context.CancelFunc
	if ds != nil {
		shutdownCtx, cancelShutdown = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		defer ds.Shutdown(shutdownCtx)
	}

	if err := tc.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("client error", "err", err)
		os.Exit(1)
	}
}
