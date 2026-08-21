package main

import (
    "context"
    "flag"
    "fmt"
    "net"
    "os"
    "os/signal"
    "syscall"
    "time"

    "porttrap/internal/config"
    "porttrap/internal/server"
    "porttrap/internal/logging"
)

var version = "v1.0.0"

func main() {
    showVersion := flag.Bool("version", false, "show version")
    healthcheck := flag.Bool("healthcheck", false, "probe a configured TCP port and exit 0 if reachable")
    flag.Parse()

    if *showVersion {
        fmt.Println("PortTrap", version)
        return
    }

    cfg, err := config.LoadFromEnv()
    if err != nil {
        fmt.Fprintln(os.Stderr, "config error:", err)
        os.Exit(2)
    }

    if *healthcheck {
        os.Exit(runHealthcheck(cfg))
    }

    fmt.Printf("PortTrap %s\n", version)
    fmt.Printf("TCP ports: %s\n", cfg.TCPPorts.String())
    fmt.Printf("UDP ports: %s\n", cfg.UDPPorts.String())

    logger, err := logging.NewLogger(cfg)
    if err != nil {
        fmt.Fprintln(os.Stderr, "logger error:", err)
        os.Exit(2)
    }
    defer logger.Close()

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    srv := server.New(cfg, logger)

    // start server
    if err := srv.Start(ctx); err != nil {
        fmt.Fprintln(os.Stderr, "server start error:", err)
        os.Exit(2)
    }

    // signal handling
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

    select {
    case s := <-sigs:
        fmt.Fprintln(os.Stderr, "signal received:", s)
        cancel()
    case <-ctx.Done():
    }

    // graceful shutdown with timeout
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer shutdownCancel()
    srv.Stop(shutdownCtx)
    logger.Sync()
}

// runHealthcheck dials every configured TCP port on the loopback interface.
// Returns 0 only when all ports accept a connection, 1 otherwise.
func runHealthcheck(cfg *config.Config) int {
    if len(cfg.TCPPorts) == 0 {
        fmt.Fprintln(os.Stderr, "healthcheck: no TCP_PORTS configured to probe")
        return 1
    }
    for _, port := range cfg.TCPPorts {
        addr := net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port))
        conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
        if err != nil {
            fmt.Fprintf(os.Stderr, "healthcheck: unhealthy: port %d: %v\n", port, err)
            return 1
        }
        conn.Close()
    }
    return 0
}
