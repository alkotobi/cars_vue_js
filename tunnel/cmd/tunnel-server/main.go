// Command tunnel-server runs the tunnel proxy and control server.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"

	"github.com/yourname/tunnel/internal/admin"
	"github.com/yourname/tunnel/internal/config"
	"github.com/yourname/tunnel/internal/control"
	"github.com/yourname/tunnel/internal/log"
	"github.com/yourname/tunnel/internal/proxy"
	"github.com/yourname/tunnel/internal/registry"
	"github.com/yourname/tunnel/internal/wsproxy"
)

func main() {
	port := flag.Int("port", config.DefaultHTTPPort, "HTTP proxy port")
	controlPort := flag.Int("control", config.DefaultControlPort, "Control port")
	adminPort := flag.Int("admin", config.DefaultAdminPort, "Admin port")
	domain := flag.String("domain", config.BaseDomain, "Base domain (e.g. merhab.com)")
	workers := flag.Int("workers", 0, "Worker count, 0=GOMAXPROCS")
	logLevel := flag.String("log-level", "info", "debug|info|warn|error")
	pprof := flag.Bool("pprof", false, "Enable pprof on admin server")
	version := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *version {
		fmt.Println(config.Version)
		os.Exit(0)
	}

	cfg := config.ServerDefaults()
	cfg.HTTPPort = *port
	cfg.ControlPort = *controlPort
	cfg.AdminPort = *adminPort
	cfg.BaseDomain = *domain
	cfg.Workers = *workers
	cfg.LogLevel = *logLevel
	cfg.PProfEnabled = *pprof
	if err := cfg.Validate(true); err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	if cfg.Workers > 0 {
		runtime.GOMAXPROCS(cfg.Workers)
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

	workersStr := strconv.Itoa(runtime.GOMAXPROCS(0))
	if cfg.Workers > 0 {
		workersStr = strconv.Itoa(cfg.Workers)
	}
	fmt.Printf("╔══════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  Tunnel Server v%s (Go)                           ║\n", config.Version)
	fmt.Printf("║  Domain:   *.%s%-35s║\n", cfg.BaseDomain, "")
	fmt.Printf("║  HTTP:     0.0.0.0:%-5d   (nginx proxies here)        ║\n", cfg.HTTPPort)
	fmt.Printf("║  Control:  0.0.0.0:%-5d   (clients connect directly)  ║\n", cfg.ControlPort)
	fmt.Printf("║  Admin:    127.0.0.1:%-5d                             ║\n", cfg.AdminPort)
	fmt.Printf("║  Workers:  %-3s                                       ║\n", workersStr)
	fmt.Printf("║  TLS:      nginx → see deploy/nginx.conf.example      ║\n")
	fmt.Printf("╚══════════════════════════════════════════════════════╝\n")

	logger.Info("firewall: ensure port 3000 is open (control channel)")
	logger.Info("firewall: ensure port 83 is NOT exposed (nginx only)")
	logger.Info("firewall: ensure port 8083 is blocked externally (admin only)")

	reg := registry.New()
	prx := proxy.New(cfg, reg, logger)
	wsp := wsproxy.New(logger)
	prx.SetWSUpgrader(wsp)
	ctrl := control.NewHandler(cfg, reg, prx, wsp, logger)
	adm := admin.New(cfg, reg, logger)
	if err := adm.Start(); err != nil {
		logger.Error("admin start", "err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("/", prx)
	srv := &http.Server{Addr: ":" + strconv.Itoa(cfg.HTTPPort), Handler: mux}
	go func() { _ = srv.ListenAndServe() }()

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.ControlPort))
	if err != nil {
		logger.Error("control listen", "err", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go ctrl.ServeConn(ctx, conn)
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownGracePeriod)
	defer cancel()
	_ = ln.Close()
	_ = srv.Shutdown(shutdownCtx)
	_ = adm.Shutdown(shutdownCtx)
	logger.Info("shutdown complete")
}
