package listener

import (
    "context"
    "net"
    "time"

    "porttrap/internal/logging"
)

func ServeUDP(ctx context.Context, conn *net.UDPConn, logger *logging.Logger) error {
    defer conn.Close()
    buf := make([]byte, 4096)
    for {
        select {
        case <-ctx.Done():
            return nil
        default:
        }
        conn.SetReadDeadline(time.Now().Add(1 * time.Second))
        n, addr, err := conn.ReadFromUDP(buf)
        if err != nil {
            netErr, ok := err.(net.Error)
            if ok && netErr.Timeout() {
                continue
            }
            return err
        }
        if n > 0 {
            // ignore loopback sources (e.g. local health/test probes) to keep honeypot logs clean
            if !logLoopback && addr.IP.IsLoopback() {
                continue
            }
            ev := logging.Event{
                Timestamp:       time.Now().UTC(),
                Protocol:        "udp",
                SourceIP:        addr.IP.String(),
                SourcePort:      addr.Port,
                DestinationPort: conn.LocalAddr().(*net.UDPAddr).Port,
            }
            logger.Log(ev)
        }
    }
}
