package listener

import (
    "context"
    "fmt"
    "net"
    "time"

    "porttrap/internal/logging"
)

func ServeTCP(ctx context.Context, l net.Listener, proto string, logger *logging.Logger, sem chan struct{}, stop <-chan struct{}) error {
    defer l.Close()
    for {
        select {
        case <-ctx.Done():
            return nil
        default:
        }
        conn, err := l.Accept()
        if err != nil {
            select {
            case <-ctx.Done():
                return nil
            default:
            }
            // transient error
            continue
        }

        // ignore loopback sources (e.g. the container healthcheck) to keep honeypot logs clean
        if isLoopback(hostPart(conn.RemoteAddr().String())) {
            conn.Close()
            continue
        }

        // try acquire semaphore
        select {
        case sem <- struct{}{}:
            go handleConn(conn, proto, logger, sem)
        default:
            // saturation: log and close immediately
            remote := conn.RemoteAddr()
            // Best effort logging
            logger.Log(logging.Event{
                Timestamp:       time.Now().UTC(),
                Protocol:        "tcp",
                SourceIP:        hostPart(remote.String()),
                SourcePort:      portPart(remote.String()),
                DestinationPort: 0,
            })
            conn.Close()
        }
    }
}

func handleConn(conn net.Conn, proto string, logger *logging.Logger, sem chan struct{}) {
    defer func() { <-sem }()
    defer conn.Close()
    remote := conn.RemoteAddr()
    local := conn.LocalAddr()
    ev := logging.Event{
        Timestamp:       time.Now().UTC(),
        Protocol:        "tcp",
        SourceIP:        hostPart(remote.String()),
        SourcePort:      portPart(remote.String()),
        DestinationPort: portPart(local.String()),
    }
    logger.Log(ev)
}

func hostPart(addr string) string {
    // net.Addr.String usually returns ip:port or [ipv6]:port
    h, _, err := net.SplitHostPort(addr)
    if err != nil {
        return addr
    }
    return h
}

func isLoopback(host string) bool {
    ip := net.ParseIP(host)
    return ip != nil && ip.IsLoopback()
}

func portPart(addr string) int {
    _, p, err := net.SplitHostPort(addr)
    if err != nil {
        return 0
    }
    var pv int
    fmt.Sscanf(p, "%d", &pv)
    return pv
}
