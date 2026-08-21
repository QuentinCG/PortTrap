package listener

import (
    "context"
    "net"
    "testing"
    "time"

    "porttrap/internal/config"
    "porttrap/internal/logging"
)

func TestTCPIntegration(t *testing.T) {
    logLoopback = true
    defer func() { logLoopback = false }()
    cfg := &config.Config{LogStdout: true, LogFile: "", LogFormat: "text", LogMaxSizeMB: 10, LogMaxFiles: 2, MaxConns: 10}
    logger, err := logging.NewLogger(cfg)
    if err != nil {
        t.Fatal(err)
    }
    defer logger.Close()
    sub := logger.Subscribe()

    l, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil {
        t.Fatal(err)
    }
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    sem := make(chan struct{}, cfg.MaxConns)

    go ServeTCP(ctx, l, "tcp4", logger, sem, nil)

    addr := l.Addr().String()
    conn, err := net.Dial("tcp", addr)
    if err != nil {
        t.Fatal(err)
    }
    conn.Close()

    select {
    case ev := <-sub:
        if ev.Protocol != "tcp" {
            t.Fatalf("expected tcp event, got %v", ev.Protocol)
        }
    case <-time.After(2 * time.Second):
        t.Fatal("timeout waiting for event")
    }
}
