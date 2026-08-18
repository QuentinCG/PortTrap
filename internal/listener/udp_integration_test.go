package listener

import (
    "context"
    "net"
    "testing"
    "time"

    "porttrap/internal/config"
    "porttrap/internal/logging"
)

func TestUDPIntegration(t *testing.T) {
    cfg := &config.Config{LogStdout: true, LogFile: "", LogFormat: "text", LogMaxSizeMB: 10, LogMaxFiles: 2, MaxConns: 10}
    logger, err := logging.NewLogger(cfg)
    if err != nil {
        t.Fatal(err)
    }
    defer logger.Close()
    sub := logger.Subscribe()

    addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:0")
    if err != nil {
        t.Fatal(err)
    }
    conn, err := net.ListenUDP("udp4", addr)
    if err != nil {
        t.Fatal(err)
    }
    defer conn.Close()

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    go ServeUDP(ctx, conn, logger)

    // send a packet
    raddr := conn.LocalAddr().String()
    c, err := net.Dial("udp4", raddr)
    if err != nil {
        t.Fatal(err)
    }
    c.Write([]byte("hello"))
    c.Close()

    select {
    case ev := <-sub:
        if ev.Protocol != "udp" {
            t.Fatalf("expected udp event, got %v", ev.Protocol)
        }
    case <-time.After(2 * time.Second):
        t.Fatal("timeout waiting for udp event")
    }
}
